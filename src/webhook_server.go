// Copyright (c) 2025 CloudContactAI LLC
// Licensed under the MIT License. See LICENSE in the project root for license information.

// Package main provides a basic webhook server example.
package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/cloudcontactai/ccai-go/src/pkg/webhook"
)

func main() {
	fmt.Println("Starting CloudContactAI Webhook Server...")

	// Create webhook handler
	webhookHandler := webhook.CreateHandler(webhook.HandlerOptions{
		LogEvents: true,

		OnMessageSent: func(event webhook.MessageSentEventData) error {
			fmt.Printf("Message Sent - To: %s, Message: %s\n",
				event.GetTo(), event.GetMessage())
			return nil
		},

		OnMessageReceived: func(event webhook.MessageReceivedEventData) error {
			fmt.Printf("Message Received - From: %s, Message: %s\n",
				event.GetFrom(), event.GetMessage())
			return nil
		},
	})

	// Set up routes
	http.Handle("/webhook", webhookHandler)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "CloudContactAI Webhook Server is running!\nEndpoint: /webhook")
	})

	fmt.Println("Webhook endpoint: http://localhost:8080/webhook")
	fmt.Println("Test page: http://localhost:8080")

	log.Fatal(http.ListenAndServe(":8080", nil))
}
