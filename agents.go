package agentphone

import(
	"context"
	"net/http"
)


// AgentsService handles the /agents endpoints: creating and configuring AI agent personas, and their attached numbers, conversations, and calls.
type AgentsService struct {
	client  *Client
}


// Agent represents an AI agent persona (voice mode, prompt, voice settings) that can have phone numbers attached to it.
type Agent struct {
	ID 						string	`json:"id"`
	Name					string	`json:"name"`
	Description 			string 	`json:"description,omitempty"`
	VoiceMode   			string 	`json:"voiceMode,omitempty"`
	SystemPrompt    		string 	`json:"systemPrompt,omitempty"`
	BeginMessage    		string 	`json:"beginMessage,omitempty"`
	Voice 					string	`json:"voice,omitempty"`
	ModelTier       		string 	`json:"modelTier,omitempty"`
	STTMode         		string 	`json:"sttMode,omitempty"`
	AmbientSound    		string 	`json:"ambientSound,omitempty"`
	DenoisingMode   		string 	`json:"denoisingMode,omitempty"`
	TransferNumber  		string 	`json:"transferNumber,omitempty"`
	VoicemailMessage 		string 	`json:"voicemailMessage,omitempty"`
	Language				string	`json:"language,omitempty"`
	VoiceSpeed              float64 `json:"voiceSpeed,omitempty"`
	InterruptionSensitivity float64 `json:"interruptionSensitivity,omitempty"`
	EnableBackchannel       bool    `json:"enableBackchannel,omitempty"`
	MaxSilenceMs            int     `json:"maxSilenceMs,omitempty"`
	EnableMessaging         bool    `json:"enableMessaging,omitempty"`
 
	CreatedAt 				string  `json:"createdAt,omitempty"`
	Numbers   []Number `json:"numbers,omitempty"`
}

// CreateAgentParams are the parameters for creating an agent. Name is required; everything else is optional and uses the platform default when omitted. VoiceSpeed, InterruptionSensitivity, EnableBackchannel, MaxSilenceMs and EnableMessaging are pointers because their zero value (0 / false) is meaningful and distinct from "unset"
type CreateAgentParams struct {
	Name 					string		`json:"name"`
	Description				string		`json:"description,omitempty"`
	VoiceMode				string		`json:"voiceMode,omitempty"`
	SystemPrompt			string		`json:"systemPrompt,omitempty"`
	BeginMessage     		string 		`json:"beginMessage,omitempty"`
	Voice            		string 		`json:"voice,omitempty"`
	ModelTier        		string 		`json:"modelTier,omitempty"`
	STTMode          		string 		`json:"sttMode,omitempty"`
	AmbientSound     		string 		`json:"ambientSound,omitempty"`
	DenoisingMode    		string 		`json:"denoisingMode,omitempty"`
	TransferNumber   		string 		`json:"transferNumber,omitempty"`
	VoicemailMessage 		string 		`json:"voicemailMessage,omitempty"`
	Language         		string 		`json:"language,omitempty"`
	VoiceSpeed              *float64 	`json:"voiceSpeed,omitempty"`
	InterruptionSensitivity *float64 	`json:"interruptionSensitivity,omitempty"`
	EnableBackchannel       *bool    	`json:"enableBackchannel,omitempty"`
	MaxSilenceMs            *int     	`json:"maxSilenceMs,omitempty"`
	EnableMessaging         *bool    	`json:"enableMessaging,omitempty"`
}


// UpdateAgentParams are the parameters for PATCH /v1/agents/{id}.
type UpdateAgentParams struct {
	Name        			string   `json:"name,omitempty"`
	Description 			string   `json:"description,omitempty"`
	VoiceMode        		string   `json:"voiceMode,omitempty"`
	SystemPrompt     		string   `json:"systemPrompt,omitempty"`
	BeginMessage     		string   `json:"beginMessage,omitempty"`
	Voice            		string   `json:"voice,omitempty"`
	ModelTier        		string   `json:"modelTier,omitempty"`
	STTMode          		string   `json:"sttMode,omitempty"`
	AmbientSound     		string   `json:"ambientSound,omitempty"`
	DenoisingMode    		string   `json:"denoisingMode,omitempty"`
	TransferNumber   		string   `json:"transferNumber,omitempty"`
	VoicemailMessage 		string   `json:"voicemailMessage,omitempty"`
	Language         		string 	 `json:"language,omitempty"`
	VoiceSpeed              *float64 `json:"voiceSpeed,omitempty"`
	InterruptionSensitivity *float64 `json:"interruptionSensitivity,omitempty"`
	EnableBackchannel       *bool    `json:"enableBackchannel,omitempty"`
	MaxSilenceMs            *int     `json:"maxSilenceMs,omitempty"`
	EnableMessaging         *bool    `json:"enableMessaging,omitempty"`
}


