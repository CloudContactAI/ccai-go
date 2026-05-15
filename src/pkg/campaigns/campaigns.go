// Copyright (c) 2025 CloudContactAI LLC
// Licensed under the MIT License. See LICENSE in the project root for license information.

// Package campaigns provides functionality for managing 10DLC campaigns.
package campaigns

import (
	"encoding/json"
	"fmt"
	"strings"
)

// ClientInterface defines the interface that the CCAI client must implement.
type ClientInterface interface {
	CustomRequest(method, endpoint string, data interface{}, customBaseURL string, headers map[string]string) ([]byte, error)
	GetComplianceBaseURL() string
}

// Service provides campaign management operations.
type Service struct {
	client ClientInterface
}

// NewService creates a new campaigns service.
func NewService(client ClientInterface) *Service {
	return &Service{client: client}
}

var (
	campaignUseCases = map[string]bool{
		"TWO_FACTOR_AUTHENTICATION": true, "ACCOUNT_NOTIFICATION": true, "CUSTOMER_CARE": true,
		"DELIVERY_NOTIFICATION": true, "FRAUD_ALERT": true, "HIGHER_EDUCATION": true,
		"LOW_VOLUME_MIXED": true, "MARKETING": true, "MIXED": true,
		"POLLING_VOTING": true, "PUBLIC_SERVICE_ANNOUNCEMENT": true, "SECURITY_ALERT": true,
	}
	campaignSubUseCases = map[string]bool{
		"TWO_FACTOR_AUTHENTICATION": true, "ACCOUNT_NOTIFICATION": true, "CUSTOMER_CARE": true,
		"DELIVERY_NOTIFICATION": true, "FRAUD_ALERT": true, "MARKETING": true, "POLLING_VOTING": true,
	}
	mixedUseCases = map[string]bool{"MIXED": true, "LOW_VOLUME_MIXED": true}
)

// CampaignRequest represents the payload for creating or updating a campaign.
// Boolean fields are pointers to allow partial updates (nil = not provided).
type CampaignRequest struct {
	BrandID          int64    `json:"brandId,omitempty"`
	UseCase          string   `json:"useCase,omitempty"`
	SubUseCases      []string `json:"subUseCases,omitempty"`
	Description      string   `json:"description,omitempty"`
	MessageFlow      string   `json:"messageFlow,omitempty"`
	TermsLink        string   `json:"termsLink,omitempty"`
	PrivacyLink      string   `json:"privacyLink,omitempty"`
	HasEmbeddedLinks *bool    `json:"hasEmbeddedLinks,omitempty"`
	HasEmbeddedPhone *bool    `json:"hasEmbeddedPhone,omitempty"`
	IsAgeGated       *bool    `json:"isAgeGated,omitempty"`
	IsDirectLending  *bool    `json:"isDirectLending,omitempty"`
	OptInKeywords    []string `json:"optInKeywords,omitempty"`
	OptInMessage     string   `json:"optInMessage,omitempty"`
	OptInProofUrl    string   `json:"optInProofUrl,omitempty"`
	HelpKeywords     []string `json:"helpKeywords,omitempty"`
	HelpMessage      string   `json:"helpMessage,omitempty"`
	OptOutKeywords   []string `json:"optOutKeywords,omitempty"`
	OptOutMessage    string   `json:"optOutMessage,omitempty"`
	SampleMessages   []string `json:"sampleMessages,omitempty"`
}

// CampaignResponse represents the API response for campaign operations.
type CampaignResponse struct {
	ID               int64    `json:"id"`
	AccountID        int64    `json:"accountId,omitempty"`
	BrandID          int64    `json:"brandId,omitempty"`
	UseCase          string   `json:"useCase,omitempty"`
	SubUseCases      []string `json:"subUseCases,omitempty"`
	Description      string   `json:"description,omitempty"`
	MessageFlow      string   `json:"messageFlow,omitempty"`
	TermsLink        string   `json:"termsLink,omitempty"`
	PrivacyLink      string   `json:"privacyLink,omitempty"`
	HasEmbeddedLinks bool     `json:"hasEmbeddedLinks,omitempty"`
	HasEmbeddedPhone bool     `json:"hasEmbeddedPhone,omitempty"`
	IsAgeGated       bool     `json:"isAgeGated,omitempty"`
	IsDirectLending  bool     `json:"isDirectLending,omitempty"`
	OptInKeywords    []string `json:"optInKeywords,omitempty"`
	OptInMessage     string   `json:"optInMessage,omitempty"`
	OptInProofUrl    string   `json:"optInProofUrl,omitempty"`
	HelpKeywords     []string `json:"helpKeywords,omitempty"`
	HelpMessage      string   `json:"helpMessage,omitempty"`
	OptOutKeywords   []string `json:"optOutKeywords,omitempty"`
	OptOutMessage    string   `json:"optOutMessage,omitempty"`
	SampleMessages   []string `json:"sampleMessages,omitempty"`
	MonthlyFee       float64  `json:"monthlyFee,omitempty"`
	CreatedAt        string   `json:"createdAt,omitempty"`
	UpdatedAt        string   `json:"updatedAt,omitempty"`
}

