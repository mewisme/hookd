package main

import (
	"log/slog"
	"net/http"
	"os"

	"hookd/internal/config"
	"hookd/internal/engine"
	"hookd/internal/logx"
	"hookd/internal/plugin"

	_ "hookd/internal/plugin/action/http"
	_ "hookd/internal/plugin/action/log"
	_ "hookd/internal/plugin/trigger/generic"
	_ "hookd/internal/plugin/trigger/resend"
)

func main() {
	logger := logx.New(env("LOG_LEVEL", "info"))
	slog.SetDefault(logger)

	port := env("PORT", "8080")
	cfgPath := env("CONFIG_PATH", "/config/workflows.yaml")
	publicURL := env("PUBLIC_URL", "")

	slog.Info("registry", "triggers", plugin.TriggerTypes(), "actions", plugin.ActionTypes())

	cfg, err := config.Load(cfgPath)
	if err != nil {
		slog.Error("load config", "path", cfgPath, "error", err)
		os.Exit(1)
	}

	eng := engine.New(cfg, publicURL)
	addr := ":" + port
	slog.Info("server starting", "addr", addr, "config", cfgPath, "public_url", publicURL)
	if err := http.ListenAndServe(addr, eng.Handler()); err != nil {
		slog.Error("server stopped", "error", err)
		os.Exit(1)
	}
}

func env(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
