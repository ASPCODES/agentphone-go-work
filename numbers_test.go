package agentphone

import(
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
)


func TestNumbersService_List_BuildsQueryAndParsesResponse(t *testing.T) {
	var gotQuery string
	client, server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"data": [{"id": "num_1", "phoneNumber": "+15551234567", "agentId": "agt_1"}],
			"hasMore": true,
			"total": 5
		}`))
	})
	defer server.Close()
 
	resp, err := client.Numbers.List(context.Background(), &ListParams{Limit: 10, Offset: 20})
	if err != nil {
		t.Fatalf("List() error: %v", err)
	}
	if gotQuery != "limit=10&offset=20" {
		t.Errorf("query = %q, want limit=10&offset=20", gotQuery)
	}
	if len(resp.Numbers) != 1 || resp.Numbers[0].PhoneNumber != "+15551234567" {
		t.Errorf("Numbers = %+v", resp.Numbers)
	}
	if !resp.HasMore || resp.Total != 5 {
		t.Errorf("HasMore=%v Total=%d, want true/5", resp.HasMore, resp.Total)
	}
}
 
func TestNumbersService_List_NilParamsOmitsQuery(t *testing.T) {
	var gotQuery string
	client, server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"data": [], "hasMore": false}`))
	})
	defer server.Close()
 
	if _, err := client.Numbers.List(context.Background(), nil); err != nil {
		t.Fatalf("List(nil) error: %v", err)
	}
	if gotQuery != "" {
		t.Errorf("query = %q, want empty for nil params", gotQuery)
	}
}
 
func TestNumbersService_Create_SendsCamelCaseBody(t *testing.T) {
	client, server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		body := readBody(t, r)
		for _, want := range []string{`"areaCode"`, `"agentId"`} {
			if !strings.Contains(body, want) {
				t.Errorf("request body %s missing key %s", body, want)
			}
		}
		// PhoneNumber wasn't set, so omitempty must drop it entirely
		// rather than sending an empty string.
		if strings.Contains(body, `"phoneNumber"`) {
			t.Errorf("request body %s should omit unset phoneNumber", body)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id": "num_new", "phoneNumber": "+14155551212"}`))
	})
	defer server.Close()
 
	num, err := client.Numbers.Create(context.Background(), &CreateNumberParams{
		AreaCode: "415",
		AgentID:  "agt_1",
	})
	if err != nil {
		t.Fatalf("Create() error: %v", err)
	}
	if num.ID != "num_new" {
		t.Errorf("ID = %q, want num_new", num.ID)
	}
}
 
func TestNumbersService_Get_BuildsCorrectPath(t *testing.T) {
	var gotPath string
	client, server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id": "num_1"}`))
	})
	defer server.Close()
 
	if _, err := client.Numbers.Get(context.Background(), "num_1"); err != nil {
		t.Fatalf("Get() error: %v", err)
	}
	if gotPath != "/numbers/num_1" {
		t.Errorf("path = %q, want /numbers/num_1", gotPath)
	}
}
 
func TestNumbersService_Delete_UsesDeleteMethod(t *testing.T) {
	var gotMethod string
	client, server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		w.WriteHeader(http.StatusOK)
	})
	defer server.Close()
 
	if err := client.Numbers.Delete(context.Background(), "num_1"); err != nil {
		t.Fatalf("Delete() error: %v", err)
	}
	if gotMethod != http.MethodDelete {
		t.Errorf("method = %q, want DELETE", gotMethod)
	}
}
 
func TestNumbersService_Lookup_UsesPhoneNumberQueryParam(t *testing.T) {
	var gotQuery string
	client, server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"phoneNumber": "+15551234567", "valid": true, "lineType": "mobile"}`))
	})
	defer server.Close()
 
	result, err := client.Numbers.Lookup(context.Background(), "+15551234567")
	if err != nil {
		t.Fatalf("Lookup() error: %v", err)
	}
	if gotQuery != "phoneNumber=%2B15551234567" {
		t.Errorf("query = %q, want URL-encoded phoneNumber param", gotQuery)
	}
	if !result.Valid || result.LineType != "mobile" {
		t.Errorf("result = %+v", result)
	}
}
 
func TestNumbersService_GetMessages_UsesCursorPaginationNotOffset(t *testing.T) {
	var gotQuery string
	client, server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"data": [{"id": "msg_1", "receivedAt": "2026-01-01T00:00:00Z"}], "hasMore": true}`))
	})
	defer server.Close()
 
	resp, err := client.Numbers.GetMessages(context.Background(), "num_1", &CursorListParams{
		Limit:  50,
		Before: "2026-01-01T00:00:00Z",
	})
	if err != nil {
		t.Fatalf("GetMessages() error: %v", err)
	}
	// Cursor pagination uses before/after, never offset — this is the bug
	// we caught and fixed against the confirmed pagination docs.
	if strings.Contains(gotQuery, "offset") {
		t.Errorf("query = %q, should never contain offset (cursor pagination)", gotQuery)
	}
	if !strings.Contains(gotQuery, "before=") {
		t.Errorf("query = %q, want a before= param", gotQuery)
	}
	if len(resp.Messages) != 1 || resp.Messages[0].ReceivedAt != "2026-01-01T00:00:00Z" {
		t.Errorf("Messages = %+v", resp.Messages)
	}
}
 

// --- small shared test helper, used across resource test files ---
 
func readBody(t *testing.T, r *http.Request) string {
	t.Helper()
	data, err := io.ReadAll(r.Body)
	if err != nil {
		t.Fatalf("reading request body: %v", err)
	}
	return string(data)
}
