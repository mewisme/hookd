package template_test

import (
	"testing"

	tpl "hookd/internal/template"
)

func TestEventTitleColor(t *testing.T) {
	if got := mustRender(t, `{{eventTitle .event}}`, map[string]any{"event": "email.bounced"}); got != "Email Bounced" {
		t.Fatalf("title: %q", got)
	}
	if got := mustRender(t, `{{eventColor .event}}`, map[string]any{"event": "email.failed"}); got != "15158332" {
		t.Fatalf("color: %q", got)
	}
	if got := mustRender(t, `{{eventTitle .event}}`, map[string]any{"event": "email.sent"}); got != "Resend Event" {
		t.Fatalf("default title: %q", got)
	}
}

func TestJoinDefaultJSON(t *testing.T) {
	data := tpl.Context("email.opened", "2024-01-01T00:00:00Z", map[string]any{
		"to":      []any{"a@x.com", "b@x.com"},
		"from":    "",
		"subject": "Hi",
		"tags":    map[string]any{"campaign": "x"},
	}, `{}`)

	if got := mustRender(t, `{{join .data.to ", "}}`, data); got != "a@x.com, b@x.com" {
		t.Fatalf("join: %q", got)
	}
	if got := mustRender(t, `{{default .data.from "N/A"}}`, data); got != "N/A" {
		t.Fatalf("default: %q", got)
	}
	if got := mustRender(t, `{{json .data.tags}}`, data); got != `{"campaign":"x"}` {
		t.Fatalf("json: %q", got)
	}
}

func mustRender(t *testing.T, tmpl string, data any) string {
	t.Helper()
	out, err := tpl.Render(tmpl, data)
	if err != nil {
		t.Fatal(err)
	}
	return out
}