// Create creates a new campaign.
func (s *Service) Create(req CampaignRequest) (*CampaignResponse, error) {
	if err := validateCampaignRequest(req, true); err != nil {
		return nil, err
	}

	body, err := s.client.CustomRequest("POST", "/v1/campaigns", req, s.client.GetComplianceBaseURL(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create campaign: %w", err)
	}

	var resp CampaignResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse campaign response: %w", err)
	}
	return &resp, nil
}

// Get retrieves a campaign by ID.
func (s *Service) Get(id int64) (*CampaignResponse, error) {
	if id == 0 {
		return nil, fmt.Errorf("id is required")
	}

	body, err := s.client.CustomRequest("GET", fmt.Sprintf("/v1/campaigns/%d", id), nil, s.client.GetComplianceBaseURL(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get campaign: %w", err)
	}

	var resp CampaignResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse campaign response: %w", err)
	}
	return &resp, nil
}

// List retrieves all campaigns.
func (s *Service) List() ([]CampaignResponse, error) {
	body, err := s.client.CustomRequest("GET", "/v1/campaigns", nil, s.client.GetComplianceBaseURL(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list campaigns: %w", err)
	}

	var resp []CampaignResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse campaign list response: %w", err)
	}
	return resp, nil
}

// Update updates an existing campaign (partial update via PATCH).
func (s *Service) Update(id int64, req CampaignRequest) (*CampaignResponse, error) {
	if id == 0 {
		return nil, fmt.Errorf("id is required")
	}
	if err := validateCampaignRequest(req, false); err != nil {
		return nil, err
	}

	body, err := s.client.CustomRequest("PATCH", fmt.Sprintf("/v1/campaigns/%d", id), req, s.client.GetComplianceBaseURL(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to update campaign: %w", err)
	}

	var resp CampaignResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse campaign response: %w", err)
	}
	return &resp, nil
}

// Delete deletes a campaign by ID.
func (s *Service) Delete(id int64) error {
	if id == 0 {
		return fmt.Errorf("id is required")
	}

	_, err := s.client.CustomRequest("DELETE", fmt.Sprintf("/v1/campaigns/%d", id), nil, s.client.GetComplianceBaseURL(), nil)
	if err != nil {
		return fmt.Errorf("failed to delete campaign: %w", err)
	}
	return nil
}

