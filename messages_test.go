package agentphone

import (
	"context"
	"net/http"
	"strings"
	"testing"
)

func TestMessagesService_Send_SendsCamelCaseBodyIncludingMediaURL(t *testing.T) {
	var gotMethod, gotPath, gotBody string
	client, server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		gotMethod, gotPath, gotBody = r.Method, r.URL.Path, readBody(t, r)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id": "msg_1", "status": "queued", "body": "hi"}`))
	})
	defer server.Close()

	msg, err := client.Messages.Send(context.Background(), &SendMessageParams{
		AgentID:  "agt_1",
		ToNumber: "+15559876543",
		Body:     "hi",
		MediaURL: "https://example.com/image.png",
	})
	if err != nil {
		t.Fatalf("Send() error: %v", err)
	}
	if gotMethod != http.MethodPost || gotPath != "/messages" {
		t.Errorf("got %s %s, want POST /messages", gotMethod, gotPath)
	}
	for _, want := range []string{`"agentId"`, `"toNumber"`, `"mediaUrl"`, `"body"`} {
		if !strings.Contains(gotBody, want) {
			t.Errorf("request body %s missing key %s", gotBody, want)
		}
	}
	if msg.ID != "msg_1" {
		t.Errorf("ID = %q, want msg_1", msg.ID)
	}
}

func TestMessagesService_Send_OmitsUnsetOptionalFields(t *testing.T) {
	var gotBody string
	client, server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		gotBody = readBody(t, r)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id": "msg_1"}`))
	})
	defer server.Close()

	_, err := client.Messages.Send(context.Background(), &SendMessageParams{
		ToNumber: "+15559876543",
		Body:     "hi",
	})
	if err != nil {
		t.Fatalf("Send() error: %v", err)
	}
	for _, unwanted := range []string{`"mediaUrl"`, `"agentId"`, `"numberId"`} {
		if strings.Contains(gotBody, unwanted) {
			t.Errorf("request body %s should omit unset %s", gotBody, unwanted)
		}
	}
}

func TestMessagesService_Send_DeliveryFailureReasonIsParsed(t *testing.T) {
	// Delivery failures are NOT API errors: the send succeeds (200/201) and the failure shows up on the message itself, later.
	client, server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id": "msg_1", "status": "failed", "failureReason": "The recipient has opted out of messages from this number."}`))
	})
	defer server.Close()

	msg, err := client.Messages.Send(context.Background(), &SendMessageParams{ToNumber: "+1", Body: "x"})
	if err != nil {
		t.Fatalf("Send() should not return an error for a message that fails delivery later: %v", err)
	}
	if msg.Status != "failed" || msg.FailureReason == "" {
		t.Errorf("Status=%q FailureReason=%q, want failed + a reason", msg.Status, msg.FailureReason)
	}
}

func TestMessagesService_SendReaction_UsesReactionFieldAndPath(t *testing.T) {
	var gotPath, gotBody string
	client, server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		gotPath, gotBody = r.URL.Path, readBody(t, r)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id": "rxn_1", "messageId": "msg_1", "reaction": "love"}`))
	})
	defer server.Close()

	rxn, err := client.Messages.SendReaction(context.Background(), "msg_1", &SendReactionParams{Reaction: "love"})
	if err != nil {
		t.Fatalf("SendReaction() error: %v", err)
	}
	if gotPath != "/messages/msg_1/reactions" {
		t.Errorf("path = %q, want /messages/msg_1/reactions", gotPath)
	}
	// The field is "reaction" (a fixed enum like "love")
	if !strings.Contains(gotBody, `"reaction":"love"`) || strings.Contains(gotBody, `"emoji"`) {
		t.Errorf("request body = %s, want reaction=love and no emoji key", gotBody)
	}
	if rxn.Reaction != "love" || rxn.MessageID != "msg_1" {
		t.Errorf("reaction = %+v", rxn)
	}
}
