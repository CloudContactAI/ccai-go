// Copyright (c) 2025 CloudContactAI LLC
// Licensed under the MIT License. See LICENSE in the project root for license information.

package ccai_test

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cloudcontactai/ccai-go/src/pkg/testutil"
	"github.com/cloudcontactai/ccai-go/src/pkg/webhook"
)

func TestWebhookRegister(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/client/test-client/integration" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]map[string]interface{}{
			{"id": 100, "url": "https://example.com/webhook", "method": "POST", "secretKey": "sk_live_auto_generated_secret"},
		})
	}))
	defer server.Close()

	tc := &testutil.TestClient{BaseURL: server.URL, FilesURL: server.URL}
	svc := webhook.NewService(tc)

	res, err := svc.Register(webhook.WebhookConfig{
		URL: "https://example.com/webhook",
		// Secret is nil - server should generate it
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if res.URL != "https://example.com/webhook" {
		t.Errorf("expected URL 'https://example.com/webhook', got '%s'", res.URL)
	}

	if res.SecretKey != "sk_live_auto_generated_secret" {
		t.Errorf("expected auto-generated secret, got '%s'", res.SecretKey)
	}
}

func TestWebhookRegisterWithCustomSecret(t *testing.T) {
	customSecret := "my-custom-secret"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/client/test-client/integration" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]map[string]interface{}{
			{"id": 101, "url": "https://example.com/custom-webhook", "method": "POST", "secretKey": customSecret},
		})
	}))
	defer server.Close()

	tc := &testutil.TestClient{BaseURL: server.URL, FilesURL: server.URL}
	svc := webhook.NewService(tc)

	res, err := svc.Register(webhook.WebhookConfig{
		URL:    "https://example.com/custom-webhook",
		Secret: &customSecret,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if res.SecretKey != customSecret {
		t.Errorf("expected custom secret '%s', got '%s'", customSecret, res.SecretKey)
	}
}

func TestWebhookUpdate(t *testing.T) {
	newSecret := "new-secret"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]map[string]interface{}{
			{"id": 100, "url": "https://example.com/updated", "method": "POST", "secretKey": newSecret},
		})
	}))
	defer server.Close()

	tc := &testutil.TestClient{BaseURL: server.URL, FilesURL: server.URL}
	svc := webhook.NewService(tc)

	res, err := svc.Update("100", webhook.WebhookConfig{
		URL:    "https://example.com/updated",
		Secret: &newSecret,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if res.URL != "https://example.com/updated" {
		t.Errorf("expected updated URL, got '%s'", res.URL)
	}

	if res.SecretKey != newSecret {
		t.Errorf("expected secret '%s', got '%s'", newSecret, res.SecretKey)
	}
}

func TestWebhookList(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("expected GET, got %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]map[string]interface{}{
			{"id": 1, "url": "https://example.com/hook1"},
			{"id": 2, "url": "https://example.com/hook2"},
		})
	}))
	defer server.Close()

	tc := &testutil.TestClient{BaseURL: server.URL, FilesURL: server.URL}
	svc := webhook.NewService(tc)

	webhooks, err := svc.List()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(webhooks) != 2 {
		t.Errorf("expected 2 webhooks, got %d", len(webhooks))
	}
}

func TestWebhookDelete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"message": "Webhook deleted",
		})
	}))
	defer server.Close()

	tc := &testutil.TestClient{BaseURL: server.URL, FilesURL: server.URL}
	svc := webhook.NewService(tc)

	res, err := svc.Delete("100")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !res.Success {
		t.Error("expected success to be true")
	}
}

func TestWebhookVerifySignatureValid(t *testing.T) {
	clientID := "client-123"
	eventHash := "event-hash-abc123"
	secret := "test-secret"
	expected := computeBase64HMAC(secret, clientID, eventHash)

	tc := &testutil.TestClient{}
	svc := webhook.NewService(tc)

	if !svc.VerifySignature(expected, clientID, eventHash, secret) {
		t.Error("expected valid signature to pass")
	}
}

func TestWebhookVerifySignatureInvalid(t *testing.T) {
	tc := &testutil.TestClient{}
	svc := webhook.NewService(tc)

	if svc.VerifySignature("bad-signature", "client-123", "event-hash", "secret") {
		t.Error("expected invalid signature to fail")
	}
}

func TestWebhookVerifySignatureEmptyParams(t *testing.T) {
	tc := &testutil.TestClient{}
	svc := webhook.NewService(tc)

	if svc.VerifySignature("", "client-123", "event-hash", "secret") {
		t.Error("expected empty signature to fail")
	}

	if svc.VerifySignature("sig", "", "event-hash", "secret") {
		t.Error("expected empty clientID to fail")
	}

	if svc.VerifySignature("sig", "client-123", "", "secret") {
		t.Error("expected empty eventHash to fail")
	}

	if svc.VerifySignature("sig", "client-123", "event-hash", "") {
		t.Error("expected empty secret to fail")
	}
}

func TestWebhookParseEvent(t *testing.T) {
	payload := []byte(`{
		"eventType":"message.sent",
		"eventHash":"hash-abc123",
		"data":{
			"To":"+15552222",
			"Message":"Hello"
		}
	}`)

	event, err := webhook.ParseEvent(payload)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if event.EventType != "message.sent" {
		t.Errorf("expected eventType 'message.sent', got '%s'", event.EventType)
	}

	if to, ok := event.Data["To"].(string); !ok || to != "+15552222" {
		t.Errorf("expected To '+15552222', got '%v'", event.Data["To"])
	}

	if message, ok := event.Data["Message"].(string); !ok || message != "Hello" {
		t.Errorf("expected Message 'Hello', got '%v'", event.Data["Message"])
	}
}

func TestWebhookParseEventInvalidJSON(t *testing.T) {
	_, err := webhook.ParseEvent([]byte("not valid json"))

	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestWebhookEventTypes(t *testing.T) {
	if webhook.MessageSentEvent != webhook.WebhookEventType("message.sent") {
		t.Errorf("unexpected MessageSentEvent: %s", webhook.MessageSentEvent)
	}

	if webhook.MessageReceivedEvent != webhook.WebhookEventType("message.received") {
		t.Errorf("unexpected MessageReceivedEvent: %s", webhook.MessageReceivedEvent)
	}
}

func computeHMAC(secret, body string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(body))
	return hex.EncodeToString(mac.Sum(nil))
}

func computeBase64HMAC(secret, clientID, eventHash string) string {
	// Compute: HMAC-SHA256(secretKey, clientId:eventHash) in Base64
	data := fmt.Sprintf("%s:%s", clientID, eventHash)
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(data))
	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}
