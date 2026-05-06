// Copyright (c) 2025 CloudContactAI LLC
// Licensed under the MIT License. See LICENSE in the project root for license information.

// Package brands provides functionality for managing 10DLC brands.
package brands

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

// ClientInterface defines the interface that the CCAI client must implement.
type ClientInterface interface {
	CustomRequest(method, endpoint string, data interface{}, customBaseURL string, headers map[string]string) ([]byte, error)
	GetComplianceBaseURL() string
}

// Service provides brand management operations.
type Service struct {
	client ClientInterface
}

// NewService creates a new brands service.
func NewService(client ClientInterface) *Service {
	return &Service{client: client}
}

var (
	entityTypes   = map[string]bool{"PRIVATE_PROFIT": true, "PUBLIC_PROFIT": true, "NON_PROFIT": true, "GOVERNMENT": true, "SOLE_PROPRIETOR": true}
	verticalTypes = map[string]bool{
		"AUTOMOTIVE": true, "AGRICULTURE": true, "BANKING": true, "COMMUNICATION": true,
		"CONSTRUCTION": true, "EDUCATION": true, "ENERGY": true, "ENTERTAINMENT": true,
		"GOVERNMENT": true, "HEALTHCARE": true, "HOSPITALITY": true, "INSURANCE": true,
		"LEGAL": true, "MANUFACTURING": true, "NON_PROFIT": true, "PROFESSIONAL": true,
		"REAL_ESTATE": true, "RETAIL": true, "TECHNOLOGY": true, "TRANSPORTATION": true,
	}
	taxIdCountries = map[string]bool{"US": true, "CA": true, "GB": true, "AU": true}
	stockExchanges = map[string]bool{"NASDAQ": true, "NYSE": true, "AMEX": true, "TSX": true, "LON": true, "JPX": true, "HKEX": true, "OTHER": true}
	emailRegex     = regexp.MustCompile(`^[^\s@]+@[^\s@]+\.[^\s@]+$`)
)

// BrandRequest represents the payload for creating or updating a brand.
// All fields are pointers so partial updates only send the provided fields.
type BrandRequest struct {
	LegalCompanyName *string `json:"legalCompanyName,omitempty"`
	Dba              *string `json:"dba,omitempty"`
	EntityType       *string `json:"entityType,omitempty"`
	TaxId            *string `json:"taxId,omitempty"`
	TaxIdCountry     *string `json:"taxIdCountry,omitempty"`
	Country          *string `json:"country,omitempty"`
	VerticalType     *string `json:"verticalType,omitempty"`
	WebsiteUrl       *string `json:"websiteUrl,omitempty"`
	StockSymbol      *string `json:"stockSymbol,omitempty"`
	StockExchange    *string `json:"stockExchange,omitempty"`
	Street           *string `json:"street,omitempty"`
	City             *string `json:"city,omitempty"`
	State            *string `json:"state,omitempty"`
	PostalCode       *string `json:"postalCode,omitempty"`
	ContactFirstName *string `json:"contactFirstName,omitempty"`
	ContactLastName  *string `json:"contactLastName,omitempty"`
	ContactEmail     *string `json:"contactEmail,omitempty"`
	ContactPhone     *string `json:"contactPhone,omitempty"`
	WebsiteMatch     *bool   `json:"websiteMatch,omitempty"`
}

// BrandResponse represents the API response for brand operations.
type BrandResponse struct {
	ID                int64   `json:"id"`
	AccountID         int64   `json:"accountId,omitempty"`
	LegalCompanyName  string  `json:"legalCompanyName,omitempty"`
	Dba               string  `json:"dba,omitempty"`
	EntityType        string  `json:"entityType,omitempty"`
	TaxId             string  `json:"taxId,omitempty"`
	TaxIdCountry      string  `json:"taxIdCountry,omitempty"`
	Country           string  `json:"country,omitempty"`
	VerticalType      string  `json:"verticalType,omitempty"`
	WebsiteUrl        string  `json:"websiteUrl,omitempty"`
	StockSymbol       string  `json:"stockSymbol,omitempty"`
	StockExchange     string  `json:"stockExchange,omitempty"`
	Street            string  `json:"street,omitempty"`
	City              string  `json:"city,omitempty"`
	State             string  `json:"state,omitempty"`
	PostalCode        string  `json:"postalCode,omitempty"`
	ContactFirstName  string  `json:"contactFirstName,omitempty"`
	ContactLastName   string  `json:"contactLastName,omitempty"`
	ContactEmail      string  `json:"contactEmail,omitempty"`
	ContactPhone      string  `json:"contactPhone,omitempty"`
	WebsiteMatchScore *int    `json:"websiteMatchScore,omitempty"`
	CreatedAt         string  `json:"createdAt,omitempty"`
	UpdatedAt         string  `json:"updatedAt,omitempty"`
}

// Create creates a new brand.
func (s *Service) Create(req BrandRequest) (*BrandResponse, error) {
	if err := validateBrandRequest(req, true); err != nil {
		return nil, err
	}

	body, err := s.client.CustomRequest("POST", "/v1/brands", req, s.client.GetComplianceBaseURL(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create brand: %w", err)
	}

	var resp BrandResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse brand response: %w", err)
	}
	return &resp, nil
}

