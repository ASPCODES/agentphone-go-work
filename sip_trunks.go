package agentphone

import(
	"context"
	"net/http"
)


// SIPTrunksService handles the /sip-trunks endpoints.
// Everything below is inferred purely from what a "SIP trunk" generally
// means for a telephony platform (a connection to an external SIP
// provider or on-prem PBX, so calls can route through it instead of — or
// alongside — AgentPhone's own numbers). Treat this file as a rough
// skeleton, not a reliable reference; confirm every field against a real
// request/response before using it.
type SIPTrunksService struct {
	client *Client
}


// SIPTrunk represents a configured SIP trunk connection.
type SIPTrunk struct {
	ID 		  string	`json:"id"`
	Name	  string	`json:"name,omitempty"`
	Host	  string	`json:"host,omitempty"`
	Port	  int		`json:"port,omitempty"`
	Username  string 	`json:"username,omitempty"`
	CreatedAt string 	`json:"createdAt,omitempty"`
}


// ListSIPTrunksResponse is the response from List.
type ListSIPTrunksResponse struct {
	SIPTrunks []SIPTrunk  `json:"data"`
	OffsetPageInfo
}


// List returns the SIP trunks on the account. params may be nil.
func (s *SIPTrunksService) List(ctx context.Context, params *ListParams) (*ListSIPTrunksResponse, error) {
	var resp ListSIPTrunksResponse
	err := s.client.request(ctx, http.MethodGet, "/sip-trunks"+params.toQuery(), nil, &resp)
	return &resp, err
}


// CreateSIPTrunkParams are the parameters for creating a SIP trunk.
type CreateSIPTrunkParams struct {
	Name     string `json:"name"`
	Host     string `json:"host"`
	Port     int    `json:"port,omitempty"`
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
}


// Create creates a new SIP trunk.
func (s *SIPTrunksService) Create(ctx context.Context, params *CreateSIPTrunkParams) (*SIPTrunk, error) {
	var trunk SIPTrunk
	err := s.client.request(ctx, http.MethodPost, "/sip-trunks", params, &trunk)
	return &trunk, err
}


// Get retrieves a single SIP trunk by ID.
func (s *SIPTrunksService) Get(ctx context.Context, trunkID string) (*SIPTrunk, error) {
	var trunk SIPTrunk
	err := s.client.request(ctx, http.MethodGet, "/sip-trunks/"+trunkID, nil, &trunk)
	return &trunk, err
}


// UpdateSIPTrunkParams are the parameters for updating a SIP trunk. All fields are optional — only set the ones you want to change.
type UpdateSIPTrunkParams struct {
	Name     string `json:"name,omitempty"`
	Host     string `json:"host,omitempty"`
	Port     int    `json:"port,omitempty"`
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
}


// Update changes a SIP trunk's configuration.
func (s *SIPTrunksService) Update(ctx context.Context, trunkID string, params *UpdateSIPTrunkParams) (*SIPTrunk, error) {
	var trunk SIPTrunk
	err := s.client.request(ctx, http.MethodPatch, "/sip-trunks/"+trunkID, params, &trunk)
	return &trunk, err
}


// Delete deletes a SIP trunk.
func (s *SIPTrunksService) Delete(ctx context.Context, trunkID string) error {
	return s.client.request(ctx, http.MethodDelete, "/sip-trunks/"+trunkID, nil, nil)
}
