package session

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"mime/multipart"
	"net/http"
	"net/http/httputil"
	"os"
	"reflect"
	"strings"
	"time"

	"github.com/golang/glog"
)

type AviResult struct {
	// Code should match the HTTP status code.
	Code int `json:"code"`

	// Message should contain a short description of the result of the requested
	// operation.
	Message *string `json:"message"`
}

// AviError represents an error resulting from a request to the Avi Controller
type AviError struct {
	// aviresult holds the standard header (code and message) that is included in
	// responses from Avi.
	AviResult

	// verb is the HTTP verb (GET, POST, PUT, PATCH, or DELETE) that was
	// used in the request that resulted in the error.
	Verb string

	// url is the URL that was used in the request that resulted in the error.
	Url string

	// HttpStatusCode is the HTTP response status code (e.g., 200, 404, etc.).
	HttpStatusCode int

	// err contains a descriptive error object for error cases other than HTTP
	// errors (i.e., non-2xx responses), such as socket errors or malformed JSON.
	err error
}

// Error implements the error interface.
func (err AviError) Error() string {
	var msg string

	if err.err != nil {
		msg = fmt.Sprintf("error: %v", err.err)
	} else if err.Message != nil {
		msg = fmt.Sprintf("HTTP code: %d; error from Avi: %s",
			err.HttpStatusCode, *err.Message)
	} else {
		msg = fmt.Sprintf("HTTP code: %d.", err.HttpStatusCode)
	}

	return fmt.Sprintf("Encountered an error on %s request to URL %s: %s",
		err.Verb, err.Url, msg)
}

//AviSession maintains a session to the specified Avi Controller
type AviSession struct {
	// host specifies the hostname or IP address of the Avi Controller
	host string

	// username specifies the username with which we should authenticate with the
	// Avi Controller.
	username string

	// password specifies the password with which we should authenticate with the
	// Avi Controller.
	password string

	// auth token generated by Django, for use in token mode
	authToken string

	// optional callback function passed in by the client which generates django auth token
	refreshAuthToken func() string

	// insecure specifies whether we should perform strict certificate validation
	// for connections to the Avi Controller.
	insecure bool

	// timeout specifies time limit for API request. Default value set to 60 seconds
	timeout time.Duration

	// optional tenant string to use for API request
	tenant string

	// optional version string to use for API request
	version string

	// internal: session id for this session
	sessionid string

	// internal: csrfToken for this session
	csrfToken string

	// internal: referer field string to use in requests
	prefix string

	// internal: re-usable transport to enable connection reuse
	transport *http.Transport

	// internal: reusable client
	client *http.Client
}

const DEFAULT_AVI_VERSION = "17.1.2"
const DEFAULT_API_TIMEOUT = time.Duration(60 * time.Second)
const DEFAULT_API_TENANT = "admin"

//NewAviSession initiates a session to AviController and returns it
func NewAviSession(host string, username string, options ...func(*AviSession) error) (*AviSession, error) {
	if flag.Parsed() == false {
		flag.Parse()
	}
	avisess := &AviSession{
		host:     host,
		username: username,
	}
	avisess.sessionid = ""
	avisess.csrfToken = ""
	avisess.prefix = "https://" + avisess.host + "/"
	avisess.tenant = ""
	avisess.insecure = false

	for _, option := range options {
		err := option(avisess)
		if err != nil {
			return avisess, err
		}
	}

    if avisess.tenant == "" {
        avisess.tenant = DEFAULT_API_TENANT
    }
	if avisess.version == "" {
		avisess.version = DEFAULT_AVI_VERSION
	}

	// create default transport object
	if avisess.transport == nil {
		avisess.transport = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
	}

	// set default timeout
	if avisess.timeout == 0 {
		avisess.timeout = DEFAULT_API_TIMEOUT
	}

	// attach transport object to client
	avisess.client = &http.Client{
		Transport: avisess.transport,
		Timeout:   avisess.timeout,
	}
	err := avisess.initiateSession()
	return avisess, err
}

