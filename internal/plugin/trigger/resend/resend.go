package resend

import (
	"encoding/json"
	"fmt"
	"net/http"

	svix "github.com/svix/svix-webhooks/go"

	"hookd/internal/plugin"
)

func init() {
	plugin.RegisterTrigger(&Trigger{})
}

type Trigger struct{}

func (t *Trigger) Type() string { return "resend" }

func (t *Trigger) Verify(headers http.Header, body []byte, secret string) error {
	if secret == "" {
		return fmt.Errorf("resend: secret required")
	}
	wh, err := svix.NewWebhook(secret)
	if err != nil {
		return fmt.Errorf("resend: %w", err)
	}
	return wh.Verify(body, headers)
}

func (t *Trigger) Parse(body []byte) (string, map[string]any, error) {
	var payload map[string]any
	if err := json.Unmarshal(body, &payload); err != nil {
		return "", nil, fmt.Errorf("resend: invalid json: %w", err)
	}
	event, _ := payload["type"].(string)
	if event == "" {
		return "", nil, fmt.Errorf("resend: missing type")
	}
	data, _ := payload["data"].(map[string]any)
	if data == nil {
		data = map[string]any{}
	}
	// stash created_at into data sibling via engine reading payload — return data only;
	// engine extracts created_at from top-level via a second unmarshal or we embed it.
	if ca, ok := payload["created_at"]; ok {
		data["_created_at"] = ca
	}
	return event, data, nil
}

func (t *Trigger) DeliveryID(headers http.Header) string {
	return headers.Get("svix-id")
}