// ListAgentsResponse is the response from List. Agents use offset-based pagination
type ListAgentsResponse struct {
	Agents []Agent	`json:"data"`
	OffsetPageInfo
}


// List returns the agents on the account. params may be nil.
func (s *AgentsService) List(ctx context.Context, params *ListParams) (*ListAgentsResponse, error) {
	var resp ListAgentsResponse
	err := s.client.request(ctx, http.MethodGet, "/agents"+params.toQuery(), nil, &resp)
	return &resp, err
}


// Create, creates a new agent.
func (s *AgentsService) Create(ctx context.Context, params *CreateAgentParams) (*Agent, error) {
	var agent Agent
	err := s.client.request(ctx, http.MethodPost, "/agents", params, &agent)
	return &agent, err
}


// Get retrieves a single agent, including its attached numbers.
func (s* AgentsService) Get(ctx context.Context, agentID string) (*Agent, error) {
	var agent Agent
	err := s.client.request(ctx, http.MethodGet, "/agents/"+agentID, nil, &agent)
	return &agent, err
}


// Update changes an agent's configuration. Only fields set on params are updated.
func (s *AgentsService) Update(ctx context.Context, agentID string, params *UpdateAgentParams) (*Agent, error) {
	var agent Agent
	err := s.client.request(ctx, http.MethodPatch, "/agents/"+agentID, params, &agent)
	return &agent, err
}


// Delete deletes an agent. Its numbers, conversations, and calls have their agent reference cleared but are not themselves deleted.
func (s *AgentsService) Delete(ctx context.Context, agentID string) error {
	return s.client.request(ctx, http.MethodDelete, "/agents/"+agentID, nil, nil)
}


// attachNumberParams is the internal request body for AttachNumber.
type attachNumberParams struct {
	NumberID 	string	`json:"numberId"`
}


// AttachNumber attaches an existing phone number to an agent. The number
// must belong to the same project and must not already be released.
func (s *AgentsService) AttachNumber(ctx context.Context, agentID, numberID string) (*Number, error) {
	var num Number
	err := s.client.request(ctx, http.MethodPost, "/agents/"+agentID+"/numbers", attachNumberParams{NumberID: numberID}, &num)
	return &num, err
}


// DetachNumber detaches a phone number from an agent.
func (s *AgentsService) DetachNumber(ctx context.Context, agentID, numberID string) error {
	return s.client.request(ctx, http.MethodDelete, "/agents/"+agentID+"/numbers/"+numberID, nil, nil)
}


// ListCalls returns the calls for a specific agent. params may be nil.
func (s *AgentsService) ListCalls(ctx context.Context, agentID string, params *ListParams) (*ListCallsResponse, error) {
	var resp ListCallsResponse
	err := s.client.request(ctx, http.MethodGet, "/agents/"+agentID+"/calls"+params.toQuery(), nil, &resp)
	return &resp, err
}


// Voice is a TTS voice that can be used with an agent's Voice field.
type Voice struct {
	VoiceID          string `json:"voice_id"`
	VoiceName        string `json:"voice_name"`
	Gender           string `json:"gender,omitempty"`
	Accent           string `json:"accent,omitempty"`
	PreviewAudioURL  string `json:"preview_audio_url,omitempty"`
}


// ListVoicesResponse is the response from ListVoices.
type ListVoicesResponse struct {
	Voices []Voice `json:"data"`
}


// ListVoices lists the TTS voices available for the Voice field on CreateAgentParams/UpdateAgentParams and CreateOutboundCallParams.
func (s *AgentsService) ListVoices(ctx context.Context) (*ListVoicesResponse, error) {
	var resp ListVoicesResponse
	err := s.client.request(ctx, http.MethodGet, "/agents/voices", nil, &resp)
	return &resp, err
}
