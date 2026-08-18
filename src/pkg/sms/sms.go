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
func (s *Service) Send(accounts []Account, message, title, senderPhone string, options *Options) (*Response, error) {
	return s.SendWithOptions(accounts, message, title, senderPhone, nil, options)
}

// SendWithOptions sends an SMS with an optional templateId.
func (s *Service) SendWithOptions(accounts []Account, message, title, senderPhone string, templateID *int64, options *Options) (*Response, error) {
	if len(accounts) == 0 {
		return nil, fmt.Errorf("at least one account is required")
	}
	if message == "" && templateID == nil {
		return nil, fmt.Errorf("message is required when templateId is not provided")
	}
	if title == "" {
		return nil, fmt.Errorf("title is required")
	}
	if options == nil {
		options = &Options{}
	}
	options.NotifyProgress("Preparing to send SMS")
	endpoint := fmt.Sprintf("/clients/%s/campaigns/direct", s.client.GetClientID())
	campaignData := Campaign{
		Accounts:    accounts,
		Message:     message,
		Title:       title,
		SenderPhone: senderPhone,
		TemplateID:  templateID,
	}
	options.NotifyProgress("Sending SMS")
	responseBody, err := s.client.Request("POST", endpoint, campaignData, nil)
	if err != nil {
		options.NotifyProgress("SMS sending failed")
		return nil, fmt.Errorf("failed to send SMS: %w", err)
	}
	var response Response
	if err := json.Unmarshal(responseBody, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	options.NotifyProgress("SMS sent successfully")
	return &response, nil
}

// SendWithTemplate sends an SMS using a pre-approved template (for template-controlled accounts).
func (s *Service) SendWithTemplate(accounts []Account, templateID int64, title, senderPhone string, options *Options) (*Response, error) {
	return s.SendWithOptions(accounts, "", title, senderPhone, &templateID, options)
}

// SendSingleWithTemplate sends an SMS to a single recipient using a pre-approved template.
func (s *Service) SendSingleWithTemplate(firstName, lastName, phone string, templateID int64, title, senderPhone string, options *Options) (*Response, error) {
	account := Account{FirstName: firstName, LastName: lastName, Phone: phone}
	return s.SendWithTemplate([]Account{account}, templateID, title, senderPhone, options)
}

// SendSingle sends a single SMS message to one recipient.
func (s *Service) SendSingle(firstName, lastName, phone, message, title, customData, senderPhone string, options *Options) (*Response, error) {
	account := Account{
		FirstName:   firstName,
		LastName:    lastName,
		Phone:       phone,
		MessageData: customData,
	}

	return s.Send([]Account{account}, message, title, senderPhone, options)
}
