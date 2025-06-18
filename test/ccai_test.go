// Copyright (c) 2025 CloudContactAI LLC
// Licensed under the MIT License. See LICENSE in the project root for license information.

package test

import (
	"testing"

	"github.com/cloudcontactai/ccai-go/pkg/ccai"
	"github.com/cloudcontactai/ccai-go/pkg/sms"
)

func TestNewClient(t *testing.T) {
	// Test with valid config
	client, err := ccai.NewClient(ccai.Config{
		ClientID: "test-client-id",
		APIKey:   "test-api-key",
	})
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if client == nil {
		t.Fatal("Expected client to be non-nil")
	}
	if client.GetClientID() != "test-client-id" {
		t.Errorf("Expected client ID to be 'test-client-id', got '%s'", client.GetClientID())
	}
	if client.GetAPIKey() != "test-api-key" {
		t.Errorf("Expected API key to be 'test-api-key', got '%s'", client.GetAPIKey())
	}
	if client.GetBaseURL() != "https://core.cloudcontactai.com/api" {
		t.Errorf("Expected base URL to be 'https://core.cloudcontactai.com/api', got '%s'", client.GetBaseURL())
	}

	// Test with custom base URL
	client, err = ccai.NewClient(ccai.Config{
		ClientID: "test-client-id",
		APIKey:   "test-api-key",
		BaseURL:  "https://custom-api.example.com",
	})
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if client.GetBaseURL() != "https://custom-api.example.com" {
		t.Errorf("Expected base URL to be 'https://custom-api.example.com', got '%s'", client.GetBaseURL())
	}

	// Test with empty client ID
	_, err = ccai.NewClient(ccai.Config{
		ClientID: "",
		APIKey:   "test-api-key",
	})
	if err == nil {
		t.Fatal("Expected error for empty client ID, got nil")
	}

	// Test with empty API key
	_, err = ccai.NewClient(ccai.Config{
		ClientID: "test-client-id",
		APIKey:   "",
	})
	if err == nil {
		t.Fatal("Expected error for empty API key, got nil")
	}
}

func TestClientServices(t *testing.T) {
	client, err := ccai.NewClient(ccai.Config{
		ClientID: "test-client-id",
		APIKey:   "test-api-key",
	})
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Test SMS service
	if client.SMS == nil {
		t.Fatal("Expected SMS service to be non-nil")
	}
	if _, ok := client.SMS.(*sms.Service); !ok {
		t.Errorf("Expected SMS service to be of type *sms.Service")
	}

	// Test MMS service
	if client.MMS == nil {
		t.Fatal("Expected MMS service to be non-nil")
	}
	if _, ok := client.MMS.(*sms.MMSService); !ok {
		t.Errorf("Expected MMS service to be of type *sms.MMSService")
	}
}