//SwitchTenant Sets tenant into the avisession.
func (avisess *AviSession) SwitchTenant(tenant string) (error) {
	avisess.tenant = tenant
	if avisess.tenant == "" {
		avisess.tenant = DEFAULT_API_TENANT
	}
	return nil
}

//GetTenant Gets tenant from the avisession.
func (avisess *AviSession) GetTenant() (string) {
	return avisess.tenant
}

//GetTenant Gets tenant from the avisession.
func GetTenant() (string, error) {
	return func(avisess *AviSession) error {
		return avisess.GetTenant()
	}
}

func (avisess *AviSession) initiateSession() error {
	if avisess.insecure == true {
		glog.Warning("Strict certificate verification is *DISABLED*")
	}

	// If refresh auth token is provided, use callback function provided
	if avisess.isTokenAuth() {
		if avisess.refreshAuthToken != nil {
			avisess.setAuthToken(avisess.refreshAuthToken())
		}
	}

	// initiate http session here
	// first set the csrf token
	var res interface{}
	rerror := avisess.Get("", res)

	// now login to get session_id, csrfToken
	cred := make(map[string]string)
	cred["username"] = avisess.username

	if avisess.isTokenAuth() {
		cred["token"] = avisess.authToken
	} else {
		cred["password"] = avisess.password
	}

	rerror = avisess.Post("login", cred, res)
	if rerror != nil {
		return rerror
	}

	glog.Infof("response: %v", res)
	if res != nil && reflect.TypeOf(res).Kind() != reflect.String {
		glog.Infof("results: %v error %v", res.(map[string]interface{}), rerror)
	}

	return nil
}

// SetPassword - Use this for NewAviSession option argument for setting password
func SetPassword(password string) func(*AviSession) error {
	return func(sess *AviSession) error {
		return sess.setPassword(password)
	}
}

func (avisess *AviSession) setPassword(password string) error {
	avisess.password = password
	return nil
}

// SetVersion - Use this for NewAviSession option argument for setting version
func SetVersion(version string) func(*AviSession) error {
	return func(sess *AviSession) error {
		return sess.setVersion(version)
	}
}

func (avisess *AviSession) setVersion(version string) error {
	avisess.version = version
	return nil
}

// SetAuthToken - Use this for NewAviSession option argument for setting authToken
func SetAuthToken(authToken string) func(*AviSession) error {
	return func(sess *AviSession) error {
		return sess.setAuthToken(authToken)
	}
}

func (avisess *AviSession) setAuthToken(authToken string) error {
	avisess.authToken = authToken
	return nil
}

// SetAuthToken - Use this for NewAviSession option argument for setting authToken
func SetRefreshAuthTokenCallback(f func() string) func(*AviSession) error {
	return func(sess *AviSession) error {
		return sess.setRefreshAuthTokenCallback(f)
	}
}

func (avisess *AviSession) setRefreshAuthTokenCallback(f func() string) error {
	avisess.refreshAuthToken = f
	return nil
}

// SetTenant - Use this for NewAviSession option argument for setting tenant
func SetTenant(tenant string) func(*AviSession) error {
	return func(sess *AviSession) error {
		return sess.setTenant(tenant)
	}
}

func (avisess *AviSession) setTenant(tenant string) error {
	avisess.tenant = tenant
	return nil
}

// SetInsecure - Use this for NewAviSession option argument for allowing insecure connection to AviController
func SetInsecure(avisess *AviSession) error {
	avisess.insecure = true
	return nil
}

// SetTransport - Use this for NewAviSession option argument for configuring http transport to enable connection
func SetTransport(transport *http.Transport) func(*AviSession) error {
	return func(sess *AviSession) error {
		return sess.setTransport(transport)
	}
}

