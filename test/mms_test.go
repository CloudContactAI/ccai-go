// Copyright (c) 2025 CloudContactAI LLC
// Licensed under the MIT License. See LICENSE in the project root for license information.

package test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/cloudcontactai/ccai-go/pkg/sms"
)

// MockClientWithCustomResponse is a mock client that returns custom responses.
type MockClientWithCustomResponse struct {
	*MockClient
	requestFunc func(method, endpoint string, data interface{}, headers map[string]string) ([]byte, error)
}

func NewMockClientWithCustomResponse(requestFunc func(method, endpoint string, data interface{}, headers map[string]string) ([]byte, error)) *MockClientWithCustomResponse {
	return &MockClientWithCustomResponse{
		MockClient:  NewMockClient(),
		requestFunc: requestFunc,
	}
}

func (m *MockClientWithCustomResponse) Request(method, endpoint string, data interface{}, headers map[string]string) ([]byte, error) {
	return m.requestFunc(method, endpoint, data, headers)
}

func TestMMSSend(t *testing.T) {
	mockClient := NewMockClient()
	mmsService := sms.NewMMSService(mockClient)

	// Test with valid inputs
	pictureFileKey := "test-client-id/campaign/test-image.jpg"
	accounts := []sms.Account{
		{
			FirstName: "John",
			LastName:  "Doe",
			Phone:     "+15551234567",
		},
	}
	message := "Hello ${firstName}, check out this image!"
	title := "Test MMS Campaign"

	response, err := mmsService.Send(pictureFileKey, accounts, message, title, nil, true)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if response == nil {
		t.Fatal("Expected response to be non-nil")
	}
	if response.ID != "msg-123" {
		t.Errorf("Expected ID to be 'msg-123', got '%s'", response.ID)
	}
	if response.Status != "sent" {
		t.Errorf("Expected status to be 'sent', got '%s'", response.Status)
	}
	if response.CampaignID != "camp-456" {
		t.Errorf("Expected campaign ID to be 'camp-456', got '%s'", response.CampaignID)
	}

	// Test with empty pictureFileKey
	_, err = mmsService.Send("", accounts, message, title, nil, true)
	if err == nil {
		t.Fatal("Expected error for empty pictureFileKey, got nil")
	}

	// Test with empty accounts
	_, err = mmsService.Send(pictureFileKey, []sms.Account{}, message, title, nil, true)
	if err == nil {
		t.Fatal("Expected error for empty accounts, got nil")
	}

	// Test with empty message
	_, err = mmsService.Send(pictureFileKey, accounts, "", title, nil, true)
	if err == nil {
		t.Fatal("Expected error for empty message, got nil")
	}

	// Test with empty title
	_, err = mmsService.Send(pictureFileKey, accounts, message, "", nil, true)
	if err == nil {
		t.Fatal("Expected error for empty title, got nil")
	}

	// Test with progress tracking
	progressUpdates := []string{}
	options := &sms.Options{
		OnProgress: func(status string) {
			progressUpdates = append(progressUpdates, status)
		},
	}

	_, err = mmsService.Send(pictureFileKey, accounts, message, title, options, true)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if len(progressUpdates) != 3 {
		t.Errorf("Expected 3 progress updates, got %d", len(progressUpdates))
	}
	if progressUpdates[0] != "Preparing to send MMS" {
		t.Errorf("Expected first progress update to be 'Preparing to send MMS', got '%s'", progressUpdates[0])
	}
	if progressUpdates[1] != "Sending MMS" {
		t.Errorf("Expected second progress update to be 'Sending MMS', got '%s'", progressUpdates[1])
	}
	if progressUpdates[2] != "MMS sent successfully" {
		t.Errorf("Expected third progress update to be 'MMS sent successfully', got '%s'", progressUpdates[2])
	}
}

func TestMMSSendSingle(t *testing.T) {
	mockClient := NewMockClient()
	mmsService := sms.NewMMSService(mockClient)

	pictureFileKey := "test-client-id/campaign/test-image.jpg"
	firstName := "Jane"
	lastName := "Smith"
	phone := "+15559876543"
	message := "Hi ${firstName}, check out this image!"
	title := "Single MMS Test"

	response, err := mmsService.SendSingle(pictureFileKey, firstName, lastName, phone, message, title, nil, true)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if response == nil {
		t.Fatal("Expected response to be non-nil")
	}
	if response.ID != "msg-123" {
		t.Errorf("Expected ID to be 'msg-123', got '%s'", response.ID)
	}
}

