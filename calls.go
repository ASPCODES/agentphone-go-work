package agentphone

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// CallsService handles the /calls endpoints: placing and managing voice calls, and streaming their transcripts.
type CallsService struct {
	client *Client
}


// Call represents a voice call (inbound, outbound, or web).
type Call struct {
	ID              string `json:"id"`
	AgentID         string `json:"agentId"`
	PhoneNumberID   string `json:"phoneNumberId"`
	PhoneNumber     string `json:"phoneNumber"`
	FromNumber      string `json:"fromNumber"`
	ToNumber        string `json:"toNumber"`
	Direction       string `json:"direction"`
	Status          string `json:"status"`    
	StartedAt       string `json:"startedAt"`
	EndedAt         string `json:"endedAt,omitempty"`
	DurationSeconds int    `json:"durationSeconds,omitempty"`
	LastTranscriptSnippet string `json:"lastTranscriptSnippet,omitempty"`

	// RecordingURL and RecordingAvailable are only meaningful when the
	// call-recording add-on is enabled.
	RecordingURL       string `json:"recordingUrl,omitempty"`
	RecordingAvailable bool   `json:"recordingAvailable,omitempty"`
	Transcripts []Transcript `json:"transcripts,omitempty"`
}


type Transcript struct {
	ID         string  `json:"id"`
	Transcript string  `json:"transcript"`
	Confidence float64 `json:"confidence,omitempty"`
	Response   string  `json:"response,omitempty"`
	CreatedAt  string  `json:"createdAt,omitempty"`
}

// ListCallsResponse is the response from List. Calls use offset-based pagination
type ListCallsResponse struct {
	Calls []Call `json:"data"`
	OffsetPageInfo
}

// ListCallsParams filters and paginates List. All fields are optional.
type ListCallsParams struct {
	Limit  int
	Offset int
	Status string
	Direction string
	Search string
}

func (p *ListCallsParams) toQuery() string {
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
	if p.Status != "" {
		q.Set("status", p.Status)
	}
	if p.Direction != "" {
		q.Set("direction", p.Direction)
	}
	if p.Search != "" {
		q.Set("search", p.Search)
	}
	if encoded := q.Encode(); encoded != "" {
		return "?" + encoded
	}
	return ""
}

// List returns the calls on the account.
func (s *CallsService) List(ctx context.Context, params *ListCallsParams) (*ListCallsResponse, error) {
	var resp ListCallsResponse
	err := s.client.request(ctx, http.MethodGet, "/calls"+params.toQuery(), nil, &resp)
	return &resp, err
}

// CreateOutboundCallParams are the parameters for placing an outbound call. AgentID and ToNumber are required.
type CreateOutboundCallParams struct {
	AgentID  string `json:"agentId"`
	ToNumber string `json:"toNumber"` // E.164, e.g. "+15559876543"


	InitialGreeting string `json:"initialGreeting,omitempty"`
	Voice string `json:"voice,omitempty"`
	SystemPrompt string `json:"systemPrompt,omitempty"`
	FromNumberID string `json:"fromNumberId,omitempty"`
}


// Create places an outbound voice call.
func (s *CallsService) Create(ctx context.Context, params *CreateOutboundCallParams) (*Call, error) {
	var call Call
	err := s.client.request(ctx, http.MethodPost, "/calls", params, &call)
	return &call, err
}

// CreateWebCallParams are the parameters for minting a web-call access
// token. AgentID is required.
type CreateWebCallParams struct {
	AgentID string `json:"agentId"`
	// Variables are template variables for hosted-mode agents, referenced
	// in the system prompt as {{var_name}}.
	Variables map[string]interface{} `json:"variables,omitempty"`
}

// CreateWebCallResponse is the response from CreateWeb.
type CreateWebCallResponse struct {
	// AccessToken is valid for only 30 seconds — pass it to the frontend
	// immediately, which uses the agentphone-web-sdk npm package to start
	// the call with it.
	AccessToken string `json:"accessToken"`
	CallID      string `json:"callId"`
}

// CreateWeb mints a short-lived access token for a browser-based call via
// the agentphone-web-sdk. Your backend calls this, then hands the token to
// the frontend. The resulting call's Direction is "web".
func (s *CallsService) CreateWeb(ctx context.Context, params *CreateWebCallParams) (*CreateWebCallResponse, error) {
	var resp CreateWebCallResponse
	err := s.client.request(ctx, http.MethodPost, "/calls/web", params, &resp)
	return &resp, err
}

// Get retrieves a call, including its full transcript.
func (s *CallsService) Get(ctx context.Context, callID string) (*Call, error) {
	var call Call
	err := s.client.request(ctx, http.MethodGet, "/calls/"+callID, nil, &call)
	return &call, err
}

// End ends an in-progress call.
//
// NOTE: this endpoint is listed in the API reference but wasn't detailed
// in the guide docs — no request/response example was available. This
// assumes it takes no body and returns the updated Call; verify against a
// live API key.
func (s *CallsService) End(ctx context.Context, callID string) (*Call, error) {
	var call Call
	err := s.client.request(ctx, http.MethodPost, "/calls/"+callID+"/end", nil, &call)
	return &call, err
}


