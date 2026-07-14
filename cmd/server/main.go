package main

import (
	"log/slog"
	"net/http"
	"os"

	"hookd/internal/config"
	"hookd/internal/engine"
	"hookd/internal/plugin"

	_ "hookd/internal/plugin/action/http"
	_ "hookd/internal/plugin/action/log"
	_ "hookd/internal/plugin/trigger/generic"
	_ "hookd/internal/plugin/trigger/resend"
)

func main() {
	port := env("PORT", "8080")
	cfgPath := env("CONFIG_PATH", "/config/workflows.yaml")
	// PUBLIC_URL = public origin for this service (e.g. https://hooks.example.com).
	publicURL := env("PUBLIC_URL", "")

	slog.Info("registry", "triggers", plugin.TriggerTypes(), "actions", plugin.ActionTypes())

	cfg, err := config.Load(cfgPath)
	if err != nil {
		slog.Error("load config", "path", cfgPath, "err", err)
		os.Exit(1)
	}

	eng := engine.New(cfg, publicURL)
	addr := ":" + port
	slog.Info("listening", "addr", addr, "config", cfgPath, "public_url", publicURL)
	if err := http.ListenAndServe(addr, eng.Handler()); err != nil {
		slog.Error("server", "err", err)
		os.Exit(1)
	}
}

func env(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
