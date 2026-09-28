package agentphone

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestSubAccountsService_List_UsesOffsetPagination(t *testing.T) {
	var gotMethod, gotPath, gotQuery string
	client, server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		gotMethod, gotPath, gotQuery = r.Method, r.URL.Path, r.URL.RawQuery
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"data": [{"id": "sub_1", "name": "Support", "status": "active"}], "hasMore": true, "total": 12}`))
	})
	defer server.Close()

	resp, err := client.SubAccounts.List(context.Background(), &ListParams{Limit: 5, Offset: 10})
	if err != nil {
		t.Fatalf("List() error: %v", err)
	}
	if gotMethod != http.MethodGet || gotPath != "/sub-accounts" || gotQuery != "limit=5&offset=10" {
		t.Errorf("request = %s %s?%s, want GET /sub-accounts?limit=5&offset=10", gotMethod, gotPath, gotQuery)
	}
	if len(resp.SubAccounts) != 1 || resp.SubAccounts[0].ID != "sub_1" || resp.SubAccounts[0].Status != "active" || !resp.HasMore || resp.Total != 12 {
		t.Errorf("response = %+v", resp)
	}
}

func TestSubAccountsService_Create_SendsName(t *testing.T) {
	client, server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/sub-accounts" {
			t.Errorf("request = %s %s, want POST /sub-accounts", r.Method, r.URL.Path)
		}
		var body CreateSubAccountParams
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		if body.Name != "Support" {
			t.Errorf("request name = %q, want Support", body.Name)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id": "sub_1", "name": "Support", "status": "active"}`))
	})
	defer server.Close()

	sub, err := client.SubAccounts.Create(context.Background(), &CreateSubAccountParams{Name: "Support"})
	if err != nil {
		t.Fatalf("Create() error: %v", err)
	}
	if sub.ID != "sub_1" || sub.Name != "Support" || sub.Status != "active" {
		t.Errorf("sub-account = %+v", sub)
	}
}

func TestSubAccountsService_Update_UsesPatchAndSubAccountPath(t *testing.T) {
	client, server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch || r.URL.Path != "/sub-accounts/sub_1" {
			t.Errorf("request = %s %s, want PATCH /sub-accounts/sub_1", r.Method, r.URL.Path)
		}
		var body UpdateSubAccountParams
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		if body.Name != "Sales" || body.Status != "suspended" {
			t.Errorf("request body = %+v, want name Sales and status suspended", body)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id": "sub_1", "name": "Sales", "status": "suspended"}`))
	})
	defer server.Close()

	sub, err := client.SubAccounts.Update(context.Background(), "sub_1", &UpdateSubAccountParams{Name: "Sales", Status: "suspended"})
	if err != nil {
		t.Fatalf("Update() error: %v", err)
	}
	if sub.ID != "sub_1" || sub.Name != "Sales" || sub.Status != "suspended" {
		t.Errorf("sub-account = %+v", sub)
	}
}

func TestSubAccountsService_Delete_UsesSubAccountPath(t *testing.T) {
	var gotMethod, gotPath string
	client, server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		gotMethod, gotPath = r.Method, r.URL.Path
		w.WriteHeader(http.StatusNoContent)
	})
	defer server.Close()

	if err := client.SubAccounts.Delete(context.Background(), "sub_1"); err != nil {
		t.Fatalf("Delete() error: %v", err)
	}
	if gotMethod != http.MethodDelete || gotPath != "/sub-accounts/sub_1" {
		t.Errorf("request = %s %s, want DELETE /sub-accounts/sub_1", gotMethod, gotPath)
	}
}
