package agentphone

import (
	"context"
	"net/http"
)

// MessagesService handles the /messages endpoints: sending SMS/iMessage
// and reacting to messages.
type MessagesService struct {
	client *Client
}


// Message represents a single SMS/iMessage message.
type Message struct {
	ID             string `json:"id"`
	ConversationID string `json:"conversationId,omitempty"`
	FromNumber     string `json:"fromNumber,omitempty"`
	ToNumber       string `json:"toNumber,omitempty"`
	Body           string `json:"body"`
	Direction      string `json:"direction,omitempty"` // "inbound" | "outbound"
	Status string `json:"status,omitempty"`
	FailureReason string `json:"failureReason,omitempty"`
	ReceivedAt string `json:"receivedAt,omitempty"`
}


// ListMessagesResponse is the response shape used by endpoints that return a page of messages (e.g. NumbersService.GetMessages). 
type ListMessagesResponse struct {
	Messages []Message `json:"data"`
	CursorPageInfo
}


// SendMessageParams are the parameters for sending a message. Exactly how
// the sender is identified (AgentID vs. NumberID)
type SendMessageParams struct {
	AgentID  string `json:"agentId,omitempty"`
	NumberID string `json:"numberId,omitempty"`
	ToNumber string `json:"toNumber"`
	Body     string `json:"body"`
	// MediaURL attaches an image for MMS/iMessage. Optional.
	MediaURL string `json:"mediaUrl,omitempty"`
}


// Send sends a new message.
func (s *MessagesService) Send(ctx context.Context, params *SendMessageParams) (*Message, error) {
	var msg Message
	err := s.client.request(ctx, http.MethodPost, "/messages", params, &msg)
	return &msg, err
}

// Reaction represents a tapback reaction sent to a message (iMessage only).
type Reaction struct {
	ID        string `json:"id"`
	MessageID string `json:"messageId"`
	Reaction  string `json:"reaction"`
	CreatedAt string `json:"createdAt,omitempty"`
}


// SendReactionParams are the parameters for reacting to a message.
// Reaction is one of: "love", "like", "dislike", "laugh", "emphasize", "question". Not a literal emoji character, despite the endpoint's name
type SendReactionParams struct {
	Reaction string `json:"reaction"`
}


// SendReaction sends a tapback reaction to an existing message.
func (s *MessagesService) SendReaction(ctx context.Context, messageID string, params *SendReactionParams) (*Reaction, error) {
	var reaction Reaction
	err := s.client.request(ctx, http.MethodPost, "/messages/"+messageID+"/reactions", params, &reaction)
	return &reaction, err
}
