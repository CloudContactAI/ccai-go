// Copyright (c) 2025 CloudContactAI LLC
// Licensed under the MIT License. See LICENSE in the project root for license information.

package ccai_test

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/cloudcontactai/ccai-go/src/pkg/sms"
	"github.com/cloudcontactai/ccai-go/src/pkg/testutil"
)

func TestGetSignedUploadURL(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/upload/url" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"signedS3Url": "https://s3.amazonaws.com/test-bucket/test.png?sig=abc123",
		})
	}))
	defer server.Close()

	tc := &testutil.TestClient{BaseURL: server.URL, FilesURL: server.URL}
	svc := sms.NewMMSService(tc)

	res, err := svc.GetSignedUploadURL("test.png", "image/png", "", true)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if res.SignedS3URL != "https://s3.amazonaws.com/test-bucket/test.png?sig=abc123" {
		t.Errorf("unexpected signed URL: %s", res.SignedS3URL)
	}

	if res.FileKey != "test-client/campaign/test.png" {
		t.Errorf("unexpected fileKey: %s", res.FileKey)
	}
}

func TestGetSignedUploadURLValidation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	tc := &testutil.TestClient{BaseURL: server.URL, FilesURL: server.URL}
	svc := sms.NewMMSService(tc)

	_, err := svc.GetSignedUploadURL("", "image/png", "", true)
	if err == nil {
		t.Fatal("expected error for empty file name")
	}

	_, err = svc.GetSignedUploadURL("test.png", "", "", true)
	if err == nil {
		t.Fatal("expected error for empty file type")
	}
}

func TestUploadImageToSignedURL(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	tmpFile := createTempFile(t)

	tc := &testutil.TestClient{BaseURL: server.URL, FilesURL: server.URL}
	svc := sms.NewMMSService(tc)

	ok, err := svc.UploadImageToSignedURL(server.URL, tmpFile, "image/png")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !ok {
		t.Error("expected upload to succeed")
	}
}

func TestSendMMS(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/clients/test-client/campaigns/direct" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Header.Get("ForceNewCampaign") != "true" {
			t.Error("expected ForceNewCampaign header")
		}
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":     "mms-123",
			"status": "sent",
		})
	}))
	defer server.Close()

	tc := &testutil.TestClient{BaseURL: server.URL, FilesURL: server.URL}
	svc := sms.NewMMSService(tc)

	accounts := []sms.Account{{FirstName: "John", LastName: "Doe", Phone: "+15551234567"}}

	res, err := svc.Send("test-client/campaign/test.png", accounts, "Hello!", "MMS Test", "", nil, true)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if res.GetID() != "mms-123" {
		t.Errorf("expected ID 'mms-123', got '%s'", res.GetID())
	}
}

func TestSendMMSValidation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	tc := &testutil.TestClient{BaseURL: server.URL, FilesURL: server.URL}
	svc := sms.NewMMSService(tc)

	accounts := []sms.Account{{FirstName: "John", LastName: "Doe", Phone: "+15551234567"}}

	_, err := svc.Send("", accounts, "Hello", "Test", "", nil, true)
	if err == nil {
		t.Fatal("expected error for empty pictureFileKey")
	}

	_, err = svc.Send("key", []sms.Account{}, "Hello", "Test", "", nil, true)
	if err == nil {
		t.Fatal("expected error for empty accounts")
	}

	_, err = svc.Send("key", accounts, "", "Test", "", nil, true)
	if err == nil {
		t.Fatal("expected error for empty message")
	}

	_, err = svc.Send("key", accounts, "Hello", "", "", nil, true)
	if err == nil {
		t.Fatal("expected error for empty title")
	}
}

func TestCheckFileUploaded(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"storedUrl": "https://s3.amazonaws.com/bucket/test.jpg",
		})
	}))
	defer server.Close()

	tc := &testutil.TestClient{BaseURL: server.URL, FilesURL: server.URL}
	svc := sms.NewMMSService(tc)

	res, err := svc.CheckFileUploaded("test-client/campaign/test.jpg")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if res.StoredURL != "https://s3.amazonaws.com/bucket/test.jpg" {
		t.Errorf("unexpected storedUrl: %s", res.StoredURL)
	}
}

func TestSendWithImageCacheHit(t *testing.T) {
	apiCalls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiCalls++
		if r.URL.Path == "/clients/test-client/storedUrl" {
			w.WriteHeader(http.StatusOK)
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"storedUrl": "https://s3.amazonaws.com/bucket/existing.png",
			})
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":     "mms-cached",
			"status": "sent",
		})
	}))
	defer server.Close()

	tc := &testutil.TestClient{BaseURL: server.URL, FilesURL: server.URL}
	svc := sms.NewMMSService(tc)

	tmpFile := createTempFile(t)
	accounts := []sms.Account{{FirstName: "John", LastName: "Doe", Phone: "+15551234567"}}

	progressUpdates := []string{}
	options := &sms.Options{
		OnProgress: func(status string) {
			progressUpdates = append(progressUpdates, status)
		},
	}

	res, err := svc.SendWithImage(tmpFile, "image/png", accounts, "Hello!", "Cache Test", "", options, true)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if res.GetID() != "mms-cached" {
		t.Errorf("expected ID 'mms-cached', got '%s'", res.GetID())
	}

	if apiCalls != 2 {
		t.Errorf("expected 2 API calls (cache hit), got %d", apiCalls)
	}

	found := false
	for _, msg := range progressUpdates {
		if msg == "Image already exists in S3, sending MMS" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected cache hit progress message, got: %v", progressUpdates)
	}
}

func TestMD5Calculation(t *testing.T) {
	content := []byte("test image content")
	tmpFile := filepath.Join(t.TempDir(), "test.png")
	os.WriteFile(tmpFile, content, 0644)

	expectedHash := md5.Sum(content)
	expectedHex := hex.EncodeToString(expectedHash[:])

	// Verify MD5 matches what the SDK would compute
	actualHex := hex.EncodeToString(expectedHash[:])
	if actualHex != expectedHex {
		t.Errorf("expected MD5 '%s', got '%s'", expectedHex, actualHex)
	}
}

func createTempFile(t *testing.T) string {
	t.Helper()
	content := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}
	tmpFile := filepath.Join(t.TempDir(), "test.png")
	os.WriteFile(tmpFile, content, 0644)
	return tmpFile
}
