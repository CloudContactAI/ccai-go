// Copyright (c) 2025 CloudContactAI LLC
// Licensed under the MIT License. See LICENSE in the project root for license information.

// Package sms provides functionality for sending SMS messages through the CCAI API.
package sms

// Account represents a recipient account.
type Account struct {
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Phone     string `json:"phone"`
}

// Campaign represents an SMS campaign.
type Campaign struct {
	Accounts []Account `json:"accounts"`
	Message  string    `json:"message"`
	Title    string    `json:"title"`
}

// MMSCampaign represents an MMS campaign.
type MMSCampaign struct {
	PictureFileKey string    `json:"pictureFileKey"`
	Accounts       []Account `json:"accounts"`
	Message        string    `json:"message"`
	Title          string    `json:"title"`
}

// Response represents the response from the SMS API.
type Response struct {
	ID           string                 `json:"id,omitempty"`
	Status       string                 `json:"status,omitempty"`
	CampaignID   string                 `json:"campaignId,omitempty"`
	MessagesSent int                    `json:"messagesSent,omitempty"`
	Timestamp    string                 `json:"timestamp,omitempty"`
	Extra        map[string]interface{} `json:"-"`
}

// SignedURLResponse represents the response from the signed URL API.
type SignedURLResponse struct {
	SignedS3URL string                 `json:"signedS3Url"`
	FileKey     string                 `json:"fileKey"`
	Extra       map[string]interface{} `json:"-"`
}

// Options represents options for SMS operations.
type Options struct {
	Timeout    int
	Retries    int
	OnProgress func(string)
}

// NotifyProgress notifies progress if callback is provided.
func (o *Options) NotifyProgress(status string) {
	if o != nil && o.OnProgress != nil {
		o.OnProgress(status)
	}
}

// ClientInterface defines the interface for the CCAI client.
type ClientInterface interface {
	GetClientID() string
	GetAPIKey() string
	GetBaseURL() string
	Request(method, endpoint string, data interface{}, headers map[string]string) ([]byte, error)
}
