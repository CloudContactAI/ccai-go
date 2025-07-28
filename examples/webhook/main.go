// Copyright (c) 2025 CloudContactAI LLC
// Licensed under the MIT License. See LICENSE in the project root for license information.

// Package main provides an example of using the CCAI Go webhook functionality.
package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/cloudcontactai/ccai-go/pkg/ccai"
	"github.com/cloudcontactai/ccai-go/pkg/webhook"
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

	// Example 1: Register a webhook
	fmt.Println("=== Registering Webhook ===")
	webhookConfig := webhook.WebhookConfig{
		URL: "https://your-domain.com/webhook/ccai",
		Events: []webhook.WebhookEventType{
			webhook.MessageSentEvent,
			webhook.MessageReceivedEvent,
		},
		Secret: "your-webhook-secret-key",
	}

	registeredWebhook, err := client.Webhook.Register(webhookConfig)
	if err != nil {
		log.Printf("Failed to register webhook: %v", err)
	} else {
		fmt.Printf("Webhook registered successfully!\n")
		fmt.Printf("Webhook ID: %s\n", registeredWebhook.ID)
		fmt.Printf("Webhook URL: %s\n", registeredWebhook.URL)
		fmt.Printf("Events: %v\n", registeredWebhook.Events)
	}

	// Example 2: List all webhooks
	fmt.Println("\n=== Listing Webhooks ===")
	webhooks, err := client.Webhook.List()
	if err != nil {
		log.Printf("Failed to list webhooks: %v", err)
	} else {
		fmt.Printf("Found %d webhook(s):\n", len(webhooks))
		for i, wh := range webhooks {
			fmt.Printf("  %d. ID: %s, URL: %s, Events: %v\n", i+1, wh.ID, wh.URL, wh.Events)
		}
	}

	// Example 3: Create webhook HTTP handler
	fmt.Println("\n=== Setting up Webhook HTTP Handler ===")

	// Create webhook handler with event processors
	webhookHandler := webhook.CreateHandler(webhook.HandlerOptions{
		Secret:    "your-webhook-secret-key",
		LogEvents: true,

		// Handler for outbound messages (messages sent from campaigns)
		OnMessageSent: func(event webhook.MessageSentEventData) error {
			fmt.Printf("✅ Message sent successfully:\n")
			fmt.Printf("   Campaign ID: %d\n", event.GetCampaign().ID)
			fmt.Printf("   Recipient: %s\n", event.GetTo())
			fmt.Printf("   Message: %s\n", event.GetMessage())

			// Example: Update your database with delivery status
			// err := updateMessageStatus(event.Campaign.ID, "sent")
			// return err

			return nil
		},

		// Handler for inbound messages (replies from recipients)
		OnMessageReceived: func(event webhook.MessageReceivedEventData) error {
			fmt.Printf("📥 Received reply:\n")
			fmt.Printf("   Sender: %s\n", event.GetFrom())
			fmt.Printf("   Message: %s\n", event.GetMessage())
			fmt.Printf("   Campaign ID: %d\n", event.GetCampaign().ID)

			// Example: Auto-respond to certain keywords
			message := event.GetMessage()
			if message != "" && contains(message, "stop") {
				fmt.Printf("🛑 Processing opt-out request from: %s\n", event.GetFrom())
				// processOptOut(event.GetFrom())
			}

			// Example: Store the reply in your database
			// err := storeInboundMessage(event)
			// return err

			return nil
		},
	})

	// Set up HTTP server with webhook endpoint
	http.Handle("/webhook/ccai", webhookHandler)

	fmt.Println("🚀 Webhook server starting on http://localhost:8080")
	fmt.Println("📡 Webhook endpoint: http://localhost:8080/webhook/ccai")
	fmt.Println("💡 Tip: Use ngrok to expose this endpoint for testing")
	fmt.Println("   Example: ngrok http 8080")
	fmt.Println("\n⏳ Server is running... (Ctrl+C to stop)")

	// Start the HTTP server
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Failed to start webhook server: %v", err)
	}
}

// Helper function to check if a string contains a substring (case-insensitive)
func contains(text, substring string) bool {
	return len(text) >= len(substring) &&
		text[:len(substring)] == substring ||
		(len(text) > len(substring) &&
			findSubstring(text, substring))
}

func findSubstring(text, substring string) bool {
	for i := 0; i <= len(text)-len(substring); i++ {
		if text[i:i+len(substring)] == substring {
			return true
		}
	}
	return false
}
