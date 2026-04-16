// Copyright (c) 2025 CloudContactAI LLC
// Licensed under the MIT License. See LICENSE in the project root for license information.

// Example demonstrating email sending functionality
package main

import (
	"fmt"
	"log"
	"os"

	"github.com/cloudcontactai/ccai-go/src/pkg/ccai"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables from .env file
	err := godotenv.Load("../../../.env") // Load from project root
	if err != nil {
		err = godotenv.Load(".env") // Try current directory
		if err != nil {
			log.Printf("Warning: Could not load .env file: %v", err)
		}
	}

	// Initialize the CCAI client
	config := ccai.Config{
		ClientID: os.Getenv("CCAI_CLIENT_ID"),
		APIKey:   os.Getenv("CCAI_API_KEY"),
	}

	if config.ClientID == "" || config.APIKey == "" {
		log.Fatal("CCAI_CLIENT_ID and CCAI_API_KEY environment variables are required")
	}

	client, err := ccai.NewClient(config)
	if err != nil {
		log.Fatalf("Failed to create CCAI client: %v", err)
	}

	// Example 1: Send a single email
	fmt.Println("Sending single email...")
	err = sendSingleEmail(client)
	if err != nil {
		log.Printf("Error sending single email: %v", err)
	}

	// Example 2: Send an email campaign
	fmt.Println("Sending email campaign...")
	err = sendEmailCampaign(client)
	if err != nil {
		log.Printf("Error sending email campaign: %v", err)
	}
}

// sendSingleEmail demonstrates sending a single email
func sendSingleEmail(client *ccai.Client) error {
	options := &ccai.EmailOptions{
		OnProgress: func(status string) {
			fmt.Printf("Progress: %s\n", status)
		},
	}

	response, err := client.Email.SendSingle(
		"Thava",                   // firstName
		"Antonio",                 // lastName
		"thavasantonio@gmail.com", // email
		"Welcome to Our Service",  // subject
		"<p>Hello Thava,</p><p>Thank you for signing up for our service!</p><p>Best regards,<br>The Team</p>", // message
		"",                            // textContent (optional plain text fallback)
		"noreply@noreply@allcode.com", // senderEmail
		"support@noreply@allcode.com", // replyEmail
		"CloudContactAI",              // senderName
		"Welcome Email",               // title
		options,
	)

	if err != nil {
		return fmt.Errorf("failed to send single email: %w", err)
	}

	fmt.Printf("Single email sent successfully: %+v\n", response)
	return nil
}

// sendEmailCampaign demonstrates sending an email campaign to multiple recipients
func sendEmailCampaign(client *ccai.Client) error {
	accounts := []ccai.EmailAccount{
		{
			FirstName: "Thava",
			LastName:  "Antonio",
			Email:     "thavasantonio@gmail.com",
			Phone:     "",
		},
	}

	campaign := &ccai.EmailCampaign{
		Subject:      "Monthly Newsletter",
		Title:        "July 2025 Newsletter",
		Message:      "<h1>Monthly Newsletter - July 2025</h1><p>Hello ${firstName},</p><p>Here are our updates for this month:</p><ul><li>New feature: Email campaigns</li><li>Improved performance</li><li>Bug fixes</li></ul><p>Thank you for being a valued customer!</p><p>Best regards,<br>The Team</p>",
		SenderEmail:  "noreply@noreply@allcode.com",
		ReplyEmail:   "support@noreply@allcode.com",
		SenderName:   "CloudContactAI",
		Accounts:     accounts,
		CampaignType: "EMAIL",
		AddToList:    "noList",
		ContactInput: "accounts",
		FromType:     "single",
		Senders:      []interface{}{},
	}

	options := &ccai.EmailOptions{
		OnProgress: func(status string) {
			fmt.Printf("Campaign Progress: %s\n", status)
		},
	}

	response, err := client.Email.SendCampaign(campaign, options)
	if err != nil {
		return fmt.Errorf("failed to send email campaign: %w", err)
	}

	fmt.Printf("Email campaign sent successfully: %+v\n", response)
	return nil
}