// Get retrieves a brand by ID.
func (s *Service) Get(id int64) (*BrandResponse, error) {
	if id == 0 {
		return nil, fmt.Errorf("id is required")
	}

	body, err := s.client.CustomRequest("GET", fmt.Sprintf("/v1/brands/%d", id), nil, s.client.GetComplianceBaseURL(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get brand: %w", err)
	}

	var resp BrandResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse brand response: %w", err)
	}
	return &resp, nil
}

// List retrieves all brands.
func (s *Service) List() ([]BrandResponse, error) {
	body, err := s.client.CustomRequest("GET", "/v1/brands", nil, s.client.GetComplianceBaseURL(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list brands: %w", err)
	}

	var resp []BrandResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse brand list response: %w", err)
	}
	return resp, nil
}

// Update updates an existing brand (partial update via PATCH).
func (s *Service) Update(id int64, req BrandRequest) (*BrandResponse, error) {
	if id == 0 {
		return nil, fmt.Errorf("id is required")
	}
	if err := validateBrandRequest(req, false); err != nil {
		return nil, err
	}

	body, err := s.client.CustomRequest("PATCH", fmt.Sprintf("/v1/brands/%d", id), req, s.client.GetComplianceBaseURL(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to update brand: %w", err)
	}

	var resp BrandResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse brand response: %w", err)
	}
	return &resp, nil
}

// Delete deletes a brand by ID.
func (s *Service) Delete(id int64) error {
	if id == 0 {
		return fmt.Errorf("id is required")
	}

	_, err := s.client.CustomRequest("DELETE", fmt.Sprintf("/v1/brands/%d", id), nil, s.client.GetComplianceBaseURL(), nil)
	if err != nil {
		return fmt.Errorf("failed to delete brand: %w", err)
	}
	return nil
}

func derefStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func validateBrandRequest(req BrandRequest, isCreate bool) error {
	var errors []string

	if isCreate {
		if derefStr(req.LegalCompanyName) == "" {
			errors = append(errors, "legalCompanyName is required")
		}
		if derefStr(req.EntityType) == "" {
			errors = append(errors, "entityType is required")
		}
		if derefStr(req.TaxId) == "" {
			errors = append(errors, "taxId is required")
		}
		if derefStr(req.TaxIdCountry) == "" {
			errors = append(errors, "taxIdCountry is required")
		}
		if derefStr(req.Country) == "" {
			errors = append(errors, "country is required")
		}
		if derefStr(req.VerticalType) == "" {
			errors = append(errors, "verticalType is required")
		}
		if derefStr(req.WebsiteUrl) == "" {
			errors = append(errors, "websiteUrl is required")
		}
		if derefStr(req.Street) == "" {
			errors = append(errors, "street is required")
		}
		if derefStr(req.City) == "" {
			errors = append(errors, "city is required")
		}
		if derefStr(req.State) == "" {
			errors = append(errors, "state is required")
		}
		if derefStr(req.PostalCode) == "" {
			errors = append(errors, "postalCode is required")
		}
		if derefStr(req.ContactFirstName) == "" {
			errors = append(errors, "contactFirstName is required")
		}
		if derefStr(req.ContactLastName) == "" {
			errors = append(errors, "contactLastName is required")
		}
		if derefStr(req.ContactEmail) == "" {
			errors = append(errors, "contactEmail is required")
		}
		if derefStr(req.ContactPhone) == "" {
			errors = append(errors, "contactPhone is required")
		}
	}

	if et := derefStr(req.EntityType); et != "" && !entityTypes[et] {
		errors = append(errors, "Invalid entity type")
	}
	if vt := derefStr(req.VerticalType); vt != "" && !verticalTypes[vt] {
		errors = append(errors, "Invalid vertical type")
	}
	if tc := derefStr(req.TaxIdCountry); tc != "" && !taxIdCountries[tc] {
		errors = append(errors, "Invalid tax ID country")
	}
	if se := derefStr(req.StockExchange); se != "" && !stockExchanges[se] {
		errors = append(errors, "Invalid stock exchange")
	}

	if wu := derefStr(req.WebsiteUrl); wu != "" {
		if !strings.HasPrefix(wu, "http://") && !strings.HasPrefix(wu, "https://") {
			errors = append(errors, "Website URL must start with http:// or https://")
		}
	}
	if ce := derefStr(req.ContactEmail); ce != "" {
		if !emailRegex.MatchString(ce) {
			errors = append(errors, "Invalid email format")
		}
	}

	taxId := derefStr(req.TaxId)
	taxIdCountry := derefStr(req.TaxIdCountry)
	if taxId != "" && (taxIdCountry == "US" || taxIdCountry == "CA") {
		digits := regexp.MustCompile(`\D`).ReplaceAllString(taxId, "")
		if len(digits) != 9 {
			errors = append(errors, fmt.Sprintf("Tax ID must be exactly 9 digits for %s", taxIdCountry))
		}
	}

	if derefStr(req.EntityType) == "PUBLIC_PROFIT" {
		if derefStr(req.StockSymbol) == "" {
			errors = append(errors, "Stock symbol is required for PUBLIC_PROFIT entities")
		}
		if derefStr(req.StockExchange) == "" {
			errors = append(errors, "Stock exchange is required for PUBLIC_PROFIT entities")
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("Brand validation failed: %s", strings.Join(errors, ", "))
	}
	return nil
}
