package engine_test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	svix "github.com/svix/svix-webhooks/go"

	"hookd/internal/config"
	"hookd/internal/engine"

	_ "hookd/internal/plugin/action/http"
	_ "hookd/internal/plugin/action/log"
	_ "hookd/internal/plugin/trigger/generic"
	_ "hookd/internal/plugin/trigger/resend"
)

func TestResendToLogAndEventMatch(t *testing.T) {
	secret := "whsec_testsecret1234567890123456789012"
	dir := t.TempDir()
	path := filepath.Join(dir, "w.yaml")
	src := `
triggers:
  - id: r
    type: resend
    path: /hooks/resend-to-discord
    secret: ` + secret + `
    hooks:
      - event: email.bounced
        actions:
          - type: log
            message: bounced-{{.event}}
      - event: "*"
        actions:
          - type: log
            message: all-{{.event}}
`
	if err := os.WriteFile(path, []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg, err := config.Load(path)
	if err != nil {
		t.Fatal(err)
	}
	eng := engine.New(cfg, "")
	srv := httptest.NewServer(eng.Handler())
	t.Cleanup(srv.Close)

	payload := map[string]any{
		"type":       "email.bounced",
		"created_at": "2024-01-01T00:00:00.000Z",
		"data": map[string]any{
			"email_id": "abc",
			"to":       []string{"a@x.com"},
			"from":     "b@x.com",
			"subject":  "Hi",
		},
	}
	body, _ := json.Marshal(payload)
	req := sign(t, secret, body, srv.URL+"/hooks/resend-to-discord")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body)
	if resp.StatusCode != 200 {
		t.Fatalf("status %d", resp.StatusCode)
	}

	// duplicate svix-id should still 200
	resp2, err := http.DefaultClient.Do(signWithID(t, secret, body, srv.URL+"/hooks/resend-to-discord", req.Header.Get("svix-id")))
	if err != nil {
		t.Fatal(err)
	}
	defer resp2.Body.Close()
	if resp2.StatusCode != 200 {
		t.Fatalf("dup status %d", resp2.StatusCode)
	}
}

func TestGenericSecretHeader(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "w.yaml")
	src := `
triggers:
  - id: g
    type: generic
    path: /hooks/partner
    secret: s3cret
    header: X-Custom-Secret
    hooks:
      - event: order.created
        actions:
          - type: log
            message: ok
`
	if err := os.WriteFile(path, []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg, err := config.Load(path)
	if err != nil {
		t.Fatal(err)
	}
	srv := httptest.NewServer(engine.New(cfg, "").Handler())
	t.Cleanup(srv.Close)

	body := []byte(`{"type":"order.created","data":{"id":1}}`)
	req, _ := http.NewRequest(http.MethodPost, srv.URL+"/hooks/partner", bytes.NewReader(body))
	req.Header.Set("X-Custom-Secret", "s3cret")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Fatalf("status %d", resp.StatusCode)
	}

	bad, _ := http.NewRequest(http.MethodPost, srv.URL+"/hooks/partner", bytes.NewReader(body))
	bad.Header.Set("X-Custom-Secret", "nope")
	respBad, _ := http.DefaultClient.Do(bad)
	defer respBad.Body.Close()
	if respBad.StatusCode != 400 {
		t.Fatalf("want 400 got %d", respBad.StatusCode)
	}
}

func TestHealthz(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "w.yaml")
	_ = os.WriteFile(path, []byte("triggers: []\n"), 0o644)
	cfg, err := config.Load(path)
	if err != nil {
		t.Fatal(err)
	}
	srv := httptest.NewServer(engine.New(cfg, "").Handler())
	t.Cleanup(srv.Close)
	resp, err := http.Get(srv.URL + "/healthz")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Fatalf("status %d", resp.StatusCode)
	}
}

func sign(t *testing.T, secret string, body []byte, url string) *http.Request {
	t.Helper()
	return signWithID(t, secret, body, url, "msg_test_"+strconv.FormatInt(time.Now().UnixNano(), 10))
}

func signWithID(t *testing.T, secret string, body []byte, url, msgID string) *http.Request {
	t.Helper()
	wh, err := svix.NewWebhook(secret)
	if err != nil {
		t.Fatal(err)
	}
	ts := time.Now()
	sig, err := wh.Sign(msgID, ts, body)
	if err != nil {
		t.Fatal(err)
	}
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("svix-id", msgID)
	req.Header.Set("svix-timestamp", strconv.FormatInt(ts.Unix(), 10))
	req.Header.Set("svix-signature", sig)
	return req
}
