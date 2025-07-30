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
	err := godotenv.Load(".env")
	if err != nil {
		err = godotenv.Load("../.env")
		if err != nil {
			log.Printf("Warning: Could not load .env file: %v", err)
		}
	}

	// Initialize the client with credentials from environment variables
	client, err := ccai.NewClient(ccai.Config{
		ClientID: os.Getenv("CCAI_CLIENT_ID"),
		APIKey:   os.Getenv("CCAI_API_KEY"),
	})
	if err != nil {
		log.Fatalf("Failed to create CCAI client: %v", err)
	}

	// Send a single email
	response, err := client.Email.SendSingle(
		"Thava",                          // firstName
		"Antonio",                        // lastName
		"thavasantonio@gmail.com",        // email
		"Test Email from CloudContactAI", // subject
		"<h1>Test Email</h1><p>Hello Thava,</p><p>This is a test email from CloudContactAI Go SDK.</p><p>If you received this, the integration is working!</p><p>Best regards,<br>CloudContactAI Team</p>", // message
		"noreply@noreply@allcode.com", // senderEmail - Use a proper domain
		"support@noreply@allcode.com", // replyEmail
		"CloudContactAI",              // senderName
		"Welcome Email",               // title
		nil,                           // options
	)

	if err != nil {
		log.Fatalf("Failed to send email: %v", err)
	}

	fmt.Printf("Email sent successfully: %+v\n", response)
}
