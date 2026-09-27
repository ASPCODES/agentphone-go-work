package agentphone

import(
	"context"
	"net/http"
	"net/url"
)


// WhatsAppService handles the /integrations/whatsapp endpoints: connecting a WhatsApp Business Account, attaching numbers to it, and managing message templates.
type WhatsAppService struct {
	client *Client
}


// Known Meta WhatsApp send-error codes. These show up in the failure detail of a WhatsApp send Meta rejected (as opposed to AgentPhone itself).
const (
	// MetaErrorWindowClosed: the 24-hour session window is closed, the recipient hasn't messaged in over a day. Send an approved template instead of a free-form message.
	MetaErrorWindowClosed = "131047"
	// MetaErrorNoPaymentMethod: the WhatsApp Business Account has no valid payment method. Session messages are free, so this usually surfaces on your first template send.
	MetaErrorNoPaymentMethod = "131042"
	// MetaErrorRecipientUnreachable: the recipient can't receive messages. often not a WhatsApp user, or has blocked you. Don't retry; fall back to SMS if you have consent.
	MetaErrorRecipientUnreachable = "131026"
	// MetaErrorUnsupportedMessageType: unsupported message type for this recipient.
	MetaErrorUnsupportedMessageType = "131051"
	// MetaErrorMalformedRequest: malformed request, usually a template whose variable shape doesn't match how it was authored.
	MetaErrorMalformedRequest = "100"
)


// WhatsAppConnection represents a connected WhatsApp Business Account.
type WhatsAppConnection struct {
	ID 		string	`json:"id"`
	Status    string `json:"status,omitempty"`
	CreatedAt string `json:"createdAt,omitempty"`
}


// ConnectWhatsAppParams are the parameters for starting a WhatsApp connection.
// NOTE: WhatsApp Business connections are typically completed via Meta's Embedded Signup flow, which hands back an authorization code your backend exchanges server-side. Code is assumed to be that value; verify against the actual flow before relying on it.
type ConnectWhatsAppParams struct {
	Code 	string	 `json:"code,omitempty"`
}


// Connect starts or completes connecting a WhatsApp Business Account.
func (s *WhatsAppService) Connect(ctx context.Context, params *ConnectWhatsAppParams) (*WhatsAppConnection, error) {
	var conn WhatsAppConnection
	err := s.client.request(ctx, http.MethodPost, "/integrations/whatsapp/connect", params, &conn)
	return &conn, err
}
 

// WhatsAppAvailableNumber is a number eligible to attach to a WhatsApp connection.
type WhatsAppAvailableNumber struct {
	PhoneNumber string `json:"phoneNumber,omitempty"`
}
 
// ListAvailableNumbersResponse is the response from AvailableNumbers.
type ListWhatsAppAvailableNumbersResponse struct {
	Numbers []WhatsAppAvailableNumber `json:"data"`
}
 
// AvailableNumbers lists numbers eligible to attach to a WhatsApp connection.
func (s *WhatsAppService) AvailableNumbers(ctx context.Context, connectionID string) (*ListWhatsAppAvailableNumbersResponse, error) {
	var resp ListWhatsAppAvailableNumbersResponse
	err := s.client.request(ctx, http.MethodGet, "/integrations/whatsapp/"+connectionID+"/available-numbers", nil, &resp)
	return &resp, err
}
 
// ConnectWhatsAppNumberParams are the parameters for attaching a number to a WhatsApp connection.
type ConnectWhatsAppNumberParams struct {
	PhoneNumber string `json:"phoneNumber"`
}
 
// ConnectNumber attaches a phone number to a WhatsApp connection, enabling WhatsApp messaging on it. Returns the updated Number.
func (s *WhatsAppService) ConnectNumber(ctx context.Context, connectionID string, params *ConnectWhatsAppNumberParams) (*Number, error) {
	var num Number
	err := s.client.request(ctx, http.MethodPost, "/integrations/whatsapp/"+connectionID+"/numbers", params, &num)
	return &num, err
}
 
// WhatsAppTemplate is a pre-approved WhatsApp message template.
type WhatsAppTemplate struct {
	ID   string `json:"id"`
	Name string `json:"name,omitempty"`
	Status   string `json:"status,omitempty"`
	Category string `json:"category,omitempty"`
	Language string `json:"language,omitempty"`
}
 
// ListWhatsAppTemplatesResponse is the response from ListTemplates.
type ListWhatsAppTemplatesResponse struct {
	Templates []WhatsAppTemplate `json:"data"`
}
 
// ListTemplates lists the WhatsApp message templates on a connection.
func (s *WhatsAppService) ListTemplates(ctx context.Context, connectionID string) (*ListWhatsAppTemplatesResponse, error) {
	var resp ListWhatsAppTemplatesResponse
	err := s.client.request(ctx, http.MethodGet, "/integrations/whatsapp/"+connectionID+"/templates", nil, &resp)
	return &resp, err
}
 
// CreateWhatsAppTemplateParams are the parameters for creating a template, submitted to Meta for approval.
type CreateWhatsAppTemplateParams struct {
	Name     string `json:"name"`
	Category string `json:"category,omitempty"`
	Language string `json:"language,omitempty"`
	Body     string `json:"body,omitempty"`
}
 
// CreateTemplate creates a new WhatsApp message template.
func (s *WhatsAppService) CreateTemplate(ctx context.Context, connectionID string, params *CreateWhatsAppTemplateParams) (*WhatsAppTemplate, error) {
	var tmpl WhatsAppTemplate
	err := s.client.request(ctx, http.MethodPost, "/integrations/whatsapp/"+connectionID+"/templates", params, &tmpl)
	return &tmpl, err
}
 
// DeleteTemplate deletes a WhatsApp message template.
func (s *WhatsAppService) DeleteTemplate(ctx context.Context, connectionID, templateID string) error {
	q := url.Values{}
	q.Set("templateId", templateID)
	return s.client.request(ctx, http.MethodDelete, "/integrations/whatsapp/"+connectionID+"/templates?"+q.Encode(), nil, nil)
}
 
// WhatsAppStatus is the response from Status.
type WhatsAppStatus struct {
	Connected bool `json:"connected,omitempty"`
	Enabled bool   `json:"enabled,omitempty"`
	Status  string `json:"status,omitempty"`
}
 

// Status returns the account's overall WhatsApp integration status.
func (s *WhatsAppService) Status(ctx context.Context) (*WhatsAppStatus, error) {
	var status WhatsAppStatus
	err := s.client.request(ctx, http.MethodGet, "/integrations/whatsapp/status", nil, &status)
	return &status, err
}
 

// Disconnect removes a WhatsApp connection.
func (s *WhatsAppService) Disconnect(ctx context.Context, connectionID string) error {
	return s.client.request(ctx, http.MethodDelete, "/integrations/whatsapp/"+connectionID, nil, nil)
}
