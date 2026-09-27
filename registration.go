package agentphone

import(
	"context"
	"net/http"
)


// RegistrationService handles the /register endpoints: US A2P 10DLC registration, required to send SMS to US numbers at scale, reference -> docs.agentphone.ai/error-handling's. This number is not registered for US A2P 10DLC messaging" error, and its note that raising a brand's daily volume limit requires brand vetting
type RegistrationService struct {
	client *Client
}

// RegistrationStatus is the response from GetStatus.
type RegistrationStatus struct {
	Status string `json:"status,omitempty"`
}
 

// GetStatus returns a lightweight registration status check.
func (s *RegistrationService) GetStatus(ctx context.Context) (*RegistrationStatus, error) {
	var status RegistrationStatus
	err := s.client.request(ctx, http.MethodGet, "/register/status", nil, &status)
	return &status, err
}

 
// RegistrationBrand identifies the business sending messages.
type RegistrationBrand struct {
	LegalName string `json:"legalName,omitempty"`
	EIN       string `json:"ein,omitempty"`
	Address   string `json:"address,omitempty"`
	Website   string `json:"website,omitempty"`
}
 

// RegistrationCampaign describes the messaging use case being registered.
type RegistrationCampaign struct {
	UseCase        string   `json:"useCase,omitempty"`
	Description    string   `json:"description,omitempty"`
	SampleMessages []string `json:"sampleMessages,omitempty"`
}
 

// Registration is the account's full A2P 10DLC registration record.
type Registration struct {
	Status    string                `json:"status,omitempty"`
	Brand     *RegistrationBrand    `json:"brand,omitempty"`
	Campaign  *RegistrationCampaign `json:"campaign,omitempty"`
	CreatedAt string                `json:"createdAt,omitempty"`
}
 

// GetDetails returns the full registration record (brand and campaign details, not just the status).
func (s *RegistrationService) GetDetails(ctx context.Context) (*Registration, error) {
	var reg Registration
	err := s.client.request(ctx, http.MethodGet, "/register", nil, &reg)
	return &reg, err
}
 

// RegisterA2PParams are the parameters for starting A2P 10DLC registration. LegalName is assumed required; the rest follow the standard brand+campaign fields.
type RegisterA2PParams struct {
	LegalName      string   `json:"legalName"`
	EIN            string   `json:"ein,omitempty"`
	Address        string   `json:"address,omitempty"`
	Website        string   `json:"website,omitempty"`
	UseCase        string   `json:"useCase,omitempty"`
	Description    string   `json:"description,omitempty"`
	SampleMessages []string `json:"sampleMessages,omitempty"`
}
 
// Register starts or submits A2P 10DLC registration for the account.
func (s *RegistrationService) Register(ctx context.Context, params *RegisterA2PParams) (*Registration, error) {
	var reg Registration
	err := s.client.request(ctx, http.MethodPost, "/register", params, &reg)
	return &reg, err
}
 
// UpdateA2PRegistrationParams are the parameters for updating a registration.
type UpdateA2PRegistrationParams struct {
	LegalName      string   `json:"legalName,omitempty"`
	EIN            string   `json:"ein,omitempty"`
	Address        string   `json:"address,omitempty"`
	Website        string   `json:"website,omitempty"`
	UseCase        string   `json:"useCase,omitempty"`
	Description    string   `json:"description,omitempty"`
	SampleMessages []string `json:"sampleMessages,omitempty"`
}
 

// Update updates the account's A2P 10DLC registration details.
func (s *RegistrationService) Update(ctx context.Context, params *UpdateA2PRegistrationParams) (*Registration, error) {
	var reg Registration
	err := s.client.request(ctx, http.MethodPatch, "/register", params, &reg)
	return &reg, err
}
