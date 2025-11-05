package yalo

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	retryablehttp "github.com/hashicorp/go-retryablehttp"
)

// Client represents a Yalo API client
type Client struct {
	BaseURL         string
	AccountID       string
	BotID           string
	Token           string
	Debug           bool
	RetryableClient *retryablehttp.Client
	rateLimiter     <-chan time.Time // Rate limiter channel (40 requests per second)
}

// ClientOption is a function that configures a Client
type ClientOption func(*Client)

// WithBaseURL sets the base URL for the Yalo API
func WithBaseURL(baseURL string) ClientOption {
	return func(c *Client) {
		c.BaseURL = baseURL
	}
}

// WithAccount sets the account ID and bot ID for the Yalo API
func WithAccount(accountID, botID string) ClientOption {
	return func(c *Client) {
		c.AccountID = accountID
		c.BotID = botID
	}
}

// WithToken sets the bearer token for authentication
func WithToken(token string) ClientOption {
	return func(c *Client) {
		c.Token = token
	}
}

// WithRetryableClient allows injecting a custom retryable HTTP client
func WithRetryableClient(client *retryablehttp.Client) ClientOption {
	return func(c *Client) {
		c.RetryableClient = client
	}
}

// WithDebug enables debug mode to print raw requests and responses
func WithDebug(debug bool) ClientOption {
	return func(c *Client) {
		c.Debug = debug
	}
}

// NewClient creates a new Yalo client with the provided options
func NewClient(opts ...ClientOption) *Client {
	// Create default retryable client
	retryClient := retryablehttp.NewClient()
	retryClient.RetryMax = 3
	retryClient.Logger = nil // Disable default logging

	// Create rate limiter: 40 requests per second as per Yalo's rate limit
	rateLimiter := time.Tick(time.Second / 40)

	client := &Client{
		BaseURL:         "https://api-global.yalochat.com",
		RetryableClient: retryClient,
		rateLimiter:     rateLimiter,
	}

	// Apply all options
	for _, opt := range opts {
		opt(client)
	}

	return client
}

// Response represents the response from Yalo
type Response struct {
	JSONData   string
	StatusCode int
	Headers    http.Header
}

// SendRequest sends a JSON request to a specific Yalo API endpoint with automatic retries
func (c *Client) SendRequest(ctx context.Context, endpoint string, jsonRequest string) (*Response, error) {
	// Wait for rate limiter (40 requests per second)
	<-c.rateLimiter

	// Construct full URL
	fullURL := c.BaseURL + endpoint

	// Debug: Print request details
	if c.Debug {
		log.Printf("[DEBUG] Request URL: %s", fullURL)
		log.Printf("[DEBUG] Request Body: %s", jsonRequest)
	}

	// Step 1: Prepare retryable HTTP request
	req, err := retryablehttp.NewRequest(http.MethodPost, fullURL, bytes.NewBufferString(jsonRequest))
	if err != nil {
		return nil, fmt.Errorf("error creating HTTP request: %w", err)
	}

	// Step 2: Set headers
	req.Header.Set("Content-Type", "application/json")

	// Add Bearer token authentication if provided
	if c.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.Token)
		if c.Debug {
			log.Printf("[DEBUG] Authorization: Bearer %s...", c.Token[:10])
		}
	}

	// Step 3: Send request with automatic retries
	resp, err := c.RetryableClient.Do(req.WithContext(ctx))
	if err != nil {
		return nil, fmt.Errorf("error sending request: %w", err)
	}
	defer resp.Body.Close()

	// Step 4: Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}

	jsonResponse := string(body)

	// Debug: Print response details
	if c.Debug {
		log.Printf("[DEBUG] Response Status: %d", resp.StatusCode)
		log.Printf("[DEBUG] Response Body: %s", jsonResponse)
	}

	// Check if response is actually JSON by trying to validate it
	// If it's not JSON (e.g., error page, HTML), return error
	var testJSON interface{}
	if err := json.Unmarshal(body, &testJSON); err != nil {
		return &Response{
			JSONData:   jsonResponse,
			StatusCode: resp.StatusCode,
			Headers:    resp.Header,
		}, fmt.Errorf("API returned non-JSON response (status %d): %s", resp.StatusCode, jsonResponse)
	}

	return &Response{
		JSONData:   jsonResponse,
		StatusCode: resp.StatusCode,
		Headers:    resp.Header,
	}, nil
}

// SendRequestWithPayload sends a request with a struct payload that will be marshaled to JSON
func (c *Client) SendRequestWithPayload(ctx context.Context, endpoint string, payload interface{}) (*Response, error) {
	jsonBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("error marshaling payload to JSON: %w", err)
	}
	return c.SendRequest(ctx, endpoint, string(jsonBytes))
}

// ParseResponse unmarshals the JSON response into the provided struct
func (r *Response) ParseResponse(v interface{}) error {
	return json.Unmarshal([]byte(r.JSONData), v)
}

// NotificationRequest represents a Yalo notification request
type NotificationRequest struct {
	Type  string `json:"type"`
	Users []User `json:"users"`
}

type User struct {
	Priority string                 `json:"priority,omitempty"`
	Phone    string                 `json:"phone"`
	Params   map[string]interface{} `json:"params"`
}

// NotificationOption configures a notification user payload
type NotificationOption func(*User)

// WithPriority sets the notification priority for the user payload
func WithPriority(priority string) NotificationOption {
	return func(u *User) {
		u.Priority = priority
	}
}

// NotificationResponse represents the response payload currently returned by Yalo.
type NotificationResponse struct {
	Success    bool     `json:"success"`
	ID         string   `json:"id"`
	MessageIDs []string `json:"message_ids"`
}

// SendNotification sends a WhatsApp notification via Yalo to the specified users
func (c *Client) SendNotification(ctx context.Context, notificationType, phone string, params map[string]interface{}, opts ...NotificationOption) (*NotificationResponse, error) {
	if c.AccountID == "" || c.BotID == "" {
		return nil, fmt.Errorf("accountID and botID are required")
	}

	endpoint := fmt.Sprintf("/notifications/api/v1/accounts/%s/bots/%s/notifications", c.AccountID, c.BotID)

	user := User{
		Priority: "1",
		Phone:    phone,
		Params:   params,
	}

	for _, opt := range opts {
		opt(&user)
	}

	payload := NotificationRequest{
		Type:  notificationType,
		Users: []User{user},
	}

	response, err := c.SendRequestWithPayload(ctx, endpoint, payload)
	if err != nil {
		return nil, err
	}

	var result NotificationResponse
	if err := response.ParseResponse(&result); err != nil {
		return nil, fmt.Errorf("error parsing response: %w", err)
	}

	return &result, nil
}