func (avisess *AviSession) setTransport(transport *http.Transport) error {
	avisess.transport = transport
	return nil
}

// SetTimeout -
func SetTimeout(timeout time.Duration) func(*AviSession) error {
	return func(sess *AviSession) error {
		return sess.setTimeout(timeout)
	}
}

func (avisess *AviSession) setTimeout(timeout time.Duration) error {
	avisess.timeout = timeout
	return nil
}

func (avisess *AviSession) isTokenAuth() bool {
	return avisess.authToken != "" || avisess.refreshAuthToken != nil
}

func (avisess *AviSession) checkRetryForSleep(retry int, verb string, url string) error {
	if retry == 0 {
		return nil
	} else if retry == 1 {
		time.Sleep(100 * time.Millisecond)
	} else if retry == 2 {
		time.Sleep(500 * time.Millisecond)
	} else if retry == 3 {
		time.Sleep(1 * time.Second)
	} else if retry > 3 {
		errorResult := AviError{Verb: verb, Url: url}
		errorResult.err = fmt.Errorf("tried 3 times and failed")
		glog.Error("Aborting after 3 times")
		return errorResult
	}
	return nil
}

func (avisess *AviSession) newAviRequest(verb string, url string, payload io.Reader) (*http.Request, AviError) {
	req, err := http.NewRequest(verb, url, payload)
	errorResult := AviError{Verb: verb, Url: url}
	if err != nil {
		errorResult.err = fmt.Errorf("http.NewRequest failed: %v", err)
		return nil, errorResult
	}
	req.Header.Set("Content-Type", "application/json")
	//req.Header.Set("Accept", "application/json")
	req.Header.Set("X-Avi-Version", avisess.version)

	if avisess.csrfToken != "" {
		req.Header["X-CSRFToken"] = []string{avisess.csrfToken}
		req.AddCookie(&http.Cookie{Name: "csrftoken", Value: avisess.csrfToken})
	}
	if avisess.prefix != "" {
		req.Header.Set("Referer", avisess.prefix)
	}
	if avisess.tenant != "" {
		req.Header.Set("X-Avi-Tenant", avisess.tenant)
	}
	if avisess.sessionid != "" {
		req.AddCookie(&http.Cookie{Name: "sessionid", Value: avisess.sessionid})
		req.AddCookie(&http.Cookie{Name: "avi-sessionid", Value: avisess.sessionid})
	}
	return req, errorResult
}

//
// Helper routines for REST calls.
//

func (avisess *AviSession) collectCookiesFromResp(resp *http.Response) {
	// collect cookies from the resp
	for _, cookie := range resp.Cookies() {
		glog.Infof("cookie: %v", cookie)
		if cookie.Name == "csrftoken" {
			avisess.csrfToken = cookie.Value
			glog.Infof("Set the csrf token to %v", avisess.csrfToken)
		}
		if cookie.Name == "sessionid" {
			avisess.sessionid = cookie.Value
		}
		if cookie.Name == "avi-sessionid" {
			avisess.sessionid = cookie.Value
		}
	}
}

