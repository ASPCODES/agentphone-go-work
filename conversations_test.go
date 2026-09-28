package agentphone

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestConversationsService_ListAndGet(t *testing.T) {
	var requests []string
	client, server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		requests = append(requests, r.Method+" "+r.URL.Path+"?"+r.URL.RawQuery)
		w.WriteHeader(http.StatusOK)
		if r.URL.Path == "/conversations" {
			w.Write([]byte(`{"data":[{"id":"conv_1","agentId":"agt_1"}],"hasMore":false,"total":1}`))
			return
		}
		w.Write([]byte(`{"id":"conv_1","metadata":{"orderId":"ORD-1"},"messages":[{"id":"msg_1","body":"Hello"}],"capabilities":{"whatsappWindowExpiresAt":"2026-01-01T00:00:00Z"}}`))
	})
	defer server.Close()

	list, err := client.Conversations.List(context.Background(), &ListParams{Limit: 5})
	if err != nil {
		t.Fatalf("List() error: %v", err)
	}
	if len(list.Conversations) != 1 || list.Conversations[0].AgentID != "agt_1" || list.Total != 1 {
		t.Errorf("list response = %+v", list)
	}
	convo, err := client.Conversations.Get(context.Background(), "conv_1", 10)
	if err != nil {
		t.Fatalf("Get() error: %v", err)
	}
	if convo.ID != "conv_1" || len(convo.Messages) != 1 || convo.Capabilities == nil || convo.Metadata["orderId"] != "ORD-1" {
		t.Errorf("conversation = %+v", convo)
	}
	want := []string{"GET /conversations?limit=5", "GET /conversations/conv_1?messageLimit=10"}
	if len(requests) != len(want) {
		t.Fatalf("requests = %v, want %v", requests, want)
	}
	for i := range want {
		if requests[i] != want[i] {
			t.Errorf("requests[%d] = %q, want %q", i, requests[i], want[i])
		}
	}
}

func TestConversationsService_UpdateAndGetMessages(t *testing.T) {
	var requests []string
	client, server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		requests = append(requests, r.Method+" "+r.URL.Path+"?"+r.URL.RawQuery)
		if r.Method == http.MethodPatch {
			var body UpdateConversationParams
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatalf("decode update body: %v", err)
			}
			if body.Metadata["orderId"] != "ORD-2" {
				t.Errorf("metadata = %v", body.Metadata)
			}
			w.Write([]byte(`{"id":"conv_1","metadata":{"orderId":"ORD-2"}}`))
			return
		}
		w.Write([]byte(`{"data":[{"id":"msg_1","body":"Hello"}],"hasMore":true}`))
	})
	defer server.Close()

	if _, err := client.Conversations.Update(context.Background(), "conv_1", &UpdateConversationParams{Metadata: map[string]interface{}{"orderId": "ORD-2"}}); err != nil {
		t.Fatalf("Update() error: %v", err)
	}
	messages, err := client.Conversations.GetMessages(context.Background(), "conv_1", &CursorListParams{Limit: 2, Before: "cursor /1"})
	if err != nil {
		t.Fatalf("GetMessages() error: %v", err)
	}
	if len(messages.Messages) != 1 || messages.Messages[0].Body != "Hello" || !messages.HasMore {
		t.Errorf("messages = %+v", messages)
	}
	want := []string{"PATCH /conversations/conv_1?", "GET /conversations/conv_1/messages?before=cursor+%2F1&limit=2"}
	if len(requests) != len(want) {
		t.Fatalf("requests = %v, want %v", requests, want)
	}
	for i := range want {
		if requests[i] != want[i] {
			t.Errorf("requests[%d] = %q, want %q", i, requests[i], want[i])
		}
	}
}

func TestConversationsService_TypingAndBackgroundPaths(t *testing.T) {
	var requests []string
	client, server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		requests = append(requests, r.Method+" "+r.URL.Path)
		w.WriteHeader(http.StatusOK)
	})
	defer server.Close()

	if err := client.Conversations.SendTypingIndicator(context.Background(), "conv_1"); err != nil {
		t.Fatalf("SendTypingIndicator() error: %v", err)
	}
	if err := client.Conversations.SetChatBackground(context.Background(), "conv_1", &SetChatBackgroundParams{Color: "#ffffff"}); err != nil {
		t.Fatalf("SetChatBackground() error: %v", err)
	}
	if err := client.Conversations.RemoveChatBackground(context.Background(), "conv_1"); err != nil {
		t.Fatalf("RemoveChatBackground() error: %v", err)
	}
	want := []string{"POST /conversations/conv_1/typing", "POST /conversations/conv_1/background", "DELETE /conversations/conv_1/background"}
	if len(requests) != len(want) {
		t.Fatalf("requests = %v, want %v", requests, want)
	}
	for i := range want {
		if requests[i] != want[i] {
			t.Errorf("requests[%d] = %q, want %q", i, requests[i], want[i])
		}
	}
}
