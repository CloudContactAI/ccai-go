// Copyright (c) 2025 CloudContactAI LLC
// Licensed under the MIT License. See LICENSE in the project root for license information.

// Package ccai provides a client for interacting with the Cloud Contact AI API.
package ccai

import (
	"github.com/cloudcontactai/ccai-go/pkg/sms"
)

// Version is the current version of the CCAI Go client.
const Version = "1.0.0"

// Account represents a recipient account.
type Account = sms.Account

// Response represents the response from the SMS API.
type Response = sms.Response

// Options represents options for SMS operations.
type Options = sms.Options

// SignedURLResponse represents the response from the signed URL API.
type SignedURLResponse = sms.SignedURLResponse
