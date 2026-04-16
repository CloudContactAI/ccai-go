// Copyright (c) 2025 CloudContactAI LLC
// Licensed under the MIT License. See LICENSE in the project root for license information.

// Package testutil provides shared test utilities for the CCAI Go SDK.
package testutil

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
)

// TestClient implements all SDK client interfaces for testing with httptest.Server
type TestClient struct {
	BaseURL  string
	FilesURL string
}

func (t *TestClient) GetClientID() string          { return "test-client" }
func (t *TestClient) GetAPIKey() string            { return "test-key" }
func (t *TestClient) GetBaseURL() string           { return t.BaseURL }
func (t *TestClient) GetEmailBaseURL() string      { return t.BaseURL }
func (t *TestClient) GetFilesBaseURL() string      { return t.FilesURL }

func (t *TestClient) Request(method, endpoint string, data interface{}, headers map[string]string) ([]byte, error) {
	return doRequest(t.BaseURL+endpoint, method, data, headers)
}

func (t *TestClient) CustomRequest(method, endpoint string, data interface{}, customBaseURL string, headers map[string]string) ([]byte, error) {
	return doRequest(customBaseURL+endpoint, method, data, headers)
}

func doRequest(url, method string, data interface{}, headers map[string]string) ([]byte, error) {
	var body io.Reader
	if data != nil {
		jsonData, _ := json.Marshal(data)
		body = bytes.NewBuffer(jsonData)
	}

	req, _ := http.NewRequest(method, url, body)
	req.Header.Set("Authorization", "Bearer test-key")
	req.Header.Set("Content-Type", "application/json")
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, &APIError{StatusCode: resp.StatusCode, Body: string(respBody)}
	}

	return respBody, nil
}

// APIError represents an API error response
type APIError struct {
	StatusCode int
	Body       string
}

func (e *APIError) Error() string {
	return e.Body
}
