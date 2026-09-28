package agentphone

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"testing"
)

func TestCallsService_List_BuildsFilterQuery(t *testing.T) {
	var gotQuery string
	client, server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"data": [{"id": "call_1", "direction": "inbound", "status": "completed", "durationSeconds": 330, "recordingAvailable": true}], "hasMore": false, "total": 1}`))
	})
	defer server.Close()

	resp, err := client.Calls.List(context.Background(), &ListCallsParams{
		Limit:     10,
		Status:    "completed",
		Direction: "inbound",
		Search:    "+1415",
	})
	if err != nil {
		t.Fatalf("List() error: %v", err)
	}
	for _, want := range []string{"limit=10", "status=completed", "direction=inbound", "search=%2B1415"} {
		if !strings.Contains(gotQuery, want) {
			t.Errorf("query %q missing %q", gotQuery, want)
		}
	}
	if len(resp.Calls) != 1 || resp.Calls[0].DurationSeconds != 330 || !resp.Calls[0].RecordingAvailable {
		t.Errorf("Calls = %+v", resp.Calls)
	}
}

func TestCallsService_Get_ParsesTranscripts(t *testing.T) {
	client, server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/calls/call_1" {
			t.Errorf("path = %q, want /calls/call_1", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id": "call_1", "transcripts": [{"id": "tr_1", "transcript": "Hello", "confidence": 0.95, "response": "Hi"}]}`))
	})
	defer server.Close()

	call, err := client.Calls.Get(context.Background(), "call_1")
	if err != nil {
		t.Fatalf("Get() error: %v", err)
	}
	if len(call.Transcripts) != 1 || call.Transcripts[0].Confidence != 0.95 {
		t.Errorf("Transcripts = %+v", call.Transcripts)
	}
}

func TestCallsService_Create_SendsOutboundCallBody(t *testing.T) {
	var gotMethod, gotPath, gotBody string
	client, server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		gotMethod, gotPath, gotBody = r.Method, r.URL.Path, readBody(t, r)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id": "call_new", "status": "in-progress", "direction": "outbound"}`))
	})
	defer server.Close()

	call, err := client.Calls.Create(context.Background(), &CreateOutboundCallParams{
		AgentID:         "agt_1",
		ToNumber:        "+15559876543",
		InitialGreeting: "Hi!",
	})
	if err != nil {
		t.Fatalf("Create() error: %v", err)
	}
	if gotMethod != http.MethodPost || gotPath != "/calls" {
		t.Errorf("got %s %s, want POST /calls", gotMethod, gotPath)
	}
	for _, want := range []string{`"agentId"`, `"toNumber"`, `"initialGreeting"`} {
		if !strings.Contains(gotBody, want) {
			t.Errorf("body %s missing %s", gotBody, want)
		}
	}
	if strings.Contains(gotBody, `"systemPrompt"`) {
		t.Errorf("body %s should omit unset systemPrompt (webhook-mode call)", gotBody)
	}
	if call.ID != "call_new" {
		t.Errorf("ID = %q", call.ID)
	}
}

func TestCallsService_CreateWeb_ReturnsAccessToken(t *testing.T) {
	client, server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/calls/web" {
			t.Errorf("path = %q, want /calls/web", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"accessToken": "tok_abc", "callId": "call_web_1"}`))
	})
	defer server.Close()

	resp, err := client.Calls.CreateWeb(context.Background(), &CreateWebCallParams{AgentID: "agt_1"})
	if err != nil {
		t.Fatalf("CreateWeb() error: %v", err)
	}
	if resp.AccessToken != "tok_abc" || resp.CallID != "call_web_1" {
		t.Errorf("resp = %+v", resp)
	}
}

func TestCallsService_EndAndListForNumber_Paths(t *testing.T) {
	var calls []string
	client, server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		calls = append(calls, r.Method+" "+r.URL.Path)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id": "call_1", "data": [], "hasMore": false}`))
	})
	defer server.Close()

	if _, err := client.Calls.End(context.Background(), "call_1"); err != nil {
		t.Fatalf("End() error: %v", err)
	}
	if _, err := client.Calls.ListForNumber(context.Background(), "num_1", nil); err != nil {
		t.Fatalf("ListForNumber() error: %v", err)
	}
	want := []string{"POST /calls/call_1/end", "GET /numbers/num_1/calls"}
	if len(calls) != 2 || calls[0] != want[0] || calls[1] != want[1] {
		t.Errorf("calls = %v, want %v", calls, want)
	}
}

