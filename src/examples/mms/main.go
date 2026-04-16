// Copyright (c) 2025 CloudContactAI LLC
// Licensed under the MIT License. See LICENSE in the project root for license information.

// Package main provides an example of using the MMS functionality of the CCAI Go client.
package main

import (
	"fmt"
	"log"

	"github.com/cloudcontactai/ccai-go/src/pkg/ccai"
	"github.com/cloudcontactai/ccai-go/src/pkg/sms"
)

func main() {
	// Initialize the client
	client, err := ccai.NewClient(ccai.Config{
		ClientID: "YOUR_CLIENT_ID",
		APIKey:   "YOUR_API_KEY",
	})
	if err != nil {
		log.Fatalf("Failed to create CCAI client: %v", err)
	}

	// Example 1: Complete MMS workflow (get URL, upload image, send MMS)
	fmt.Println("=== Example 1: Complete MMS workflow ===")
	sendMMSWithImage(client)

	// Example 2: Step-by-step MMS workflow
	fmt.Println("\n=== Example 2: Step-by-step MMS workflow ===")
	sendMMSStepByStep(client)

	// Example 3: Send a single MMS
	fmt.Println("\n=== Example 3: Send a single MMS ===")
	sendSingleMMS(client)
}

func sendMMSWithImage(client *ccai.Client) {
	// Path to your image file
	imagePath := "path/to/your/image.jpg"
	contentType := "image/jpeg"

	// Define recipient
	account := sms.Account{
		FirstName: "John",
		LastName:  "Doe",
		Phone:     "+14156566694", // Use E.164 format
	}

	// Message content and campaign title
	message := "Hello ${firstName}, check out this image!"
	title := "MMS Campaign Example"

	// Define progress tracking
	progressUpdates := []string{}
	options := &sms.Options{
		Timeout: 60,
		OnProgress: func(status string) {
			fmt.Printf("Progress: %s\n", status)
			progressUpdates = append(progressUpdates, status)
		},
	}

	// Send MMS with image in one step
	response, err := client.MMS.SendWithImage(
		imagePath,
		contentType,
		[]sms.Account{account},
		message,
		title,
		"",
		options,
		true,
	)
	if err != nil {
		log.Printf("Error sending MMS: %v", err)
		return
	}

	fmt.Printf("MMS sent! Campaign ID: %s\n", response.CampaignID)
	fmt.Printf("Messages sent: %d\n", response.MessagesSent)
	fmt.Printf("Status: %s\n", response.Status)
}

func sendMMSStepByStep(client *ccai.Client) {
	// Path to your image file
	imagePath := "path/to/your/image.jpg"
	fileName := "image.jpg" // You would normally extract this from the path
	contentType := "image/jpeg"

	// Step 1: Get a signed URL for uploading
	fmt.Println("Getting signed upload URL...")
	uploadResponse, err := client.MMS.GetSignedUploadURL(
		fileName,
		contentType,
		"",
		true,
	)
	if err != nil {
		log.Printf("Error getting signed URL: %v", err)
		return
	}

	signedURL := uploadResponse.SignedS3URL
	fileKey := uploadResponse.FileKey

	fmt.Printf("Got signed URL: %s\n", signedURL)
	fmt.Printf("File key: %s\n", fileKey)

	// Step 2: Upload the image to the signed URL
	fmt.Println("Uploading image...")
	uploadSuccess, err := client.MMS.UploadImageToSignedURL(
		signedURL,
		imagePath,
		contentType,
	)
	if err != nil {
		log.Printf("Error uploading image: %v", err)
		return
	}

	if !uploadSuccess {
		fmt.Println("Failed to upload image")
		return
	}

	fmt.Println("Image uploaded successfully")

	// Step 3: Send the MMS with the uploaded image
	fmt.Println("Sending MMS...")

	// Define recipients
	accounts := []sms.Account{
		{
			FirstName: "John",
			LastName:  "Doe",
			Phone:     "+14156566694",
		},
		{
			FirstName: "Jane",
			LastName:  "Smith",
			Phone:     "+15559876543",
		},
	}

	// Message content and campaign title
	message := "Hello ${firstName}, check out this image!"
	title := "MMS Campaign Example"

	// Send the MMS
	response, err := client.MMS.Send(
		fileKey,
		accounts,
		message,
		title,
		"",
		nil,
		true,
	)
	if err != nil {
		log.Printf("Error sending MMS: %v", err)
		return
	}

	fmt.Printf("MMS sent! Campaign ID: %s\n", response.CampaignID)
	fmt.Printf("Messages sent: %d\n", response.MessagesSent)
	fmt.Printf("Status: %s\n", response.Status)
}

func sendSingleMMS(client *ccai.Client) {
	// Define the file key of an already uploaded image
	pictureFileKey := "YOUR_CLIENT_ID/campaign/your-image.jpg"

	// Send a single MMS
	response, err := client.MMS.SendSingle(
		pictureFileKey,
		"John",
		"Doe",
		"+14156566694",
		"Hello ${firstName}, check out this image!",
		"Single MMS Example",
		"",
		"",
		nil,
		true,
	)
	if err != nil {
		log.Printf("Error sending single MMS: %v", err)
		return
	}

	fmt.Printf("MMS sent! Campaign ID: %s\n", response.CampaignID)
	fmt.Printf("Status: %s\n", response.Status)
}
