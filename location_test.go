package agentphone

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestLocationService_Get_EscapesPhoneNumber(t *testing.T) {
	var gotPath string
	client, server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.EscapedPath()
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"phoneNumber":"+14155550100","city":"San Francisco","state":"CA","country":"US","timezone":"America/Los_Angeles"}`))
	})
	defer server.Close()

	loc, err := client.Location.Get(context.Background(), "+14155550100")
	if err != nil {
		t.Fatalf("Get() error: %v", err)
	}
	if gotPath != "/location/+14155550100" {
		t.Errorf("path = %q, want /location/+14155550100", gotPath)
	}
	if loc.City != "San Francisco" || loc.Timezone != "America/Los_Angeles" {
		t.Errorf("location = %+v", loc)
	}
}

func TestLocationService_Refresh_SendsNumberIDs(t *testing.T) {
	client, server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/location/refresh" {
			t.Errorf("request = %s %s, want POST /location/refresh", r.Method, r.URL.Path)
		}
		var body RefreshLocationsParams
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		if len(body.NumberIDs) != 2 || body.NumberIDs[0] != "num_1" || body.NumberIDs[1] != "num_2" {
			t.Errorf("numberIds = %v", body.NumberIDs)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"refreshed":2}`))
	})
	defer server.Close()

	resp, err := client.Location.Refresh(context.Background(), &RefreshLocationsParams{NumberIDs: []string{"num_1", "num_2"}})
	if err != nil {
		t.Fatalf("Refresh() error: %v", err)
	}
	if resp.Refreshed != 2 {
		t.Errorf("refreshed = %d, want 2", resp.Refreshed)
	}
}
