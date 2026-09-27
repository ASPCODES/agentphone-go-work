package agentphone

import(
	"context"
	"net/http"
	"strconv"
)


// ConversationsService handles the /conversations endpoints: threaded SMS/iMessage conversations between a number and a contact.
type ConversationsService struct {
	client *Client
}


// ConversationCapabilities describes what's currently possible on a conversation. WhatsappWindowExpiresAt is confirmed by the error-handling
type ConversationCapabilities struct {
	WhatsappWindowExpiresAt string `json:"whatsappWindowExpiresAt,omitempty"`
}


// Conversation represents a threaded conversation with a contact.
type Conversation struct {
	ID          		string `json:"id"`
	NumberID    		string `json:"numberId,omitempty"`
	AgentID     		string `json:"agentId,omitempty"`
	ContactID   		string `json:"contactId,omitempty"`
	PhoneNumber 		string `json:"phoneNumber,omitempty"` // the contact's number
	LastMessageSnippet 	string `json:"lastMessageSnippet,omitempty"`
	LastMessageAt       string `json:"lastMessageAt,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
	Capabilities *ConversationCapabilities `json:"capabilities,omitempty"`
	Messages []Message `json:"messages,omitempty"`
	CreatedAt string `json:"createdAt,omitempty"`
}


// ListConversationsResponse is the response from List. Conversations use offset-based pagination,
type ListConversationsResponse struct {
	Conversations []Conversation `json:"data"`
	OffsetPageInfo
}


// List returns the conversations on the account.
func (s *ConversationsService) List(ctx context.Context, params *ListParams) (*ListConversationsResponse, error) {
	var resp ListConversationsResponse
	err := s.client.request(ctx, http.MethodGet, "/conversations"+params.toQuery(), nil, &resp)
	return &resp, err
}


// Get retrieves a conversation. messageLimit caps how many recent messages are embedded in the response.
func (s *ConversationsService) Get(ctx context.Context, conversationID string, messageLimit int) (*Conversation, error) {
	path := "/conversations/" + conversationID
	if messageLimit != 0 {
		path += "?messageLimit=" + strconv.Itoa(messageLimit)
	}

	var convo Conversation
	err := s.client.request(ctx, http.MethodGet, path, nil, &convo)
	return &convo, err
}


// UpdateConversationParams are the parameters for updating a conversation's metadata.
type UpdateConversationParams struct {
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}


// Update sets a conversation's metadata.
func (s *ConversationsService) Update(ctx context.Context, conversationID string, params *UpdateConversationParams) (*Conversation, error) {
	var convo Conversation
	err := s.client.request(ctx, http.MethodPatch, "/conversations/"+conversationID, params, &convo)
	return &convo, err
}


// GetMessages returns a conversation's messages, newest first.
func (s *ConversationsService) GetMessages(ctx context.Context, conversationID string, params *CursorListParams) (*ListMessagesResponse, error) {
	var resp ListMessagesResponse
	err := s.client.request(ctx, http.MethodGet, "/conversations/"+conversationID+"/messages"+params.toQuery(), nil, &resp)
	return &resp, err
}


// SendTypingIndicator shows a "typing..." indicator to the contact.
func (s *ConversationsService) SendTypingIndicator(ctx context.Context, conversationID string) error {
	return s.client.request(ctx, http.MethodPost, "/conversations/"+conversationID+"/typing", nil, nil)
}

// SetChatBackgroundParams sets a conversation's chat background (iMessage).
type SetChatBackgroundParams struct {
	URL   string `json:"url,omitempty"`
	Color string `json:"color,omitempty"`
}


func (s *ConversationsService) SetChatBackground(ctx context.Context, conversationID string, params *SetChatBackgroundParams) error {
	return s.client.request(ctx, http.MethodPost, "/conversations/"+conversationID+"/background", params, nil)
}
 

// RemoveChatBackground clears a conversation's chat background.
func (s *ConversationsService) RemoveChatBackground(ctx context.Context, conversationID string) error {
	return s.client.request(ctx, http.MethodDelete, "/conversations/"+conversationID+"/background", nil, nil)
}
