package engine

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"

	"hookd/internal/config"
	"hookd/internal/plugin"
	"hookd/internal/plugin/trigger/generic"
)

// Engine routes webhooks to trigger plugins and runs matching actions.
type Engine struct {
	cfg       *config.Config
	publicURL string // optional; when set, mount logs include full webhook URL
	// ponytail: process-local dedup; multi-replica needs Redis.
	seen   map[string]time.Time
	seenMu sync.Mutex
	ttl    time.Duration
}

func New(cfg *config.Config, publicURL string) *Engine {
	return &Engine{
		cfg:       cfg,
		publicURL: strings.TrimRight(publicURL, "/"),
		seen:      map[string]time.Time{},
		ttl:       10 * time.Minute,
	}
}

func (e *Engine) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	for i := range e.cfg.Triggers {
		tc := e.cfg.Triggers[i]
		trig, err := plugin.GetTrigger(tc.Type)
		if err != nil {
			panic(err) // validated at load
		}
		path := tc.Path
		mux.HandleFunc("POST "+path, e.handleTrigger(tc, trig))
		attrs := []any{"id", tc.ID, "type", tc.Type, "path", path}
		if e.publicURL != "" {
			attrs = append(attrs, "url", e.publicURL+path)
		}
		slog.Info("mounted trigger", attrs...)
	}
	return mux
}

func (e *Engine) handleTrigger(tc config.Trigger, trig plugin.Trigger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "read body", http.StatusBadRequest)
			return
		}

		var verifyErr error
		if tc.Type == "generic" && tc.Header != "" {
			verifyErr = generic.VerifyWithHeader(r.Header, tc.Secret, tc.Header)
		} else {
			verifyErr = trig.Verify(r.Header, body, tc.Secret)
		}
		if verifyErr != nil {
			slog.Warn("verify failed", "trigger", tc.ID, "err", verifyErr)
			http.Error(w, "invalid webhook", http.StatusBadRequest)
			return
		}

		if id := trig.DeliveryID(r.Header); id != "" {
			if e.duplicate(id) {
				slog.Info("duplicate delivery skipped", "id", id, "trigger", tc.ID)
				w.WriteHeader(http.StatusOK)
				return
			}
		}

		eventName, data, err := trig.Parse(body)
		if err != nil {
			slog.Warn("parse failed", "trigger", tc.ID, "err", err)
			http.Error(w, "invalid payload", http.StatusBadRequest)
			return
		}

		createdAt := ""
		if ca, ok := data["_created_at"]; ok {
			createdAt = stringify(ca)
			delete(data, "_created_at")
		} else {
			var top map[string]any
			if json.Unmarshal(body, &top) == nil {
				createdAt = stringify(top["created_at"])
			}
		}

		ev := plugin.Event{
			Event:     eventName,
			CreatedAt: createdAt,
			Data:      data,
			Raw:       string(body),
		}

		ctx := r.Context()
		var actionErr error
		for _, h := range tc.Hooks {
			if h.Event != "*" && h.Event != eventName {
				continue
			}
			for _, acfg := range h.Actions {
				typ, _ := acfg["type"].(string)
				act, err := plugin.GetAction(typ)
				if err != nil {
					actionErr = err
					slog.Error("action lookup", "err", err)
					continue
				}
				if err := act.Run(ctx, acfg, ev); err != nil {
					actionErr = err
					slog.Error("action failed", "trigger", tc.ID, "action", typ, "err", err)
				}
			}
		}

		if actionErr != nil && tc.FailOnError {
			http.Error(w, "action failed", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}

func (e *Engine) duplicate(id string) bool {
	e.seenMu.Lock()
	defer e.seenMu.Unlock()
	now := time.Now()
	for k, t := range e.seen {
		if now.Sub(t) > e.ttl {
			delete(e.seen, k)
		}
	}
	if _, ok := e.seen[id]; ok {
		return true
	}
	e.seen[id] = now
	return false
}

func stringify(v any) string {
	if v == nil {
		return ""
	}
	switch t := v.(type) {
	case string:
		return t
	default:
		b, err := json.Marshal(t)
		if err != nil {
			return ""
		}
		s := string(b)
		if len(s) >= 2 && s[0] == '"' {
			return s[1 : len(s)-1]
		}
		return s
	}
}