// restRequest makes a REST request to the Avi Controller's REST API.
// Returns a byte[] if successful
func (avisess *AviSession) restRequest(verb string, uri string, payload interface{}, tenant string, retryNum ...int) ([]byte, error) {
	var result []byte
	url := avisess.prefix + uri
	if tenant == "" {
		tenant = avisess.tenant
	}
	// If optional retryNum arg is provided, then count which retry number this is
	retry := 0
	if len(retryNum) > 0 {
		retry = retryNum[0]
	}
	if errorResult := avisess.checkRetryForSleep(retry, verb, url); errorResult != nil {
		return nil, errorResult
	}
	var payloadIO io.Reader
	if payload != nil {
		jsonStr, err := json.Marshal(payload)
		if err != nil {
			return result, AviError{Verb: verb, Url: url, err: err}
		}
		payloadIO = bytes.NewBuffer(jsonStr)
	}

	req, errorResult := avisess.newAviRequest(verb, url, payloadIO)
	if errorResult.err != nil {
		return result, errorResult
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-Avi-Tenant", tenant)
	glog.Infof("Sending req for uri %v", url)
	resp, err := avisess.client.Do(req)
	if err != nil {
		errorResult.err = fmt.Errorf("client.Do failed: %v", err)
		dump, err := httputil.DumpRequestOut(req, true)
		debug(dump, err)
		return result, errorResult
	}

	errorResult.HttpStatusCode = resp.StatusCode
	avisess.collectCookiesFromResp(resp)

	retryReq := false
	if resp.StatusCode == 401 && len(avisess.sessionid) != 0 && uri != "login" {
		resp.Body.Close()
		err := avisess.initiateSession()
		if err != nil {
			return nil, err
		}
		retryReq = true
	} else if resp.StatusCode == 419 || (resp.StatusCode >= 500 && resp.StatusCode < 599) {
		resp.Body.Close()
		retryReq = true
		glog.Infof("Retrying %d due to Status Code %d", retry, resp.StatusCode)
	}

	if retryReq {
		check, err := avisess.CheckControllerStatus()
		if check == false {
			glog.Errorf("restRequest Error during checking controller state %v", err)
			return nil, err
		}
		// Doing this so that a new request is made to the
		return avisess.restRequest(verb, uri, payload, tenant, retry+1)
	}

	defer resp.Body.Close()

	if resp.StatusCode == 204 {
		// no content in the response
		return result, nil
	}
	result, err = ioutil.ReadAll(resp.Body)
	if err == nil {
		if resp.StatusCode < 200 || resp.StatusCode > 299 {
			mres, _ := convertAviResponseToMapInterface(result)
			glog.Infof("Error resp: %v", mres)
			emsg := fmt.Sprintf("%v", mres)
			errorResult.Message = &emsg
		} else {
			return result, nil
		}
	} else {
		errmsg := fmt.Sprintf("Response body read failed: %v", err)
		errorResult.Message = &errmsg
		glog.Errorf("Error in reading uri %v %v", uri, err)
	}
	return result, errorResult
}

// restMultipartUploadRequest makes a REST request to the Avi Controller's REST API using POST to upload a file.
// Return status of multipart upload.
func (avisess *AviSession) restMultipartUploadRequest(verb string, uri string, file_path_ptr *os.File, retryNum ...int) error {
	url := avisess.prefix + "/api/fileservice/" + uri

	// If optional retryNum arg is provided, then count which retry number this is
	retry := 0
	if len(retryNum) > 0 {
		retry = retryNum[0]
	}

	if errorResult := avisess.checkRetryForSleep(retry, verb, url); errorResult != nil {
		return errorResult
	}

	errorResult := AviError{Verb: verb, Url: url}
	//Prepare a file that you will submit to an URL.
	values := map[string]io.Reader{
		"file": file_path_ptr,
	}

	var b bytes.Buffer
	w := multipart.NewWriter(&b)
	for key, r := range values {
		var fw io.Writer
		var err error
		if x, ok := r.(io.Closer); ok {
			defer x.Close()
		}
		// Add an file
		if x, ok := r.(*os.File); ok {
			if fw, err = w.CreateFormFile(key, x.Name()); err != nil {
				if err != nil {
					glog.Errorf("restMultipartUploadRequest Error in adding file: %v ", err)
					return err
				}
			}
		}
		if _, err := io.Copy(fw, r); err != nil {
			if err != nil {
				glog.Errorf("restMultipartUploadRequest Error io.Copy %v ", err)
				return err
			}
		}

	}
	// Closing the multipart writer.
	// If you don't close it, your request will be missing the terminating boundary.
	w.Close()
	uri_temp := "controller://" + strings.Split(uri, "?")[0]
	err := w.WriteField("uri", uri_temp)
	if err != nil {
		errorResult.err = fmt.Errorf("restMultipartUploadRequest Adding URI field failed: %v", err)
		return errorResult
	}
	req, errorResult := avisess.newAviRequest(verb, url, &b)
	if errorResult.err != nil {
		return errorResult
	}
	req.Header.Set("Content-Type", w.FormDataContentType())
	resp, err := avisess.client.Do(req)
	if err != nil {
		glog.Errorf("restMultipartUploadRequest Error during client request: %v ", err)
		dump, err := httputil.DumpRequestOut(req, true)
		debug(dump, err)
		return err
	}

	defer resp.Body.Close()

	errorResult.HttpStatusCode = resp.StatusCode
	avisess.collectCookiesFromResp(resp)
	glog.Infof("Response code: %v", resp.StatusCode)

	retryReq := false
	if resp.StatusCode == 401 && len(avisess.sessionid) != 0 && uri != "login" {
		resp.Body.Close()
		err := avisess.initiateSession()
		if err != nil {
			return err
		}
		retryReq = true
	} else if resp.StatusCode == 419 || (resp.StatusCode >= 500 && resp.StatusCode < 599) {
		resp.Body.Close()
		retryReq = true
		glog.Infof("Retrying %d due to Status Code %d", retry, resp.StatusCode)
	}

	if retryReq {
		check, err := avisess.CheckControllerStatus()
		if check == false {
			glog.Errorf("restMultipartUploadRequest Error during checking controller state")
			return err
		}
		// Doing this so that a new request is made to the
		return avisess.restMultipartUploadRequest(verb, uri, file_path_ptr, retry+1)
	}

	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		glog.Errorf("Error: %v", resp)
		bres, berr := ioutil.ReadAll(resp.Body)
		if berr == nil {
			mres, _ := convertAviResponseToMapInterface(bres)
			glog.Infof("Error resp: %v", mres)
			emsg := fmt.Sprintf("%v", mres)
			errorResult.Message = &emsg
		}
		return errorResult
	}

	if resp.StatusCode == 201 {
		// File Created and upload to server
		fmt.Printf("restMultipartUploadRequest Response: %v", resp.Status)
		return nil
	}

	return err
}

