/***************************************************************************
 *
 * AVI CONFIDENTIAL
 * __________________
 *
 * [2013] - [2018] Avi Networks Incorporated
 * All Rights Reserved.
 *
 * NOTICE: All information contained herein is, and remains the property
 * of Avi Networks Incorporated and its suppliers, if any. The intellectual
 * and technical concepts contained herein are proprietary to Avi Networks
 * Incorporated, and its suppliers and are covered by U.S. and Foreign
 * Patents, patents in process, and are protected by trade secret or
 * copyright law, and other laws. Dissemination of this information or
 * reproduction of this material is strictly forbidden unless prior written
 * permission is obtained from Avi Networks Incorporated.
 */

package clients

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

import (
	"github.com/avinetworks/sdk/go/models"
	"github.com/avinetworks/sdk/go/session"
)

// WafCRSClient is a client for avi WafCRS resource
type WafCRSClient struct {
	aviSession *session.AviSession
}

// NewWafCRSClient creates a new client for WafCRS resource
func NewWafCRSClient(aviSession *session.AviSession) *WafCRSClient {
	return &WafCRSClient{aviSession: aviSession}
}

func (client *WafCRSClient) getAPIPath(uuid string) string {
	path := "api/wafcrs"
	if uuid != "" {
		path += "/" + uuid
	}
	return path
}

// GetAll is a collection API to get a list of WafCRS objects
func (client *WafCRSClient) GetAll() ([]*models.WafCRS, error) {
	var plist []*models.WafCRS
	err := client.aviSession.GetCollection(client.getAPIPath(""), &plist)
	return plist, err
}

// Get an existing WafCRS by uuid
func (client *WafCRSClient) Get(uuid string) (*models.WafCRS, error) {
	var obj *models.WafCRS
	err := client.aviSession.Get(client.getAPIPath(uuid), &obj)
	return obj, err
}

// GetByName - Get an existing WafCRS by name
func (client *WafCRSClient) GetByName(name string) (*models.WafCRS, error) {
	var obj *models.WafCRS
	err := client.aviSession.GetObjectByName("wafcrs", name, &obj)
	return obj, err
}

// GetObject - Get an existing WafCRS by filters like name, cloud, tenant
// Api creates WafCRS object with every call.
func (client *WafCRSClient) GetObject(options ...session.ApiOptionsParams) (*models.WafCRS, error) {
	var obj *models.WafCRS
	newOptions := make([]session.ApiOptionsParams, len(options)+1)
	for i, p := range options {
		newOptions[i] = p
	}
	newOptions[len(options)] = session.SetResult(&obj)
	err := client.aviSession.GetObject("wafcrs", newOptions...)
	return obj, err
}

// Create a new WafCRS object
func (client *WafCRSClient) Create(obj *models.WafCRS) (*models.WafCRS, error) {
	var robj *models.WafCRS
	err := client.aviSession.Post(client.getAPIPath(""), obj, &robj)
	return robj, err
}

// Update an existing WafCRS object
func (client *WafCRSClient) Update(obj *models.WafCRS) (*models.WafCRS, error) {
	var robj *models.WafCRS
	path := client.getAPIPath(*obj.UUID)
	err := client.aviSession.Put(path, obj, &robj)
	return robj, err
}

// Patch an existing WafCRS object specified using uuid
// patchOp: Patch operation - add, replace, or delete
// patch: Patch payload should be compatible with the models.WafCRS
// or it should be json compatible of form map[string]interface{}
func (client *WafCRSClient) Patch(uuid string, patch interface{}, patchOp string) (*models.WafCRS, error) {
	var robj *models.WafCRS
	path := client.getAPIPath(uuid)
	err := client.aviSession.Patch(path, patch, patchOp, &robj)
	return robj, err
}

// Delete an existing WafCRS object with a given UUID
func (client *WafCRSClient) Delete(uuid string) error {
	return client.aviSession.Delete(client.getAPIPath(uuid))
}

// DeleteByName - Delete an existing WafCRS object with a given name
func (client *WafCRSClient) DeleteByName(name string) error {
	res, err := client.GetByName(name)
	if err != nil {
		return err
	}
	return client.Delete(*res.UUID)
}

// GetAviSession
func (client *WafCRSClient) GetAviSession() *session.AviSession {
	return client.aviSession
}
