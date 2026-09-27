package agentphone

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
)

// ContactsService handles the /contacts endpoints: a simple address book
// of phone numbers with names, emails, and notes.
type ContactsService struct {
	client *Client
}

// Contact represents a saved contact.
type Contact struct {
	ID          string `json:"id"`
	PhoneNumber string `json:"phoneNumber"`
	Name        string `json:"name,omitempty"`
	Email       string `json:"email,omitempty"`
	Notes       string `json:"notes,omitempty"`
	CreatedAt   string `json:"createdAt,omitempty"`
}

// ListContactsResponse is the response from List.
//
// NOTE: the pagination docs name numbers/conversations/agents as offset-
// paginated examples ("endpoints like ...", not an exhaustive list).
// Contacts is assumed to follow the same convention; verify against a
// live response.
type ListContactsResponse struct {
	Contacts []Contact `json:"data"`
	OffsetPageInfo
}

// ListContactsParams filters and paginates List.
type ListContactsParams struct {
	Limit  int
	Offset int
	Search string
}

func (p *ListContactsParams) toQuery() string {
	if p == nil {
		return ""
	}
	q := url.Values{}
	if p.Limit != 0 {
		q.Set("limit", strconv.Itoa(p.Limit))
	}
	if p.Offset != 0 {
		q.Set("offset", strconv.Itoa(p.Offset))
	}
	if p.Search != "" {
		q.Set("search", p.Search)
	}
	if encoded := q.Encode(); encoded != "" {
		return "?" + encoded
	}
	return ""
}

// List returns the contacts on the account.
func (s *ContactsService) List(ctx context.Context, params *ListContactsParams) (*ListContactsResponse, error) {
	var resp ListContactsResponse
	err := s.client.request(ctx, http.MethodGet, "/contacts"+params.toQuery(), nil, &resp)
	return &resp, err
}

// CreateContactParams are the parameters for creating a contact.
// PhoneNumber is required.
type CreateContactParams struct {
	PhoneNumber string `json:"phoneNumber"`
	Name        string `json:"name,omitempty"`
	Email       string `json:"email,omitempty"`
	Notes       string `json:"notes,omitempty"`
}

// Create creates a new contact.
func (s *ContactsService) Create(ctx context.Context, params *CreateContactParams) (*Contact, error) {
	var contact Contact
	err := s.client.request(ctx, http.MethodPost, "/contacts", params, &contact)
	return &contact, err
}

// Get retrieves a single contact by ID.
func (s *ContactsService) Get(ctx context.Context, contactID string) (*Contact, error) {
	var contact Contact
	err := s.client.request(ctx, http.MethodGet, "/contacts/"+contactID, nil, &contact)
	return &contact, err
}

// UpdateContactParams are the parameters for updating a contact.
type UpdateContactParams struct {
	Name  string `json:"name,omitempty"`
	Email string `json:"email,omitempty"`
	Notes string `json:"notes,omitempty"`
}


// Update changes a contact's name, email, or notes.
func (s *ContactsService) Update(ctx context.Context, contactID string, params *UpdateContactParams) (*Contact, error) {
	var contact Contact
	err := s.client.request(ctx, http.MethodPatch, "/contacts/"+contactID, params, &contact)
	return &contact, err
}


// Delete deletes a contact.
func (s *ContactsService) Delete(ctx context.Context, contactID string) error {
	return s.client.request(ctx, http.MethodDelete, "/contacts/"+contactID, nil, nil)
}

// ContactCapabilities describes which channels a phone number can be reached on.
type ContactCapabilities struct {
	PhoneNumber string `json:"phoneNumber,omitempty"`
	SMS         bool   `json:"sms,omitempty"`
	IMessage    bool   `json:"imessage,omitempty"`
	WhatsApp    bool   `json:"whatsapp,omitempty"`
	Voice       bool   `json:"voice,omitempty"`
}

// GetCapabilities checks which channels (SMS, iMessage, WhatsApp, voice)
// a phone number supports — useful before creating a contact or sending
// to a new number.
func (s *ContactsService) GetCapabilities(ctx context.Context, phoneNumber string) (*ContactCapabilities, error) {
	q := url.Values{}
	q.Set("phoneNumber", phoneNumber)

	var caps ContactCapabilities
	err := s.client.request(ctx, http.MethodGet, "/contacts/capabilities?"+q.Encode(), nil, &caps)
	return &caps, err
}
