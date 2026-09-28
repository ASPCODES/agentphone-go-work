package agentphone

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestRegistrationService_GetStatusAndDetails(t *testing.T) {
	var requests []string
	client, server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		requests = append(requests, r.Method+" "+r.URL.Path)
		w.WriteHeader(http.StatusOK)
		if r.URL.Path == "/register/status" {
			w.Write([]byte(`{"status":"approved"}`))
			return
		}
		w.Write([]byte(`{"status":"approved","brand":{"legalName":"Example Inc","ein":"12-3456789"},"campaign":{"useCase":"customer-care","sampleMessages":["Hello"]}}`))
	})
	defer server.Close()

	status, err := client.Registration.GetStatus(context.Background())
	if err != nil {
		t.Fatalf("GetStatus() error: %v", err)
	}
	reg, err := client.Registration.GetDetails(context.Background())
	if err != nil {
		t.Fatalf("GetDetails() error: %v", err)
	}
	if status.Status != "approved" || reg.Brand == nil || reg.Brand.LegalName != "Example Inc" || reg.Campaign == nil || reg.Campaign.SampleMessages[0] != "Hello" {
		t.Errorf("status = %+v, registration = %+v", status, reg)
	}
	want := []string{"GET /register/status", "GET /register"}
	if len(requests) != len(want) {
		t.Fatalf("requests = %v, want %v", requests, want)
	}
	for i := range want {
		if requests[i] != want[i] {
			t.Errorf("requests[%d] = %q, want %q", i, requests[i], want[i])
		}
	}
}

func TestRegistrationService_RegisterAndUpdate(t *testing.T) {
	var requests []string
	client, server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		requests = append(requests, r.Method+" "+r.URL.Path)
		var body map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		if body["legalName"] != "Example Inc" || body["useCase"] != "customer-care" {
			t.Errorf("request body = %v", body)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"pending"}`))
	})
	defer server.Close()

	if _, err := client.Registration.Register(context.Background(), &RegisterA2PParams{LegalName: "Example Inc", UseCase: "customer-care"}); err != nil {
		t.Fatalf("Register() error: %v", err)
	}
	if _, err := client.Registration.Update(context.Background(), &UpdateA2PRegistrationParams{LegalName: "Example Inc", UseCase: "customer-care"}); err != nil {
		t.Fatalf("Update() error: %v", err)
	}
	want := []string{"POST /register", "PATCH /register"}
	if len(requests) != len(want) {
		t.Fatalf("requests = %v, want %v", requests, want)
	}
	for i := range want {
		if requests[i] != want[i] {
			t.Errorf("requests[%d] = %q, want %q", i, requests[i], want[i])
		}
	}
}
