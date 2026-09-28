package agentphone

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestWebhooksService_GetCreateOrUpdateDelete(t *testing.T) {
	var requests []string
	client, server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		requests = append(requests, r.Method+" "+r.URL.Path)
		if r.Method == http.MethodPost {
			var body CreateOrUpdateWebhookParams
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatalf("decode webhook body: %v", err)
			}
			if body.URL != "https://example.com/hooks" || body.ContextLimit == nil || *body.ContextLimit != 0 || body.Timeout == nil || *body.Timeout != 0 {
				t.Errorf("webhook request = %+v; expected explicit zero values", body)
			}
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"url":"https://example.com/hooks","secret":"whsec_test","contextLimit":0,"timeout":0}`))
	})
	defer server.Close()

	webhook, err := client.Webhooks.Get(context.Background())
	if err != nil {
		t.Fatalf("Get() error: %v", err)
	}
	if webhook.URL != "https://example.com/hooks" {
		t.Errorf("webhook = %+v", webhook)
	}
	if _, err := client.Webhooks.CreateOrUpdate(context.Background(), &CreateOrUpdateWebhookParams{URL: "https://example.com/hooks", ContextLimit: Int(0), Timeout: Int(0)}); err != nil {
		t.Fatalf("CreateOrUpdate() error: %v", err)
	}
	if err := client.Webhooks.Delete(context.Background()); err != nil {
		t.Fatalf("Delete() error: %v", err)
	}
	want := []string{"GET /webhooks", "POST /webhooks", "DELETE /webhooks"}
	if len(requests) != len(want) {
		t.Fatalf("requests = %v, want %v", requests, want)
	}
	for i := range want {
		if requests[i] != want[i] {
			t.Errorf("requests[%d] = %q, want %q", i, requests[i], want[i])
		}
	}
}

func TestWebhooksService_DeliveryListsAndStats(t *testing.T) {
	var requests []string
	client, server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		requests = append(requests, r.Method+" "+r.URL.Path+"?"+r.URL.RawQuery)
		w.WriteHeader(http.StatusOK)
		switch r.URL.Path {
		case "/webhooks/deliveries":
			w.Write([]byte(`{"data":[{"id":"del_1","event":"message.received","statusCode":200,"success":true}],"hasMore":false,"total":1}`))
		case "/webhooks/deliveries/stats":
			w.Write([]byte(`{"successRate":0.9,"total":10,"successful":9,"failed":1,"hours":24}`))
		case "/webhooks/deliveries/all-time":
			w.Write([]byte(`{"successRate":0.95,"total":100,"successful":95,"failed":5}`))
		}
	})
	defer server.Close()

	deliveries, err := client.Webhooks.ListDeliveries(context.Background(), &ListParams{Limit: 5, Offset: 10})
	if err != nil {
		t.Fatalf("ListDeliveries() error: %v", err)
	}
	stats, err := client.Webhooks.GetDeliveryStats(context.Background(), 24)
	if err != nil {
		t.Fatalf("GetDeliveryStats() error: %v", err)
	}
	allTime, err := client.Webhooks.GetAllTimeStats(context.Background())
	if err != nil {
		t.Fatalf("GetAllTimeStats() error: %v", err)
	}
	if len(deliveries.Deliveries) != 1 || deliveries.Deliveries[0].StatusCode != 200 || deliveries.Total != 1 {
		t.Errorf("deliveries = %+v", deliveries)
	}
	if stats.SuccessRate != 0.9 || stats.Hours != 24 || allTime.Total != 100 || allTime.Successful != 95 {
		t.Errorf("stats = %+v, allTime = %+v", stats, allTime)
	}
	want := []string{"GET /webhooks/deliveries?limit=5&offset=10", "GET /webhooks/deliveries/stats?hours=24", "GET /webhooks/deliveries/all-time?"}
	if len(requests) != len(want) {
		t.Fatalf("requests = %v, want %v", requests, want)
	}
	for i := range want {
		if requests[i] != want[i] {
			t.Errorf("requests[%d] = %q, want %q", i, requests[i], want[i])
		}
	}
}

func TestWebhooksService_TestSendsOptionalAgent(t *testing.T) {
	client, server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/webhooks/test" {
			t.Errorf("request = %s %s, want POST /webhooks/test", r.Method, r.URL.Path)
		}
		var body TestWebhookParams
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		if body.AgentID != "agt_1" {
			t.Errorf("agentId = %q, want agt_1", body.AgentID)
		}
		w.WriteHeader(http.StatusNoContent)
	})
	defer server.Close()

	if err := client.Webhooks.Test(context.Background(), &TestWebhookParams{AgentID: "agt_1"}); err != nil {
		t.Fatalf("Test() error: %v", err)
	}
}
