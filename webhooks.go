package agentphone

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
)

// WebhooksService handles the project-level /webhooks endpoints: the
// single webhook URL that receives SMS/iMessage and (in webhook voice
// mode) voice events for the whole account, unless overridden per-agent.
type WebhooksService struct {
	client *Client
}

// Webhook is the project's webhook configuration.
type Webhook struct {
	URL          string `json:"url"`
	Secret       string `json:"secret,omitempty"`
	ContextLimit int    `json:"contextLimit,omitempty"`
	Timeout      int    `json:"timeout.omitempty"`
	CreatedAt    string `json:"createdAt,omitempty"`
	UpdatedAt    string `json:"updatedAt,omitempty"`
}

// Get retrieves the project's webhook configuration.
func (s *WebhooksService) Get(ctx context.Context) (*Webhook, error) {
	var wh Webhook
	err := s.client.request(ctx, http.MethodGet, "/webhooks", nil, &wh)
	return &wh, err
}

// CreateOrUpdateWebhookParams are the parameters for setting the
// project's webhook. URL is required; ContextLimit and Timeout are
// pointers since 0 is a meaningful value (e.g. ContextLimit: 0 means "no
// history"), distinct from "leave unset".
type CreateOrUpdateWebhookParams struct {
	URL          string `json:"url"`
	ContextLimit *int   `json:"contextLimit,omitempty"`
	Timeout      *int   `json:"timeout,omitempty"`
}

// CreateOrUpdate sets (or replaces) the project's webhook configuration.The response includes Secret — save it, since it's needed to verify incoming deliveries.
func (s *WebhooksService) CreateOrUpdate(ctx context.Context, params *CreateOrUpdateWebhookParams) (*Webhook, error) {
	var wh Webhook
	err := s.client.request(ctx, http.MethodPost, "/webhooks", params, &wh)
	return &wh, err
}

// Delete removes the project's webhook configuration.
func (s *WebhooksService) Delete(ctx context.Context) error {
	return s.client.request(ctx, http.MethodDelete, "/webhooks", nil, nil)
}

// WebhookDelivery is one attempted delivery to the webhook URL.
type WebhookDelivery struct {
	ID           string `json:"id"`
	Event        string `json:"event,omitempty"`
	StatusCode   int    `json:"statusCode,omitempty"`
	Success      bool   `json:"success,omitempty"`
	DeliveredAt  string `json:"deliveredAt,omitempty"`
	ResponseBody string `json:"responseBody,omitempty"`
}

// ListDeliveriesResponse is the response from ListDeliveries.
type ListDeliveriesResponse struct {
	Deliveries []WebhookDelivery `json:"data"`
	OffsetPageInfo
}

// ListDeliveries returns the webhook's recent delivery attempts.
func (s *WebhooksService) ListDeliveries(ctx context.Context, params *ListParams) (*ListDeliveriesResponse, error) {
	var resp ListDeliveriesResponse
	err := s.client.request(ctx, http.MethodGet, "/webhooks/deliveries"+params.toQuery(), nil, &resp)
	return &resp, err
}

// DeliveryStats summarizes delivery success over a recent time window.
type DeliveryStats struct {
	SuccessRate float64 `json:"successRate,omitempty"`
	Total       int     `json:"total,omitempty"`
	Successful  int     `json:"successful,omitempty"`
	Failed      int     `json:"failed,omitempty"`
	Hours       int     `json:"hours,omitempty"`
}


// GetDeliveryStats returns delivery stats for the last N hours.
func (s *WebhooksService) GetDeliveryStats(ctx context.Context, hours int) (*DeliveryStats, error) {
	path := "/webhooks/deliveries/stats"
	if hours != 0 {
		q := url.Values{}
		q.Set("hours", strconv.Itoa(hours))
		path += "?" + q.Encode()
	}

	var stats DeliveryStats
	err := s.client.request(ctx, http.MethodGet, path, nil, &stats)
	return &stats, err
}


// AllTimeStats summarizes delivery success across the webhook's entire history.
type AllTimeStats struct {
	SuccessRate float64 `json:"successRate,omitempty"`
	Total       int     `json:"total,omitempty"`
	Successful  int     `json:"successful,omitempty"`
	Failed      int     `json:"failed,omitempty"`
}


// GetAllTimeStats returns all-time delivery stats for the webhook.
func (s *WebhooksService) GetAllTimeStats(ctx context.Context) (*AllTimeStats, error) {
	var stats AllTimeStats
	err := s.client.request(ctx, http.MethodGet, "/webhooks/deliveries/all-time", nil, &stats)
	return &stats, err
}


// TestWebhookParams are the parameters for Test. AgentID is optional.
type TestWebhookParams struct {
	AgentID 	string `json:"agentId,omitempty"`
}

// Test sends a test event to the webhook URL to confirm it's reachable and responding correctly.
func (s *WebhooksService) Test(ctx context.Context, params *TestWebhookParams) error {
	return s.client.request(ctx, http.MethodPost, "webhooks/test", params, nil)
}
