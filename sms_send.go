package main

import (
	"fmt"
	"log"

	"github.com/cloudcontactai/ccai-go/pkg/ccai"
)

func main() {
	// Initialize the client with your credentials
	client, err := ccai.NewClient(ccai.Config{
		ClientID: "YOUR-CLIENT-ID",
		APIKey:   "YOUR-API-KEY"
	})
	if err != nil {
		log.Fatalf("Failed to create CCAI client: %v", err)
	}

	// Send a single SMS
	response, err := client.SMS.SendSingle(
		"John",
		"Doe",
		"+15551234567", // Replace with your phone number
		"Hello ${firstName}, this is a test message from Go!",
		"Go Test Campaign",
		nil,
	)
	if err != nil {
		log.Fatalf("Failed to send SMS: %v", err)
	}

	fmt.Printf("Message sent successfully!\n")
	fmt.Printf("Message ID: %s\n", response.GetID())
	fmt.Printf("Status: %s\n", response.Status)
	fmt.Printf("Campaign ID: %s\n", response.CampaignID)
}