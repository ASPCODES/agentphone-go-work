package agentphone

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestContactCardsService_GetPutDelete(t *testing.T) {
	var requests []string
	client, server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		requests = append(requests, r.Method+" "+r.URL.Path)
		if r.Method == http.MethodPut {
			var body PutContactCardParams
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatalf("decode request body: %v", err)
			}
			if body.Name != "AgentPhone" || body.AvatarURL != "https://example.com/avatar.png" {
				t.Errorf("request body = %+v", body)
			}
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"numberId":"num_1","name":"AgentPhone","organization":"Support","avatarUrl":"https://example.com/avatar.png"}`))
	})
	defer server.Close()

	card, err := client.ContactCards.Get(context.Background(), "num_1")
	if err != nil {
		t.Fatalf("Get() error: %v", err)
	}
	if card.NumberID != "num_1" || card.Organization != "Support" {
		t.Errorf("card = %+v", card)
	}
	if _, err := client.ContactCards.Put(context.Background(), "num_1", &PutContactCardParams{
		Name: "AgentPhone", Organization: "Support", AvatarURL: "https://example.com/avatar.png",
	}); err != nil {
		t.Fatalf("Put() error: %v", err)
	}
	if err := client.ContactCards.Delete(context.Background(), "num_1"); err != nil {
		t.Fatalf("Delete() error: %v", err)
	}
	want := []string{"GET /numbers/num_1/contact-card", "PUT /numbers/num_1/contact-card", "DELETE /numbers/num_1/contact-card"}
	if len(requests) != len(want) {
		t.Fatalf("requests = %v, want %v", requests, want)
	}
	for i := range want {
		if requests[i] != want[i] {
			t.Errorf("requests[%d] = %q, want %q", i, requests[i], want[i])
		}
	}
}
