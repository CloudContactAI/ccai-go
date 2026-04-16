// Copyright (c) 2025 CloudContactAI LLC
// Licensed under the MIT License. See LICENSE in the project root for license information.

package ccai_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cloudcontactai/ccai-go/src/pkg/email"
	"github.com/cloudcontactai/ccai-go/src/pkg/testutil"
)

func TestEmailSendSingle(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		id := int64(123)
		status := "PENDING"
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":     id,
			"status": status,
		})
	}))
	defer server.Close()

	tc := &testutil.TestClient{BaseURL: server.URL, FilesURL: server.URL}
	svc := email.NewService(tc)

	res, err := svc.SendSingle("John", "Doe", "john@example.com",
		"Test Subject", "<p>Test message</p>",
		"sender@example.com", "reply@example.com", "Test Sender", "Test Campaign", nil)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if res.ID == nil || *res.ID != 123 {
		t.Errorf("expected ID 123, got %v", res.ID)
	}
}

func TestEmailSend(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		id := int64(456)
		status := "SENT"
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":     id,
			"status": status,
		})
	}))
	defer server.Close()

	tc := &testutil.TestClient{BaseURL: server.URL, FilesURL: server.URL}
	svc := email.NewService(tc)

	accounts := []email.EmailAccount{
		{FirstName: "John", LastName: "Doe", Email: "john@example.com", Phone: ""},
	}

	res, err := svc.Send(accounts, "Subject", "<p>Message</p>",
		"sender@example.com", "reply@example.com", "Sender", "Campaign", nil)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if res.ID == nil || *res.ID != 456 {
		t.Errorf("expected ID 456, got %v", res.ID)
	}
}

func TestEmailCampaignValidation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	tc := &testutil.TestClient{BaseURL: server.URL, FilesURL: server.URL}
	svc := email.NewService(tc)

	// Empty accounts
	_, err := svc.SendCampaign(&email.EmailCampaign{
		Accounts:    []email.EmailAccount{},
		Subject:     "Test",
		Title:       "Test",
		Message:     "Test",
		SenderEmail: "test@test.com",
		ReplyEmail:  "test@test.com",
		SenderName:  "Test",
	}, nil)
	if err == nil {
		t.Fatal("expected error for empty accounts")
	}

	// Missing subject
	_, err = svc.SendCampaign(&email.EmailCampaign{
		Accounts:    []email.EmailAccount{{FirstName: "John", LastName: "Doe", Email: "j@e.com"}},
		Title:       "Test",
		Message:     "Test",
		SenderEmail: "test@test.com",
		ReplyEmail:  "test@test.com",
		SenderName:  "Test",
	}, nil)
	if err == nil {
		t.Fatal("expected error for missing subject")
	}
}

func TestEmailAccountValidation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	tc := &testutil.TestClient{BaseURL: server.URL, FilesURL: server.URL}
	svc := email.NewService(tc)

	// Missing first name
	_, err := svc.SendCampaign(&email.EmailCampaign{
		Accounts:    []email.EmailAccount{{LastName: "Doe", Email: "j@e.com"}},
		Subject:     "Test",
		Title:       "Test",
		Message:     "Test",
		SenderEmail: "test@test.com",
		ReplyEmail:  "test@test.com",
		SenderName:  "Test",
	}, nil)
	if err == nil {
		t.Fatal("expected error for missing first name")
	}

	// Missing email
	_, err = svc.SendCampaign(&email.EmailCampaign{
		Accounts:    []email.EmailAccount{{FirstName: "John", LastName: "Doe"}},
		Subject:     "Test",
		Title:       "Test",
		Message:     "Test",
		SenderEmail: "test@test.com",
		ReplyEmail:  "test@test.com",
		SenderName:  "Test",
	}, nil)
	if err == nil {
		t.Fatal("expected error for missing email in account")
	}
}

func TestEmailAccountCustomFieldsAndID(t *testing.T) {
	var capturedBody []byte

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedBody = make([]byte, r.ContentLength)
		r.Body.Read(capturedBody)
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		id := int64(999)
		status := "PENDING"
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":         id,
			"status":     status,
			"message":    "Email campaign sent successfully",
			"responseId": "resp-email-xyz",
		})
	}))
	defer server.Close()

	tc := &testutil.TestClient{BaseURL: server.URL, FilesURL: server.URL}
	svc := email.NewService(tc)

	accounts := []email.EmailAccount{
		{
			FirstName:       "John",
			LastName:        "Doe",
			Email:           "john@example.com",
			Phone:           "",
			CustomAccountID: "ext-id-123",
			Data:            map[string]string{"tier": "gold", "locale": "en-US"},
		},
	}

	res, err := svc.Send(accounts, "Subject", "<p>Hi</p>",
		"sender@test.com", "reply@test.com", "Sender", "Campaign", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify request body contains customAccountId and data keys
	bodyStr := string(capturedBody)
	for _, expected := range []string{`"customAccountId"`, `"ext-id-123"`, `"data"`, `"gold"`} {
		found := false
		for i := 0; i <= len(bodyStr)-len(expected); i++ {
			if bodyStr[i:i+len(expected)] == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected request body to contain %q, got: %s", expected, bodyStr)
		}
	}

	// Verify response fields
	if res.Message != "Email campaign sent successfully" {
		t.Errorf("expected message 'Email campaign sent successfully', got '%s'", res.Message)
	}
	if res.ResponseID != "resp-email-xyz" {
		t.Errorf("expected responseId 'resp-email-xyz', got '%s'", res.ResponseID)
	}
}

func TestEmailProgress(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		id := int64(789)
		json.NewEncoder(w).Encode(map[string]interface{}{"id": id})
	}))
	defer server.Close()

	tc := &testutil.TestClient{BaseURL: server.URL, FilesURL: server.URL}
	svc := email.NewService(tc)

	progressUpdates := []string{}
	options := &email.EmailOptions{
		OnProgress: func(status string) {
			progressUpdates = append(progressUpdates, status)
		},
	}

	accounts := []email.EmailAccount{{FirstName: "John", LastName: "Doe", Email: "j@e.com"}}

	_, err := svc.SendCampaign(&email.EmailCampaign{
		Accounts:     accounts,
		Subject:      "Test",
		Title:        "Test",
		Message:      "Test",
		SenderEmail:  "test@test.com",
		ReplyEmail:   "test@test.com",
		SenderName:   "Test",
		CampaignType: "EMAIL",
		AddToList:    "noList",
		ContactInput: "accounts",
		FromType:     "single",
		Senders:      []interface{}{},
	}, options)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := []string{"Preparing to send email campaign", "Sending email campaign", "Email campaign sent successfully"}
	for i, exp := range expected {
		if progressUpdates[i] != exp {
			t.Errorf("update[%d]: expected '%s', got '%s'", i, exp, progressUpdates[i])
		}
	}
}
