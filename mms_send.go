package main

import (
	"fmt"
	"log"

	"github.com/cloudcontactai/ccai-go/pkg/ccai"
	"github.com/cloudcontactai/ccai-go/pkg/sms"
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

	// Define recipient
	accounts := []sms.Account{
		{
			FirstName: "John",
			LastName:  "Doe",
			Phone:     "+14156566694",
		},
	}

	// Progress tracking
	options := &sms.Options{
		OnProgress: func(status string) {
			fmt.Printf("Progress: %s\n", status)
		},
	}

	// Send MMS with image
	response, err := client.MMS.SendWithImage(
		"imageGO.jpg",
		"image/jpeg",
		accounts,
		"Hello ${firstName}, check out this image from Go!",
		"Go MMS Test Campaign",
		options,
		true,
	)
	if err != nil {
		log.Fatalf("Failed to send MMS: %v", err)
	}

	fmt.Printf("MMS sent successfully!\n")
	fmt.Printf("Message ID: %s\n", response.GetID())
	fmt.Printf("Status: %s\n", response.Status)
	fmt.Printf("Campaign ID: %s\n", response.CampaignID)
}
