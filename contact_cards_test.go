package agentphone

import (
	"context"
	"net/http"
)

// ContactCardsService handles the /numbers/{id}/contact-card endpoints, the iMessage Business "contact card" shown to recipients for a number (name, organization, avatar).
type ContactCardsService struct {
	client *Client
}


// ContactCard is a number's iMessage Business contact card.
type ContactCard struct {
	NumberID     string `json:"numberId,omitempty"`
	Name         string `json:"name,omitempty"`
	Organization string `json:"organization,omitempty"`
	AvatarURL    string `json:"avatarUrl,omitempty"`
	UpdatedAt    string `json:"updatedAt,omitempty"`
}

// Get retrieves a number's contact card.
func (s *ContactCardsService) Get(ctx context.Context, numberID string) (*ContactCard, error) {
	var card ContactCard
	err := s.client.request(ctx, http.MethodGet, "/numbers/"+numberID+"/contact-card", nil, &card)
	return &card, err
}

// PutContactCardParams sets a number's contact card. Being a PUT, this likely replaces the whole card rather than patching individual fields
type PutContactCardParams struct {
	Name         string `json:"name,omitempty"`
	Organization string `json:"organization,omitempty"`
	AvatarURL    string `json:"avatarUrl,omitempty"`
}

// Put sets (replaces) a number's contact card.
func (s *ContactCardsService) Put(ctx context.Context, numberID string, params *PutContactCardParams) (*ContactCard, error) {
	var card ContactCard
	err := s.client.request(ctx, http.MethodPut, "/numbers/"+numberID+"/contact-card", params, &card)
	return &card, err
}

// Delete removes a number's contact card.
func (s *ContactCardsService) Delete(ctx context.Context, numberID string) error {
	return s.client.request(ctx, http.MethodDelete, "/numbers/"+numberID+"/contact-card", nil, nil)
}
