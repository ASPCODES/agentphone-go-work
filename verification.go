package agentphone

import(
	"context"
	"net/http"
)

// VerificationService handles the /verify endpoints: one-time-code phone number verification (send a code, then check what the user entered).
type VerificationService struct {
	client *Client
}


// SendVerificationParams are the parameters for sending a verification code.
type SendVerificationParams struct {
	PhoneNumber		string   `json:"phoneNumber"`
	Channel			string	 `json:"channel,omitempty"`
}


// SendVerificationResponse is the response from Send.
type SendVerificationResponse struct {
	Status		string	`json:"status,omitempty"`
	PhoneNumber	string	`json:"phoneNumber,omitempty"`
}


// Send, sends a one-time verification code to a phone number.
func (s *VerificationService) Send(ctx context.Context, params *SendVerificationParams) (*SendVerificationResponse, error) {
	var resp SendVerificationResponse
	err := s.client.request(ctx, http.MethodPost, "/verify/send", params, &resp)
	return &resp, err
}


// CheckVerificationParams are the parameters for checking a code the user entered.
type CheckVerificationParams struct {
	PhoneNumber 	string `json:"phoneNumber"`
	Code        	string `json:"code"`
}


// CheckVerificationResponse is the response from Check.
type CheckVerificationResponse struct {
	Status 	string 	`json:"status,omitempty"`
	Valid	bool	`json:"valid,omitempty"`
}


// Check verifies the code a user entered against the one most recently sent to their number.
func (s *VerificationService) Check(ctx context.Context, params *CheckVerificationParams) (*CheckVerificationResponse, error) {
	var resp CheckVerificationResponse
	err := s.client.request(ctx, http.MethodPost, "/verify/check", params, &resp)
	return &resp, err
}

