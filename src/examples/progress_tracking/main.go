// Copyright (c) 2025 CloudContactAI LLC
// Licensed under the MIT License. See LICENSE in the project root for license information.

// Package main provides an example of using progress tracking with the CCAI Go client.
package main

import (
	"fmt"
	"log"
	"time"

	"github.com/cloudcontactai/ccai-go/src/pkg/ccai"
	"github.com/cloudcontactai/ccai-go/src/pkg/sms"
)

func main() {
	// Initialize the client
	client, err := ccai.NewClient(ccai.Config{
		ClientID: "YOUR_CLIENT_ID",
		APIKey:   "YOUR_API_KEY",
	})
	if err != nil {
		log.Fatalf("Failed to create CCAI client: %v", err)
	}

	// Create options with progress tracking
	options := &sms.Options{
		Timeout: 60,
		Retries: 3,
		OnProgress: func(status string) {
			fmt.Printf("%s - %s\n", time.Now().Format("2006-01-02 15:04:05"), status)
		},
	}

	// Define recipients
	accounts := []sms.Account{
		{
			FirstName: "John",
			LastName:  "Doe",
			Phone:     "+14156566694",
		},
		{
			FirstName: "Jane",
			LastName:  "Smith",
			Phone:     "+15559876543",
		},
	}

	// Send SMS with progress tracking
	response, err := client.SMS.Send(
		accounts,
		"Hello ${firstName}, this is a test message with progress tracking!",
		"Progress Tracking Test",
		options,
	)
	if err != nil {
		log.Fatalf("Failed to send SMS: %v", err)
	}

	fmt.Printf("Campaign sent with ID: %s\n", response.CampaignID)
}
