package agentphone

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"
)

// APIError represents an error response from the AgentPhone API. It's the
// base type embedded in every typed error below, and satisfies the
// standard error interface on its own.
//
// AgentPhone's API uses two response shapes for errors.
// Reference -> docs.agentphone.ai/error-handling):

type APIError struct {
	StatusCode int
	Message    string
	Code 	   string
	Type 	   string
	Details   []ErrorDetail
	RetryAfter time.Duration
}

func (e *APIError) Error() string {
	if e.Code != "" {
		return fmt.Sprintf("agentphone: %d [%s] %s", e.StatusCode, e.Code, e.Message)
	}
	return fmt.Sprintf("agentphone: %d %s", e.StatusCode, e.Message)
}


type ErrorDetail struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Type    string `json:"type"`
}


// AuthenticationError: missing or invalid API key (401).
type AuthenticationError struct{ *APIError }

// ForbiddenError: the account or line isn't allowed to perform this
// action (403) — e.g. Code "WHATSAPP_NOT_ENABLED".
type ForbiddenError struct{ *APIError }

// PaymentRequiredError: account balance is too low for a paid action
// (402), e.g. provisioning a number.
type PaymentRequiredError struct{ *APIError }

// NotFoundError: the resource doesn't exist or you don't have access to
// it (404) — includes Code "PHONE_NUMBER_NOT_FOUND" style errors.
type NotFoundError struct{ *APIError }

// ConflictError: the account's number limit was reached, the requested
// number is unavailable, or a number hit its per-number concurrent-call
// limit (409).
type ConflictError struct{ *APIError }

// ValidationError: invalid request parameters or data (400 or 422) —
// includes Code "VALIDATION_ERROR" (check Details) and "INBOUND_ONLY".
type ValidationError struct{ *APIError }

type RateLimitError struct{ *APIError }


var nonRetriableRateLimitCodes = map[string]bool{
	"CONVERSATION_STREAK_LIMIT":      true,
	"CONVERSATION_AWAITING_REPLY":    true,
	"CONVERSATION_INACTIVE":          true,
	"OUTBOUND_LIMIT_REACHED":         true,
	"NEW_CONVERSATION_LIMIT_REACHED": true,
}


func (e *RateLimitError) Retriable() bool {
	return !nonRetriableRateLimitCodes[e.Code]
}


type ServerError struct{ *APIError }


type errorEnvelope struct {
	Error struct {
		Message string        `json:"message"`
		Code    string        `json:"code"`
		Type    string        `json:"type"`
		Details []ErrorDetail `json:"details"`
	} `json:"error"`
}


type plainDetailBody struct {
	Detail string `json:"detail"`
}


// parseAPIError converts a non-2xx HTTP response into a typed error,
// trying the envelope shape first, then the plain-detail shape, then
// falling back to the raw body so the SDK never returns an empty message.
func parseAPIError(statusCode int, body []byte, retryAfterHeader string) error {
	base := &APIError{StatusCode: statusCode}

	var envelope errorEnvelope
	if err := json.Unmarshal(body, &envelope); err == nil && envelope.Error.Message != "" {
		base.Message = envelope.Error.Message
		base.Code = envelope.Error.Code
		base.Type = envelope.Error.Type
		base.Details = envelope.Error.Details
	} else {
		var plain plainDetailBody
		if err := json.Unmarshal(body, &plain); err == nil && plain.Detail != "" {
			base.Message = plain.Detail
		} else if len(body) > 0 {
			base.Message = string(body)
		} else {
			base.Message = "unknown error"
		}
	}

	if retryAfterHeader != "" {
		if secs, err := strconv.Atoi(retryAfterHeader); err == nil {
			base.RetryAfter = time.Duration(secs) * time.Second
		}
	}

	switch statusCode {
	case 401:
		return &AuthenticationError{base}
	case 402:
		return &PaymentRequiredError{base}
	case 403:
		return &ForbiddenError{base}
	case 404:
		return &NotFoundError{base}
	case 409:
		return &ConflictError{base}
	case 400, 422:
		return &ValidationError{base}
	case 429:
		return &RateLimitError{base}
	case 500, 502, 503, 504:
		return &ServerError{base}
	default:
		return base
	}
}
