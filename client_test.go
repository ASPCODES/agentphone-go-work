package agentphone

import(
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// newTestServer spins up a fake API server and returns a *Client pointed at it, plus the server itself so each test can control what it returns.Every resource file's test uses this same helper — it's the foundation everything else builds on.
func newTestServer(handler http.HandlerFunc) (*Client, *httptest.Server) {
	server := httptest.NewServer(handler)
	client := NewClient("test-api-key", WithBaseURL(server.URL))
	return client, server
}


func TestNewClient_Defaults(t *testing.T) {
	client := NewClient("test-api-key")
	if client.apiKey != "test-api-key" {
		t.Errorf("apiKey = %q, want %q", client.apiKey, "test-api-key")
	}
	if client.baseURL != defaultBaseURL {
		t.Errorf("baseURL = %q, want %q", client.baseURL, defaultBaseURL)
	}
	if client.httpClient.Timeout != defaultTimeout {
		t.Errorf("timeout = %v, want %v", client.httpClient.Timeout, defaultTimeout)
	}

	// Every service should be wired up, not nil. A nil service would panic on first use rather than failing cleanly.
	if client.Agents == nil || client.Numbers == nil || client.Calls == nil {
		t.Error("expected all services to be initialized, got a nil service")
	}
}


func TestWithBaseURL_TrimsTrailingSlash(t *testing.T) {
	client := NewClient("key", WithBaseURL("https://api.agentphone.ai/v1/"))
	if client.baseURL != "https://api.agentphone.ai/v1" {
		t.Errorf("baseURL = %q, want trailing slash trimmed", client.baseURL)
	}
}


func TestWithTimeout_OrderIndependent(t *testing.T) {
	custom := &http.Client{Timeout: 99 * time.Second}

	clientA := NewClient("key", WithTimeout(5*time.Second), WithHTTPClient(custom))
	clientB := NewClient("key", WithHTTPClient(custom), WithTimeout(5*time.Second))

	if clientA.httpClient.Timeout != 5*time.Second {
		t.Errorf("timeout(WithTimeout then WithHTTPClient) = %v, want 5s", clientA.httpClient.Timeout)
	}
	if clientB.httpClient.Timeout != 5*time.Second {
		t.Errorf("timeout(WithHTTPClient then WithTimeout) = %v, want 5s", clientB.httpClient.Timeout)
	}
}


func TestRequest_SendAuthHeaderAndPath(t *testing.T) {
	var gotPath, gotAuth, gotMethod string

	client, server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotAuth = r.Header.Get("Authorization")
		gotMethod = r.Method
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{}`))
	})
	defer server.Close()

	err := client.request(context.Background(), http.MethodGet, "/agents", nil, nil)
	if err != nil {
		t.Fatalf("request() returned error: %v", err)
	}
	if gotMethod != http.MethodGet {
		t.Errorf("method = %q, want GET", gotMethod)
	}
	if gotPath != "/agents" {
		t.Errorf("path = %q, want /agents", gotPath)
	}
	if gotAuth != "Bearer test-api-key" {
		t.Errorf("Authorization header = %q, want %q", gotAuth, "Bearer test-api-key")
	}
}

func TestRequest_EncodesBodyAndDecodesResponse(t *testing.T) {
	type reqBody struct {
		Name string `json:"name"`
	}
	type respBody struct {
		ID string `json:"id"`
	}

	client, server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		var got reqBody
		if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
			t.Fatalf("server failed to decode request body: %v", err)
		}
		if got.Name != "Support Bot" {
			t.Errorf("request body name = %q, want %q", got.Name, "Support Bot")
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(respBody{ID: "agt_123"})
	})
	defer server.Close()

	var result respBody
	err := client.request(context.Background(), http.MethodPost, "/agents", reqBody{Name: "Support Bot"}, &result)

	if err != nil {
		t.Fatalf("request() returned error: %v", err)
	}
	if result.ID != "agt_123" {
		t.Errorf("decoded ID = %q, want %q", result.ID, "agt_123")
	}
}


func TestRequest_MissingLeadingSlashIsFixed(t *testing.T) {
	var gotPath string
	client, server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.WriteHeader(http.StatusOK)
	})
	defer server.Close()
 
	// Deliberately pass a path with no leading "/" — request() should
	// still produce a single "/", never "//" and never a missing slash.
	if err := client.request(context.Background(), http.MethodGet, "agents", nil, nil); err != nil {
		t.Fatalf("request() returned error: %v", err)
	}
	if gotPath != "/agents" {
		t.Errorf("path = %q, want /agents (with exactly one leading slash)", gotPath)
	}
}


func TestRequest_NonJSONErrorBodyStillReturnsAnError(t *testing.T) {
	client, server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("upstream is on fire")) // not JSON at all
	})
	defer server.Close()

	err := client.request(context.Background(), http.MethodGet, "/agents", nil, nil)
	if err == nil {
		t.Fatal("expected an error for a 500 response, got nil")
	}

	var serverErr *ServerError
	if !errors.As(err, &serverErr) {
		t.Fatalf("expected a *ServerError, got %T: %v", err, err)
	}
	if serverErr.Message != "upstream is on fire" {
		t.Errorf("Message = %q, want the raw body as a fallback", serverErr.Message)
	}
}
