package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"hookd/internal/config"

	_ "hookd/internal/plugin/action/http"
	_ "hookd/internal/plugin/action/log"
	_ "hookd/internal/plugin/trigger/generic"
	_ "hookd/internal/plugin/trigger/resend"
)

func TestLoadAndEnvExpand(t *testing.T) {
	t.Setenv("RESEND_WEBHOOK_SECRET", "whsec_test")
	t.Setenv("DISCORD_WEBHOOK_URL", "https://discord.example/hook")

	dir := t.TempDir()
	path := filepath.Join(dir, "w.yaml")
	src := `
triggers:
  - id: r
    type: resend
    path: hooks/resend-to-discord
    secret: ${RESEND_WEBHOOK_SECRET}
    hooks:
      - event: "*"
        actions:
          - type: http
            method: POST
            url: ${DISCORD_WEBHOOK_URL}
            body: '{"e":"{{.event}}"}'
`
	if err := os.WriteFile(path, []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg, err := config.Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Triggers[0].Path != "/hooks/resend-to-discord" {
		t.Fatalf("path: %q", cfg.Triggers[0].Path)
	}
	if cfg.Triggers[0].Secret != "whsec_test" {
		t.Fatalf("secret: %q", cfg.Triggers[0].Secret)
	}
	url, _ := cfg.Triggers[0].Hooks[0].Actions[0]["url"].(string)
	if url != "https://discord.example/hook" {
		t.Fatalf("url: %q", url)
	}
}

func TestUnknownPluginFails(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.yaml")
	src := `
triggers:
  - id: x
    type: notaplugin
    path: /x
    secret: s
    hooks:
      - event: "*"
        actions:
          - type: log
            message: hi
`
	if err := os.WriteFile(path, []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := config.Load(path); err == nil {
		t.Fatal("expected error")
	}
}
