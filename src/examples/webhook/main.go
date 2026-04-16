// Copyright (c) 2025 CloudContactAI LLC
// Licensed under the MIT License. See LICENSE in the project root for license information.

// Package main provides an example of using the CCAI Go webhook functionality.
package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/cloudcontactai/ccai-go/src/pkg/ccai"
	"github.com/cloudcontactai/ccai-go/src/pkg/webhook"
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
	customSecret := "your-webhook-secret-key"
	webhookConfig := webhook.WebhookConfig{
		URL: "https://your-domain.com/webhook/ccai",
		Events: []webhook.WebhookEventType{
			webhook.MessageSentEvent,
			webhook.MessageReceivedEvent,
		},
		Secret: &customSecret,
	}

	registeredWebhook, err := client.Webhook.Register(webhookConfig)
	if err != nil {
		log.Printf("Failed to register webhook: %v", err)
	} else {
		fmt.Printf("Webhook registered successfully!\n")
		fmt.Printf("Webhook ID: %s\n", registeredWebhook.ID)
		fmt.Printf("Webhook URL: %s\n", registeredWebhook.URL)
		fmt.Printf("Secret Key: %s\n", registeredWebhook.SecretKey)
	}

	// Example 2: List all webhooks
	fmt.Println("\n=== Listing Webhooks ===")
	webhooks, err := client.Webhook.List()
	if err != nil {
		log.Printf("Failed to list webhooks: %v", err)
	} else {
		fmt.Printf("Found %d webhook(s):\n", len(webhooks))
		for i, wh := range webhooks {
			fmt.Printf("  %d. ID: %s, URL: %s\n", i+1, wh.ID, wh.URL)
		}
	}

	// Example 3: Create webhook HTTP handler
	fmt.Println("\n=== Setting up Webhook HTTP Handler ===")

	// Create webhook handler with event processor
	webhookHandler := webhook.CreateHandler(webhook.HandlerOptions{
		ClientID:  "YOUR_CLIENT_ID",
		Secret:    "your-webhook-secret-key",
		LogEvents: true,

		// Handler for all webhook events
		OnEvent: func(event *webhook.WebhookEvent) error {
			fmt.Printf("📨 Webhook Event Type: %s\n", event.EventType)

			switch event.EventType {
			case "message.sent":
				fmt.Printf("✅ Message sent successfully\n")
				if to, ok := event.Data["To"].(string); ok {
					fmt.Printf("   Recipient: %s\n", to)
				}
				if price, ok := event.Data["TotalPrice"].(float64); ok {
					fmt.Printf("   Cost: $%v\n", price)
				}
				if segments, ok := event.Data["Segments"].(float64); ok {
					fmt.Printf("   Segments: %v\n", int(segments))
				}

			case "message.incoming":
				fmt.Printf("📥 Received reply\n")
				if from, ok := event.Data["From"].(string); ok {
					fmt.Printf("   Sender: %s\n", from)
				}
				if msg, ok := event.Data["Message"].(string); ok {
					fmt.Printf("   Message: %s\n", msg)
				}

			case "message.excluded":
				fmt.Printf("⚠️ Message excluded\n")
				if reason, ok := event.Data["ExcludedReason"].(string); ok {
					fmt.Printf("   Reason: %s\n", reason)
				}

			case "message.error.carrier":
				fmt.Printf("❌ Carrier error\n")
				if code, ok := event.Data["ErrorCode"].(string); ok {
					fmt.Printf("   Code: %s\n", code)
				}
				if msg, ok := event.Data["ErrorMessage"].(string); ok {
					fmt.Printf("   Message: %s\n", msg)
				}

			case "message.error.cloudcontact":
				fmt.Printf("🚨 System error\n")
				if code, ok := event.Data["ErrorCode"].(string); ok {
					fmt.Printf("   Code: %s\n", code)
				}
				if msg, ok := event.Data["ErrorMessage"].(string); ok {
					fmt.Printf("   Message: %s\n", msg)
				}
			}

			// Handle custom data if present
			if customData, ok := event.Data["CustomData"].(string); ok && customData != "" {
				fmt.Printf("📌 Custom Data: %s\n", customData)
			}

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
