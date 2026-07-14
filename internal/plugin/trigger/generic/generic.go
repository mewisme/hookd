package generic

import (
	"crypto/subtle"
	"encoding/json"
	"fmt"
	"net/http"

	"hookd/internal/plugin"
)

func init() {
	plugin.RegisterTrigger(&Trigger{})
}

// Trigger compares a configurable header to the secret (shared-secret auth).
type Trigger struct{}

func (t *Trigger) Type() string { return "generic" }

func (t *Trigger) Verify(headers http.Header, body []byte, secret string) error {
	_ = body
	if secret == "" {
		return fmt.Errorf("generic: secret required")
	}
	// Header name comes from trigger config; engine sets X-Webhook-Secret default via
	// VerifyWithHeader when Header is set. Default path uses X-Webhook-Secret.
	got := headers.Get("X-Webhook-Secret")
	if got == "" {
		got = headers.Get("x-webhook-secret")
	}
	if subtle.ConstantTimeCompare([]byte(got), []byte(secret)) != 1 {
		return fmt.Errorf("generic: invalid secret")
	}
	return nil
}

// VerifyWithHeader allows engine to pass a custom header name.
func VerifyWithHeader(headers http.Header, secret, header string) error {
	if secret == "" {
		return fmt.Errorf("generic: secret required")
	}
	if header == "" {
		header = "X-Webhook-Secret"
	}
	got := headers.Get(header)
	if subtle.ConstantTimeCompare([]byte(got), []byte(secret)) != 1 {
		return fmt.Errorf("generic: invalid secret")
	}
	return nil
}

func (t *Trigger) Parse(body []byte) (string, map[string]any, error) {
	var payload map[string]any
	if err := json.Unmarshal(body, &payload); err != nil {
		return "", nil, fmt.Errorf("generic: invalid json: %w", err)
	}
	event, _ := payload["type"].(string)
	if event == "" {
		event, _ = payload["event"].(string)
	}
	if event == "" {
		return "", nil, fmt.Errorf("generic: missing type/event")
	}
	data, _ := payload["data"].(map[string]any)
	if data == nil {
		data = payload
	}
	if ca, ok := payload["created_at"]; ok {
		data["_created_at"] = ca
	}
	return event, data, nil
}

func (t *Trigger) DeliveryID(headers http.Header) string {
	if id := headers.Get("Idempotency-Key"); id != "" {
		return id
	}
	return headers.Get("idempotency-key")
}
