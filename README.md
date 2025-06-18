# CCAI Go Client

A Go client for interacting with the Cloud Contact AI API that allows you to easily send SMS and MMS messages.

## Requirements

- Go 1.18 or higher

## Installation

```bash
go get github.com/cloudcontactai/ccai-go
```

## Usage

### SMS

```go
package main

import (
	"fmt"
	"log"

	"github.com/cloudcontactai/ccai-go/pkg/ccai"
	"github.com/cloudcontactai/ccai-go/pkg/sms"
)

func main() {
	// Initialize the client
	client, err := ccai.NewClient(ccai.Config{
		ClientID: "YOUR-CLIENT-ID",
		APIKey:   "YOUR-API-KEY",
	})
	if err != nil {
		log.Fatalf("Failed to create CCAI client: %v", err)
	}

	// Send a single SMS
	response, err := client.SMS.SendSingle(
		"John",
		"Doe",
		"+15551234567",
		"Hello ${firstName}, this is a test message!",
		"Test Campaign",
		nil,
	)
	if err != nil {
		log.Fatalf("Failed to send SMS: %v", err)
	}

	fmt.Printf("Message sent with ID: %s\n", response.ID)

	// Send to multiple recipients
	accounts := []sms.Account{
		{
			FirstName: "John",
			LastName:  "Doe",
			Phone:     "+15551234567",
		},
		{
			FirstName: "Jane",
			LastName:  "Smith",
			Phone:     "+15559876543",
		},
	}

	campaignResponse, err := client.SMS.Send(
		accounts,
		"Hello ${firstName} ${lastName}, this is a test message!",
		"Bulk Test Campaign",
		nil,
	)
	if err != nil {
		log.Fatalf("Failed to send bulk SMS: %v", err)
	}

	fmt.Printf("Campaign sent with ID: %s\n", campaignResponse.CampaignID)
}
```

### MMS

```go
package main

import (
	"fmt"
	"log"

	"github.com/cloudcontactai/ccai-go/pkg/ccai"
	"github.com/cloudcontactai/ccai-go/pkg/sms"
)

func main() {
	// Initialize the client
	client, err := ccai.NewClient(ccai.Config{
		ClientID: "YOUR-CLIENT-ID",
		APIKey:   "YOUR-API-KEY",
	})
	if err != nil {
		log.Fatalf("Failed to create CCAI client: %v", err)
	}

	// Define progress tracking
	options := &sms.Options{
		Timeout: 60,
		OnProgress: func(status string) {
			fmt.Printf("Progress: %s\n", status)
		},
	}

	// Complete MMS workflow (get URL, upload image, send MMS)
	imagePath := "path/to/your/image.jpg"
	contentType := "image/jpeg"

	// Define recipient
	account := sms.Account{
		FirstName: "John",
		LastName:  "Doe",
		Phone:     "+15551234567",  // Use E.164 format
	}

	// Send MMS with image in one step
	response, err := client.MMS.SendWithImage(
		imagePath,
		contentType,
		[]sms.Account{account},
		"Hello ${firstName}, check out this image!",
		"MMS Campaign Example",
		options,
		true,
	)
	if err != nil {
		log.Fatalf("Error sending MMS: %v", err)
	}

	fmt.Printf("MMS sent! Campaign ID: %s\n", response.CampaignID)
}
```

### Step-by-Step MMS Workflow

```go
// Step 1: Get a signed URL for uploading
uploadResponse, err := client.MMS.GetSignedUploadURL(
	"image.jpg",
	"image/jpeg",
	"",
	true,
)
if err != nil {
	log.Fatalf("Error getting signed URL: %v", err)
}

signedURL := uploadResponse.SignedS3URL
fileKey := uploadResponse.FileKey

// Step 2: Upload the image to the signed URL
uploadSuccess, err := client.MMS.UploadImageToSignedURL(
	signedURL,
	"path/to/your/image.jpg",
	"image/jpeg",
)
if err != nil {
	log.Fatalf("Error uploading image: %v", err)
}

if uploadSuccess {
	// Step 3: Send the MMS with the uploaded image
	response, err := client.MMS.Send(
		fileKey,
		accounts,
		"Hello ${firstName}, check out this image!",
		"MMS Campaign Example",
		nil,
		true,
	)
	if err != nil {
		log.Fatalf("Error sending MMS: %v", err)
	}

	fmt.Printf("MMS sent! Campaign ID: %s\n", response.CampaignID)
}
```

### With Progress Tracking

```go
// Create options with progress tracking
options := &sms.Options{
	Timeout: 60,
	Retries: 3,
	OnProgress: func(status string) {
		fmt.Printf("%s - %s\n", time.Now().Format("2006-01-02 15:04:05"), status)
	},
}

// Send SMS with progress tracking
response, err := client.SMS.Send(
	accounts,
	message,
	title,
	options,
)
```

## Project Structure

- `pkg/` - Package code
  - `ccai/` - Main CCAI client package
    - `client.go` - Main CCAI client implementation
    - `ccai.go` - Type definitions and exports
  - `sms/` - SMS-related functionality
    - `models.go` - Data models
    - `sms.go` - SMS service implementation
    - `mms.go` - MMS service implementation
- `examples/` - Example usage
  - `basic/` - Basic SMS example
  - `mms/` - MMS examples
  - `progress_tracking/` - Progress tracking example
- `test/` - Test files
- `cmd/` - Command-line tools (if any)

## Features

- Send SMS messages to single or multiple recipients
- Send MMS messages with images
- Upload images to S3 with signed URLs
- Variable substitution in messages
- Progress tracking via callbacks
- Comprehensive error handling
- Full test coverage

## License

MIT © 2025 CloudContactAI LLC
