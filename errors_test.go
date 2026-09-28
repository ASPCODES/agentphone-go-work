package agentphone

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"
)


func TestParseAPIError_EnvelopeShape(t *testing.T) {
	client, server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnprocessableEntity)
		w.Write([]byte(`{
			"error": {
				"message": "Validation error",
				"code": "VALIDATION_ERROR",
				"type": "validation_error",
				"details": [{"field": "country", "message": "Country must be a 2-letter ISO code", "type": "value_error"}]
			}
		}`))
	})
	defer server.Close()

	err := client.request(context.Background(), http.MethodPost, "/numbers", nil, nil)

	var validationErr *ValidationError
	if !errors.As(err, &validationErr) {
		t.Fatalf("expected a *ValidationError, got %T: %v", err, err)
	}
	if validationErr.Code != "VALIDATION_ERROR" {
		t.Errorf("Code = %q, want VALIDATION_ERROR", validationErr.Code)
	}
	if len(validationErr.Details) != 1 || validationErr.Details[0].Field != "country" {
		t.Errorf("Details = %+v, want one entry for field \"country\"", validationErr.Details)
	}
}

func TestParseAPIError_PlainDetailShape(t *testing.T) {
	client, server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"detail": "Invalid API key"}`))
	})
	defer server.Close()

	err := client.request(context.Background(), http.MethodGet, "/agents", nil, nil)

	var authErr *AuthenticationError
	if !errors.As(err, &authErr) {
		t.Fatalf("expected an *AuthenticationError, got %T: %v", err, err)
	}
	if authErr.Message != "Invalid API key" {
		t.Errorf("Message = %q, want %q", authErr.Message, "Invalid API key")
	}
	// The plain-detail shape carries no machine-readable code.
	if authErr.Code != "" {
		t.Errorf("Code = %q, want empty for the plain-detail shape", authErr.Code)
	}
}

func TestParseAPIError_StatusCodeMapping(t *testing.T) {
	cases := []struct {
		status  int
		checkAs func(error) bool
	}{
		{http.StatusUnauthorized, func(err error) bool { var e *AuthenticationError; return errors.As(err, &e) }},
		{http.StatusPaymentRequired, func(err error) bool { var e *PaymentRequiredError; return errors.As(err, &e) }},
		{http.StatusForbidden, func(err error) bool { var e *ForbiddenError; return errors.As(err, &e) }},
		{http.StatusNotFound, func(err error) bool { var e *NotFoundError; return errors.As(err, &e) }},
		{http.StatusConflict, func(err error) bool { var e *ConflictError; return errors.As(err, &e) }},
		{http.StatusBadRequest, func(err error) bool { var e *ValidationError; return errors.As(err, &e) }},
		{http.StatusUnprocessableEntity, func(err error) bool { var e *ValidationError; return errors.As(err, &e) }},
		{http.StatusTooManyRequests, func(err error) bool { var e *RateLimitError; return errors.As(err, &e) }},
		{http.StatusInternalServerError, func(err error) bool { var e *ServerError; return errors.As(err, &e) }},
		{http.StatusBadGateway, func(err error) bool { var e *ServerError; return errors.As(err, &e) }},
		{http.StatusServiceUnavailable, func(err error) bool { var e *ServerError; return errors.As(err, &e) }},
		{http.StatusGatewayTimeout, func(err error) bool { var e *ServerError; return errors.As(err, &e) }},
	}

	for _, tc := range cases {
		err := parseAPIError(tc.status, []byte(`{"detail": "x"}`), "")
		if !tc.checkAs(err) {
			t.Errorf("status %d: got wrong error type %T", tc.status, err)
		}
	}
}

func TestParseAPIError_RetryAfterHeader(t *testing.T) {
	err := parseAPIError(http.StatusTooManyRequests, []byte(`{"detail": "slow down"}`), "30")

	var rateLimitErr *RateLimitError
	if !errors.As(err, &rateLimitErr) {
		t.Fatalf("expected a *RateLimitError, got %T", err)
	}
	if rateLimitErr.RetryAfter != 30*time.Second {
		t.Errorf("RetryAfter = %v, want 30s", rateLimitErr.RetryAfter)
	}
}

func TestRateLimitError_Retriable(t *testing.T) {
	cases := []struct {
		code string
		want bool
	}{
		{"", true},             
		{"RATE_LIMITED", true},
		{"CONVERSATION_STREAK_LIMIT", false},
		{"CONVERSATION_AWAITING_REPLY", false},
		{"CONVERSATION_INACTIVE", false},
		{"OUTBOUND_LIMIT_REACHED", false},
		{"NEW_CONVERSATION_LIMIT_REACHED", false},
	}

	for _, tc := range cases {
		body := `{"detail": "x"}`
		if tc.code != "" {
			body = `{"error": {"message": "x", "code": "` + tc.code + `"}}`
		}
		err := parseAPIError(http.StatusTooManyRequests, []byte(body), "")

		var rateLimitErr *RateLimitError
		if !errors.As(err, &rateLimitErr) {
			t.Fatalf("code %q: expected a *RateLimitError, got %T", tc.code, err)
		}
		if got := rateLimitErr.Retriable(); got != tc.want {
			t.Errorf("code %q: Retriable() = %v, want %v", tc.code, got, tc.want)
		}
	}
}

func TestParseAPIError_FallsBackToRawBodyWhenNeitherShapeMatches(t *testing.T) {
	err := parseAPIError(http.StatusInternalServerError, []byte("plain text, not JSON at all"), "")

	var serverErr *ServerError
	if !errors.As(err, &serverErr) {
		t.Fatalf("expected a *ServerError, got %T", err)
	}
	if serverErr.Message != "plain text, not JSON at all" {
		t.Errorf("Message = %q, want the raw body", serverErr.Message)
	}
}

func TestParseAPIError_EmptyBody(t *testing.T) {
	err := parseAPIError(http.StatusInternalServerError, []byte{}, "")

	var serverErr *ServerError
	if !errors.As(err, &serverErr) {
		t.Fatalf("expected a *ServerError, got %T", err)
	}
	if serverErr.Message != "unknown error" {
		t.Errorf("Message = %q, want the \"unknown error\" fallback", serverErr.Message)
	}
}

func TestAPIError_ErrorStringIncludesCodeWhenPresent(t *testing.T) {
	withCode := &APIError{StatusCode: 422, Code: "VALIDATION_ERROR", Message: "bad input"}
	if got := withCode.Error(); got != "agentphone: 422 [VALIDATION_ERROR] bad input" {
		t.Errorf("Error() = %q", got)
	}

	withoutCode := &APIError{StatusCode: 401, Message: "invalid key"}
	if got := withoutCode.Error(); got != "agentphone: 401 invalid key" {
		t.Errorf("Error() = %q", got)
	}
}

