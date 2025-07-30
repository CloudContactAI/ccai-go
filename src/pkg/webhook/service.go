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
}

// NewService creates a new webhook service
func NewService(client HTTPClient) *Service {
	return &Service{
		client: client,
	}
}

// Register registers a new webhook endpoint
func (s *Service) Register(config WebhookConfig) (*WebhookResponse, error) {
	data, err := s.client.Request("POST", "/webhooks", config, nil)
	if err != nil {
		return nil, err
	}

	var response WebhookResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, err
	}

	return &response, nil
}

// Update updates an existing webhook configuration
func (s *Service) Update(id string, config WebhookConfig) (*WebhookResponse, error) {
	path := fmt.Sprintf("/webhooks/%s", id)
	data, err := s.client.Request("PUT", path, config, nil)
	if err != nil {
		return nil, err
	}

	var response WebhookResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, err
	}

	return &response, nil
}

// List lists all registered webhooks
func (s *Service) List() ([]WebhookResponse, error) {
	data, err := s.client.Request("GET", "/webhooks", nil, nil)
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
	path := fmt.Sprintf("/webhooks/%s", id)
	data, err := s.client.Request("DELETE", path, nil, nil)
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
func (s *Service) VerifySignature(signature, body, secret string) bool {
	client := &Client{}
	return client.VerifySignature(signature, body, secret)
}
