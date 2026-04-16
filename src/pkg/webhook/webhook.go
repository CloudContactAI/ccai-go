// Copyright (c) 2025 CloudContactAI LLC
// Licensed under the MIT License. See LICENSE in the project root for license information.

// Package webhook provides functionality to configure and manage webhooks for CCAI events.
package webhook

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
)

// Client represents a webhook client for managing CloudContactAI webhooks
type Client struct {
	baseURL string
	client  *http.Client
	apiKey  string
}

// NewClient creates a new webhook client
func NewClient(baseURL, apiKey string) *Client {
	return &Client{
		baseURL: baseURL,
		client:  &http.Client{},
		apiKey:  apiKey,
	}
}

// Register registers a new webhook endpoint
func (w *Client) Register(config WebhookConfig) (*WebhookResponse, error) {
	url := fmt.Sprintf("%s/webhooks", w.baseURL)
	return w.makeRequest("POST", url, config)
}

// Update updates an existing webhook configuration
func (w *Client) Update(id string, config WebhookConfig) (*WebhookResponse, error) {
	url := fmt.Sprintf("%s/webhooks/%s", w.baseURL, id)
	return w.makeRequest("PUT", url, config)
}

// List lists all registered webhooks
func (w *Client) List() ([]WebhookResponse, error) {
	url := fmt.Sprintf("%s/webhooks", w.baseURL)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+w.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := w.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to list webhooks: status %d", resp.StatusCode)
	}

	var webhooks []WebhookResponse
	if err := json.NewDecoder(resp.Body).Decode(&webhooks); err != nil {
		return nil, err
	}

	return webhooks, nil
}

// Delete deletes a webhook
func (w *Client) Delete(id string) (*WebhookDeleteResponse, error) {
	url := fmt.Sprintf("%s/webhooks/%s", w.baseURL, id)

	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+w.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := w.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to delete webhook: status %d", resp.StatusCode)
	}

	var deleteResp WebhookDeleteResponse
	if err := json.NewDecoder(resp.Body).Decode(&deleteResp); err != nil {
		return nil, err
	}

	return &deleteResp, nil
}

// VerifySignature verifies a webhook signature using HMAC-SHA256
// Signature is computed as: HMAC-SHA256(secretKey, clientId:eventHash) encoded in Base64
func (w *Client) VerifySignature(signature string, clientID string, eventHash string, secret string) bool {
	// Compute: HMAC-SHA256(secretKey, "$clientId:$eventHash")
	data := fmt.Sprintf("%s:%s", clientID, eventHash)
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(data))

	// Encode the result in Base64
	expectedSignature := base64.StdEncoding.EncodeToString(h.Sum(nil))

	// Compare signatures (constant time comparison)
	return hmac.Equal([]byte(signature), []byte(expectedSignature))
}

// makeRequest is a helper method to make HTTP requests with webhook responses
func (w *Client) makeRequest(method, url string, data interface{}) (*WebhookResponse, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(method, url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+w.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := w.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("request failed: status %d", resp.StatusCode)
	}

	var webhookResp WebhookResponse
	if err := json.NewDecoder(resp.Body).Decode(&webhookResp); err != nil {
		return nil, err
	}

	return &webhookResp, nil
}
