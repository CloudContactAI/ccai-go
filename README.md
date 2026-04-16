# CCAI Go Client

A Go client for interacting with the Cloud Contact AI API that allows you to easily send SMS and MMS messages, send email campaigns, manage webhooks, and manage contact opt-out preferences.

## Requirements

- Go 1.18 or higher

## Installation

```bash
go get github.com/cloudcontactai/ccai-go
```

## Usage

### Environment Variables

Create a `.env` file in your project root:

```env
CCAI_CLIENT_ID=your_client_id
CCAI_API_KEY=your_api_key
```

### Email

```go
package main

import (
    "fmt"
    "log"
    "os"

    "github.com/cloudcontactai/ccai-go/pkg/ccai"
    "github.com/joho/godotenv"
)

func main() {
    // Load environment variables
    err := godotenv.Load()
    if err != nil {
        log.Printf("Warning: Could not load .env file: %v", err)
    }

    // Initialize the client
    client, err := ccai.NewClient(ccai.Config{
        ClientID: os.Getenv("CCAI_CLIENT_ID"),
        APIKey:   os.Getenv("CCAI_API_KEY"),
    })
    if err != nil {
        log.Fatalf("Failed to create CCAI client: %v", err)
    }

    // Send a single email
    response, err := client.Email.SendSingle(
        "John",                              // firstName
        "Doe",                               // lastName
        "recipient@example.com",             // email
        "Test Email Subject",                // subject
        "<p>Hello John, this is a test!</p>", // message (HTML)
        "noreply@cloudcontactai.com",        // senderEmail
        "support@cloudcontactai.com",        // replyEmail
        "CloudContactAI",                    // senderName
        "Test Campaign",                     // title
        nil,                                 // options
    )
    if err != nil {
        log.Fatalf("Failed to send email: %v", err)
    }

    fmt.Printf("Email sent with ID: %d\n", response.ID)
}
```

### SMS

```go
package main

import (
    "fmt"
    "log"
    "os"

    "github.com/cloudcontactai/ccai-go/pkg/ccai"
    "github.com/cloudcontactai/ccai-go/pkg/sms"
    "github.com/joho/godotenv"
)

func main() {
    // Load environment variables
    err := godotenv.Load()
    if err != nil {
        log.Printf("Warning: Could not load .env file: %v", err)
    }

    // Initialize the client
    client, err := ccai.NewClient(ccai.Config{
        ClientID: os.Getenv("CCAI_CLIENT_ID"),
        APIKey:   os.Getenv("CCAI_API_KEY"),
    })
    if err != nil {
        log.Fatalf("Failed to create CCAI client: %v", err)
    }

	// Send a single SMS
	response, err := client.SMS.SendSingle(
		"John",
		"Doe",
		"+14156566694",
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
			Phone:     "+14156566694",
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
    "os"

    "github.com/cloudcontactai/ccai-go/pkg/ccai"
    "github.com/cloudcontactai/ccai-go/pkg/sms"
    "github.com/joho/godotenv"
)

func main() {
    // Load environment variables
    err := godotenv.Load()
    if err != nil {
        log.Printf("Warning: Could not load .env file: %v", err)
    }

    // Initialize the client
    client, err := ccai.NewClient(ccai.Config{
        ClientID: os.Getenv("CCAI_CLIENT_ID"),
        APIKey:   os.Getenv("CCAI_API_KEY"),
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
		Phone:     "+14156566694",  // Use E.164 format
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

### Contact

Manage opt-out preferences for contacts.

```go
package main

import (
    "fmt"
    "log"
    "os"

    "github.com/cloudcontactai/ccai-go/pkg/ccai"
    "github.com/joho/godotenv"
)

func main() {
    err := godotenv.Load()
    if err != nil {
        log.Printf("Warning: Could not load .env file: %v", err)
    }

    client, err := ccai.NewClient(ccai.Config{
        ClientID: os.Getenv("CCAI_CLIENT_ID"),
        APIKey:   os.Getenv("CCAI_API_KEY"),
    })
    if err != nil {
        log.Fatalf("Failed to create CCAI client: %v", err)
    }

    // Opt a contact out of text messages (by phone)
    result, err := client.Contact.SetDoNotText(true, "", "+15551234567")
    if err != nil {
        log.Fatalf("Failed to set do-not-text: %v", err)
    }
    fmt.Printf("Opted out contact: %s\n", result.Phone)

    // Opt a contact back in
    _, err = client.Contact.SetDoNotText(false, "", "+15551234567")
    if err != nil {
        log.Fatalf("Failed to opt in: %v", err)
    }

    // Opt out by contactId
    _, err = client.Contact.SetDoNotText(true, "contact-abc-123", "")
    if err != nil {
        log.Fatalf("Failed to set do-not-text by ID: %v", err)
    }
}
```

### Webhooks

```go
package main

