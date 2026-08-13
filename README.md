# CCAI Go Client

A Go client for interacting with the Cloud Contact AI API that allows you to easily send SMS and MMS messages, send email campaigns, manage webhooks, and manage contact opt-out preferences.

## Requirements

- Go 1.21.6 or higher

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

    "github.com/cloudcontactai/ccai-go/src/pkg/ccai"
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
        "",                                   // textContent (optional plain-text alternative)
        "noreply@cloudcontactai.com",        // senderEmail
        "support@cloudcontactai.com",        // replyEmail
        "CloudContactAI",                    // senderName
        "Test Campaign",                     // title
        nil,                                 // options
    )
    if err != nil {
        log.Fatalf("Failed to send email: %v", err)
    }

    if response.ID != nil {
        fmt.Printf("Email sent with ID: %d\n", *response.ID)
    }
}
```

### SMS

```go
package main

import (
    "fmt"
    "log"
    "os"

    "github.com/cloudcontactai/ccai-go/src/pkg/ccai"
    "github.com/cloudcontactai/ccai-go/src/pkg/sms"
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
		"", // customData (optional)
		"", // senderPhone (optional)
		nil, // options
	)
	if err != nil {
		log.Fatalf("Failed to send SMS: %v", err)
	}

	fmt.Printf("Message sent with ID: %s\n", response.GetID())

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
		"", // senderPhone (optional)
		nil, // options
	)
	if err != nil {
		log.Fatalf("Failed to send bulk SMS: %v", err)
	}

	fmt.Printf("Campaign sent with ID: %s\n", campaignResponse.CampaignID)
}
```

### SMS — Template-Controlled Accounts

If an account has been configured to enforce template-only messaging, all campaigns must reference a pre-approved template ID. Sending a free-text message to such an account will result in a `422` error.

```go
templateID := int64(12345)

// Send to multiple recipients using a template
response, err := client.SMS.SendWithTemplate(
    accounts,
    templateID,
    "My Campaign",
    "",    // senderPhone (optional)
    nil,   // options
)

// Send to a single recipient using a template
response, err := client.SMS.SendSingleWithTemplate(
    "John",
    "Doe",
    "+15551234567",
    templateID,
    "My Campaign",
    "",  // senderPhone (optional)
    nil, // options
)
```

The message body is resolved server-side from the template. Variable substitution (e.g. `${firstName}`) is applied automatically using the recipient's account data.

### MMS

```go
package main

