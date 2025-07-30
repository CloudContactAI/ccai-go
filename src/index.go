// Copyright (c) 2025 CloudContactAI LLC
// Licensed under the MIT License. See LICENSE in the project root for license information.

// Package main provides the main export for the CCAI Go module
package main

import (
	"github.com/cloudcontactai/ccai-go/src/pkg/ccai"
	"github.com/cloudcontactai/ccai-go/src/pkg/email"
	"github.com/cloudcontactai/ccai-go/src/pkg/sms"
)

// Main CCAI types
type (
	// Config represents the configuration for the CCAI client
	Config = ccai.Config

	// Client is the main client for interacting with the CloudContactAI API
	Client = ccai.Client

	// Account represents a recipient account
	Account = ccai.Account

	// SMS types
	Response          = sms.Response
	Options           = sms.Options
	SignedURLResponse = sms.SignedURLResponse

	// Email types
	EmailAccount  = email.EmailAccount
	EmailCampaign = email.EmailCampaign
	EmailResponse = email.EmailResponse
	EmailOptions  = email.EmailOptions

	// Webhook types can be added here when needed
)

// NewClient creates a new CCAI client instance
var NewClient = ccai.NewClient

// Version is the current version of the CCAI Go client
const Version = "1.0.0"
