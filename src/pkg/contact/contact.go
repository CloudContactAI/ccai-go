// Copyright (c) 2025 CloudContactAI LLC
// Licensed under the MIT License. See LICENSE in the project root for license information.

// Package contact provides functionality for managing contact preferences (opt-out).
package contact

import (
	"encoding/json"
	"fmt"
)

// ClientInterface defines the interface that the CCAI client must implement.
type ClientInterface interface {
	Request(method, endpoint string, data interface{}, headers map[string]string) ([]byte, error)
	GetClientID() string
}

// Service provides contact preference management.
type Service struct {
	client ClientInterface
}

// SetDoNotTextRequest represents the request payload.
type SetDoNotTextRequest struct {
	ClientID  string `json:"clientId"`
	ContactID string `json:"contactId,omitempty"`
	Phone     string `json:"phone,omitempty"`
	DoNotText bool   `json:"doNotText"`
}

// SetDoNotTextResponse represents the API response.
type SetDoNotTextResponse struct {
	ContactID string `json:"contactId,omitempty"`
	Phone     string `json:"phone,omitempty"`
	DoNotText bool   `json:"doNotText,omitempty"`
}

// NewService creates a new contact service.
func NewService(client ClientInterface) *Service {
	return &Service{
		client: client,
	}
}

// SetDoNotText sets the do-not-text preference for a contact.
// Either contactID or phone should be provided.
func (s *Service) SetDoNotText(doNotText bool, contactID string, phone string) (*SetDoNotTextResponse, error) {
	req := SetDoNotTextRequest{
		ClientID:  s.client.GetClientID(),
		DoNotText: doNotText,
	}

	if contactID != "" {
		req.ContactID = contactID
	}
	if phone != "" {
		req.Phone = phone
	}

	responseData, err := s.client.Request("PUT", "/account/do-not-text", req, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to set do-not-text: %w", err)
	}

	var resp SetDoNotTextResponse
	if err := json.Unmarshal(responseData, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &resp, nil
}
