package agentphone

import (
	"errors"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestVerifyWebhook_AcceptsValidSignature(t *testing.T) {
	payload := []byte(`{"event":"message.received"}`)
	timestamp := time.Now().Unix()
	timestampString := strconvFormatInt(timestamp)
	signature := signWebhookPayload("whsec_test", timestampString, payload)

	if err := VerifyWebhook(payload, signature, "whsec_test", timestampString, time.Minute); err != nil {
		t.Fatalf("VerifyWebhook() error: %v", err)
	}
}

func TestVerifyWebhook_RejectsBadSignatureAndStaleTimestamp(t *testing.T) {
	payload := []byte(`{"event":"message.received"}`)
	timestampString := strconvFormatInt(time.Now().Unix())
	err := VerifyWebhook(payload, "invalid", "whsec_test", timestampString, time.Minute)
	var verifyErr *WebhookVerificationError
	if !errors.As(err, &verifyErr) || !strings.Contains(verifyErr.Message, "signature mismatch") {
		t.Fatalf("error = %T %v, want signature mismatch verification error", err, err)
	}

	stale := strconvFormatInt(time.Now().Add(-10 * time.Minute).Unix())
	err = VerifyWebhook(payload, signWebhookPayload("whsec_test", stale, payload), "whsec_test", stale, time.Minute)
	if !errors.As(err, &verifyErr) || !strings.Contains(verifyErr.Message, "outside the allowed tolerance") {
		t.Fatalf("error = %T %v, want stale timestamp verification error", err, err)
	}
}

func TestVerifyWebhook_RejectsInvalidTimestampAndSupportsDisabledTolerance(t *testing.T) {
	payload := []byte(`payload`)
	err := VerifyWebhook(payload, "signature", "secret", "not-a-timestamp", time.Minute)
	var verifyErr *WebhookVerificationError
	if !errors.As(err, &verifyErr) || verifyErr.Message != "invalid or missing timestamp" {
		t.Fatalf("error = %T %v, want invalid timestamp verification error", err, err)
	}

	signature := signWebhookPayload("secret", "", payload)
	if err := VerifyWebhook(payload, signature, "secret", "", -1); err != nil {
		t.Fatalf("VerifyWebhook() with disabled tolerance error: %v", err)
	}
}

func TestConstructEvent_VerifiesThenParsesPayload(t *testing.T) {
	payload := []byte(`{"event":"message.received","channel":"sms","data":{"message":"Hi","fromNumber":"+14155550100"},"recentHistory":[{"direction":"inbound","content":"Hello"}],"conversationState":{"intent":"support"}}`)
	timestamp := strconvFormatInt(time.Now().Unix())
	signature := signWebhookPayload("whsec_test", timestamp, payload)

	event, err := ConstructEvent(payload, signature, "whsec_test", timestamp, time.Minute)
	if err != nil {
		t.Fatalf("ConstructEvent() error: %v", err)
	}
	if event.Event != "message.received" || event.Channel != "sms" || event.Data.Message != "Hi" || event.Data.FromNumber != "+14155550100" || len(event.RecentHistory) != 1 || event.ConversationState["intent"] != "support" {
		t.Errorf("event = %+v", event)
	}

	if _, err := ConstructEvent(payload, "bad-signature", "whsec_test", timestamp, time.Minute); err == nil {
		t.Fatal("ConstructEvent() accepted an invalid signature")
	}
	invalidJSON := []byte("not-json")
	invalidSignature := signWebhookPayload("whsec_test", timestamp, invalidJSON)
	if _, err := ConstructEvent(invalidJSON, invalidSignature, "whsec_test", timestamp, time.Minute); err == nil || !strings.Contains(err.Error(), "decoding webhook event") {
		t.Fatalf("error = %v, want verified JSON decode error", err)
	}
}

func strconvFormatInt(value int64) string {
	return strconv.FormatInt(value, 10)
}