// --- SSE transcript streaming ---

// sseHandler writes raw SSE text and flushes, like the real endpoint.
func sseHandler(body string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, body)
		if f, ok := w.(http.Flusher); ok {
			f.Flush()
		}
	}
}

const sampleStream = `event: connected
data: {"callId": "call_1", "status": "in-progress", "agentId": "agt_1"}

event: turn
data: {"role": "user", "content": "I need help.", "createdAt": "2026-01-01T00:00:01Z"}

: heartbeat

event: turn
data: {"role": "agent", "content": "Sure!", "createdAt": "2026-01-01T00:00:02Z"}

event: ended
data: {"callId": "call_1", "status": "completed", "durationSeconds": 42}

`

func TestStreamTranscript_DeliversEventsInOrder(t *testing.T) {
	client, server := newTestServer(sseHandler(sampleStream))
	defer server.Close()

	var types []string
	var lastTurn string
	var duration int

	err := client.Calls.StreamTranscript(context.Background(), "call_1", func(e TranscriptEvent) error {
		types = append(types, e.Type)
		if e.Turn != nil {
			lastTurn = e.Turn.Content
		}
		if e.Ended != nil {
			duration = e.Ended.DurationSeconds
		}
		return nil
	})
	if err != nil {
		t.Fatalf("StreamTranscript() error: %v", err)
	}

	want := "connected,turn,turn,ended"
	if got := strings.Join(types, ","); got != want {
		t.Errorf("event order = %s, want %s (heartbeat comment must be ignored)", got, want)
	}
	if lastTurn != "Sure!" || duration != 42 {
		t.Errorf("lastTurn=%q duration=%d", lastTurn, duration)
	}
}

func TestStreamTranscript_HandlerErrorStopsStream(t *testing.T) {
	client, server := newTestServer(sseHandler(sampleStream))
	defer server.Close()

	stop := errors.New("enough")
	seen := 0
	err := client.Calls.StreamTranscript(context.Background(), "call_1", func(e TranscriptEvent) error {
		seen++
		return stop // bail out on the very first event
	})

	if !errors.Is(err, stop) {
		t.Errorf("err = %v, want the handler's own error returned as-is", err)
	}
	if seen != 1 {
		t.Errorf("handler called %d times, want 1 (stream should stop immediately)", seen)
	}
}

func TestStreamTranscript_HTTPErrorBecomesTypedError(t *testing.T) {
	client, server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"detail": "Call not found"}`))
	})
	defer server.Close()

	err := client.Calls.StreamTranscript(context.Background(), "bad", func(TranscriptEvent) error { return nil })

	var notFound *NotFoundError
	if !errors.As(err, &notFound) {
		t.Fatalf("err = %T %v, want *NotFoundError", err, err)
	}
}

func TestStreamTranscript_SendsAcceptHeader(t *testing.T) {
	var gotAccept string
	client, server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		gotAccept = r.Header.Get("Accept")
		sseHandler(sampleStream)(w, r)
	})
	defer server.Close()

	err := client.Calls.StreamTranscript(context.Background(), "call_1", func(TranscriptEvent) error { return nil })
	if err != nil {
		t.Fatalf("StreamTranscript() error: %v", err)
	}
	if gotAccept != "text/event-stream" {
		t.Errorf("Accept = %q, want text/event-stream", gotAccept)
	}
}

// A transcript turn can be long. bufio.Scanner's default 64KB line limit would make this fail with "token too long" unless the buffer is raised.
func TestStreamTranscript_HandlesVeryLongLines(t *testing.T) {
	long := strings.Repeat("a", 200*1024)  // 200KB single turn
	body := fmt.Sprintf("event: turn\ndata: {\"role\": \"user\", \"content\": %q}\n\n", long)

	client, server := newTestServer(sseHandler(body))
	defer server.Close()

	var gotLen int
	err := client.Calls.StreamTranscript(context.Background(), "call_1", func(e TranscriptEvent) error {
		if e.Turn != nil {
			gotLen = len(e.Turn.Content)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("StreamTranscript() failed on a long line: %v", err)
	}
	if gotLen != len(long) {
		t.Errorf("content length = %d, want %d", gotLen, len(long))
	}
}

