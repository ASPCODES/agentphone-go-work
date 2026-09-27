package agentphone

import(
	"context"
	"net/url"
	"net/http"
)


// LocationService handles the /location endpoints: geographic lookup for a phone number, with a way to refresh cached results.
type LocationService struct {
	client *Client
}


// Location is geographic info for a phone number.
type Location struct {
	PhoneNumber string `json:"phoneNumber,omitempty"`
	City        string `json:"city,omitempty"`
	State       string `json:"state,omitempty"`
	Country     string `json:"country,omitempty"`
	Timezone    string `json:"timezone,omitempty"`
}


// Get looks up geographic info for a phone number.
func (s *LocationService) Get(ctx context.Context, phoneNumber string) (*Location, error) {
	var loc Location
	err := s.client.request(ctx, http.MethodGet, "/location/"+url.PathEscape(phoneNumber), nil, &loc)
	return &loc, err
}

// RefreshLocationsParams are the parameters for Refresh.
type RefreshLocationsParams struct {
	NumberIDs	[]string  `json:"numberIds,omitempty"`
}


// RefreshLocationsResponse is the response from Refresh.
type RefreshLocationsResponse struct {
	Refreshed  int 	`json:"refreshed,omitempty"`
}

// Refresh re-fetches cached location data for the account's numbers (or a subset, via params.NumberIDs).
func (s *LocationService) Refresh(ctx context.Context, params *RefreshLocationsParams) (*RefreshLocationsResponse, error) {
	var resp RefreshLocationsResponse
	err := s.client.request(ctx, http.MethodPost, "location/refresh", params, &resp)
	return &resp, err
}