// restMultipartDownloadRequest makes a REST request to the Avi Controller's REST API.
// Returns multipart download and write data to file
func (avisess *AviSession) restMultipartDownloadRequest(verb string, uri string, file_path_ptr *os.File, retryNum ...int) error {
	url := avisess.prefix + "/api/fileservice/" + uri

	// If optional retryNum arg is provided, then count which retry number this is
	retry := 0
	if len(retryNum) > 0 {
		retry = retryNum[0]
	}

	if errorResult := avisess.checkRetryForSleep(retry, verb, url); errorResult != nil {
		return errorResult
	}

	req, errorResult := avisess.newAviRequest(verb, url, nil)
	if errorResult.err != nil {
		return errorResult
	}
	req.Header.Set("Accept", "application/json")
	resp, err := avisess.client.Do(req)
	if err != nil {
		errorResult.err = fmt.Errorf("restMultipartDownloadRequest Error for during client request: %v", err)
		dump, err := httputil.DumpRequestOut(req, true)
		debug(dump, err)
		return errorResult
	}

	errorResult.HttpStatusCode = resp.StatusCode
	avisess.collectCookiesFromResp(resp)
	glog.Infof("Response code: %v", resp.StatusCode)

	retryReq := false
	if resp.StatusCode == 401 && len(avisess.sessionid) != 0 && uri != "login" {
		resp.Body.Close()
		err := avisess.initiateSession()
		if err != nil {
			return err
		}
		retryReq = true
	} else if resp.StatusCode == 419 || (resp.StatusCode >= 500 && resp.StatusCode < 599) {
		resp.Body.Close()
		retryReq = true
		glog.Infof("Retrying %d due to Status Code %d", retry, resp.StatusCode)
	}

	if retryReq {
		check, err := avisess.CheckControllerStatus()
		if check == false {
			glog.Errorf("restMultipartDownloadRequest Error during checking controller state")
			return err
		}
		return avisess.restMultipartDownloadRequest(verb, uri, file_path_ptr)
	}

	defer resp.Body.Close()

	if resp.StatusCode == 204 {
		// no content in the response
		return nil
	}

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		glog.Errorf("Error: %v", resp)
		bres, berr := ioutil.ReadAll(resp.Body)
		if berr == nil {
			mres, _ := convertAviResponseToMapInterface(bres)
			glog.Infof("Error resp: %v", mres)
			emsg := fmt.Sprintf("%v", mres)
			errorResult.Message = &emsg
		}
		return errorResult
	}

	_, err = io.Copy(file_path_ptr, resp.Body)
	defer file_path_ptr.Close()

	if err != nil {
		glog.Errorf("Error while downloading %v", err)
	}
	return err
}

