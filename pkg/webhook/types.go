// Copyright (c) 2025 CloudContactAI LLC
// Licensed under the MIT License. See LICENSE in the project root for license information.

// Package webhook provides types and utilities for handling CloudContactAI webhooks.
package webhook

import "time"

// WebhookEventType represents the types of webhook events
type WebhookEventType string

const (
	// MessageSentEvent represents an outbound message event
	MessageSentEvent WebhookEventType = "message.sent"
	// MessageReceivedEvent represents an inbound message event
	MessageReceivedEvent WebhookEventType = "message.received"
)

// WebhookCampaign contains campaign information included in webhook events
type WebhookCampaign struct {
	ID          int       `json:"id"`
	Title       string    `json:"title"`
	Message     string    `json:"message"`
	SenderPhone string    `json:"senderPhone"`
	CreatedAt   time.Time `json:"createdAt"`
	RunAt       time.Time `json:"runAt"`
}

// WebhookEventBase is the base interface for all webhook events
type WebhookEventBase struct {
	Campaign WebhookCampaign `json:"campaign"`
	From     string          `json:"from"`
	To       string          `json:"to"`
	Message  string          `json:"message"`
}

// MessageSentEventData represents a message sent (outbound) webhook event
type MessageSentEventData struct {
	WebhookEventBase
	Type WebhookEventType `json:"type"`
}

// MessageReceivedEventData represents a message received (inbound) webhook event
type MessageReceivedEventData struct {
	WebhookEventBase
	Type WebhookEventType `json:"type"`
}

// WebhookEvent is a union type for all webhook events
type WebhookEvent interface {
	GetType() WebhookEventType
	GetCampaign() WebhookCampaign
	GetFrom() string
	GetTo() string
	GetMessage() string
}

// GetType returns the event type for MessageSentEventData
func (e MessageSentEventData) GetType() WebhookEventType {
	return e.Type
}

// GetCampaign returns the campaign for MessageSentEventData
func (e MessageSentEventData) GetCampaign() WebhookCampaign {
	return e.Campaign
}

// GetFrom returns the from field for MessageSentEventData
func (e MessageSentEventData) GetFrom() string {
	return e.From
}

// GetTo returns the to field for MessageSentEventData
func (e MessageSentEventData) GetTo() string {
	return e.To
}

// GetMessage returns the message for MessageSentEventData
func (e MessageSentEventData) GetMessage() string {
	return e.Message
}

// GetType returns the event type for MessageReceivedEventData
func (e MessageReceivedEventData) GetType() WebhookEventType {
	return e.Type
}

// GetCampaign returns the campaign for MessageReceivedEventData
func (e MessageReceivedEventData) GetCampaign() WebhookCampaign {
	return e.Campaign
}

// GetFrom returns the from field for MessageReceivedEventData
func (e MessageReceivedEventData) GetFrom() string {
	return e.From
}

// GetTo returns the to field for MessageReceivedEventData
func (e MessageReceivedEventData) GetTo() string {
	return e.To
}

// GetMessage returns the message for MessageReceivedEventData
func (e MessageReceivedEventData) GetMessage() string {
	return e.Message
}

// WebhookConfig represents the configuration for webhook integration
type WebhookConfig struct {
	URL    string             `json:"url"`
	Events []WebhookEventType `json:"events"`
	Secret string             `json:"secret,omitempty"` // Optional secret for webhook signature verification
}

// WebhookResponse represents the response when registering/updating webhooks
type WebhookResponse struct {
	ID     string             `json:"id"`
	URL    string             `json:"url"`
	Events []WebhookEventType `json:"events"`
}

// WebhookDeleteResponse represents the response when deleting a webhook
type WebhookDeleteResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}
