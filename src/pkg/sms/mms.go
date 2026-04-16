// Copyright (c) 2025 CloudContactAI LLC
// Licensed under the MIT License. See LICENSE in the project root for license information.

package sms

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// MMSService is the MMS service for sending multimedia messages through the CCAI API.
type MMSService struct {
	client ClientInterface
}

// StoredURLResponse represents the response from the stored URL check.
type StoredURLResponse struct {
	StoredURL string `json:"storedUrl"`
}

// NewMMSService creates a new MMS service instance.
func NewMMSService(client ClientInterface) *MMSService {
	return &MMSService{
		client: client,
	}
}

// GetSignedUploadURL gets a signed S3 URL to upload an image file.
func (m *MMSService) GetSignedUploadURL(fileName, fileType string, fileBasePath string, publicFile bool) (*SignedURLResponse, error) {
	if fileName == "" {
		return nil, fmt.Errorf("file name is required")
	}

	if fileType == "" {
		return nil, fmt.Errorf("file type is required")
	}

	// Use default fileBasePath if not provided
	if fileBasePath == "" {
		fileBasePath = fmt.Sprintf("%s/campaign", m.client.GetClientID())
	}

	// Define fileKey explicitly as clientId/campaign/filename
	fileKey := fmt.Sprintf("%s/campaign/%s", m.client.GetClientID(), fileName)

	data := map[string]interface{}{
		"fileName":     fileName,
		"fileType":     fileType,
		"fileBasePath": fileBasePath,
		"publicFile":   publicFile,
	}

	// Create a new HTTP client for this request
	httpClient := &http.Client{}

	// Marshal the data to JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request data: %w", err)
	}

	// Create the request
	req, err := http.NewRequest("POST", m.client.GetFilesBaseURL()+"/upload/url", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Authorization", "Bearer "+m.client.GetAPIKey())
	req.Header.Set("Content-Type", "application/json")

	// Send the request
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get signed upload URL: %w", err)
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("API error: %d - %s", resp.StatusCode, string(body))
	}

	// Parse the response
	var response SignedURLResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Override the fileKey with our explicitly defined one
	response.FileKey = fileKey

	return &response, nil
}

// UploadImageToSignedURL uploads an image file to a signed S3 URL.
func (m *MMSService) UploadImageToSignedURL(signedURL, filePath, contentType string) (bool, error) {
	if signedURL == "" {
		return false, fmt.Errorf("signed URL is required")
	}

	if filePath == "" {
		return false, fmt.Errorf("file path is required")
	}

	if contentType == "" {
		return false, fmt.Errorf("content type is required")
	}

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return false, fmt.Errorf("file does not exist: %s", filePath)
	}

	// Read the file
	fileContent, err := os.ReadFile(filePath)
	if err != nil {
		return false, fmt.Errorf("failed to read file: %w", err)
	}

	// Create a new HTTP client for this request
	httpClient := &http.Client{}

	// Create the request
	req, err := http.NewRequest("PUT", signedURL, bytes.NewBuffer(fileContent))
	if err != nil {
		return false, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", contentType)

	// Send the request
	resp, err := httpClient.Do(req)
	if err != nil {
		return false, fmt.Errorf("failed to upload file: %w", err)
	}
	defer resp.Body.Close()

	return resp.StatusCode >= 200 && resp.StatusCode < 300, nil
}