func convertAviResponseToMapInterface(resbytes []byte) (interface{}, error) {
	var result interface{}
	err := json.Unmarshal(resbytes, &result)
	return result, err
}

// AviCollectionResult for representing the collection type results from Avi
type AviCollectionResult struct {
	Count   int
	Results json.RawMessage
}

func debug(data []byte, err error) {
	if err == nil {
		glog.Infof("%s\n\n", data)
	} else {
		glog.Errorf("%s\n\n", err)
	}
}

//Checking for controller up state.
//This is an infinite loop till the controller is in up state.
//Return true when controller is in up state.
func (avisess *AviSession) CheckControllerStatus() (bool, error) {
	url := avisess.prefix + "/api/cluster/status"
	//This is an infinite loop. Generating http request for a login URI till controller is in up state.
	for round := 0; round < 10; round++ {
		checkReq, err := http.NewRequest("GET", url, nil)
		if err != nil {
			glog.Errorf("CheckControllerStatus Error %v while generating http request.", err)
			return false, err
		}
		//Getting response from controller's API
		if stateResp, err := avisess.client.Do(checkReq); err == nil {
			defer stateResp.Body.Close()
			//Checking controller response
			if stateResp.StatusCode != 503 && stateResp.StatusCode != 502 && stateResp.StatusCode != 500 {
				break
			} else {
				glog.Infof("CheckControllerStatus Error while generating http request %d %v",
					stateResp.StatusCode, err)
			}
		} else {
			glog.Errorf("CheckControllerStatus Error while generating http request %v %v", url, err)
		}
		//wait before retry
		time.Sleep(time.Duration(math.Exp(float64(round))*3) * time.Second)
		glog.Errorf("CheckControllerStatus Controller %v Retrying. round %v..!", url, round)
	}
	return true, nil
}

func (avisess *AviSession) restRequestInterfaceResponse(verb string, url string,
	payload interface{}, response interface{}, tenant ...string) error {

	res, rerror := avisess.restRequest(verb, url, payload, tenant[0])
	if rerror != nil || res == nil {
		return rerror
	}
	return json.Unmarshal(res, &response)
}

// Get issues a GET request against the avisess REST API.
func (avisess *AviSession) Get(uri string, response interface{}, tenant ...string) error {
	loc_tenant := ""
	if len(tenant) == 0 {
		loc_tenant = avisess.tenant
	} else {
		loc_tenant = tenant[0]
	}
	return avisess.restRequestInterfaceResponse("GET", uri, nil, response, loc_tenant)
}

// Post issues a POST request against the avisess REST API.
func (avisess *AviSession) Post(uri string, payload interface{}, response interface{}, tenant ...string) error {
	loc_tenant := ""
	if len(tenant) == 0 {
		loc_tenant = avisess.tenant
	} else {
		loc_tenant = tenant[0]
	}
	return avisess.restRequestInterfaceResponse("POST", uri, payload, response, loc_tenant)
}

