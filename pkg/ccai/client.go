// Copyright (c) 2025 CloudContactAI LLC
// Licensed under the MIT License. See LICENSE in the project root for license information.

// Package ccai provides a client for interacting with the Cloud Contact AI API.
package ccai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/cloudcontactai/ccai-go/pkg/sms"
)

// Config represents the configuration for the CCAI client.
type Config struct {
	ClientID string
	APIKey   string
	BaseURL  string
}

// Client is the main client for interacting with the CloudContactAI API.
type Client struct {
	config     Config
	httpClient *http.Client
	SMS        *sms.Service
	MMS        *sms.MMSService
}

// NewClient creates a new CCAI client instance.
func NewClient(config Config) (*Client, error) {
	if config.ClientID == "" {
		return nil, fmt.Errorf("client ID is required")
	}

	if config.APIKey == "" {
		return nil, fmt.Errorf("API key is required")
	}

	// Set default base URL if not provided
	if config.BaseURL == "" {
		config.BaseURL = "https://core.cloudcontactai.com/api"
	}

	httpClient := &http.Client{
		Timeout: 30 * time.Second,
	}

	client := &Client{
		config:     config,
		httpClient: httpClient,
	}

	// Initialize the SMS service
	client.SMS = sms.NewService(client)
	
	// Initialize the MMS service
	client.MMS = sms.NewMMSService(client)

	return client, nil
}

// GetClientID returns the client ID.
func (c *Client) GetClientID() string {
	return c.config.ClientID
}

// GetAPIKey returns the API key.
func (c *Client) GetAPIKey() string {
	return c.config.APIKey
}

// GetBaseURL returns the base URL.
func (c *Client) GetBaseURL() string {
	return c.config.BaseURL
}

// Request makes an authenticated API request to the CCAI API.
func (c *Client) Request(method, endpoint string, data interface{}, headers map[string]string) ([]byte, error) {
	url := c.config.BaseURL + endpoint

	var reqBody io.Reader
	if data != nil {
		jsonData, err := json.Marshal(data)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request data: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set default headers
	req.Header.Set("Authorization", "Bearer "+c.config.APIKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "*/*")

	// Set additional headers if provided
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("API error: %d - %s", resp.StatusCode, string(body))
	}

	return body, nil
}
