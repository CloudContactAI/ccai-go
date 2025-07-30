// Copyright (c) 2025 CloudContactAI LLC
// Licensed under the MIT License. See LICENSE in the project root for license information.

package sms

import (
	"encoding/json"
	"fmt"
)

// Service is the SMS service for sending messages through the CCAI API.
type Service struct {
	client ClientInterface
}

// NewService creates a new SMS service instance.
func NewService(client ClientInterface) *Service {
	return &Service{
		client: client,
	}
}

// Send sends an SMS message to one or more recipients.
func (s *Service) Send(accounts []Account, message, title string, options *Options) (*Response, error) {
	// Validate inputs
	if len(accounts) == 0 {
		return nil, fmt.Errorf("at least one account is required")
	}

	if message == "" {
		return nil, fmt.Errorf("message is required")
	}

	if title == "" {
		return nil, fmt.Errorf("title is required")
	}

	// Create options if not provided
	if options == nil {
		options = &Options{}
	}

	// Notify progress if callback provided
	options.NotifyProgress("Preparing to send SMS")

	// Prepare the endpoint and data
	endpoint := fmt.Sprintf("/clients/%s/campaigns/direct", s.client.GetClientID())

	campaignData := Campaign{
		Accounts: accounts,
		Message:  message,
		Title:    title,
	}

	// Notify progress if callback provided
	options.NotifyProgress("Sending SMS")

	// Make the API request
	responseBody, err := s.client.Request("POST", endpoint, campaignData, nil)
	if err != nil {
		// Notify progress if callback provided
		options.NotifyProgress("SMS sending failed")
		return nil, fmt.Errorf("failed to send SMS: %w", err)
	}

	// Parse the response
	var response Response
	if err := json.Unmarshal(responseBody, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Notify progress if callback provided
	options.NotifyProgress("SMS sent successfully")

	return &response, nil
}

// SendSingle sends a single SMS message to one recipient.
func (s *Service) SendSingle(firstName, lastName, phone, message, title string, options *Options) (*Response, error) {
	account := Account{
		FirstName: firstName,
		LastName:  lastName,
		Phone:     phone,
	}

	return s.Send([]Account{account}, message, title, options)
}