// Put issues a PUT request against the avisess REST API.
func (avisess *AviSession) Put(uri string, payload interface{}, response interface{}, tenant ...string) error {
	loc_tenant := ""
	if len(tenant) == 0 {
		loc_tenant = avisess.tenant
	} else {
		loc_tenant = tenant[0]
	}
	return avisess.restRequestInterfaceResponse("PUT", uri, payload, response, loc_tenant)
}

// Post issues a PATCH request against the avisess REST API.
// allowed patchOp - add, replace, remove
func (avisess *AviSession) Patch(uri string, payload interface{}, patchOp string, response interface{}, tenant ...string) error {
	var patchPayload = make(map[string]interface{})
	patchPayload[patchOp] = payload
	glog.Info(" PATCH OP %v data %v", patchOp, payload)
	loc_tenant := ""
	if len(tenant) == 0 {
		loc_tenant = avisess.tenant
	} else {
		loc_tenant = tenant[0]
	}
	return avisess.restRequestInterfaceResponse("PATCH", uri, patchPayload, response, loc_tenant)
}

// Delete issues a DELETE request against the avisess REST API.
func (avisess *AviSession) Delete(uri string, tenant string, params ...interface{}) error {
	var payload, response interface{}
	if len(params) > 0 {
		payload = params[0]
		if len(params) == 2 {
			response = params[1]
		}
	}
	return avisess.restRequestInterfaceResponse("DELETE", uri, payload, response, tenant)
}

// GetCollectionRaw issues a GET request and returns a AviCollectionResult with unmarshaled (raw) results section.
func (avisess *AviSession) GetCollectionRaw(uri string, tenant string) (AviCollectionResult, error) {
	var result AviCollectionResult
	res, rerror := avisess.restRequest("GET", uri, nil, tenant)
	if rerror != nil || res == nil {
		return result, rerror
	}
	err := json.Unmarshal(res, &result)
	return result, err
}

// GetCollection performs a collection API call and unmarshals the results into objList, which should be an array type
func (avisess *AviSession) GetCollection(uri string, objList interface{}, tenant string) error {
	result, err := avisess.GetCollectionRaw(uri, tenant)
	if err != nil {
		return err
	}
	if result.Count == 0 {
		return nil
	}
	return json.Unmarshal(result.Results, &objList)
}

// GetRaw performs a GET API call and returns raw data
func (avisess *AviSession) GetRaw(uri string, tenant string) ([]byte, error) {
	return avisess.restRequest("GET", uri, nil, tenant)
}

// PostRaw performs a POST API call and returns raw data
func (avisess *AviSession) PostRaw(uri string, payload interface{}, tenant string) ([]byte, error) {
	return avisess.restRequest("POST", uri, payload, tenant)
}

// GetMultipartRaw performs a GET API call and returns multipart raw data (File Download)
func (avisess *AviSession) GetMultipartRaw(verv string, uri string, file_loc_ptr *os.File) error {
	return avisess.restMultipartDownloadRequest("GET", uri, file_loc_ptr)
}

// PostMultipartRequest performs a POST API call and uploads multipart data
func (avisess *AviSession) PostMultipartRequest(verb string, uri string, file_loc_ptr *os.File) error {
	return avisess.restMultipartUploadRequest("POST", uri, file_loc_ptr)
}

type ApiOptions struct {
	name        string
	cloud       string
	cloudUUID   string
	skipDefault bool
	includeName bool
	result      interface{}
}

func SetName(name string) func(*ApiOptions) error {
	return func(opts *ApiOptions) error {
		return opts.setName(name)
	}
}

func (opts *ApiOptions) setName(name string) error {
	opts.name = name
	return nil
}

func SetCloud(cloud string) func(*ApiOptions) error {
	return func(opts *ApiOptions) error {
		return opts.setCloud(cloud)
	}
}

