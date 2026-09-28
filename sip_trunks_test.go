package agentphone

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestSIPTrunksService_ListCreateGetUpdateDelete(t *testing.T) {
	var requests []string
	client, server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		requests = append(requests, r.Method+" "+r.URL.Path+"?"+r.URL.RawQuery)
		if r.Method == http.MethodPost {
			var body CreateSIPTrunkParams
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatalf("decode create body: %v", err)
			}
			if body.Name != "Primary" || body.Host != "sip.example.com" || body.Password != "secret" {
				t.Errorf("create body = %+v", body)
			}
		}
		if r.Method == http.MethodPatch {
			var body UpdateSIPTrunkParams
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatalf("decode update body: %v", err)
			}
			if body.Name != "Backup" || body.Port != 5061 {
				t.Errorf("update body = %+v", body)
			}
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"data":[{"id":"sip_1","name":"Primary","host":"sip.example.com","port":5060}],"hasMore":false,"total":1}`))
	})
	defer server.Close()

	list, err := client.SIPTrunks.List(context.Background(), &ListParams{Limit: 3})
	if err != nil {
		t.Fatalf("List() error: %v", err)
	}
	if len(list.SIPTrunks) != 1 || list.SIPTrunks[0].Port != 5060 || list.Total != 1 {
		t.Errorf("list response = %+v", list)
	}
	if _, err := client.SIPTrunks.Create(context.Background(), &CreateSIPTrunkParams{Name: "Primary", Host: "sip.example.com", Port: 5060, Password: "secret"}); err != nil {
		t.Fatalf("Create() error: %v", err)
	}
	if _, err := client.SIPTrunks.Get(context.Background(), "sip_1"); err != nil {
		t.Fatalf("Get() error: %v", err)
	}
	if _, err := client.SIPTrunks.Update(context.Background(), "sip_1", &UpdateSIPTrunkParams{Name: "Backup", Port: 5061}); err != nil {
		t.Fatalf("Update() error: %v", err)
	}
	if err := client.SIPTrunks.Delete(context.Background(), "sip_1"); err != nil {
		t.Fatalf("Delete() error: %v", err)
	}
	want := []string{"GET /sip-trunks?limit=3", "POST /sip-trunks?", "GET /sip-trunks/sip_1?", "PATCH /sip-trunks/sip_1?", "DELETE /sip-trunks/sip_1?"}
	if len(requests) != len(want) {
		t.Fatalf("requests = %v, want %v", requests, want)
	}
	for i := range want {
		if requests[i] != want[i] {
			t.Errorf("requests[%d] = %q, want %q", i, requests[i], want[i])
		}
	}
}
