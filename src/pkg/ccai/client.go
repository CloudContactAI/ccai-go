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
	"os"
	"time"

	"github.com/cloudcontactai/ccai-go/src/pkg/contact"
	"github.com/cloudcontactai/ccai-go/src/pkg/contactvalidator"
	"github.com/cloudcontactai/ccai-go/src/pkg/email"
	"github.com/cloudcontactai/ccai-go/src/pkg/sms"
	"github.com/cloudcontactai/ccai-go/src/pkg/webhook"
)

// Production URLs
const (
	ProdBaseURL  = "https://core.cloudcontactai.com/api"
	ProdEmailURL = "https://email-campaigns.cloudcontactai.com/api/v1"
	ProdFilesURL = "https://files.cloudcontactai.com"
)

// Test environment URLs
const (
	TestBaseURL  = "https://core-test-cloudcontactai.allcode.com/api"
	TestEmailURL = "https://email-campaigns-test-cloudcontactai.allcode.com/api/v1"
	TestFilesURL = "https://files-test-cloudcontactai.allcode.com"
)

// Config represents the configuration for the CCAI client.
type Config struct {
	ClientID           string
	APIKey             string
	UseTestEnvironment bool
	BaseURL            string
	EmailBaseURL       string
	FilesBaseURL       string
}

// Account represents a recipient account.
type Account = sms.Account

// EmailAccount represents an email recipient account.
type EmailAccount = email.EmailAccount

// EmailCampaign represents the email campaign configuration.
type EmailCampaign = email.EmailCampaign

// EmailResponse represents the response from the email API.
type EmailResponse = email.EmailResponse

// EmailOptions represents options for email operations.
type EmailOptions = email.EmailOptions

// Client is the main client for interacting with the CloudContactAI API.
type Client struct {
	config           Config
	httpClient       *http.Client
	SMS              *sms.Service
	MMS              *sms.MMSService
	Webhook          *webhook.Service
	Email            *email.Service
	Contact          *contact.Service
	ContactValidator *contactvalidator.Service
}

// NewClient creates a new CCAI client instance.
func NewClient(config Config) (*Client, error) {
	if config.ClientID == "" {
		return nil, fmt.Errorf("client ID is required")
	}

	if config.APIKey == "" {
		return nil, fmt.Errorf("API key is required")
	}

	// Resolve URLs: explicit override > env var > test/prod default
	config.BaseURL = resolveURL(config.BaseURL, "CCAI_BASE_URL", TestBaseURL, ProdBaseURL, config.UseTestEnvironment)
	config.EmailBaseURL = resolveURL(config.EmailBaseURL, "CCAI_EMAIL_BASE_URL", TestEmailURL, ProdEmailURL, config.UseTestEnvironment)
	config.FilesBaseURL = resolveURL(config.FilesBaseURL, "CCAI_FILES_BASE_URL", TestFilesURL, ProdFilesURL, config.UseTestEnvironment)

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

	// Initialize the Webhook service
	client.Webhook = webhook.NewService(client)

	// Initialize the Email service
	client.Email = email.NewService(client)

	// Initialize the Contact service
	client.Contact = contact.NewService(client)

	// Initialize the ContactValidator service
	client.ContactValidator = contactvalidator.NewService(client)

	return client, nil
}

// resolveURL picks the URL in order: explicit > env > default based on test/prod
func resolveURL(explicit, envVar, testDefault, prodDefault string, useTest bool) string {
	if explicit != "" {
		return explicit
	}
	if env := os.Getenv(envVar); env != "" {
		return env
	}
	if useTest {
		return testDefault
	}
	return prodDefault
}

// GetClientID returns the client ID.
func (c *Client) GetClientID() string {
	return c.config.ClientID
}

// GetAPIKey returns the API key.
func (c *Client) GetAPIKey() string {
	return c.config.APIKey
}

// GetBaseURL returns the base URL for the core API.
func (c *Client) GetBaseURL() string {
	return c.config.BaseURL
}

// GetEmailBaseURL returns the base URL for the Email API.
func (c *Client) GetEmailBaseURL() string {
	return c.config.EmailBaseURL
}

// GetFilesBaseURL returns the base URL for the Files API.
func (c *Client) GetFilesBaseURL() string {
	return c.config.FilesBaseURL
}

// IsTestEnvironment returns whether test environment is active.
func (c *Client) IsTestEnvironment() bool {
	return c.config.UseTestEnvironment
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

// CustomRequest makes an authenticated API request to a custom base URL endpoint.
func (c *Client) CustomRequest(method, endpoint string, data interface{}, customBaseURL string, headers map[string]string) ([]byte, error) {
	url := customBaseURL + endpoint

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