// CallRecording is the response from GetRecording.
type CallRecording struct {
	RecordingURL       string `json:"recordingUrl"`
	RecordingAvailable bool   `json:"recordingAvailable"`
}


// GetRecording retrieves recording info for a call.
func (s *CallsService) GetRecording(ctx context.Context, callID string) (*CallRecording, error) {
	var rec CallRecording
	err := s.client.request(ctx, http.MethodGet, "/calls/"+callID+"/recording", nil, &rec)
	return &rec, err
}


// ListTranscriptsResponse is the response from GetTranscript.
type ListTranscriptsResponse struct {
	Transcripts []Transcript `json:"transcripts"`
}


// GetTranscript retrieves just a call's transcript turns, without the rest of the call's fields.
func (s *CallsService) GetTranscript(ctx context.Context, callID string) (*ListTranscriptsResponse, error) {
	var resp ListTranscriptsResponse
	err := s.client.request(ctx, http.MethodGet, "/calls/"+callID+"/transcript", nil, &resp)
	return &resp, err
}


// ListForNumber returns the calls associated with a specific phone number (GET /v1/numbers/{number_id}/calls).
func (s *CallsService) ListForNumber(ctx context.Context, numberID string, params *ListCallsParams) (*ListCallsResponse, error) {
	var resp ListCallsResponse
	err := s.client.request(ctx, http.MethodGet, "/numbers/"+numberID+"/calls"+params.toQuery(), nil, &resp)
	return &resp, err
}

// --- Transcript streaming (Server-Sent Events) ---

type TranscriptEvent struct {
	Type      string // "connected", "turn", or "ended"
	Connected *TranscriptConnectedEvent
	Turn      *TranscriptTurnEvent
	Ended     *TranscriptEndedEvent
}

// TranscriptConnectedEvent is sent once when the stream connects.
type TranscriptConnectedEvent struct {
	CallID     string `json:"callId"`
	Status     string `json:"status"`
	AgentID    string `json:"agentId"`
	AgentName  string `json:"agentName"`
	Direction  string `json:"direction"`
	FromNumber string `json:"fromNumber"`
	ToNumber   string `json:"toNumber"`
	StartedAt  string `json:"startedAt"`
}


// TranscriptTurnEvent is one transcript turn, replayed from history or arriving live.
type TranscriptTurnEvent struct {
	Role      string `json:"role"` // "user" | "agent"
	Content   string `json:"content"`
	CreatedAt string `json:"createdAt"`
}


// TranscriptEndedEvent is sent once the call has ended; the stream closes immediately after.
type TranscriptEndedEvent struct {
	CallID          string `json:"callId"`
	Status          string `json:"status"`
	EndedAt         string `json:"endedAt"`
	DurationSeconds int    `json:"durationSeconds"`
}


// TranscriptStreamHandler is called once, synchronously, for each event as it arrives (in order). Return a non-nil error to stop the stream early, that error is then returned from StreamTranscript.
type TranscriptStreamHandler func(event TranscriptEvent) error

// StreamTranscript streams a call's transcript in real time via Server-Sent Events. On connect, the server replays every existing turn, then for a live call, keeps streaming new turns until the call ends;for an already-completed call, it sends the "ended" event immediately after replaying history and the stream closes. 
func (s *CallsService) StreamTranscript(ctx context.Context, callID string, handler TranscriptStreamHandler) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.client.baseURL+"/calls/"+callID+"/transcript/stream", nil)
	if err != nil {
		return fmt.Errorf("agentphone: building request: %w", err)
	}
	req.Header.Set("Accept", "text/event-stream")
	s.client.setAuthHeader(req)

	resp, err := s.client.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("agentphone: sending request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return parseAPIError(resp.StatusCode, body, resp.Header.Get("Retry-After"))
	}

	scanner := bufio.NewScanner(resp.Body)
	var eventType string

	for scanner.Scan() {
		line := scanner.Text()

		switch {
		case strings.HasPrefix(line, "event:"):
			eventType = strings.TrimSpace(strings.TrimPrefix(line, "event:"))
		case strings.HasPrefix(line, "data:"):
			data := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
			event, err := parseTranscriptEvent(eventType, []byte(data))
			if err != nil {
				return fmt.Errorf("agentphone: decoding stream event: %w", err)
			}
			if err := handler(event); err != nil {
				return err
			}
		}
		// Blank lines (SSE event separators) and ": heartbeat" comment lines match neither prefix above and are silently ignored.
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("agentphone: reading stream: %w", err)
	}
	return nil
}

func parseTranscriptEvent(eventType string, data []byte) (TranscriptEvent, error) {
	event := TranscriptEvent{Type: eventType}

	switch eventType {
	case "connected":
		var e TranscriptConnectedEvent
		if err := json.Unmarshal(data, &e); err != nil {
			return event, err
		}
		event.Connected = &e
	case "turn":
		var e TranscriptTurnEvent
		if err := json.Unmarshal(data, &e); err != nil {
			return event, err
		}
		event.Turn = &e
	case "ended":
		var e TranscriptEndedEvent
		if err := json.Unmarshal(data, &e); err != nil {
			return event, err
		}
		event.Ended = &e
	}

	return event, nil
}
