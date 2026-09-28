package agentphone

import (
	"context"
	"net/http"
	"testing"
)

func TestUsageService_GetDailyMonthlyAndBreakdowns(t *testing.T) {
	var requests []string
	client, server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		requests = append(requests, r.Method+" "+r.URL.Path+"?"+r.URL.RawQuery)
		w.WriteHeader(http.StatusOK)
		switch r.URL.Path {
		case "/usage":
			w.Write([]byte(`{"plan":{"name":"Pro"},"numbers":{"used":2,"limit":10,"remaining":8},"stats":{"messagesLast24h":3,"messagesLast30d":12,"callsLast30d":4}}`))
		case "/usage/daily":
			w.Write([]byte(`{"data":[{"date":"2026-09-27","messages":5,"calls":2}]}`))
		case "/usage/monthly":
			w.Write([]byte(`{"data":[{"month":"2026-09","messages":50,"calls":20}]}`))
		case "/usage/by-number":
			w.Write([]byte(`{"data":[{"numberId":"num_1","messages":7,"calls":1}]}`))
		case "/usage/by-agent":
			w.Write([]byte(`{"data":[{"agentId":"agt_1","messages":8,"calls":2}]}`))
		}
	})
	defer server.Close()

	usage, err := client.Usage.Get(context.Background())
	if err != nil {
		t.Fatalf("Get() error: %v", err)
	}
	daily, err := client.Usage.GetDaily(context.Background(), 7)
	if err != nil {
		t.Fatalf("GetDaily() error: %v", err)
	}
	monthly, err := client.Usage.GetMonthly(context.Background(), 3)
	if err != nil {
		t.Fatalf("GetMonthly() error: %v", err)
	}
	byNumber, err := client.Usage.GetByNumber(context.Background())
	if err != nil {
		t.Fatalf("GetByNumber() error: %v", err)
	}
	byAgent, err := client.Usage.GetByAgent(context.Background())
	if err != nil {
		t.Fatalf("GetByAgent() error: %v", err)
	}
	if usage.Plan.Name != "Pro" || usage.Numbers.Remaining != 8 || usage.Stats.CallsLast30d != 4 {
		t.Errorf("usage = %+v", usage)
	}
	if len(daily.Data) != 1 || daily.Data[0].Messages != 5 || len(monthly.Data) != 1 || monthly.Data[0].Month != "2026-09" {
		t.Errorf("daily = %+v, monthly = %+v", daily, monthly)
	}
	if len(byNumber.Data) != 1 || byNumber.Data[0].NumberID != "num_1" || len(byAgent.Data) != 1 || byAgent.Data[0].AgentID != "agt_1" {
		t.Errorf("byNumber = %+v, byAgent = %+v", byNumber, byAgent)
	}
	want := []string{"GET /usage?", "GET /usage/daily?days=7", "GET /usage/monthly?months=3", "GET /usage/by-number?", "GET /usage/by-agent?"}
	if len(requests) != len(want) {
		t.Fatalf("requests = %v, want %v", requests, want)
	}
	for i := range want {
		if requests[i] != want[i] {
			t.Errorf("requests[%d] = %q, want %q", i, requests[i], want[i])
		}
	}
}

func TestUsageService_GetDailyZeroOmitsDays(t *testing.T) {
	var gotQuery string
	client, server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"data":[]}`))
	})
	defer server.Close()

	if _, err := client.Usage.GetDaily(context.Background(), 0); err != nil {
		t.Fatalf("GetDaily() error: %v", err)
	}
	if gotQuery != "" {
		t.Errorf("query = %q, want empty default query", gotQuery)
	}
}
