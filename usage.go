package agentphone

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
)

// UsageService handles the /usage endpoints: account plan info, number limits, and message/call usage stats.
type UsageService struct {
	client *Client
}


// UsagePlan describes the account's current plan.
type UsagePlan struct {
	Name string `json:"name,omitempty"`
}


// UsageNumbers is the account's phone-number usage against its plan limit. 
type UsageNumbers struct {
	Used      int `json:"used,omitempty"`
	Limit     int `json:"limit,omitempty"`
	Remaining int `json:"remaining,omitempty"`
}


// UsageStats summarizes recent messaging/call activity.
type UsageStats struct {
	MessagesLast24h int `json:"messagesLast24h,omitempty"`
	MessagesLast30d int `json:"messagesLast30d,omitempty"`
	CallsLast30d    int `json:"callsLast30d,omitempty"`
}


// Usage is the account's overall usage summary.
type Usage struct {
	Plan    UsagePlan    `json:"plan,omitempty"`
	Numbers UsageNumbers `json:"numbers,omitempty"`
	Stats   UsageStats   `json:"stats,omitempty"`
}


// Get retrieves the account's current usage summary (plan, number limits, recent stats).
func (s *UsageService) Get(ctx context.Context) (*Usage, error) {
	var usage Usage
	err := s.client.request(ctx, http.MethodGet, "/usage", nil, &usage)
	return &usage, err
}


// DailyUsageEntry is one day's usage, get_daily(days=30) example (day.date, day.messages, day.calls).
type DailyUsageEntry struct {
	Date     string `json:"date"`
	Messages int    `json:"messages,omitempty"`
	Calls    int    `json:"calls,omitempty"`
}


// DailyUsageResponse is the response from GetDaily.
type DailyUsageResponse struct {
	Data []DailyUsageEntry `json:"data"`
}


// GetDaily returns a daily usage breakdown for the last N days. Pass 0 to omit the parameter and use the API's default window.
func (s *UsageService) GetDaily(ctx context.Context, days int) (*DailyUsageResponse, error) {
	path := "/usage/daily"
	if days != 0 {
		q := url.Values{}
		q.Set("days", strconv.Itoa(days))
		path += "?" + q.Encode()
	}

	var resp DailyUsageResponse
	err := s.client.request(ctx, http.MethodGet, path, nil, &resp)
	return &resp, err
}


// MonthlyUsageEntry is one month's usage.
type MonthlyUsageEntry struct {
	Month    string `json:"month,omitempty"`
	Messages int    `json:"messages,omitempty"`
	Calls    int    `json:"calls,omitempty"`
}


// MonthlyUsageResponse is the response from GetMonthly.
type MonthlyUsageResponse struct {
	Data []MonthlyUsageEntry `json:"data"`
}


// GetMonthly returns a monthly usage breakdown for the last N months. Pass 0 to omit the parameter and use the API's default window.
func (s *UsageService) GetMonthly(ctx context.Context, months int) (*MonthlyUsageResponse, error) {
	path := "/usage/monthly"
	if months != 0 {
		q := url.Values{}
		q.Set("months", strconv.Itoa(months))
		path += "?" + q.Encode()
	}

	var resp MonthlyUsageResponse
	err := s.client.request(ctx, http.MethodGet, path, nil, &resp)
	return &resp, err
}


// UsageByNumberEntry is one number's usage breakdown.
type UsageByNumberEntry struct {
	NumberID string `json:"numberId,omitempty"`
	Messages int    `json:"messages,omitempty"`
	Calls    int    `json:"calls,omitempty"`
}


// UsageByNumberResponse is the response from GetByNumber.
type UsageByNumberResponse struct {
	Data []UsageByNumberEntry `json:"data"`
}


// GetByNumber returns a usage breakdown per phone number.
func (s *UsageService) GetByNumber(ctx context.Context) (*UsageByNumberResponse, error) {
	var resp UsageByNumberResponse
	err := s.client.request(ctx, http.MethodGet, "/usage/by-number", nil, &resp)
	return &resp, err
}


// UsageByAgentEntry is one agent's usage breakdown.
type UsageByAgentEntry struct {
	AgentID  string `json:"agentId,omitempty"`
	Messages int    `json:"messages,omitempty"`
	Calls    int    `json:"calls,omitempty"`
}


// UsageByAgentResponse is the response from GetByAgent.
type UsageByAgentResponse struct {
	Data []UsageByAgentEntry `json:"data"`
}


// GetByAgent returns a usage breakdown per agent.
func (s *UsageService) GetByAgent(ctx context.Context) (*UsageByAgentResponse, error) {
	var resp UsageByAgentResponse
	err := s.client.request(ctx, http.MethodGet, "/usage/by-agent", nil, &resp)
	return &resp, err
}
