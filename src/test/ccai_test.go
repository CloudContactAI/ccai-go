// Copyright (c) 2025 CloudContactAI LLC
// Licensed under the MIT License. See LICENSE in the project root for license information.

package ccai_test

import (
	"os"
	"testing"

	"github.com/cloudcontactai/ccai-go/src/pkg/ccai"
)

func TestNewClientValidConfig(t *testing.T) {
	client, err := ccai.NewClient(ccai.Config{
		ClientID: "test-client-id",
		APIKey:   "test-api-key",
	})

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if client.GetClientID() != "test-client-id" {
		t.Errorf("expected client ID 'test-client-id', got '%s'", client.GetClientID())
	}

	if client.GetAPIKey() != "test-api-key" {
		t.Errorf("expected API key 'test-api-key', got '%s'", client.GetAPIKey())
	}
}

func TestNewClientMissingClientID(t *testing.T) {
	_, err := ccai.NewClient(ccai.Config{
		APIKey: "test-api-key",
	})

	if err == nil {
		t.Fatal("expected error for missing client ID")
	}
}

func TestNewClientMissingAPIKey(t *testing.T) {
	_, err := ccai.NewClient(ccai.Config{
		ClientID: "test-client-id",
	})

	if err == nil {
		t.Fatal("expected error for missing API key")
	}
}

func TestDefaultProductionURLs(t *testing.T) {
	client, _ := ccai.NewClient(ccai.Config{
		ClientID: "test-client-id",
		APIKey:   "test-api-key",
	})

	if client.GetBaseURL() != ccai.ProdBaseURL {
		t.Errorf("expected base URL '%s', got '%s'", ccai.ProdBaseURL, client.GetBaseURL())
	}

	if client.GetEmailBaseURL() != ccai.ProdEmailURL {
		t.Errorf("expected email URL '%s', got '%s'", ccai.ProdEmailURL, client.GetEmailBaseURL())
	}

	if client.GetFilesBaseURL() != ccai.ProdFilesURL {
		t.Errorf("expected files URL '%s', got '%s'", ccai.ProdFilesURL, client.GetFilesBaseURL())
	}

	if client.IsTestEnvironment() {
		t.Error("expected IsTestEnvironment to be false")
	}
}

func TestTestEnvironmentURLs(t *testing.T) {
	client, _ := ccai.NewClient(ccai.Config{
		ClientID:           "test-client-id",
		APIKey:             "test-api-key",
		UseTestEnvironment: true,
	})

	if client.GetBaseURL() != ccai.TestBaseURL {
		t.Errorf("expected base URL '%s', got '%s'", ccai.TestBaseURL, client.GetBaseURL())
	}

	if client.GetEmailBaseURL() != ccai.TestEmailURL {
		t.Errorf("expected email URL '%s', got '%s'", ccai.TestEmailURL, client.GetEmailBaseURL())
	}

	if client.GetFilesBaseURL() != ccai.TestFilesURL {
		t.Errorf("expected files URL '%s', got '%s'", ccai.TestFilesURL, client.GetFilesBaseURL())
	}

	if !client.IsTestEnvironment() {
		t.Error("expected IsTestEnvironment to be true")
	}
}

func TestCustomURLsOverride(t *testing.T) {
	client, _ := ccai.NewClient(ccai.Config{
		ClientID:     "test-client-id",
		APIKey:       "test-api-key",
		BaseURL:      "https://custom.example.com/api",
		EmailBaseURL: "https://email-custom.example.com/api/v1",
		FilesBaseURL: "https://files-custom.example.com",
	})

	if client.GetBaseURL() != "https://custom.example.com/api" {
		t.Errorf("expected custom base URL, got '%s'", client.GetBaseURL())
	}

	if client.GetEmailBaseURL() != "https://email-custom.example.com/api/v1" {
		t.Errorf("expected custom email URL, got '%s'", client.GetEmailBaseURL())
	}

	if client.GetFilesBaseURL() != "https://files-custom.example.com" {
		t.Errorf("expected custom files URL, got '%s'", client.GetFilesBaseURL())
	}
}

func TestEnvVarURLs(t *testing.T) {
	os.Setenv("CCAI_BASE_URL", "https://env-base.example.com")
	defer os.Unsetenv("CCAI_BASE_URL")

	client, _ := ccai.NewClient(ccai.Config{
		ClientID: "test-client-id",
		APIKey:   "test-api-key",
	})

	if client.GetBaseURL() != "https://env-base.example.com" {
		t.Errorf("expected env base URL, got '%s'", client.GetBaseURL())
	}
}

func TestServicesInitialized(t *testing.T) {
	client, _ := ccai.NewClient(ccai.Config{
		ClientID: "test-client-id",
		APIKey:   "test-api-key",
	})

	if client.SMS == nil {
		t.Error("SMS service not initialized")
	}
	if client.MMS == nil {
		t.Error("MMS service not initialized")
	}
	if client.Webhook == nil {
		t.Error("Webhook service not initialized")
	}
	if client.Email == nil {
		t.Error("Email service not initialized")
	}
	if client.Contact == nil {
		t.Error("Contact service not initialized")
	}
}
