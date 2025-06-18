// Copyright (c) 2025 CloudContactAI LLC
// Licensed under the MIT License. See LICENSE in the project root for license information.

// Package main provides a command-line tool for interacting with the CCAI API.
package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/cloudcontactai/ccai-go/pkg/ccai"
	"github.com/cloudcontactai/ccai-go/pkg/sms"
)

func main() {
	// Define command-line flags
	clientID := flag.String("client-id", "", "CCAI client ID")
	apiKey := flag.String("api-key", "", "CCAI API key")
	firstName := flag.String("first-name", "", "Recipient's first name")
	lastName := flag.String("last-name", "", "Recipient's last name")
	phone := flag.String("phone", "", "Recipient's phone number (E.164 format)")
	message := flag.String("message", "", "Message content")
	title := flag.String("title", "CLI Test", "Campaign title")
	imagePath := flag.String("image", "", "Path to image file for MMS (optional)")
	contentType := flag.String("content-type", "image/jpeg", "Content type of the image file")
	showVersion := flag.Bool("version", false, "Show version information")

	// Parse flags
	flag.Parse()

	// Show version information if requested
	if *showVersion {
		fmt.Printf("CCAI Go Client v%s\n", ccai.Version)
		os.Exit(0)
	}

	// Check required flags
	if *clientID == "" || *apiKey == "" || *firstName == "" || *lastName == "" || *phone == "" || *message == "" {
		fmt.Println("Error: Missing required parameters")
		fmt.Println("Usage:")
		flag.PrintDefaults()
		os.Exit(1)
	}

	// Initialize the client
	client, err := ccai.NewClient(ccai.Config{
		ClientID: *clientID,
		APIKey:   *apiKey,
	})
	if err != nil {
		log.Fatalf("Failed to create CCAI client: %v", err)
	}

	// Create options with progress tracking
	options := &sms.Options{
		OnProgress: func(status string) {
			fmt.Printf("Progress: %s\n", status)
		},
	}

	// Create account
	account := sms.Account{
		FirstName: *firstName,
		LastName:  *lastName,
		Phone:     *phone,
	}

	// Send message
	var response *sms.Response
	if *imagePath != "" {
		// Send MMS
		fmt.Println("Sending MMS...")
		response, err = client.MMS.SendWithImage(
			*imagePath,
			*contentType,
			[]sms.Account{account},
			*message,
			*title,
			options,
			true,
		)
	} else {
		// Send SMS
		fmt.Println("Sending SMS...")
		response, err = client.SMS.Send(
			[]sms.Account{account},
			*message,
			*title,
			options,
		)
	}

	if err != nil {
		log.Fatalf("Failed to send message: %v", err)
	}

	fmt.Printf("Message sent successfully!\n")
	fmt.Printf("ID: %s\n", response.ID)
	fmt.Printf("Status: %s\n", response.Status)
	fmt.Printf("Campaign ID: %s\n", response.CampaignID)
	fmt.Printf("Messages sent: %d\n", response.MessagesSent)
}