func (opts *ApiOptions) setCloud(cloud string) error {
	opts.cloud = cloud
	return nil
}

func SetCloudUUID(cloudUUID string) func(*ApiOptions) error {
	return func(opts *ApiOptions) error {
		return opts.setCloudUUID(cloudUUID)
	}
}

func (opts *ApiOptions) setCloudUUID(cloudUUID string) error {
	opts.cloudUUID = cloudUUID
	return nil
}

func SetSkipDefault(skipDefault bool) func(*ApiOptions) error {
	return func(opts *ApiOptions) error {
		return opts.setSkipDefault(skipDefault)
	}
}

func (opts *ApiOptions) setSkipDefault(skipDefault bool) error {
	opts.skipDefault = skipDefault
	return nil
}

func SetIncludeName(includeName bool) func(*ApiOptions) error {
	return func(opts *ApiOptions) error {
		return opts.setIncludeName(includeName)
	}
}

func (opts *ApiOptions) setIncludeName(includeName bool) error {
	opts.includeName = includeName
	return nil
}

func SetResult(result interface{}) func(*ApiOptions) error {
	return func(opts *ApiOptions) error {
		return opts.setResult(result)
	}
}

func (opts *ApiOptions) setResult(result interface{}) error {
	opts.result = result
	return nil
}

type ApiOptionsParams func(*ApiOptions) error

func (avisess *AviSession) GetUri(obj string, options ...ApiOptionsParams) (string, error) {
	opts := &ApiOptions{}
	for _, opt := range options {
		err := opt(opts)
		if err != nil {
			return "", err
		}
	}
	if opts.result == nil {
		return "", errors.New("reference to result provided")
	}

	if opts.name == "" {
		return "", errors.New("Name not specified")
	}

	uri := "api/" + obj + "?name=" + opts.name
	if opts.cloud != "" {
		uri = uri + "&cloud=" + opts.cloud
	} else if opts.cloudUUID != "" {
		uri = uri + "&cloud_ref.uuid=" + opts.cloudUUID
	}
	if opts.skipDefault {
		uri = uri + "&skip_default=true"
	}
	if opts.includeName {
		uri = uri + "&include_name=true"
	}
	return uri, nil
}

func (avisess *AviSession) GetObject(obj string, tenant string, options ...ApiOptionsParams) error {
	opts := &ApiOptions{}
	for _, opt := range options {
		err := opt(opts)
		if err != nil {
			return err
		}
	}
	uri, err := avisess.GetUri(obj, options...)
	if err != nil {
		return err
	}
	res, err := avisess.GetCollectionRaw(uri, tenant)
	if err != nil {
		return err
	}
	if res.Count == 0 {
		return errors.New("No object of type " + obj + " with name " + opts.name + "is found")
	} else if res.Count > 1 {
		return errors.New("More than one object of type " + obj + " with name " + opts.name + "is found")
	}
	elems := make([]json.RawMessage, 1)
	err = json.Unmarshal(res.Results, &elems)
	if err != nil {
		return err
	}
	return json.Unmarshal(elems[0], &opts.result)

}

// GetObjectByName performs GET with name filter
func (avisess *AviSession) GetObjectByName(obj string, name string, result interface{}, tenant ...string) error {
	loc_tenant := ""
	if len(tenant) == 0 {
		loc_tenant = avisess.tenant
	} else {
		loc_tenant = tenant[0]
	}
	return avisess.GetObject(obj, loc_tenant, SetName(name), SetResult(result))
}

// Utility functions

// GetControllerVersion gets the version number from the Avi Controller
func (avisess *AviSession) GetControllerVersion() (string, error) {
	var resp interface{}

	err := avisess.Get("/api/initial-data", &resp)
	if err != nil {
		return "", err
	}
	version := resp.(map[string]interface{})["version"].(map[string]interface{})["Version"].(string)
	return version, nil
}
