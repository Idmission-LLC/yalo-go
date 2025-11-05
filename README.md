# Yalo Go Client

A Go module for interfacing with the Yalo API to send WhatsApp messages and notifications.

## Features

- Direct JSON request/response handling
- Bearer token authentication
- Automatic retries with exponential backoff (using HashiCorp retryablehttp)
- Rate limiting (40 requests per second) to respect Yalo's API limits
- Configurable retry settings
- Send raw JSON strings or Go structs
- Parse JSON responses into Go structs
- Support for single and multiple user notifications
- Generic parameter support - works with any WhatsApp template configuration
- Debug mode for troubleshooting
- Notification options for customizing delivery priority
- Clean, simple API

## Installation

```bash
go get github.com/Idmission-LLC/yalo-go
```

## Configuration

The module uses Viper to load configuration from a `.env` file or environment variables.

Create a `.env` file in the project root:

```bash
YALO_ACCOUNT_ID=your-account-id
YALO_BOT_ID=your-bot-id
YALO_TOKEN=your-bearer-token
YALO_NOTIFICATION_TYPE=your-notification-type
YALO_BASE_URL=https://api-global.yalochat.com  # Optional, defaults to this value
YALO_DEBUG=true  # Optional, enables debug logging of requests and responses
```

Note: The `YALO_BASE_URL` should be set to the base URL only. Specific endpoints are handled by the client methods.

## Usage

### Basic Example

```go
package main

import (
    "context"
    "fmt"
    "log"

    yalo "github.com/Idmission-LLC/yalo-go"
)

func main() {
    // Create a new client
    client := yalo.NewClient(
        yalo.WithAccount("your-account-id", "your-bot-id"),
        yalo.WithToken("your-bearer-token"),
    )

    // Create notification parameters
    // These parameters should match your WhatsApp template configuration
    params := map[string]interface{}{
        "name":   "John Doe",
        "param1": "value1",
        "param2": "value2",
    }

    // Send notification
    result, err := client.SendNotification(
        context.Background(),
        "your-template-name",
        "+1234567890",
        params,
        // Optionally override the default priority of "1"
        yalo.WithPriority("0"),
    )
    if err != nil {
        log.Fatalf("Error: %v", err)
    }

    fmt.Printf("Status: %s\n", result.Status)
    fmt.Printf("Message: %s\n", result.Message)
}
```

**Note:** The `params` map should contain key-value pairs that match the variables defined in your WhatsApp template. The template name (`type`) should be the name of a template authorized by WhatsApp in your Yalo account.

### Complex Parameters (Buttons, Arrays, Nested Objects)

The params map supports any data structure that can be marshaled to JSON:

```go
params := map[string]interface{}{
    "name": "Jane Smith",
    "link": "https://example.com/track",
    "buttons": []map[string]interface{}{
        {
            "sub_type": "url",
            "index":    0,
            "parameters": []map[string]interface{}{
                {"text": "/track/12345"},
            },
        },
    },
    "service_phone": "+1800555000",
    "custom_field":  "any value",
}

result, err := client.SendNotification(
    context.Background(),
    "your-template-name",
    "+1234567890",
    params,
    yalo.WithPriority("0"),
)
```

### Sending to Multiple Users

If you need to send notifications to multiple users, simply call `SendNotification` multiple times. The client automatically handles rate limiting (40 requests/second):

```go
phoneNumbers := []string{"+1234567890", "+1234567891", "+1234567892"}

for _, phone := range phoneNumbers {
    params := map[string]interface{}{
        "name":   "User",
        "param1": "value1",
    }

    result, err := client.SendNotification(
        context.Background(),
        "your-template-name",
        phone,
        params,
    )
    if err != nil {
        log.Printf("Error sending to %s: %v", phone, err)
        continue
    }

    fmt.Printf("Sent to %s: %s\n", phone, result.Status)
}
```

### Debug Mode

Enable debug mode to see the raw requests and responses:

```go
client := yalo.NewClient(
    yalo.WithAccount("account-id", "bot-id"),
    yalo.WithToken("bearer-token"),
    yalo.WithDebug(true),  // Enable debug logging
)
```

When debug mode is enabled, the client will log:
- Request URL
- Request body (JSON payload)
- Authorization header (first 10 characters only)
- Response status code
- Response body (raw JSON)

### Rate Limiting

The client automatically enforces Yalo's rate limit of **40 requests per second**. If you make rapid successive calls, the client will automatically throttle requests to stay within this limit. This is handled transparently - you don't need to implement any rate limiting yourself.

### Custom Retry Configuration

You can customize the retry behavior by injecting your own retryable HTTP client:

```go
import (
    "time"
    retryablehttp "github.com/hashicorp/go-retryablehttp"
    yalo "github.com/Idmission-LLC/yalo-go"
)

// Create custom retryable client
retryClient := retryablehttp.NewClient()
retryClient.RetryMax = 5                      // Max number of retries
retryClient.RetryWaitMin = 1 * time.Second   // Min wait between retries
retryClient.RetryWaitMax = 30 * time.Second  // Max wait between retries
retryClient.Logger = someLogger               // Custom logger

// Inject the custom client
client := yalo.NewClient(
    yalo.WithAccount("account-id", "bot-id"),
    yalo.WithToken("bearer-token"),
    yalo.WithRetryableClient(retryClient),
)
```

