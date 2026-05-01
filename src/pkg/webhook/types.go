// Copyright (c) 2025 CloudContactAI LLC
// Licensed under the MIT License. See LICENSE in the project root for license information.

// Package webhook provides types and utilities for handling CloudContactAI webhooks.
package webhook

import (
	"encoding/json"
	"fmt"
)

// WebhookEventType represents the types of webhook events
type WebhookEventType string

const (
	// MessageSentEvent represents an outbound message event
	MessageSentEvent WebhookEventType = "message.sent"
	// MessageIncomingEvent represents an inbound message event
	MessageIncomingEvent WebhookEventType = "message.incoming"
	// MessageReceivedEvent represents an inbound message event (legacy name)
	MessageReceivedEvent WebhookEventType = "message.received"
	// MessageExcludedEvent represents a message excluded during campaign
	MessageExcludedEvent WebhookEventType = "message.excluded"
	// MessageErrorCarrierEvent represents a carrier-level delivery failure
	MessageErrorCarrierEvent WebhookEventType = "message.error.carrier"
	// MessageErrorCloudcontactEvent represents a CloudContact system error
	MessageErrorCloudcontactEvent WebhookEventType = "message.error.cloudcontact"
)

// WebhookEvent represents the webhook payload sent by the server
// This is the unified structure for all event types
type WebhookEvent struct {
	EventType string                 `json:"eventType"` // Type of the event (e.g., "message.sent")
	Data      map[string]interface{} `json:"data"`      // Event-specific data
	EventHash string                 `json:"eventHash"` // Hash computed by the backend for signature verification
}

// ParseEvent parses a raw webhook JSON payload into a WebhookEvent
func ParseEvent(payload []byte) (*WebhookEvent, error) {
	var event WebhookEvent
	if err := json.Unmarshal(payload, &event); err != nil {
		return nil, fmt.Errorf("failed to parse webhook event: %w", err)
	}
	return &event, nil
}

// WebhookConfig represents the configuration for webhook integration
type WebhookConfig struct {
	URL    string             `json:"url"`
	Events []WebhookEventType `json:"events"`
	Secret *string            `json:"secretKey,omitempty"` // Optional secret - if nil, server generates one automatically
}

// WebhookResponse represents the response when registering/updating webhooks
type WebhookResponse struct {
	ID              interface{}        `json:"id"`
	URL             string             `json:"url"`
	Method          string             `json:"method"`
	IntegrationType string             `json:"integrationType"`
	SecretKey       string             `json:"secretKey,omitempty"`
}

// WebhookDeleteResponse represents the response when deleting a webhook
type WebhookDeleteResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}
