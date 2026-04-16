// Copyright (c) 2025 CloudContactAI LLC
// Licensed under the MIT License. See LICENSE in the project root for license information.

package ccai_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cloudcontactai/ccai-go/src/pkg/sms"
	"github.com/cloudcontactai/ccai-go/src/pkg/testutil"
)

func TestSendSingle(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/clients/test-client/campaigns/direct" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":           "msg-123",
			"status":       "sent",
			"campaignId":   "camp-456",
			"messagesSent": 1,
		})
	}))
	defer server.Close()

	tc := &testutil.TestClient{BaseURL: server.URL, FilesURL: server.URL}
	svc := sms.NewService(tc)

	res, err := svc.SendSingle("John", "Doe", "+15551234567",
		"Hello ${firstName}!", "Test Campaign", "", "", nil)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if res.GetID() != "msg-123" {
		t.Errorf("expected ID 'msg-123', got '%s'", res.GetID())
	}

	if res.Status != "sent" {
		t.Errorf("expected status 'sent', got '%s'", res.Status)
	}
}

func TestSendMultiple(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":     "msg-789",
			"status": "sent",
		})
	}))
	defer server.Close()

	tc := &testutil.TestClient{BaseURL: server.URL, FilesURL: server.URL}
	svc := sms.NewService(tc)

	accounts := []sms.Account{
		{FirstName: "John", LastName: "Doe", Phone: "+15551234567"},
		{FirstName: "Jane", LastName: "Smith", Phone: "+15559876543"},
	}

	res, err := svc.Send(accounts, "Hello all!", "Bulk Campaign", "", nil)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if res.GetID() != "msg-789" {
		t.Errorf("expected ID 'msg-789', got '%s'", res.GetID())
	}
}

func TestSendValidationEmptyAccounts(t *testing.T) {
	svc := sms.NewService(&testutil.TestClient{BaseURL: "http://test"})

	_, err := svc.Send([]sms.Account{}, "Hello", "Test", "", nil)
	if err == nil {
		t.Fatal("expected error for empty accounts")
	}
}

func TestSendValidationEmptyMessage(t *testing.T) {
	svc := sms.NewService(&testutil.TestClient{BaseURL: "http://test"})
	accounts := []sms.Account{{FirstName: "John", LastName: "Doe", Phone: "+15551234567"}}

	_, err := svc.Send(accounts, "", "Test", "", nil)
	if err == nil {
		t.Fatal("expected error for empty message")
	}
}

func TestSendValidationEmptyTitle(t *testing.T) {
	svc := sms.NewService(&testutil.TestClient{BaseURL: "http://test"})
	accounts := []sms.Account{{FirstName: "John", LastName: "Doe", Phone: "+15551234567"}}

	_, err := svc.Send(accounts, "Hello", "", "", nil)
	if err == nil {
		t.Fatal("expected error for empty title")
	}
}

func TestSendWithError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"error": "Bad request"})
	}))
	defer server.Close()

	tc := &testutil.TestClient{BaseURL: server.URL, FilesURL: server.URL}
	svc := sms.NewService(tc)

	accounts := []sms.Account{{FirstName: "John", LastName: "Doe", Phone: "+15551234567"}}

	_, err := svc.Send(accounts, "Hello", "Test", "", nil)

	if err == nil {
		t.Fatal("expected error from API")
	}
}

func TestSendProgressTracking(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"id": "msg-123", "status": "sent"})
	}))
	defer server.Close()

	tc := &testutil.TestClient{BaseURL: server.URL, FilesURL: server.URL}
	svc := sms.NewService(tc)

	progressUpdates := []string{}
	options := &sms.Options{
		OnProgress: func(status string) {
			progressUpdates = append(progressUpdates, status)
		},
	}

	accounts := []sms.Account{{FirstName: "John", LastName: "Doe", Phone: "+15551234567"}}

	_, err := svc.Send(accounts, "Hello", "Test", "", options)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(progressUpdates) != 3 {
		t.Errorf("expected 3 progress updates, got %d", len(progressUpdates))
	}

	expected := []string{"Preparing to send SMS", "Sending SMS", "SMS sent successfully"}
	for i, exp := range expected {
		if progressUpdates[i] != exp {
			t.Errorf("update[%d]: expected '%s', got '%s'", i, exp, progressUpdates[i])
		}
	}
}

func TestResponseGetID(t *testing.T) {
	r1 := sms.Response{ID: "msg-123"}
	if r1.GetID() != "msg-123" {
		t.Errorf("expected 'msg-123', got '%s'", r1.GetID())
	}

	r2 := sms.Response{ID: float64(146403)}
	if r2.GetID() != "146403" {
		t.Errorf("expected '146403', got '%s'", r2.GetID())
	}
}

func TestSendWithCustomFieldsAndCustomData(t *testing.T) {
	var capturedBody []byte

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedBody = make([]byte, r.ContentLength)
		r.Body.Read(capturedBody)
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":         "msg-cf-123",
			"status":     "sent",
			"message":    "SMS sent successfully",
			"responseId": "resp-abc-456",
		})
	}))
	defer server.Close()

	tc := &testutil.TestClient{BaseURL: server.URL, FilesURL: server.URL}
	svc := sms.NewService(tc)

	accounts := []sms.Account{
		{
			FirstName:   "John",
			LastName:    "Doe",
			Phone:       "+15551234567",
			Data:        map[string]string{"city": "Miami", "country": "USA", "plan": "premium"},
			MessageData: `{"source":"go-sdk-test"}`,
		},
	}

	res, err := svc.Send(accounts, "Hello ${firstName} from ${city}!", "Test customFields", "", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify request body contains "data" key (wire format)
	bodyStr := string(capturedBody)
	if !contains(bodyStr, `"data"`) {
		t.Errorf("expected request body to contain \"data\", got: %s", bodyStr)
	}
	if !contains(bodyStr, `"Miami"`) {
		t.Errorf("expected request body to contain city value, got: %s", bodyStr)
	}
	if !contains(bodyStr, `"messageData"`) {
		t.Errorf("expected request body to contain \"messageData\", got: %s", bodyStr)
	}

	// Verify response fields
	if res.Message != "SMS sent successfully" {
		t.Errorf("expected message 'SMS sent successfully', got '%s'", res.Message)
	}
	if res.ResponseID != "resp-abc-456" {
		t.Errorf("expected responseId 'resp-abc-456', got '%s'", res.ResponseID)
	}
}

func TestSMSResponseMessageAndResponseID(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":         "msg-123",
			"status":     "sent",
			"message":    "SMS sent successfully",
			"responseId": "resp-id-xyz",
		})
	}))
	defer server.Close()

	tc := &testutil.TestClient{BaseURL: server.URL, FilesURL: server.URL}
	svc := sms.NewService(tc)

	accounts := []sms.Account{{FirstName: "John", LastName: "Doe", Phone: "+15551234567"}}
	res, err := svc.Send(accounts, "Hello!", "Test", "", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if res.Message != "SMS sent successfully" {
		t.Errorf("expected message 'SMS sent successfully', got '%s'", res.Message)
	}
	if res.ResponseID != "resp-id-xyz" {
		t.Errorf("expected responseId 'resp-id-xyz', got '%s'", res.ResponseID)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		func() bool {
			for i := 0; i <= len(s)-len(substr); i++ {
				if s[i:i+len(substr)] == substr {
					return true
				}
			}
			return false
		}())
}
