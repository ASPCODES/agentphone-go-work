package agentphone

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestWhatsAppService_ConnectionAndNumberOperations(t *testing.T) {
	var requests []string
	client, server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		requests = append(requests, r.Method+" "+r.URL.Path+"?"+r.URL.RawQuery)
		if r.URL.Path == "/integrations/whatsapp/connect" {
			var body ConnectWhatsAppParams
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatalf("decode connect body: %v", err)
			}
			if body.Code != "oauth-code" {
				t.Errorf("code = %q, want oauth-code", body.Code)
			}
			w.Write([]byte(`{"id":"wa_1","status":"connected"}`))
			return
		}
		if r.Method == http.MethodPost {
			var body ConnectWhatsAppNumberParams
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatalf("decode number body: %v", err)
			}
			if body.PhoneNumber != "+14155550100" {
				t.Errorf("phoneNumber = %q", body.PhoneNumber)
			}
			w.Write([]byte(`{"id":"num_1","phoneNumber":"+14155550100"}`))
			return
		}
		w.Write([]byte(`{"data":[{"phoneNumber":"+14155550100"}]}`))
	})
	defer server.Close()

	conn, err := client.WhatsApp.Connect(context.Background(), &ConnectWhatsAppParams{Code: "oauth-code"})
	if err != nil {
		t.Fatalf("Connect() error: %v", err)
	}
	available, err := client.WhatsApp.AvailableNumbers(context.Background(), "wa_1")
	if err != nil {
		t.Fatalf("AvailableNumbers() error: %v", err)
	}
	number, err := client.WhatsApp.ConnectNumber(context.Background(), "wa_1", &ConnectWhatsAppNumberParams{PhoneNumber: "+14155550100"})
	if err != nil {
		t.Fatalf("ConnectNumber() error: %v", err)
	}
	if conn.ID != "wa_1" || conn.Status != "connected" || len(available.Numbers) != 1 || number.ID != "num_1" || number.PhoneNumber != "+14155550100" {
		t.Errorf("connection = %+v, available = %+v, number = %+v", conn, available, number)
	}
	want := []string{
		"POST /integrations/whatsapp/connect?",
		"GET /integrations/whatsapp/wa_1/available-numbers?",
		"POST /integrations/whatsapp/wa_1/numbers?",
	}
	if len(requests) != len(want) {
		t.Fatalf("requests = %v, want %v", requests, want)
	}
	for i := range want {
		if requests[i] != want[i] {
			t.Errorf("requests[%d] = %q, want %q", i, requests[i], want[i])
		}
	}
}

func TestWhatsAppService_TemplateAndStatusOperations(t *testing.T) {
	var requests []string
	client, server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		requests = append(requests, r.Method+" "+r.URL.Path+"?"+r.URL.RawQuery)
		if r.Method == http.MethodPost {
			var body CreateWhatsAppTemplateParams
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatalf("decode template body: %v", err)
			}
			if body.Name != "order_update" || body.Category != "UTILITY" || body.Language != "en_US" || body.Body != "Order {{1}} shipped" {
				t.Errorf("template body = %+v", body)
			}
		}
		w.WriteHeader(http.StatusOK)
		switch r.URL.Path {
		case "/integrations/whatsapp/wa_1/templates":
			if r.Method == http.MethodGet {
				w.Write([]byte(`{"data":[{"id":"tmpl_1","name":"order_update","status":"approved","category":"UTILITY","language":"en_US"}]}`))
			} else {
				w.Write([]byte(`{"id":"tmpl_2","name":"order_update","status":"pending"}`))
			}
		case "/integrations/whatsapp/status":
			w.Write([]byte(`{"connected":true,"enabled":true,"status":"active"}`))
		}
	})
	defer server.Close()

	templates, err := client.WhatsApp.ListTemplates(context.Background(), "wa_1")
	if err != nil {
		t.Fatalf("ListTemplates() error: %v", err)
	}
	created, err := client.WhatsApp.CreateTemplate(context.Background(), "wa_1", &CreateWhatsAppTemplateParams{Name: "order_update", Category: "UTILITY", Language: "en_US", Body: "Order {{1}} shipped"})
	if err != nil {
		t.Fatalf("CreateTemplate() error: %v", err)
	}
	if err := client.WhatsApp.DeleteTemplate(context.Background(), "wa_1", "tmpl 1"); err != nil {
		t.Fatalf("DeleteTemplate() error: %v", err)
	}
	status, err := client.WhatsApp.Status(context.Background())
	if err != nil {
		t.Fatalf("Status() error: %v", err)
	}
	if err := client.WhatsApp.Disconnect(context.Background(), "wa_1"); err != nil {
		t.Fatalf("Disconnect() error: %v", err)
	}
	if len(templates.Templates) != 1 || templates.Templates[0].Status != "approved" || created.ID != "tmpl_2" || !status.Connected || !status.Enabled || status.Status != "active" {
		t.Errorf("templates = %+v, created = %+v, status = %+v", templates, created, status)
	}
	want := []string{
		"GET /integrations/whatsapp/wa_1/templates?",
		"POST /integrations/whatsapp/wa_1/templates?",
		"DELETE /integrations/whatsapp/wa_1/templates?templateId=tmpl+1",
		"GET /integrations/whatsapp/status?",
		"DELETE /integrations/whatsapp/wa_1?",
	}
	if len(requests) != len(want) {
		t.Fatalf("requests = %v, want %v", requests, want)
	}
	for i := range want {
		if requests[i] != want[i] {
			t.Errorf("requests[%d] = %q, want %q", i, requests[i], want[i])
		}
	}
}
