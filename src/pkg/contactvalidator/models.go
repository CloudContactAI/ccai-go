// Copyright (c) 2025 CloudContactAI LLC
// Licensed under the MIT License. See LICENSE in the project root for license information.

// Package contactvalidator provides functionality for validating email addresses and phone numbers.
package contactvalidator

// EmailValidationResult represents the validation result for an email address.
type EmailValidationResult struct {
	Contact  string                 `json:"contact"`
	Type     string                 `json:"type"`
	Status   string                 `json:"status"`
	Metadata map[string]interface{} `json:"metadata"`
}

// PhoneValidationResult represents the validation result for a phone number.
type PhoneValidationResult struct {
	Contact  string                 `json:"contact"`
	Type     string                 `json:"type"`
	Status   string                 `json:"status"`
	Metadata map[string]interface{} `json:"metadata"`
}

// ValidationSummary holds aggregate counts for a bulk validation response.
type ValidationSummary struct {
	Total    int `json:"total"`
	Valid    int `json:"valid"`
	Invalid  int `json:"invalid"`
	Risky    int `json:"risky"`
	Landline int `json:"landline"`
}

// BulkEmailValidationResult represents the response for a bulk email validation request.
type BulkEmailValidationResult struct {
	Results []EmailValidationResult `json:"results"`
	Summary ValidationSummary       `json:"summary"`
}

// BulkPhoneValidationResult represents the response for a bulk phone validation request.
type BulkPhoneValidationResult struct {
	Results []PhoneValidationResult `json:"results"`
	Summary ValidationSummary       `json:"summary"`
}

// PhoneInput represents a phone number with an optional country code.
type PhoneInput struct {
	Phone       string `json:"phone"`
	CountryCode string `json:"countryCode,omitempty"`
}
