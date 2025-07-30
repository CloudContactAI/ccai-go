// Copyright (c) 2025 CloudContactAI LLC
// Licensed under the MIT License. See LICENSE in the project root for license information.

// Package webhook provides HTTP handler utilities for CloudContactAI webhooks.
package webhook

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

// HandlerOptions represents configuration options for the webhook handler
type HandlerOptions struct {
	// Secret used to verify webhook signatures
	Secret string

	// Handler for Message Sent events
	OnMessageSent func(event MessageSentEventData) error

	// Handler for Message Received events
	OnMessageReceived func(event MessageReceivedEventData) error

	// Whether to log events to console
	LogEvents bool
}

// CreateHandler creates an HTTP handler for CloudContactAI webhooks
func CreateHandler(options HandlerOptions) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Only allow POST requests
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Read the request body
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Failed to read request body", http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		// Verify signature if secret is provided
		if options.Secret != "" {
			signature := r.Header.Get("X-CCAI-Signature")
			if signature == "" {
				http.Error(w, "Missing signature header", http.StatusBadRequest)
				return
			}

			client := &Client{}
			if !client.VerifySignature(signature, string(body), options.Secret) {
				http.Error(w, "Invalid signature", http.StatusUnauthorized)
				return
			}
		}

		// Parse the webhook event
		var rawEvent map[string]interface{}
		if err := json.Unmarshal(body, &rawEvent); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		// Get event type
		eventType, ok := rawEvent["type"].(string)
		if !ok {
			http.Error(w, "Missing or invalid event type", http.StatusBadRequest)
			return
		}

		// Log event if enabled
		if options.LogEvents {
			log.Printf("Webhook event: %s", eventType)
		}

		// Handle the event based on type
		switch WebhookEventType(eventType) {
		case MessageSentEvent:
			if options.OnMessageSent != nil {
				var event MessageSentEventData
				if err := json.Unmarshal(body, &event); err != nil {
					http.Error(w, "Failed to parse message sent event", http.StatusBadRequest)
					return
				}

				if err := options.OnMessageSent(event); err != nil {
					log.Printf("Error handling message sent event: %v", err)
					http.Error(w, "Internal server error", http.StatusInternalServerError)
					return
				}
			}

		case MessageReceivedEvent:
			if options.OnMessageReceived != nil {
				var event MessageReceivedEventData
				if err := json.Unmarshal(body, &event); err != nil {
					http.Error(w, "Failed to parse message received event", http.StatusBadRequest)
					return
				}

				if err := options.OnMessageReceived(event); err != nil {
					log.Printf("Error handling message received event: %v", err)
					http.Error(w, "Internal server error", http.StatusInternalServerError)
					return
				}
			}

		default:
			if options.LogEvents {
				log.Printf("Unknown event type: %s", eventType)
			}
		}

		// Send success response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"success": true}`)
	}
}
