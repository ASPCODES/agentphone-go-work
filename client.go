package agentphone

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	defaultBaseURL = "https://api.agentphone.ai/v1"
	defaultTimeout = 30 * time.Second
)

// Client is the AgentPhone API client. Create one with NewClient.
type Client struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client

	Agents        *AgentsService
	Numbers       *NumbersService
	Calls         *CallsService
	Messages      *MessagesService
	Conversations *ConversationsService
	Contacts      *ContactsService
	ContactCards  *ContactCardsService
	Webhooks      *WebhooksService
	Verification  *VerificationService
	Usage         *UsageService
	SubAccounts   *SubAccountsService
	SIPTrunks     *SIPTrunksService
	WhatsApp      *WhatsAppService
	Registration  *RegistrationService
	Location      *LocationService
}

// clientOptions collects everything the Option functions configure, before
// a Client is actually built. Using an intermediate struct (rather than
// mutating *Client directly) is what makes options order-independent: e.g.
// WithTimeout and WithHTTPClient can be passed in either order and produce
// the same result, since both are only resolved once, after every option
// has run.
type clientOptions struct {
	baseURL    string
	httpClient *http.Client
	timeout    *time.Duration
}


// Option configures optional Client behavior. Pass zero or more Options to
// NewClient. Options can be passed in any order.
type Option func(*clientOptions)


func WithBaseURL(baseURL string) Option {
	return func(o *clientOptions) {
		o.baseURL = baseURL
	}
}


// WithHTTPClient overrides the default *http.Client used to send requests.
// Use this if you need custom transport behavior (proxies, retries, etc).
// Its Timeout field is preserved unless WithTimeout is also passed, in
// which case WithTimeout always wins, regardless of option order.
func WithHTTPClient(httpClient *http.Client) Option {
	return func(o *clientOptions) {
		o.httpClient = httpClient
	}
}

// WithTimeout overrides the request timeout. Applies regardless of whether
// it's passed before or after WithHTTPClient.
func WithTimeout(timeout time.Duration) Option {
	return func(o *clientOptions) {
		o.timeout = &timeout
	}
}

// NewClient creates a new AgentPhone API client. apiKey is required; every
// other setting has a sensible default and can be overridden with Option
// functions, in any order.
func NewClient(apiKey string, opts ...Option) *Client {
	cfg := &clientOptions{baseURL: defaultBaseURL}

	for _, opt := range opts {
		opt(cfg)
	}

	httpClient := cfg.httpClient
	if httpClient == nil {
		httpClient = &http.Client{Timeout: defaultTimeout}
	}
	if cfg.timeout != nil {
		httpClient.Timeout = *cfg.timeout
	}

	c := &Client{
		apiKey:     apiKey,
		baseURL:    strings.TrimRight(cfg.baseURL, "/"),
		httpClient: httpClient,
	}


	c.Agents = 			&AgentsService{client: c}
	c.Numbers = 		&NumbersService{client: c}
	c.Calls = 			&CallsService{client: c}
	c.Messages = 		&MessagesService{client: c}
	c.Conversations = 	&ConversationsService{client: c}
	c.Contacts = 		&ContactsService{client: c}
	c.ContactCards = 	&ContactCardsService{client: c}
	c.Webhooks = 		&WebhooksService{client: c}
	c.Verification = 	&VerificationService{client: c}
	c.Usage = 			&UsageService{client: c}
	c.SubAccounts = 	&SubAccountsService{client: c}
	c.SIPTrunks = 		&SIPTrunksService{client: c}
	c.WhatsApp = 		&WhatsAppService{client: c}
	c.Registration = 	&RegistrationService{client: c}
	c.Location = 		&LocationService{client: c}

	return c
}

// request builds and sends an HTTP request against the AgentPhone API. If
// body is non-nil, it's JSON-encoded as the request body. If result is
// non-nil, the JSON response is decoded into it on success. Non-2xx
// responses are converted into a typed error (see errors.go).
//
// Every resource file (agents.go, calls.go, ...) calls this method so that
// header handling, JSON encoding/decoding, and error parsing live in one
// place.
func (c *Client) request(ctx context.Context, method, path string, body, result interface{}) error {
	var reqBody io.Reader

	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("agentphone: encoding request body: %w", err)
		}
		reqBody = bytes.NewReader(data)
	}

	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, reqBody)
	if err != nil {
		return fmt.Errorf("agentphone: building request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	c.setAuthHeader(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("agentphone: sending request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("agentphone: reading response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return parseAPIError(resp.StatusCode, respBody, resp.Header.Get("Retry-After"))
	}

	if result != nil && len(respBody) > 0 {
		if err := json.Unmarshal(respBody, result); err != nil {
			return fmt.Errorf("agentphone: decoding response: %w", err)
		}
	}

	return nil
}
