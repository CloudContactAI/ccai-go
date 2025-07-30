// Copyright (c) 2025 CloudContactAI LLC
// Licensed under the MIT License. See LICENSE in the project root for license information.

// Package email provides email campaign functionality for the CCAI API.
package email

import (
	"encoding/json"
	"fmt"
)

// ClientInterface defines the interface that the CCAI client must implement
type ClientInterface interface {
	CustomRequest(method, endpoint string, data interface{}, customBaseURL string, headers map[string]string) ([]byte, error)
}

// Service provides email campaign functionality
type Service struct {
	client  ClientInterface
	baseURL string
}

// NewService creates a new email service instance
func NewService(client ClientInterface) *Service {
	return &Service{
		client:  client,
		baseURL: "https://email-campaigns-test-cloudcontactai.allcode.com/api/v1",
	}
}

// SendCampaign sends an email campaign to one or more recipients
func (s *Service) SendCampaign(campaign *EmailCampaign, options *EmailOptions) (*EmailResponse, error) {
	// Validate inputs
	if len(campaign.Accounts) == 0 {
		return nil, fmt.Errorf("at least one account is required")
	}

	if campaign.Subject == "" {
		return nil, fmt.Errorf("subject is required")
	}
	if campaign.Title == "" {
		return nil, fmt.Errorf("campaign title is required")
	}
	if campaign.Message == "" {
		return nil, fmt.Errorf("message content is required")
	}
	if campaign.SenderEmail == "" {
		return nil, fmt.Errorf("sender email is required")
	}
	if campaign.ReplyEmail == "" {
		return nil, fmt.Errorf("reply email is required")
	}
	if campaign.SenderName == "" {
		return nil, fmt.Errorf("sender name is required")
	}

	// Validate each account has the required fields
	for i, account := range campaign.Accounts {
		if account.FirstName == "" {
			return nil, fmt.Errorf("first name is required for account at index %d", i)
		}
		if account.LastName == "" {
			return nil, fmt.Errorf("last name is required for account at index %d", i)
		}
		if account.Email == "" {
			return nil, fmt.Errorf("email is required for account at index %d", i)
		}
	}

	// Notify progress if callback provided
	if options != nil && options.OnProgress != nil {
		options.OnProgress("Preparing to send email campaign")
	}

	endpoint := "/campaigns"

	// Notify progress if callback provided
	if options != nil && options.OnProgress != nil {
		options.OnProgress("Sending email campaign")
	}

	// Make the API request to the email campaigns API
	responseData, err := s.client.CustomRequest("POST", endpoint, campaign, s.baseURL, nil)
	if err != nil {
		// Notify progress if callback provided
		if options != nil && options.OnProgress != nil {
			options.OnProgress("Email campaign sending failed")
		}
		return nil, fmt.Errorf("failed to send email campaign: %w", err)
	}

	// Parse the response
	var response EmailResponse
	if err := json.Unmarshal(responseData, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Notify progress if callback provided
	if options != nil && options.OnProgress != nil {
		options.OnProgress("Email campaign sent successfully")
	}

	return &response, nil
}

// SendSingle sends a single email to one recipient
func (s *Service) SendSingle(
	firstName, lastName, email, subject, message, senderEmail, replyEmail, senderName, title string,
	options *EmailOptions,
) (*EmailResponse, error) {
	account := EmailAccount{
		FirstName: firstName,
		LastName:  lastName,
		Email:     email,
		Phone:     "", // Required by Account type but not used for email
	}

	campaign := &EmailCampaign{
		Subject:      subject,
		Title:        title,
		Message:      message,
		SenderEmail:  senderEmail,
		ReplyEmail:   replyEmail,
		SenderName:   senderName,
		Accounts:     []EmailAccount{account},
		CampaignType: "EMAIL",
		AddToList:    "noList",
		ContactInput: "accounts",
		FromType:     "single",
		Senders:      []interface{}{},
	}

	return s.SendCampaign(campaign, options)
}