## API Reference

### Client

```go
type Client struct {
    BaseURL         string
    AccountID       string
    BotID           string
    Token           string
    RetryableClient *retryablehttp.Client
}
```

### NewClient

```go
func NewClient(opts ...ClientOption) *Client
```

Creates a new Yalo client with the provided options.

### Client Options

```go
func WithBaseURL(baseURL string) ClientOption
func WithAccount(accountID, botID string) ClientOption
func WithToken(token string) ClientOption
func WithDebug(debug bool) ClientOption
func WithRetryableClient(client *retryablehttp.Client) ClientOption
```

- `WithBaseURL`: Sets the base URL for the Yalo API (defaults to `https://api-global.yalochat.com`)
- `WithAccount`: Sets the account ID and bot ID for the Yalo API
- `WithToken`: Sets the bearer token for authentication
- `WithDebug`: Enables debug mode to print raw requests and responses
- `WithRetryableClient`: Injects a custom retryable HTTP client (overrides default)

### SendRequest

```go
func (c *Client) SendRequest(ctx context.Context, endpoint string, jsonRequest string) (*Response, error)
```

Sends a JSON string request to a specific Yalo API endpoint and returns the response.

### SendRequestWithPayload

```go
func (c *Client) SendRequestWithPayload(ctx context.Context, endpoint string, payload interface{}) (*Response, error)
```

Marshals a Go struct to JSON and sends it to a specific Yalo API endpoint.

### SendNotification

```go
func (c *Client) SendNotification(ctx context.Context, notificationType, phone string, params map[string]interface{}, opts ...NotificationOption) (*NotificationResponse, error)
```

Sends a WhatsApp notification to a single user via Yalo. The `params` map should contain key-value pairs matching your WhatsApp template variables. Optional `NotificationOption` values allow advanced control (e.g., delivery priority). When no priority is provided, the client defaults to `"1"`.

### Notification Options

```go
func WithPriority(priority string) NotificationOption
```

- `WithPriority`: Overrides the default priority of `"1"` for the notification request.

### Response Types

```go
type Response struct {
    JSONData   string
    StatusCode int
    Headers    http.Header
}

type NotificationResponse struct {
    Status  string
    Message string
    Data    interface{}
    Errors  interface{}
}
```

## Running the Example

The example application demonstrates how to use the Yalo client by accepting a phone number and JSON parameters:

```bash
cd example
go run main.go +1234567890 '{"name":"John Doe","param1":"value1","param2":"value2"}' 0
```

The final argument is optional. If omitted, the client sends the notification with the default priority of `"1"`:

```bash
go run main.go +1234567890 '{"name":"John Doe","param1":"value1","param2":"value2"}'
```

Make sure you have a `.env` file configured with your Yalo credentials.

The example will:
1. Parse the JSON parameters from the command line
2. Send a notification to the specified phone number
3. Display the response from Yalo

## How It Works

The Go module sends JSON directly to the Yalo API:

1. **Request Preparation**
   - Accepts raw JSON string or Go struct (with generic `map[string]interface{}` params)
   - Marshals struct to JSON if needed
   - Adds `Content-Type: application/json` header
   - Adds `Authorization: Bearer <token>` header

2. **Rate Limiting**
   - Automatically throttles requests to 40 per second
   - Uses a token bucket approach for smooth rate limiting

3. **Request Execution**
   - Sends POST request to Yalo API endpoint
   - Uses retryable HTTP client with configurable timeout and retries (default: 3 retries)

4. **Response Handling**
   - Returns JSON response as string
   - Validates response is valid JSON
   - Provides helper to parse into Go structs

## Template Parameters

According to Yalo's documentation, the API expects:

| Parameter | Description |
|-----------|-------------|
| `type` | Template Name or HSM authorized by WhatsApp |
| `phone` | Message recipient's phone number |
| `params` | Variables and values to be interpreted by the template (as key-value pairs) |

This library provides a generic `map[string]interface{}` for the `params` field, allowing you to pass any parameters your specific WhatsApp template requires.

## API Endpoint

The Yalo notification endpoint follows this pattern:

```
POST /notifications/api/v1/accounts/{accountID}/bots/{botID}/notifications
```

The client automatically constructs the full URL based on the provided account ID and bot ID.

## Example cURL Request

For reference, here's what an equivalent cURL request looks like:

```bash
curl --location 'https://api-global.yalochat.com/notifications/api/v1/accounts/account-id/bots/bot-id/notifications' \
--header 'Content-Type: application/json' \
--header 'Authorization: Bearer [TOKEN]' \
--data '{
    "type": "template-name",
    "phone": "+1234567890",
    "params": {
        "param1": "value1",
        "param2": "value2"
    }
}'
```

The `params` object should match the variables defined in your WhatsApp template. The API sends one notification per request.

## Dependencies

- `github.com/spf13/viper` - Configuration management
- `github.com/hashicorp/go-retryablehttp` - Automatic HTTP retries with exponential backoff

## License

Apache License 2.0
