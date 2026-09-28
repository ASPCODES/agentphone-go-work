package agentphone

import (
	"context"
	"net/http"
	"strings"
	"testing"
)

func TestAgentsService_Create_SendsCamelCaseBody(t *testing.T) {
	var gotMethod, gotPath, gotBody string
	client, server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		gotMethod, gotPath, gotBody = r.Method, r.URL.Path, readBody(t, r)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id": "agt_1", "name": "Support Bot", "voiceMode": "hosted", "numbers": []}`))
	})
	defer server.Close()

	agent, err := client.Agents.Create(context.Background(), &CreateAgentParams{
		Name:         "Support Bot",
		VoiceMode:    "hosted",
		SystemPrompt: "You are helpful.",
		BeginMessage: "Hello!",
	})
	if err != nil {
		t.Fatalf("Create() error: %v", err)
	}
	if gotMethod != http.MethodPost || gotPath != "/agents" {
		t.Errorf("got %s %s, want POST /agents", gotMethod, gotPath)
	}
	for _, want := range []string{`"voiceMode"`, `"systemPrompt"`, `"beginMessage"`} {
		if !strings.Contains(gotBody, want) {
			t.Errorf("request body %s missing key %s", gotBody, want)
		}
	}
	if agent.ID != "agt_1" || agent.VoiceMode != "hosted" {
		t.Errorf("agent = %+v", agent)
	}
}

// The heart of the pointer-field design: a nil pointer means "leave this
// alone" (field omitted), while a pointer to false/0 means "explicitly set
// it to false/0" (field sent). A plain bool/int can't tell those apart.
func TestAgentsService_Update_PointerFieldsAreTriState(t *testing.T) {
	var gotBody string
	client, server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		gotBody = readBody(t, r)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id": "agt_1"}`))
	})
	defer server.Close()

	// Explicitly false / zero — must be SENT.
	_, err := client.Agents.Update(context.Background(), "agt_1", &UpdateAgentParams{
		EnableBackchannel:       Bool(false),
		InterruptionSensitivity: Float64(0),
	})
	if err != nil {
		t.Fatalf("Update() error: %v", err)
	}
	if !strings.Contains(gotBody, `"enableBackchannel":false`) {
		t.Errorf("body %s should send enableBackchannel:false explicitly", gotBody)
	}
	if !strings.Contains(gotBody, `"interruptionSensitivity":0`) {
		t.Errorf("body %s should send interruptionSensitivity:0 explicitly", gotBody)
	}

	// Not set at all (nil) — must be OMITTED.
	_, err = client.Agents.Update(context.Background(), "agt_1", &UpdateAgentParams{Name: "Renamed"})
	if err != nil {
		t.Fatalf("Update() error: %v", err)
	}
	for _, unwanted := range []string{"enableBackchannel", "interruptionSensitivity", "maxSilenceMs", "voiceSpeed", "enableMessaging"} {
		if strings.Contains(gotBody, unwanted) {
			t.Errorf("body %s should omit unset %s", gotBody, unwanted)
		}
	}
}

func TestAgentsService_Update_UsesPatchAndAgentPath(t *testing.T) {
	var gotMethod, gotPath string
	client, server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		gotMethod, gotPath = r.Method, r.URL.Path
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id": "agt_1"}`))
	})
	defer server.Close()

	if _, err := client.Agents.Update(context.Background(), "agt_1", &UpdateAgentParams{Name: "x"}); err != nil {
		t.Fatalf("Update() error: %v", err)
	}
	if gotMethod != http.MethodPatch || gotPath != "/agents/agt_1" {
		t.Errorf("got %s %s, want PATCH /agents/agt_1", gotMethod, gotPath)
	}
}

func TestAgentsService_List_UsesOffsetPagination(t *testing.T) {
	var gotQuery string
	client, server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"data": [{"id": "agt_1", "name": "A"}], "hasMore": false, "total": 1}`))
	})
	defer server.Close()

	resp, err := client.Agents.List(context.Background(), &ListParams{Limit: 5})
	if err != nil {
		t.Fatalf("List() error: %v", err)
	}
	if gotQuery != "limit=5" {
		t.Errorf("query = %q, want limit=5", gotQuery)
	}
	if len(resp.Agents) != 1 || resp.Total != 1 {
		t.Errorf("resp = %+v", resp)
	}
}

func TestAgentsService_AttachAndDetachNumber(t *testing.T) {
	var calls []string
	var attachBody string
	client, server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		calls = append(calls, r.Method+" "+r.URL.Path)
		if r.Method == http.MethodPost {
			attachBody = readBody(t, r)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id": "num_1"}`))
	})
	defer server.Close()

	if _, err := client.Agents.AttachNumber(context.Background(), "agt_1", "num_1"); err != nil {
		t.Fatalf("AttachNumber() error: %v", err)
	}
	if err := client.Agents.DetachNumber(context.Background(), "agt_1", "num_1"); err != nil {
		t.Fatalf("DetachNumber() error: %v", err)
	}

	want := []string{"POST /agents/agt_1/numbers", "DELETE /agents/agt_1/numbers/num_1"}
	if len(calls) != 2 || calls[0] != want[0] || calls[1] != want[1] {
		t.Errorf("calls = %v, want %v", calls, want)
	}
	if !strings.Contains(attachBody, `"numberId":"num_1"`) {
		t.Errorf("attach body = %s, want numberId", attachBody)
	}
}

// ListVoices is the one endpoint the docs describe in snake_case, unlike
// the rest of the camelCase API — this pins that exception down.
func TestAgentsService_ListVoices_ParsesSnakeCaseFields(t *testing.T) {
	client, server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"data": [{"voice_id": "v1", "voice_name": "Brian", "gender": "male", "accent": "US", "preview_audio_url": "https://x/p.mp3"}]}`))
	})
	defer server.Close()

	resp, err := client.Agents.ListVoices(context.Background())
	if err != nil {
		t.Fatalf("ListVoices() error: %v", err)
	}
	if len(resp.Voices) != 1 || resp.Voices[0].VoiceID != "v1" || resp.Voices[0].PreviewAudioURL != "https://x/p.mp3" {
		t.Errorf("Voices = %+v", resp.Voices)
	}
}

func TestAgentsService_ListCallsAndConversations_Paths(t *testing.T) {
	var paths []string
	client, server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		paths = append(paths, r.URL.Path)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"data": [], "hasMore": false}`))
	})
	defer server.Close()

	if _, err := client.Agents.ListCalls(context.Background(), "agt_1", nil); err != nil {
		t.Fatalf("ListCalls() error: %v", err)
	}
	if _, err := client.Agents.ListConversations(context.Background(), "agt_1", nil); err != nil {
		t.Fatalf("ListConversations() error: %v", err)
	}
	if len(paths) != 2 || paths[0] != "/agents/agt_1/calls" || paths[1] != "/agents/agt_1/conversations" {
		t.Errorf("paths = %v", paths)
	}
}
