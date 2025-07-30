package main

import (
	"fmt"
	"log"

	"github.com/cloudcontactai/ccai-go/src/pkg/ccai"
)

func main() {
	// Initialize the client with your credentials
	client, err := ccai.NewClient(ccai.Config{
		ClientID: "YOUR_CLIENT_ID",
		APIKey:   "YOUR_API_KEY",
	})
	if err != nil {
		log.Fatalf("Failed to create CCAI client: %v", err)
	}

	// Send a single SMS
	response, err := client.SMS.SendSingle(
		"John",
		"Doe",
		"+14156566694", // Replace this with YOUR actual phone number
		"Hello ${firstName}, this is a NEW test message from Go!",
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
	fmt.Printf("Full Response: %+v\n", response)
}