func validateCampaignRequest(req CampaignRequest, isCreate bool) error {
	var errors []string

	if isCreate {
		if req.BrandID == 0 {
			errors = append(errors, "brandId is required")
		}
		if req.UseCase == "" {
			errors = append(errors, "useCase is required")
		}
		if req.Description == "" {
			errors = append(errors, "description is required")
		}
		if req.MessageFlow == "" {
			errors = append(errors, "messageFlow is required")
		}
		if req.HasEmbeddedLinks == nil {
			errors = append(errors, "hasEmbeddedLinks is required")
		}
		if req.HasEmbeddedPhone == nil {
			errors = append(errors, "hasEmbeddedPhone is required")
		}
		if req.IsAgeGated == nil {
			errors = append(errors, "isAgeGated is required")
		}
		if req.IsDirectLending == nil {
			errors = append(errors, "isDirectLending is required")
		}
		if len(req.OptInKeywords) == 0 {
			errors = append(errors, "optInKeywords is required")
		}
		if req.OptInMessage == "" {
			errors = append(errors, "optInMessage is required")
		}
		if req.OptInProofUrl == "" {
			errors = append(errors, "optInProofUrl is required")
		}
		if len(req.HelpKeywords) == 0 {
			errors = append(errors, "helpKeywords is required")
		}
		if req.HelpMessage == "" {
			errors = append(errors, "helpMessage is required")
		}
		if len(req.OptOutKeywords) == 0 {
			errors = append(errors, "optOutKeywords is required")
		}
		if req.OptOutMessage == "" {
			errors = append(errors, "optOutMessage is required")
		}
		if len(req.SampleMessages) == 0 {
			errors = append(errors, "sampleMessages is required")
		}
	}

	if req.UseCase != "" && !campaignUseCases[req.UseCase] {
		errors = append(errors, "Invalid use case")
	}

	if req.UseCase != "" && mixedUseCases[req.UseCase] {
		if len(req.SubUseCases) < 2 || len(req.SubUseCases) > 3 {
			errors = append(errors, "MIXED/LOW_VOLUME_MIXED requires 2-3 sub use cases")
		} else {
			for _, sub := range req.SubUseCases {
				if !campaignSubUseCases[sub] {
					errors = append(errors, fmt.Sprintf("Invalid sub use case: %s", sub))
				}
			}
		}
	} else if req.UseCase != "" && len(req.SubUseCases) > 0 {
		errors = append(errors, "subUseCases should be empty for non-MIXED use cases")
	}

	if len(req.SampleMessages) > 0 {
		msgs := req.SampleMessages
		if len(msgs) < 2 || len(msgs) > 5 {
			errors = append(errors, "sampleMessages must contain 2-5 items")
		} else {
			optOutKws := req.OptOutKeywords
			helpKws := req.HelpKeywords

			hasOptOut := false
			for _, msg := range msgs {
				if strings.Contains(msg, "Reply STOP") {
					hasOptOut = true
					break
				}
				for _, kw := range optOutKws {
					if strings.Contains(msg, "Reply "+kw) {
						hasOptOut = true
						break
					}
				}
				if hasOptOut {
					break
				}
			}
			if !hasOptOut {
				errors = append(errors, "At least one sample must contain 'Reply STOP' or 'Reply {optOutKeyword}'")
			}

			hasHelp := false
			for _, msg := range msgs {
				if strings.Contains(msg, "Reply HELP") {
					hasHelp = true
					break
				}
				for _, kw := range helpKws {
					if strings.Contains(msg, "Reply "+kw) {
						hasHelp = true
						break
					}
				}
				if hasHelp {
					break
				}
			}
			if !hasHelp {
				errors = append(errors, "At least one sample must contain 'Reply HELP' or 'Reply {helpKeyword}'")
			}
		}
	}

	if req.OptOutMessage != "" {
		msg := req.OptOutMessage
		hasKw := strings.Contains(msg, "STOP")
		if !hasKw {
			for _, kw := range req.OptOutKeywords {
				if strings.Contains(msg, kw) {
					hasKw = true
					break
				}
			}
		}
		if !hasKw {
			errors = append(errors, "optOutMessage must contain 'STOP' or at least one optOutKeyword")
		}
	}

	if req.HelpMessage != "" {
		msg := req.HelpMessage
		hasKw := strings.Contains(msg, "HELP")
		if !hasKw {
			for _, kw := range req.HelpKeywords {
				if strings.Contains(msg, kw) {
					hasKw = true
					break
				}
			}
		}
		if !hasKw {
			errors = append(errors, "helpMessage must contain 'HELP' or at least one helpKeyword")
		}
	}

	if req.OptInProofUrl != "" {
		if !strings.HasPrefix(req.OptInProofUrl, "http://") && !strings.HasPrefix(req.OptInProofUrl, "https://") {
			errors = append(errors, "Opt-in proof URL must start with http:// or https://")
		}
	}
	if req.TermsLink != "" {
		if !strings.HasPrefix(req.TermsLink, "http://") && !strings.HasPrefix(req.TermsLink, "https://") {
			errors = append(errors, "Terms link must start with http:// or https://")
		}
	}
	if req.PrivacyLink != "" {
		if !strings.HasPrefix(req.PrivacyLink, "http://") && !strings.HasPrefix(req.PrivacyLink, "https://") {
			errors = append(errors, "Privacy link must start with http:// or https://")
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("Campaign validation failed: %s", strings.Join(errors, ", "))
	}
	return nil
}
