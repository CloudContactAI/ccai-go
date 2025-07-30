// Copyright (c) 2025 CloudContactAI LLC
// Licensed under the MIT License. See LICENSE in the project root for license information.

package email

import "time"

// EmailAccount represents an email recipient account, extending the base Account type
type EmailAccount struct {
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Phone     string `json:"phone"`
	Email     string `json:"email"`
}

// EmailCampaign represents the email campaign configuration
type EmailCampaign struct {
	Subject            string         `json:"subject"`
	Title              string         `json:"title"`
	Message            string         `json:"message"`
	Editor             *string        `json:"editor,omitempty"`
	FileKey            *string        `json:"fileKey,omitempty"`
	SenderEmail        string         `json:"senderEmail"`
	ReplyEmail         string         `json:"replyEmail"`
	SenderName         string         `json:"senderName"`
	Accounts           []EmailAccount `json:"accounts"`
	CampaignType       string         `json:"campaignType"`
	ScheduledTimestamp *string        `json:"scheduledTimestamp,omitempty"`
	ScheduledTimezone  *string        `json:"scheduledTimezone,omitempty"`
	AddToList          string         `json:"addToList"`
	SelectedList       *SelectedList  `json:"selectedList,omitempty"`
	ListID             *string        `json:"listId,omitempty"`
	ContactInput       string         `json:"contactInput"`
	ReplaceContacts    *bool          `json:"replaceContacts,omitempty"`
	EmailTemplateID    *string        `json:"emailTemplateId,omitempty"`
	FluxID             *string        `json:"fluxId,omitempty"`
	FromType           string         `json:"fromType"`
	Senders            []interface{}  `json:"senders"`
}

// SelectedList represents a list selection
type SelectedList struct {
	Value *string `json:"value"`
}

// EmailResponse represents the response from the email API
type EmailResponse struct {
	ID           *int64                 `json:"id,omitempty"`
	Status       *string                `json:"status,omitempty"`
	CampaignID   *int64                 `json:"campaignId,omitempty"`
	MessagesSent *int                   `json:"messagesSent,omitempty"`
	Timestamp    *time.Time             `json:"timestamp,omitempty"`
	Extra        map[string]interface{} `json:"-"`
}

// EmailOptions represents optional settings for email operations
type EmailOptions struct {
	// Timeout in seconds for the request
	Timeout *int `json:"timeout,omitempty"`

	// Number of retries for failed requests
	Retries *int `json:"retries,omitempty"`

	// Callback function for progress tracking
	OnProgress func(status string) `json:"-"`
}
