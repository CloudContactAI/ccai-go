// Copyright (c) 2025 CloudContactAI LLC
// Licensed under the MIT License. See LICENSE in the project root for license information.

package test

import (
	"testing"

	"github.com/cloudcontactai/ccai-go/pkg/sms"
)

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
}

func TestMMSSendSingle(t *testing.T) {
	mockClient := NewMockClient()
	mmsService := sms.NewMMSService(mockClient)

	pictureFileKey := "test-client-id/campaign/test-image.jpg"
	firstName := "Jane"
	lastName := "Smith"
	phone := "+14156961732"
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