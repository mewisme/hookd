package logaction

import (
	"context"
	"fmt"
	"log/slog"

	"hookd/internal/plugin"
	tpl "hookd/internal/template"
)

func init() {
	plugin.RegisterAction(&Action{})
}

type Action struct{}

func (a *Action) Type() string { return "log" }

func (a *Action) Run(ctx context.Context, cfg map[string]any, event plugin.Event) error {
	_ = ctx
	msg, _ := cfg["message"].(string)
	if msg == "" {
		msg = "{{.event}}"
	}
	data := tpl.Context(event.Event, event.CreatedAt, event.Data, event.Raw)
	out, err := tpl.Render(msg, data)
	if err != nil {
		return fmt.Errorf("log: template: %w", err)
	}
	slog.Info(out, "trigger_event", event.Event)
	return nil
}
