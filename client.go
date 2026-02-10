// Package splox provides a Go client for the Splox API.
//
// Create a client with [NewClient], then call methods on its
// Workflows, Chats, and Events fields:
//
//	client := splox.NewClient("your-api-key")
//
//	workflows, err := client.Workflows.List(ctx, nil)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	for _, wf := range workflows.Workflows {
//	    fmt.Println(wf.ID, wf.LatestVersion.Name)
//	}
package splox

import (
	"net/http"
	"os"
	"time"
)

const (
	DefaultBaseURL = "https://app.splox.io/api/v1"
	DefaultTimeout = 30 * time.Second
)

// Client is the Splox API client.
type Client struct {
	Workflows *WorkflowService
	Chats     *ChatService
	Events    *EventService
	Billing   *BillingService
	Memory    *MemoryService
	MCP       *MCPService

	baseURL    string
	apiKey     string
	httpClient *http.Client
}

// Option configures the Client.
type Option func(*Client)

// WithBaseURL overrides the default API base URL.
func WithBaseURL(url string) Option {
	return func(c *Client) { c.baseURL = url }
}

// WithHTTPClient sets a custom *http.Client (e.g. for proxies or custom TLS).
func WithHTTPClient(hc *http.Client) Option {
	return func(c *Client) { c.httpClient = hc }
}

// WithTimeout sets the HTTP request timeout.
func WithTimeout(d time.Duration) Option {
	return func(c *Client) { c.httpClient.Timeout = d }
}

// NewClient creates a new Splox API client.
//
// If apiKey is empty, it falls back to the SPLOX_API_KEY environment variable.
func NewClient(apiKey string, opts ...Option) *Client {
	if apiKey == "" {
		apiKey = os.Getenv("SPLOX_API_KEY")
	}

	c := &Client{
		baseURL: DefaultBaseURL,
		apiKey:  apiKey,
		httpClient: &http.Client{
			Timeout: DefaultTimeout,
		},
	}

	for _, opt := range opts {
		opt(c)
	}

	c.Workflows = &WorkflowService{client: c}
	c.Chats = &ChatService{client: c}
	c.Events = &EventService{client: c}
	c.Billing = &BillingService{client: c}
	c.Memory = &MemoryService{client: c}
	c.MCP = &MCPService{client: c}

	return c
}