func TestGetSignedUploadURL(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check request method
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}

		// Check authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader != "Bearer test-api-key" {
			t.Errorf("Expected Authorization header 'Bearer test-api-key', got '%s'", authHeader)
		}

		// Return a mock response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, `{"signedS3Url":"https://s3.amazonaws.com/bucket/signed-url","fileKey":"original/file/key"}`)
	}))
	defer server.Close()

	// Create a mock client that uses the test server URL
	mockClient := &MockClient{
		clientID: "test-client-id",
		apiKey:   "test-api-key",
		baseURL:  server.URL,
	}

	mmsService := sms.NewMMSService(mockClient)

	// Test with valid inputs
	fileName := "test-image.jpg"
	fileType := "image/jpeg"

	response, err := mmsService.GetSignedUploadURL(fileName, fileType, "", true)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if response == nil {
		t.Fatal("Expected response to be non-nil")
	}
	if response.SignedS3URL != "https://s3.amazonaws.com/bucket/signed-url" {
		t.Errorf("Expected SignedS3URL to be 'https://s3.amazonaws.com/bucket/signed-url', got '%s'", response.SignedS3URL)
	}
	if response.FileKey != "test-client-id/campaign/test-image.jpg" {
		t.Errorf("Expected FileKey to be 'test-client-id/campaign/test-image.jpg', got '%s'", response.FileKey)
	}

	// Test with empty fileName
	_, err = mmsService.GetSignedUploadURL("", fileType, "", true)
	if err == nil {
		t.Fatal("Expected error for empty fileName, got nil")
	}

	// Test with empty fileType
	_, err = mmsService.GetSignedUploadURL(fileName, "", "", true)
	if err == nil {
		t.Fatal("Expected error for empty fileType, got nil")
	}
}

func TestMMSSendWithImage(t *testing.T) {
	// Create a mock client that simulates the complete workflow
	mockClient := NewMockClientWithCustomResponse(func(method, endpoint string, data interface{}, headers map[string]string) ([]byte, error) {
		return []byte(`{"id":"msg-123","status":"sent","campaignId":"camp-456","messagesSent":1,"timestamp":"2025-06-06T12:00:00Z"}`), nil
	})

	mmsService := sms.NewMMSService(mockClient)

	// Create a temporary test file
	tempDir := t.TempDir()
	imagePath := filepath.Join(tempDir, "test-image.jpg")
	if err := os.WriteFile(imagePath, []byte("test image data"), 0644); err != nil {
		t.Fatalf("Failed to create test image file: %v", err)
	}

	// Mock the GetSignedUploadURL and UploadImageToSignedURL methods
	originalGetSignedUploadURL := mmsService.GetSignedUploadURL
	originalUploadImageToSignedURL := mmsService.UploadImageToSignedURL
	originalSend := mmsService.Send

	// Replace with mock implementations for this test
	mmsService.GetSignedUploadURL = func(fileName, fileType, fileBasePath string, publicFile bool) (*sms.SignedURLResponse, error) {
		return &sms.SignedURLResponse{
			SignedS3URL: "https://s3.amazonaws.com/bucket/signed-url",
			FileKey:     "test-client-id/campaign/test-image.jpg",
		}, nil
	}

	mmsService.UploadImageToSignedURL = func(signedURL, filePath, contentType string) (bool, error) {
		return true, nil
	}

	mmsService.Send = func(pictureFileKey string, accounts []sms.Account, message, title string, options *sms.Options, forceNewCampaign bool) (*sms.Response, error) {
		return &sms.Response{
			ID:           "msg-123",
			Status:       "sent",
			CampaignID:   "camp-456",
			MessagesSent: 1,
			Timestamp:    "2025-06-06T12:00:00Z",
		}, nil
	}

	// Restore original methods after the test
	defer func() {
		mmsService.GetSignedUploadURL = originalGetSignedUploadURL
		mmsService.UploadImageToSignedURL = originalUploadImageToSignedURL
		mmsService.Send = originalSend
	}()

	// Test with valid inputs
	accounts := []sms.Account{
		{
			FirstName: "John",
			LastName:  "Doe",
			Phone:     "+15551234567",
		},
	}
	message := "Hello ${firstName}, check out this image!"
	title := "Test MMS Campaign"
	contentType := "image/jpeg"

	// Track progress
	progressUpdates := []string{}
	options := &sms.Options{
		OnProgress: func(status string) {
			progressUpdates = append(progressUpdates, status)
		},
	}

	response, err := mmsService.SendWithImage(imagePath, contentType, accounts, message, title, options, true)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if response == nil {
		t.Fatal("Expected response to be non-nil")
	}
	if response.ID != "msg-123" {
		t.Errorf("Expected ID to be 'msg-123', got '%s'", response.ID)
	}
	if response.Status != "sent" {
		t.Errorf("Expected status to be 'sent', got '%s'", response.Status)
	}
	if response.CampaignID != "camp-456" {
		t.Errorf("Expected campaign ID to be 'camp-456', got '%s'", response.CampaignID)
	}

	// Check progress updates
	expectedUpdates := []string{
		"Getting signed upload URL",
		"Uploading image to S3",
		"Image uploaded successfully, sending MMS",
	}
	if len(progressUpdates) != len(expectedUpdates) {
		t.Errorf("Expected %d progress updates, got %d", len(expectedUpdates), len(progressUpdates))
	}
	for i, update := range progressUpdates {
		if i < len(expectedUpdates) && update != expectedUpdates[i] {
			t.Errorf("Expected progress update %d to be '%s', got '%s'", i, expectedUpdates[i], update)
		}
	}
}
