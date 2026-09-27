package agentphone

import(
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"
	"time"
)
// This file verifies INCOMING webhook deliveries (the requests AgentPhone
// sends to your server) — a different concern from WebhooksService in
// webhooks.go, which manages your webhook *configuration* via the API.


// WebhookVerificationError indicates a webhook delivery's signature is
// invalid, or its timestamp fell outside the allowed replay window.
type WebhookVerificationError struct {
	Message		string
}


func (e *WebhookVerificationError) Error() string {
	return "agentphone: webhook verification failed: " + e.Message
}


// DefaultWebhookTolerance is the default max age (5 minutes) allowed between a delivery's timestamp and now
const DefaultWebhookTolerance = 5 * time.Minute


// VerifyWebhook checks a webhook delivery's signature and, unless disabled, that its timestamp is recent enough to rule out a replayed request.
func VerifyWebhook(payload []byte, signature, secret, timestamp string, tolerance time.Duration) error {
	if tolerance == 0 {
		tolerance = DefaultWebhookTolerance
	}

	if tolerance > 0 {
		ts, err := strconv.ParseInt(timestamp, 10, 64)
		if err != nil {
			return &WebhookVerificationError{Message: "invalid or missing timestamp"}
		}
		age := time.Since(time.Unix(ts, 0))
		if age < 0 {
			age = -age
		}
		if age > tolerance {
			return &WebhookVerificationError{Message: "timestamp outside the allowed tolerance (possible replay)"}
		}
	}

	expected := signWebhookPayload(secret, timestamp, payload)
	if !hmac.Equal([]byte(expected), []byte(signature)) {
		return &WebhookVerificationError{Message: "signature mismatch"}
	}
	return nil
}


// signWebhookPayload computes the expected signature for a delivery.
func signWebhookPayload(secret, timestamp string, payload []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(timestamp))
	mac.Write([]byte(""))
	mac.Write(payload)
	return hex.EncodeToString(mac.Sum(nil))
}


// WebhookHistoryItem is one turn of recent conversation context included on a webhook event, for use as LLM context.
type WebhookHistoryItem struct {
	Direction string `json:"direction"`  // "inbound" or "outbound".
	Content   string `json:"content"`
}


// WebhookEventData carries the event-specific payload. Which fields are set depends on Channel: voice events carry Transcript
type WebhookEventData struct {
	Message    string `json:"message,omitempty"`
	Transcript string `json:"transcript,omitempty"`
	FromNumber string `json:"fromNumber,omitempty"`
}

// WebhookEvent is an incoming webhook delivery, parsed by ConstructEvent.
type WebhookEvent struct {
	Event 			  string				 `json:"event"`
	Channel           string                 `json:"channel"`
	Data              WebhookEventData       `json:"data"`
	RecentHistory     []WebhookHistoryItem   `json:"recentHistory,omitempty"`
	ConversationState map[string]interface{} `json:"conversationState,omitempty"`
}


// ConstructEvent verifies a webhook delivery (see VerifyWebhook) and, only if verification succeeds, parses its body into a WebhookEvent. Use this instead of decoding the body yourself so you never process an unverified payload.
func ConstructEvent(payload []byte, signature, secret, timestamp string, tolerance time.Duration) (*WebhookEvent, error) {
	if err := VerifyWebhook(payload, signature, secret, timestamp, tolerance); err != nil {
		return nil, err
	}

	var event WebhookEvent
	if err := json.Unmarshal(payload, &event); err != nil {
		return nil, fmt.Errorf("agentphone: decoding webhook event: %w", err)
	}
	return &event, nil
}