import (
    "fmt"
    "log"
    "os"

    "github.com/cloudcontactai/ccai-go/src/pkg/ccai"
    "github.com/cloudcontactai/ccai-go/src/pkg/sms"
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
		"", // senderPhone (optional)
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
		"", // senderPhone (optional)
		nil, // options
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

    "github.com/cloudcontactai/ccai-go/src/pkg/ccai"
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

### Contact Validator

Validate email addresses and phone numbers.

> Bulk endpoints accept up to 50 contacts per request and are processed server-side in chunks.

```go
package main

import (
    "fmt"
    "log"
    "os"

    "github.com/cloudcontactai/ccai-go/src/pkg/ccai"
    "github.com/cloudcontactai/ccai-go/src/pkg/contactvalidator"
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

    // Validate a single email
    emailResult, err := client.ContactValidator.ValidateEmail("user@example.com")
    if err != nil {
        log.Fatalf("Failed to validate email: %v", err)
    }
    fmt.Printf("Email status: %s\n", emailResult.Status) // "valid" | "invalid" | "risky"

    // Validate multiple emails (up to 50, processed server-side in chunks)
    bulkEmails, err := client.ContactValidator.ValidateEmails([]string{
        "user@example.com",
        "bad@invalid.xyz",
    })
    if err != nil {
        log.Fatalf("Failed to validate emails: %v", err)
    }
    fmt.Printf("Email summary: %+v\n", bulkEmails.Summary)

    // Validate a single phone number
    phoneResult, err := client.ContactValidator.ValidatePhone("+15551234567", "US")
    if err != nil {
        log.Fatalf("Failed to validate phone: %v", err)
    }
    fmt.Printf("Phone status: %s\n", phoneResult.Status) // "valid" | "invalid" | "landline"

    // Validate multiple phone numbers (up to 50, processed server-side in chunks)
    bulkPhones, err := client.ContactValidator.ValidatePhones([]contactvalidator.PhoneInput{
        {Phone: "+15551234567"},
        {Phone: "+15559876543", CountryCode: "US"},
    })
    if err != nil {
        log.Fatalf("Failed to validate phones: %v", err)
    }
    fmt.Printf("Phone summary: %+v\n", bulkPhones.Summary)
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

    "github.com/cloudcontactai/ccai-go/src/pkg/ccai"
    "github.com/cloudcontactai/ccai-go/src/pkg/webhook"
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
    webhookID := fmt.Sprintf("%v", wh.ID) // wh.ID is interface{}; convert once for Update/Delete
    fmt.Printf("Webhook registered with ID: %s\n", webhookID)
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
    updated, err := client.Webhook.Update(webhookID, webhook.WebhookConfig{
        URL: "https://your-app.com/api/new-webhook",
    })
    if err != nil {
        log.Fatalf("Failed to update webhook: %v", err)
    }
    fmt.Printf("Updated webhook URL: %s\n", updated.URL)

    // Delete a webhook
    _, err = client.Webhook.Delete(webhookID)
    if err != nil {
        log.Fatalf("Failed to delete webhook: %v", err)
    }

    // Handle incoming webhook events with the built-in HTTP handler helper.
    // It rejects non-POST requests, verifies the signature (when Secret is set),
    // and dispatches the parsed event to OnEvent — no manual JSON parsing needed.
    handler := webhook.CreateHandler(webhook.HandlerOptions{
        ClientID: os.Getenv("CCAI_CLIENT_ID"),
        Secret:   "your-webhook-secret", // the secret from webhook registration
        OnEvent: func(event *webhook.WebhookEvent) error {
            fmt.Printf("Event received: %s\n", event.EventType)
            fmt.Printf("Data: %v\n", event.Data)
            return nil
        },
        LogEvents: true,
    })

    http.HandleFunc("/api/ccai-webhook", handler)
    log.Fatal(http.ListenAndServe(":8080", nil))
}
```

### Brand Registration

Register and manage brands for TCR (The Campaign Registry) business verification.

`BrandRequest` fields are all pointers so that partial updates only send the fields you set. A small `strPtr` helper makes that ergonomic:

```go
package main

import (
    "fmt"
    "log"
    "os"

    "github.com/cloudcontactai/ccai-go/src/pkg/brands"
    "github.com/cloudcontactai/ccai-go/src/pkg/ccai"
    "github.com/joho/godotenv"
)

func strPtr(s string) *string { return &s }

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

    // Create a brand
    brand, err := client.Brands.Create(brands.BrandRequest{
        LegalCompanyName: strPtr("Collect.org Inc."),
        Dba:              strPtr("Collect"),
        EntityType:       strPtr("NON_PROFIT"),
        TaxId:            strPtr("123456789"),
        TaxIdCountry:     strPtr("US"),
        Country:          strPtr("US"),
        VerticalType:     strPtr("NON_PROFIT"),
        WebsiteUrl:       strPtr("https://www.collect.org"),
        Street:           strPtr("123 Main Street"),
        City:             strPtr("San Francisco"),
        State:            strPtr("CA"),
        PostalCode:       strPtr("94105"),
        ContactFirstName: strPtr("Jane"),
        ContactLastName:  strPtr("Doe"),
        ContactEmail:     strPtr("jane@collect.org"),
        ContactPhone:     strPtr("+14155551234"),
    })
    if err != nil {
        log.Fatalf("Failed to create brand: %v", err)
    }
    fmt.Printf("Brand created with ID: %d\n", brand.ID)

    // Get a brand by ID
    fetched, err := client.Brands.Get(brand.ID)
    if err != nil {
        log.Fatalf("Failed to get brand: %v", err)
    }
    fmt.Printf("Website match score: %v\n", fetched.WebsiteMatchScore)

    // List all brands
    allBrands, err := client.Brands.List()
    if err != nil {
        log.Fatalf("Failed to list brands: %v", err)
    }
    fmt.Printf("Found %d brand(s)\n", len(allBrands))

    // Update a brand (partial update)
    updated, err := client.Brands.Update(brand.ID, brands.BrandRequest{
        Street: strPtr("456 Oak Avenue"),
        City:   strPtr("Los Angeles"),
    })
    if err != nil {
        log.Fatalf("Failed to update brand: %v", err)
    }
    fmt.Printf("Brand updated: %s, %s\n", updated.Street, updated.City)

    // Delete a brand
    if err := client.Brands.Delete(brand.ID); err != nil {
        log.Fatalf("Failed to delete brand: %v", err)
    }
}
```

#### Entity Types

`PRIVATE_PROFIT`, `PUBLIC_PROFIT`, `NON_PROFIT`, `GOVERNMENT`, `SOLE_PROPRIETOR`

> Note: `PUBLIC_PROFIT` entities require `StockSymbol` and `StockExchange` fields.

#### Vertical Types

`AUTOMOTIVE`, `AGRICULTURE`, `BANKING`, `COMMUNICATION`, `CONSTRUCTION`, `EDUCATION`, `ENERGY`, `ENTERTAINMENT`, `GOVERNMENT`, `HEALTHCARE`, `HOSPITALITY`, `INSURANCE`, `LEGAL`, `MANUFACTURING`, `NON_PROFIT`, `PROFESSIONAL`, `REAL_ESTATE`, `RETAIL`, `TECHNOLOGY`, `TRANSPORTATION`

### Campaign Registration

Register and manage campaigns for TCR (The Campaign Registry) carrier vetting. Each campaign must be linked to a verified brand.

```go
package main

import (
    "fmt"
    "log"
    "os"

    "github.com/cloudcontactai/ccai-go/src/pkg/campaigns"
    "github.com/cloudcontactai/ccai-go/src/pkg/ccai"
    "github.com/joho/godotenv"
)

func boolPtr(b bool) *bool { return &b }

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

    // Create a campaign
    campaign, err := client.Campaigns.Create(campaigns.CampaignRequest{
        BrandID:          1,
        UseCase:          "MIXED",
        SubUseCases:      []string{"CUSTOMER_CARE", "TWO_FACTOR_AUTHENTICATION", "ACCOUNT_NOTIFICATION"},
        Description:      "Security codes and support messaging.",
        MessageFlow:      "Users opt-in via signup form at https://example.com/signup",
        TermsLink:        "https://example.com/terms",
        PrivacyLink:      "https://example.com/privacy",
        HasEmbeddedLinks: boolPtr(true),
        HasEmbeddedPhone: boolPtr(false),
        IsAgeGated:       boolPtr(false),
        IsDirectLending:  boolPtr(false),
        OptInKeywords:    []string{"START"},
        OptInMessage:     "Welcome! Reply STOP to cancel.",
        OptInProofUrl:    "https://example.com/opt-in-proof.png",
        HelpKeywords:     []string{"HELP"},
        HelpMessage:      "For HELP email support@example.com.",
        OptOutKeywords:   []string{"STOP"},
        OptOutMessage:    "STOP received. You are unsubscribed.",
        SampleMessages: []string{
            "Your code is 554321. Reply STOP to cancel.",
            "Your ticket has been updated. Reply HELP for info.",
        },
    })
    if err != nil {
        log.Fatalf("Failed to create campaign: %v", err)
    }
    fmt.Printf("Campaign created with ID: %d\n", campaign.ID)

    // Get a campaign by ID
    fetched, err := client.Campaigns.Get(campaign.ID)
    if err != nil {
        log.Fatalf("Failed to get campaign: %v", err)
    }
    fmt.Printf("Campaign use case: %s\n", fetched.UseCase)

    // List all campaigns
    allCampaigns, err := client.Campaigns.List()
    if err != nil {
        log.Fatalf("Failed to list campaigns: %v", err)
    }
    fmt.Printf("Found %d campaign(s)\n", len(allCampaigns))

    // Update a campaign (partial update)
    updated, err := client.Campaigns.Update(campaign.ID, campaigns.CampaignRequest{
        Description: "Updated description.",
    })
    if err != nil {
        log.Fatalf("Failed to update campaign: %v", err)
    }
    fmt.Printf("Campaign updated: %s\n", updated.Description)

    // Delete a campaign
    if err := client.Campaigns.Delete(campaign.ID); err != nil {
        log.Fatalf("Failed to delete campaign: %v", err)
    }
}
```

#### Use Cases

`TWO_FACTOR_AUTHENTICATION`, `ACCOUNT_NOTIFICATION`, `CUSTOMER_CARE`, `DELIVERY_NOTIFICATION`, `FRAUD_ALERT`, `HIGHER_EDUCATION`, `LOW_VOLUME_MIXED`, `MARKETING`, `MIXED`, `POLLING_VOTING`, `PUBLIC_SERVICE_ANNOUNCEMENT`, `SECURITY_ALERT`

> Note: `MIXED` and `LOW_VOLUME_MIXED` campaigns require 2–3 `SubUseCases`.

#### Sub-Use Cases

`TWO_FACTOR_AUTHENTICATION`, `ACCOUNT_NOTIFICATION`, `CUSTOMER_CARE`, `DELIVERY_NOTIFICATION`, `FRAUD_ALERT`, `MARKETING`, `POLLING_VOTING`

### With Progress Tracking

```go
// Create options with progress tracking
options := &sms.Options{
	OnProgress: func(status string) {
		fmt.Printf("%s - %s\n", time.Now().Format("2006-01-02 15:04:05"), status)
	},
}

// Send SMS with progress tracking
response, err := client.SMS.Send(
	accounts,
	message,
	title,
	"", // senderPhone (optional)
	options,
)
```

## Project Structure

- `src/` - Source code
  - `index.go` - Root-package re-exports (`ccai`, `sms`, `email` types only)
  - `pkg/` - Package code
    - `ccai/` - Main CCAI client package
      - `client.go` - Main CCAI client implementation
    - `sms/` - SMS and MMS functionality
      - `models.go` - Data models
      - `sms.go` - SMS service implementation
      - `mms.go` - MMS service implementation
    - `email/` - Email functionality
      - `models.go` - Email data models
      - `email.go` - Email service implementation
    - `contact/` - Contact management
      - `contact.go` - Contact service (opt-out)
    - `contactvalidator/` - Email/phone validation
      - `service.go` - Contact validator service
      - `models.go` - Validation models
    - `brands/` - Brand registration (TCR)
      - `brands.go` - Brand service implementation
    - `campaigns/` - Campaign registration (TCR)
      - `campaigns.go` - Campaign service implementation
    - `webhook/` - Webhook functionality
      - `service.go` - Webhook CRUD service (used by `ccai.Client.Webhook`)
      - `types.go` - Webhook type definitions
      - `handler.go` - `CreateHandler` HTTP handler helper
  - `examples/` - Example usage
- `.env` - Environment variables
- `.env.example` - Environment variables template

> Note: import services directly from `github.com/cloudcontactai/ccai-go/src/pkg/...` (e.g. `.../src/pkg/webhook`, `.../src/pkg/brands`) — the root `src/index.go` wrapper only re-exports the `ccai`, `sms`, and `email` types.

## Features

- Send SMS messages to single or multiple recipients
- Send MMS messages with images (automatic S3 upload)
- Send Email campaigns with HTML content
- Brand registration and management for TCR verification
- Campaign registration and management for TCR carrier vetting
- Manage contact opt-out preferences (SetDoNotText)
- Validate email addresses (valid/invalid/risky) and phone numbers (valid/invalid/landline)
- Webhook management: register, list, update, delete
- Webhook signature verification (HMAC-SHA256)
- Template variable substitution (`${firstName}`, `${lastName}`)
- Progress tracking via callbacks
- Environment variable support with .env files
- Comprehensive error handling
- Full test coverage

## License

MIT © 2025 CloudContactAI LLC
