// Copyright (c) 2025 CloudContactAI LLC
// Licensed under the MIT License. See LICENSE in the project root for license information.

package test

import (
	"fmt"
	"testing"

	"github.com/cloudcontactai/ccai-go/pkg/sms"
)

// MockClient is a mock implementation of the ClientInterface for testing.
type MockClient struct {
	clientID string
	apiKey   string
	baseURL  string
}

func NewMockClient() *MockClient {
	return &MockClient{
		clientID: "test-client-id",
		apiKey:   "test-api-key",
		baseURL:  "https://test-api.example.com",
	}
}

func (m *MockClient) GetClientID() string {
	return m.clientID
}

func (m *MockClient) GetAPIKey() string {
	return m.apiKey
}

func (m *MockClient) GetBaseURL() string {
	return m.baseURL
}

func (m *MockClient) Request(method, endpoint string, data interface{}, headers map[string]string) ([]byte, error) {
	// Mock successful response
	return []byte(`{"id":"msg-123","status":"sent","campaignId":"camp-456","messagesSent":1,"timestamp":"2025-06-06T12:00:00Z"}`), nil
}

func TestSMSSend(t *testing.T) {
	mockClient := NewMockClient()
	smsService := sms.NewService(mockClient)

	// Test with valid inputs
	accounts := []sms.Account{
		{
			FirstName: "John",
			LastName:  "Doe",
			Phone:     "+15551234567",
		},
	}
	message := "Hello ${firstName}, this is a test message!"
	title := "Test Campaign"

	response, err := smsService.Send(accounts, message, title, nil)
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
	if response.MessagesSent != 1 {
		t.Errorf("Expected messages sent to be 1, got %d", response.MessagesSent)
	}
	if response.Timestamp != "2025-06-06T12:00:00Z" {
		t.Errorf("Expected timestamp to be '2025-06-06T12:00:00Z', got '%s'", response.Timestamp)
	}

	// Test with empty accounts
	_, err = smsService.Send([]sms.Account{}, message, title, nil)
	if err == nil {
		t.Fatal("Expected error for empty accounts, got nil")
	}

	// Test with empty message
	_, err = smsService.Send(accounts, "", title, nil)
	if err == nil {
		t.Fatal("Expected error for empty message, got nil")
	}

	// Test with empty title
	_, err = smsService.Send(accounts, message, "", nil)
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

	_, err = smsService.Send(accounts, message, title, options)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if len(progressUpdates) != 3 {
		t.Errorf("Expected 3 progress updates, got %d", len(progressUpdates))
	}
	if progressUpdates[0] != "Preparing to send SMS" {
		t.Errorf("Expected first progress update to be 'Preparing to send SMS', got '%s'", progressUpdates[0])
	}
	if progressUpdates[1] != "Sending SMS" {
		t.Errorf("Expected second progress update to be 'Sending SMS', got '%s'", progressUpdates[1])
	}
	if progressUpdates[2] != "SMS sent successfully" {
		t.Errorf("Expected third progress update to be 'SMS sent successfully', got '%s'", progressUpdates[2])
	}
}

func TestSMSSendSingle(t *testing.T) {
	mockClient := NewMockClient()
	smsService := sms.NewService(mockClient)

	firstName := "Jane"
	lastName := "Smith"
	phone := "+15559876543"
	message := "Hi ${firstName}, thanks for your interest!"
	title := "Single Message Test"

	response, err := smsService.SendSingle(firstName, lastName, phone, message, title, nil)
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
