// Copyright (c) 2025 CloudContactAI LLC
// Licensed under the MIT License. See LICENSE in the project root for license information.

// Package sms provides functionality for sending SMS messages through the CCAI API.
package sms

import "fmt"

// Account represents a recipient account.
type Account struct {
	FirstName   string            `json:"firstName"`
	LastName    string            `json:"lastName"`
	Phone       string            `json:"phone"`
	// Data holds additional key-value pairs for variable substitution in message templates.
	// Define any keys you want and use them as ${key} in your message.
	// Example: Data: map[string]string{"city": "Miami"}, message: "Hello from ${city}!"
	// Sent to the API as "data" (wire format).
	Data        map[string]string `json:"data,omitempty"`
	// MessageData is an arbitrary string forwarded as-is to your webhook handler.
	// Not used in the message body. Sent to the API as "messageData" (wire format).
	MessageData string            `json:"messageData,omitempty"`
}

// Campaign represents an SMS campaign.
type Campaign struct {
	Accounts    []Account `json:"accounts"`
	Message     string    `json:"message"`
	Title       string    `json:"title"`
	SenderPhone string    `json:"senderPhone,omitempty"`
}

// MMSCampaign represents an MMS campaign.
type MMSCampaign struct {
	PictureFileKey string    `json:"pictureFileKey"`
	Accounts       []Account `json:"accounts"`
	Message        string    `json:"message"`
	Title          string    `json:"title"`
	SenderPhone    string    `json:"senderPhone,omitempty"`
}

// Response represents the response from the SMS API.
type Response struct {
	ID           interface{}            `json:"id,omitempty"`
	Status       string                 `json:"status,omitempty"`
	CampaignID   string                 `json:"campaignId,omitempty"`
	MessagesSent int                    `json:"messagesSent,omitempty"`
	Timestamp    string                 `json:"timestamp,omitempty"`
	Message      string                 `json:"message,omitempty"`
	ResponseID   string                 `json:"responseId,omitempty"`
	Extra        map[string]interface{} `json:"-"`
}

// GetID returns the ID as a string, handling both string and number types.
func (r *Response) GetID() string {
	switch v := r.ID.(type) {
	case string:
		return v
	case float64:
		return fmt.Sprintf("%.0f", v)
	case int:
		return fmt.Sprintf("%d", v)
	default:
		return fmt.Sprintf("%v", v)
	}
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
	GetFilesBaseURL() string
	Request(method, endpoint string, data interface{}, headers map[string]string) ([]byte, error)
}
