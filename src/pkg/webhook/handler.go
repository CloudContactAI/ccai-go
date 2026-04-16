// Copyright (c) 2025 CloudContactAI LLC
// Licensed under the MIT License. See LICENSE in the project root for license information.

// Package webhook provides HTTP handler utilities for CloudContactAI webhooks.
package webhook

import (
	"fmt"
	"io"
	"log"
	"net/http"
)

// HandlerOptions represents configuration options for the webhook handler
type HandlerOptions struct {
	// ClientID is the CCAI client ID (required for signature verification)
	ClientID string

	// Secret used to verify webhook signatures
	Secret string

	// Handler for all webhook events
	OnEvent func(event *WebhookEvent) error

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

		// Parse the webhook event first to get eventHash
		event, err := ParseEvent(body)
		if err != nil {
			http.Error(w, "Invalid JSON or missing required fields", http.StatusBadRequest)
			return
		}

		// Verify signature if secret is provided
		if options.Secret != "" {
			signature := r.Header.Get("X-CCAI-Signature")
			if signature == "" {
				http.Error(w, "Missing signature header", http.StatusBadRequest)
				return
			}

			if options.ClientID == "" {
				http.Error(w, "ClientID is required for signature verification", http.StatusInternalServerError)
				return
			}

			client := &Client{}
			if !client.VerifySignature(signature, options.ClientID, event.EventHash, options.Secret) {
				http.Error(w, "Invalid signature", http.StatusUnauthorized)
				return
			}
		}

		// Log event if enabled
		if options.LogEvents {
			log.Printf("✅ Webhook event verified: %s", event.EventType)
		}

		// Handle the event
		if options.OnEvent != nil {
			if err := options.OnEvent(event); err != nil {
				log.Printf("Error handling webhook event: %v", err)
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}
		}

		// Send success response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"success": true}`)
	}
}
