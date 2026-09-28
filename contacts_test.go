package agentphone

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

func TestContactsService_List_BuildsFilterQuery(t *testing.T) {
	var gotQuery string
	client, server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"data":[{"id":"con_1","phoneNumber":"+14155550100","name":"Jane"}],"hasMore":true,"total":4}`))
	})
	defer server.Close()

	resp, err := client.Contacts.List(context.Background(), &ListContactsParams{Limit: 2, Offset: 4, Search: "+1415"})
	if err != nil {
		t.Fatalf("List() error: %v", err)
	}
	for _, want := range []string{"limit=2", "offset=4", "search=%2B1415"} {
		if !strings.Contains(gotQuery, want) {
			t.Errorf("query %q missing %q", gotQuery, want)
		}
	}
	if len(resp.Contacts) != 1 || resp.Contacts[0].Name != "Jane" || !resp.HasMore || resp.Total != 4 {
		t.Errorf("response = %+v", resp)
	}
}

func TestContactsService_CreateGetUpdateDelete(t *testing.T) {
	var requests []string
	client, server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		requests = append(requests, r.Method+" "+r.URL.Path)
		if r.Method == http.MethodPost {
			var body CreateContactParams
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatalf("decode create body: %v", err)
			}
			if body.PhoneNumber != "+14155550100" || body.Name != "Jane" {
				t.Errorf("create body = %+v", body)
			}
		}
		if r.Method == http.MethodPatch {
			var body UpdateContactParams
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatalf("decode update body: %v", err)
			}
			if body.Name != "Janet" || body.Notes != "VIP" {
				t.Errorf("update body = %+v", body)
			}
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":"con_1","phoneNumber":"+14155550100","name":"Janet","notes":"VIP"}`))
	})
	defer server.Close()

	if _, err := client.Contacts.Create(context.Background(), &CreateContactParams{PhoneNumber: "+14155550100", Name: "Jane"}); err != nil {
		t.Fatalf("Create() error: %v", err)
	}
	contact, err := client.Contacts.Get(context.Background(), "con_1")
	if err != nil {
		t.Fatalf("Get() error: %v", err)
	}
	if contact.ID != "con_1" {
		t.Errorf("contact = %+v", contact)
	}
	if _, err := client.Contacts.Update(context.Background(), "con_1", &UpdateContactParams{Name: "Janet", Notes: "VIP"}); err != nil {
		t.Fatalf("Update() error: %v", err)
	}
	if err := client.Contacts.Delete(context.Background(), "con_1"); err != nil {
		t.Fatalf("Delete() error: %v", err)
	}
	want := []string{"POST /contacts", "GET /contacts/con_1", "PATCH /contacts/con_1", "DELETE /contacts/con_1"}
	if len(requests) != len(want) {
		t.Fatalf("requests = %v, want %v", requests, want)
	}
	for i := range want {
		if requests[i] != want[i] {
			t.Errorf("requests[%d] = %q, want %q", i, requests[i], want[i])
		}
	}
}

func TestContactsService_GetCapabilities_EncodesPhoneNumber(t *testing.T) {
	var gotQuery string
	client, server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"phoneNumber":"+14155550100","sms":true,"imessage":false,"whatsapp":true,"voice":true}`))
	})
	defer server.Close()

	caps, err := client.Contacts.GetCapabilities(context.Background(), "+14155550100")
	if err != nil {
		t.Fatalf("GetCapabilities() error: %v", err)
	}
	if gotQuery != "phoneNumber=%2B14155550100" {
		t.Errorf("query = %q, want phoneNumber=%%2B14155550100", gotQuery)
	}
	if caps.PhoneNumber != "+14155550100" || !caps.SMS || caps.IMessage || !caps.WhatsApp || !caps.Voice {
		t.Errorf("capabilities = %+v", caps)
	}
}
