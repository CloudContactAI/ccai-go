# Email Package

The Email package provides functionality for sending email campaigns through the Cloud Contact AI platform.

## Features

- Send single emails to individual recipients
- Send email campaigns to multiple recipients
- Progress tracking with callback functions
- Input validation for required fields
- Support for HTML email content
- Custom sender information

## Usage

### Basic Single Email

```go
package main

import (
    "fmt"
    "log"

    "github.com/cloudcontactai/ccai-go/pkg/ccai"
)

func main() {
    // Initialize client
    client, err := ccai.NewClient(ccai.Config{
        ClientID: "your-client-id",
        APIKey:   "your-api-key",
    })
    if err != nil {
        log.Fatal(err)
    }

    // Send single email
    response, err := client.Email.SendSingle(
        "John",                                // firstName
        "Doe",                                 // lastName
        "john@example.com",                    // email
        "Welcome to Our Service",              // subject
        "<p>Hello John, welcome!</p>",         // message (HTML)
        "noreply@yourcompany.com",             // senderEmail
        "support@yourcompany.com",             // replyEmail
        "Your Company",                        // senderName
        "Welcome Email",                       // title
        nil,                                   // options
    )
    
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Email sent: %+v\n", response)
}
```

### Email Campaign

```go
// Create recipients
accounts := []ccai.EmailAccount{
    {
        FirstName: "John",
        LastName:  "Doe",
        Email:     "john@example.com",
    },
    {
        FirstName: "Jane",
        LastName:  "Smith",
        Email:     "jane@example.com",
    },
}

// Create campaign
campaign := &ccai.EmailCampaign{
    Subject:      "Monthly Newsletter",
    Title:        "July 2025 Newsletter",
    Message:      "<h1>Newsletter</h1><p>Hello ${firstName}!</p>",
    SenderEmail:  "noreply@company.com",
    ReplyEmail:   "support@company.com",
    SenderName:   "Company Name",
    Accounts:     accounts,
    CampaignType: "EMAIL",
    AddToList:    "noList",
    ContactInput: "accounts",
    FromType:     "single",
    Senders:      []interface{}{},
}

// Send campaign with progress tracking
options := &ccai.EmailOptions{
    OnProgress: func(status string) {
        fmt.Printf("Progress: %s\n", status)
    },
}

response, err := client.Email.SendCampaign(campaign, options)
```

## API Endpoint

The email service uses the endpoint:
- **Base URL**: `https://email-campaigns.cloudcontactai.com/api/v1`
- **Endpoint**: `/campaigns`
- **Method**: `POST`

## Required Fields

### For Single Emails:
- firstName, lastName, email (recipient)
- subject, message, title
- senderEmail, replyEmail, senderName

### For Email Campaigns:
- All single email fields
- accounts array with at least one recipient
- campaignType (set to "EMAIL")
- addToList, contactInput, fromType

## Error Handling

The package validates all required fields and returns descriptive error messages:

```go
response, err := client.Email.SendSingle("", "Doe", "john@example.com", ...)
if err != nil {
    // Error: "first name is required for account at index 0"
    log.Fatal(err)
}
```

## Authentication

Authentication is handled automatically using the API key from the CCAI client configuration. The service uses Bearer token authentication with the Cloud Contact AI email campaigns API.