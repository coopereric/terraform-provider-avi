package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// AppLearningParams app learning params
// swagger:model AppLearningParams
type AppLearningParams struct {

	// Learn the params per URI path. Field introduced in 18.2.3.
	EnablePerURILearning *bool `json:"enable_per_uri_learning,omitempty"`

	// Maximum number of params to learn for an application. Allowed values are 10-1000. Field introduced in 18.2.3.
	MaxParams *int32 `json:"max_params,omitempty"`

	// Maximum number of URI paths to learn for an application. Allowed values are 10-1000. Field introduced in 18.2.3.
	MaxUris *int32 `json:"max_uris,omitempty"`

	// Percent of the requests subjected to Application learning. Allowed values are 1-100. Field introduced in 18.2.3.
	SamplingPercent *int32 `json:"sampling_percent,omitempty"`

	// Frequency with which SE publishes Application learning data to controller. Allowed values are 1-10080. Field introduced in 18.2.3.
	UpdateInterval *int32 `json:"update_interval,omitempty"`
}
