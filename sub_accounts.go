package agentphone

import(
	"context"
	"net/http"
)

// SubAccountsService handles the /sub-accounts endpoints: isolated environments (own agents, numbers, conversations, calls, webhooks, and 10DLC registration) under one master account, sharing its billing wallet, API key, and team members. Up to 500 per master account; sub-accounts can't create sub-accounts of their own (one level only).
type SubAccountsService struct {
	client *Client
}


// SubAccount represents a sub-account.
type SubAccount struct {
	ID 		  string	`json:"id"`
	Name	  string	`json:"name,omitempty"`
	Status    string 	`json:"status,omitempty"`
	CreatedAt string 	`json:"createdAt,omitempty"`
}


// ListSubAccountsResponse is the response from List.
type ListSubAccountsResponse struct {
	SubAccounts []SubAccount `json:"data"`
	OffsetPageInfo
}


// List returns the sub-accounts under the project. params may be nil.
func (s *SubAccountsService) List(ctx context.Context, params *ListParams) (*ListSubAccountsResponse, error) {
	var resp ListSubAccountsResponse
	err := s.client.request(ctx, http.MethodGet, "/sub-accounts"+params.toQuery(), nil, &resp)
	return &resp, err
}


// CreateSubAccountParams are the parameters for creating a sub-account. Name is assumed required; no other fields are confirmed.
type CreateSubAccountParams struct {
	Name string `json:"name"`
}


// Create creates a new sub-account.
func (s *SubAccountsService) Create(ctx context.Context, params *CreateSubAccountParams) (*SubAccount, error) {
	var sub SubAccount
	err := s.client.request(ctx, http.MethodPost, "/sub-accounts", params, &sub)
	return &sub, err
}


// UpdateSubAccountParams are the parameters for updating a sub-account.
type UpdateSubAccountParams struct {
	Name   string `json:"name,omitempty"`
	Status string `json:"status,omitempty"`
}



// Update changes a sub-account's name or status.
func (s *SubAccountsService) Update(ctx context.Context, subAccountID string, params *UpdateSubAccountParams) (*SubAccount, error) {
	var sub SubAccount
	err := s.client.request(ctx, http.MethodPatch, "/sub-accounts/"+subAccountID, params, &sub)
	return &sub, err
}


// Delete, deletes a sub-account.
func (s *SubAccountsService) Delete(ctx context.Context, subAccountID string) error {
	return s.client.request(ctx, http.MethodDelete, "/sub-accounts/"+subAccountID, nil, nil)
}

