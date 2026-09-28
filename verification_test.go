package agentphone

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestVerificationService_SendAndCheck(t *testing.T) {
	var requests []string
	client, server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		requests = append(requests, r.Method+" "+r.URL.Path)
		var body map[string]string
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		if body["phoneNumber"] != "+14155550100" {
			t.Errorf("phoneNumber = %q", body["phoneNumber"])
		}
		w.WriteHeader(http.StatusOK)
		if r.URL.Path == "/verify/send" {
			if body["channel"] != "sms" {
				t.Errorf("channel = %q, want sms", body["channel"])
			}
			w.Write([]byte(`{"status":"pending","phoneNumber":"+14155550100"}`))
			return
		}
		if body["code"] != "123456" {
			t.Errorf("code = %q, want 123456", body["code"])
		}
		w.Write([]byte(`{"status":"approved","valid":true}`))
	})
	defer server.Close()

	sent, err := client.Verification.Send(context.Background(), &SendVerificationParams{PhoneNumber: "+14155550100", Channel: "sms"})
	if err != nil {
		t.Fatalf("Send() error: %v", err)
	}
	checked, err := client.Verification.Check(context.Background(), &CheckVerificationParams{PhoneNumber: "+14155550100", Code: "123456"})
	if err != nil {
		t.Fatalf("Check() error: %v", err)
	}
	if sent.Status != "pending" || sent.PhoneNumber != "+14155550100" || checked.Status != "approved" || !checked.Valid {
		t.Errorf("send response = %+v, check response = %+v", sent, checked)
	}
	want := []string{"POST /verify/send", "POST /verify/check"}
	if len(requests) != len(want) {
		t.Fatalf("requests = %v, want %v", requests, want)
	}
	for i := range want {
		if requests[i] != want[i] {
			t.Errorf("requests[%d] = %q, want %q", i, requests[i], want[i])
		}
	}
}
