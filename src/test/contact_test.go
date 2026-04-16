// Copyright (c) 2025 CloudContactAI LLC
// Licensed under the MIT License. See LICENSE in the project root for license information.

package ccai_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cloudcontactai/ccai-go/src/pkg/contact"
	"github.com/cloudcontactai/ccai-go/src/pkg/testutil"
)

func TestContactSetDoNotTextOptOut(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/account/do-not-text" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != "PUT" {
			t.Errorf("expected PUT, got %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"contactId": "12345",
			"phone":     "+15551234567",
			"doNotText": true,
		})
	}))
	defer server.Close()

	tc := &testutil.TestClient{BaseURL: server.URL, FilesURL: server.URL}
	svc := contact.NewService(tc)

	res, err := svc.SetDoNotText(true, "", "+15551234567")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !res.DoNotText {
		t.Error("expected DoNotText to be true")
	}

	if res.Phone != "+15551234567" {
		t.Errorf("expected phone '+15551234567', got '%s'", res.Phone)
	}
}

func TestContactSetDoNotTextOptIn(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"contactId": "12345",
			"phone":     "+15551234567",
			"doNotText": false,
		})
	}))
	defer server.Close()

	tc := &testutil.TestClient{BaseURL: server.URL, FilesURL: server.URL}
	svc := contact.NewService(tc)

	res, err := svc.SetDoNotText(false, "", "+15551234567")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if res.DoNotText {
		t.Error("expected DoNotText to be false")
	}
}

func TestContactSetDoNotTextWithContactID(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"contactId": "98765",
			"phone":     "",
			"doNotText": true,
		})
	}))
	defer server.Close()

	tc := &testutil.TestClient{BaseURL: server.URL, FilesURL: server.URL}
	svc := contact.NewService(tc)

	res, err := svc.SetDoNotText(true, "98765", "")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if res.ContactID != "98765" {
		t.Errorf("expected contactId '98765', got '%s'", res.ContactID)
	}
}

func TestContactSetDoNotTextAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"error": "Bad request"})
	}))
	defer server.Close()

	tc := &testutil.TestClient{BaseURL: server.URL, FilesURL: server.URL}
	svc := contact.NewService(tc)

	_, err := svc.SetDoNotText(true, "", "+15551234567")

	if err == nil {
		t.Fatal("expected error from API")
	}
}

func TestContactSetDoNotTextPayloadContainsClientID(t *testing.T) {
	var receivedBody contact.SetDoNotTextRequest
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&receivedBody)
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"contactId": "1",
			"phone":     "+15551234567",
			"doNotText": true,
		})
	}))
	defer server.Close()

	tc := &testutil.TestClient{BaseURL: server.URL, FilesURL: server.URL}
	svc := contact.NewService(tc)

	_, _ = svc.SetDoNotText(true, "", "+15551234567")

	if receivedBody.ClientID != "test-client" {
		t.Errorf("expected clientId 'test-client', got '%s'", receivedBody.ClientID)
	}

	if !receivedBody.DoNotText {
		t.Error("expected doNotText to be true in payload")
	}
}