// Send sends an MMS message to one or more recipients.
func (m *MMSService) Send(pictureFileKey string, accounts []Account, message, title, senderPhone string, options *Options, forceNewCampaign bool) (*Response, error) {
	// Validate inputs
	if pictureFileKey == "" {
		return nil, fmt.Errorf("picture file key is required")
	}

	if len(accounts) == 0 {
		return nil, fmt.Errorf("at least one account is required")
	}

	if message == "" {
		return nil, fmt.Errorf("message is required")
	}

	if title == "" {
		return nil, fmt.Errorf("title is required")
	}

	// Create options if not provided
	if options == nil {
		options = &Options{}
	}

	// Notify progress if callback provided
	options.NotifyProgress("Preparing to send MMS")

	// Prepare the endpoint and data
	endpoint := fmt.Sprintf("/clients/%s/campaigns/direct", m.client.GetClientID())

	campaignData := MMSCampaign{
		PictureFileKey: pictureFileKey,
		Accounts:       accounts,
		Message:        message,
		Title:          title,
		SenderPhone:    senderPhone,
	}

	// Set up headers for force new campaign if needed
	headers := make(map[string]string)
	if forceNewCampaign {
		headers["ForceNewCampaign"] = "true"
	}

	// Notify progress if callback provided
	options.NotifyProgress("Sending MMS")

	// Make the API request
	responseBody, err := m.client.Request("POST", endpoint, campaignData, headers)
	if err != nil {
		// Notify progress if callback provided
		options.NotifyProgress("MMS sending failed")
		return nil, fmt.Errorf("failed to send MMS: %w", err)
	}

	// Parse the response
	var response Response
	if err := json.Unmarshal(responseBody, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Notify progress if callback provided
	options.NotifyProgress("MMS sent successfully")

	return &response, nil
}

// SendSingle sends a single MMS message to one recipient.
func (m *MMSService) SendSingle(pictureFileKey, firstName, lastName, phone, message, title, customData, senderPhone string, options *Options, forceNewCampaign bool) (*Response, error) {
	account := Account{
		FirstName:   firstName,
		LastName:    lastName,
		Phone:       phone,
		MessageData: customData,
	}

	return m.Send(pictureFileKey, []Account{account}, message, title, senderPhone, options, forceNewCampaign)
}

// SendWithImage completes the MMS workflow: check cache, optionally upload, and send MMS.
func (m *MMSService) SendWithImage(imagePath, contentType string, accounts []Account, message, title, senderPhone string, options *Options, forceNewCampaign bool) (*Response, error) {
	// Create options if not provided
	if options == nil {
		options = &Options{}
	}

	// Step 1: Compute MD5 of the image file for caching
	md5Hash, err := m.md5File(imagePath)
	if err != nil {
		return nil, fmt.Errorf("failed to compute MD5: %w", err)
	}

	extension := strings.ToLower(strings.TrimPrefix(filepath.Ext(imagePath), "."))
	fileName := fmt.Sprintf("%s.%s", md5Hash, extension)
	fileKey := fmt.Sprintf("%s/campaign/%s", m.client.GetClientID(), fileName)

	// Step 2: Check if the same image has already been uploaded
	options.NotifyProgress("Checking if image already uploaded")
	storedURLResp, err := m.CheckFileUploaded(fileKey)
	if err == nil && storedURLResp != nil && storedURLResp.StoredURL != "" {
		// Image already uploaded, skip upload and send directly
		options.NotifyProgress("Image already exists in S3, sending MMS")
		return m.Send(fileKey, accounts, message, title, senderPhone, options, forceNewCampaign)
	}

	// Step 3: Get a signed URL for uploading
	options.NotifyProgress("Getting signed upload URL")
	uploadResponse, err := m.GetSignedUploadURL(fileName, contentType, "", true)
	if err != nil {
		return nil, fmt.Errorf("failed to get signed upload URL: %w", err)
	}

	signedURL := uploadResponse.SignedS3URL

	// Step 4: Upload the image to the signed URL
	options.NotifyProgress("Uploading image to S3")
	uploadSuccess, err := m.UploadImageToSignedURL(signedURL, imagePath, contentType)
	if err != nil {
		return nil, fmt.Errorf("failed to upload image: %w", err)
	}

	if !uploadSuccess {
		return nil, fmt.Errorf("failed to upload image to S3")
	}

	// Step 5: Send the MMS with the uploaded image
	options.NotifyProgress("Image uploaded successfully, sending MMS")
	return m.Send(fileKey, accounts, message, title, senderPhone, options, forceNewCampaign)
}

// md5File calculates the MD5 hash of a file.
func (m *MMSService) md5File(filePath string) (string, error) {
	fileContent, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	hash := md5.Sum(fileContent)
	return hex.EncodeToString(hash[:]), nil
}

// CheckFileUploaded checks if a file has already been uploaded to S3.
func (m *MMSService) CheckFileUploaded(fileKey string) (*StoredURLResponse, error) {
	endpoint := fmt.Sprintf("/clients/%s/storedUrl?fileKey=%s", m.client.GetClientID(), fileKey)
	responseBody, err := m.client.Request("GET", endpoint, nil, nil)
	if err != nil {
		return &StoredURLResponse{StoredURL: ""}, nil
	}

	var response StoredURLResponse
	if err := json.Unmarshal(responseBody, &response); err != nil {
		return &StoredURLResponse{StoredURL: ""}, nil
	}

	return &response, nil
}

// MD5FileForTest exports md5File for testing purposes.
func (m *MMSService) MD5FileForTest(filePath string) (string, error) {
	return m.md5File(filePath)
}