import (
    "fmt"
    "log"
    "net/http"
    "os"

    "github.com/cloudcontactai/ccai-go/pkg/ccai"
    "github.com/cloudcontactai/ccai-go/pkg/webhook"
    "github.com/joho/godotenv"
)

func main() {
    err := godotenv.Load()
    if err != nil {
        log.Printf("Warning: Could not load .env file: %v", err)
    }

    client, err := ccai.NewClient(ccai.Config{
        ClientID: os.Getenv("CCAI_CLIENT_ID"),
        APIKey:   os.Getenv("CCAI_API_KEY"),
    })
    if err != nil {
        log.Fatalf("Failed to create CCAI client: %v", err)
    }

    // Register a new webhook - server generates secret automatically
    wh, err := client.Webhook.Register(webhook.WebhookConfig{
        URL: "https://your-app.com/api/ccai-webhook",
        // Secret is optional - if not provided, server generates one automatically
    })
    if err != nil {
        log.Fatalf("Failed to register webhook: %v", err)
    }
    fmt.Printf("Webhook registered with ID: %s\n", wh.ID)
    fmt.Printf("Secret Key: %s\n", wh.SecretKey)  // Save this securely!

    // Or provide a custom secret if needed
    customSecret := "your-custom-secret"
    wh2, err := client.Webhook.Register(webhook.WebhookConfig{
        URL:    "https://your-app.com/api/custom-webhook",
        Secret: &customSecret,
    })
    if err != nil {
        log.Fatalf("Failed to register webhook: %v", err)
    }
    fmt.Printf("Custom secret webhook registered: %s\n", wh2.SecretKey)

    // List all webhooks
    webhooks, err := client.Webhook.List()
    if err != nil {
        log.Fatalf("Failed to list webhooks: %v", err)
    }
    fmt.Printf("Registered webhooks: %d\n", len(webhooks))

    // Update a webhook
    updated, err := client.Webhook.Update(wh.ID, webhook.WebhookConfig{
        URL: "https://your-app.com/api/new-webhook",
    })
    if err != nil {
        log.Fatalf("Failed to update webhook: %v", err)
    }
    fmt.Printf("Updated webhook URL: %s\n", updated.URL)

    // Delete a webhook
    _, err = client.Webhook.Delete(wh.ID)
    if err != nil {
        log.Fatalf("Failed to delete webhook: %v", err)
    }

    // Verify webhook signature (in your HTTP handler)
    http.HandleFunc("/api/ccai-webhook", func(w http.ResponseWriter, r *http.Request) {
        signature := r.Header.Get("X-CCAI-Signature")
        
        // Parse the JSON body to get eventHash
        var payload map[string]interface{}
        if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
            http.Error(w, "Invalid JSON", http.StatusBadRequest)
            return
        }
        defer r.Body.Close()

        clientID := os.Getenv("CCAI_CLIENT_ID")
        eventHash, ok := payload["eventHash"].(string)
        if !ok {
            http.Error(w, "Missing eventHash", http.StatusBadRequest)
            return
        }

        valid := client.Webhook.VerifySignature(signature, clientID, eventHash, "your-webhook-secret")
        if !valid {
            http.Error(w, "Invalid signature", http.StatusUnauthorized)
            return
        }

        // Process the event
        eventData := payload["data"].(map[string]interface{})
        fmt.Printf("Event received: %v\n", eventData)
        
        w.WriteHeader(http.StatusOK)
    })
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

- `src/` - Source code
  - `pkg/` - Package code
    - `ccai/` - Main CCAI client package
      - `client.go` - Main CCAI client implementation
      - `ccai.go` - Type definitions and exports
    - `sms/` - SMS and MMS functionality
      - `models.go` - Data models
      - `sms.go` - SMS service implementation
      - `mms.go` - MMS service implementation
    - `email/` - Email functionality
      - `models.go` - Email data models
      - `email.go` - Email service implementation
    - `contact/` - Contact management
      - `contact.go` - Contact service (opt-out)
    - `webhook/` - Webhook functionality
      - `service.go` - Webhook CRUD service
      - `webhook.go` - Webhook client and signature verification
      - `types.go` - Webhook type definitions
      - `handler.go` - Webhook event handler
  - `examples/` - Example usage
- `.env` - Environment variables
- `.env.example` - Environment variables template

## Features

- Send SMS messages to single or multiple recipients
- Send MMS messages with images (automatic S3 upload)
- Send Email campaigns with HTML content
- Manage contact opt-out preferences (SetDoNotText)
- Webhook management: register, list, update, delete
- Webhook signature verification (HMAC-SHA256)
- Template variable substitution (`${firstName}`, `${lastName}`)
- Progress tracking via callbacks
- Environment variable support with .env files
- Comprehensive error handling
- Full test coverage

## License

MIT © 2025 CloudContactAI LLC
