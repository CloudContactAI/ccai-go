// Copyright (c) 2025 CloudContactAI LLC
// Licensed under the MIT License. See LICENSE in the project root for license information.

// Package webhook provides a service for managing CloudContactAI webhooks.
package webhook

import (
	"encoding/json"
	"fmt"
)

// Service represents the webhook service for managing CloudContactAI webhooks
type Service struct {
	client HTTPClient
}

// HTTPClient interface for making HTTP requests
type HTTPClient interface {
	Request(method, path string, data interface{}, headers map[string]string) ([]byte, error)
	GetClientID() string
}

// NewService creates a new webhook service
func NewService(client HTTPClient) *Service {
	return &Service{
		client: client,
	}
}

// registerPayload wraps a WebhookConfig for the API request
type registerPayload struct {
	URL             string `json:"url"`
	Method          string `json:"method"`
	IntegrationType string `json:"integrationType"`
	SecretKey       string `json:"secretKey,omitempty"`
}

// updatePayload wraps a WebhookConfig for the update API request
type updatePayload struct {
	ID              int    `json:"id"`
	URL             string `json:"url"`
	Method          string `json:"method"`
	IntegrationType string `json:"integrationType"`
	SecretKey       string `json:"secretKey,omitempty"`
}

// Register registers a new webhook endpoint
// If config.Secret is nil, the server will generate a secret automatically
func (s *Service) Register(config WebhookConfig) (*WebhookResponse, error) {
	webhookPayload := registerPayload{
		URL:             config.URL,
		Method:          "POST",
		IntegrationType: "ALL",
	}

	// Only include SecretKey if explicitly provided
	if config.Secret != nil {
		webhookPayload.SecretKey = *config.Secret
	}

	payload := []registerPayload{webhookPayload}

	endpoint := fmt.Sprintf("/v1/client/%s/integration", s.client.GetClientID())
	data, err := s.client.Request("POST", endpoint, payload, nil)
	if err != nil {
		return nil, err
	}

	// API returns an array — return the first element
	var responses []WebhookResponse
	if err := json.Unmarshal(data, &responses); err != nil {
		return nil, err
	}

	if len(responses) > 0 {
		return &responses[0], nil
	}

	return nil, fmt.Errorf("empty response from register webhook")
}

// Update updates an existing webhook configuration
func (s *Service) Update(id string, config WebhookConfig) (*WebhookResponse, error) {
	webhookID := 0
	fmt.Sscanf(id, "%d", &webhookID)

	webhookPayload := updatePayload{
		ID:              webhookID,
		URL:             config.URL,
		Method:          "POST",
		IntegrationType: "ALL",
	}

	// Only include SecretKey if explicitly provided
	if config.Secret != nil {
		webhookPayload.SecretKey = *config.Secret
	}

	payload := []updatePayload{webhookPayload}

	endpoint := fmt.Sprintf("/v1/client/%s/integration", s.client.GetClientID())
	data, err := s.client.Request("POST", endpoint, payload, nil)
	if err != nil {
		return nil, err
	}

	// API returns an array — return the first element
	var responses []WebhookResponse
	if err := json.Unmarshal(data, &responses); err != nil {
		return nil, err
	}

	if len(responses) > 0 {
		return &responses[0], nil
	}

	return nil, fmt.Errorf("empty response from update webhook")
}

// List lists all registered webhooks
func (s *Service) List() ([]WebhookResponse, error) {
	endpoint := fmt.Sprintf("/v1/client/%s/integration", s.client.GetClientID())
	data, err := s.client.Request("GET", endpoint, nil, nil)
	if err != nil {
		return nil, err
	}

	var webhooks []WebhookResponse
	if err := json.Unmarshal(data, &webhooks); err != nil {
		return nil, err
	}

	return webhooks, nil
}

// Delete deletes a webhook
func (s *Service) Delete(id string) (*WebhookDeleteResponse, error) {
	endpoint := fmt.Sprintf("/v1/client/%s/integration/%s", s.client.GetClientID(), id)
	data, err := s.client.Request("DELETE", endpoint, nil, nil)
	if err != nil {
		return nil, err
	}

	var response WebhookDeleteResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, err
	}

	return &response, nil
}

// VerifySignature verifies a webhook signature using HMAC-SHA256
// Signature is computed as: HMAC-SHA256(secretKey, clientId:eventHash) encoded in Base64
func (s *Service) VerifySignature(signature, clientID, eventHash, secret string) bool {
	client := &Client{}
	return client.VerifySignature(signature, clientID, eventHash, secret)
}
