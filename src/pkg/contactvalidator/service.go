// Copyright (c) 2025 CloudContactAI LLC
// Licensed under the MIT License. See LICENSE in the project root for license information.

// Package contactvalidator provides functionality for validating email addresses and phone numbers.
package contactvalidator

import (
	"encoding/json"
	"fmt"
)

// ClientInterface defines the interface that the CCAI client must implement.
type ClientInterface interface {
	Request(method, endpoint string, data interface{}, headers map[string]string) ([]byte, error)
}

// Service provides contact validation functionality.
type Service struct {
	client ClientInterface
}

// NewService creates a new contact validator service.
func NewService(client ClientInterface) *Service {
	return &Service{client: client}
}

// ValidateEmail validates a single email address.
func (s *Service) ValidateEmail(email string) (*EmailValidationResult, error) {
	data := map[string]string{"email": email}
	body, err := s.client.Request("POST", "/v1/contact-validator/email", data, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to validate email: %w", err)
	}
	var result EmailValidationResult
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return &result, nil
}

// ValidateEmails validates multiple email addresses (up to 50).
func (s *Service) ValidateEmails(emails []string) (*BulkEmailValidationResult, error) {
	data := map[string]interface{}{"emails": emails}
	body, err := s.client.Request("POST", "/v1/contact-validator/emails", data, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to validate emails: %w", err)
	}
	var result BulkEmailValidationResult
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return &result, nil
}

// ValidatePhone validates a single phone number. countryCode is optional ("" to omit).
func (s *Service) ValidatePhone(phone, countryCode string) (*PhoneValidationResult, error) {
	data := map[string]interface{}{"phone": phone}
	if countryCode != "" {
		data["countryCode"] = countryCode
	}
	body, err := s.client.Request("POST", "/v1/contact-validator/phone", data, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to validate phone: %w", err)
	}
	var result PhoneValidationResult
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return &result, nil
}

// ValidatePhones validates multiple phone numbers (up to 50).
func (s *Service) ValidatePhones(phones []PhoneInput) (*BulkPhoneValidationResult, error) {
	data := map[string]interface{}{"phones": phones}
	body, err := s.client.Request("POST", "/v1/contact-validator/phones", data, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to validate phones: %w", err)
	}
	var result BulkPhoneValidationResult
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return &result, nil
}
