package httpaction

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"hookd/internal/plugin"
	tpl "hookd/internal/template"
)

func init() {
	plugin.RegisterAction(&Action{client: &http.Client{Timeout: 15 * time.Second}})
}

type Action struct {
	client *http.Client
}

func (a *Action) Type() string { return "http" }

func (a *Action) Run(ctx context.Context, cfg map[string]any, event plugin.Event) error {
	method, _ := cfg["method"].(string)
	if method == "" {
		method = "POST"
	}
	urlStr, _ := cfg["url"].(string)
	if urlStr == "" {
		return fmt.Errorf("http: url required")
	}
	bodyTmpl, _ := cfg["body"].(string)

	data := tpl.Context(event.Event, event.CreatedAt, event.Data, event.Raw)
	urlStr, err := tpl.Render(urlStr, data)
	if err != nil {
		return fmt.Errorf("http: url template: %w", err)
	}
	var body string
	if bodyTmpl != "" {
		body, err = tpl.Render(bodyTmpl, data)
		if err != nil {
			return fmt.Errorf("http: body template: %w", err)
		}
	}

	req, err := http.NewRequestWithContext(ctx, method, urlStr, strings.NewReader(body))
	if err != nil {
		return err
	}
	if headers, ok := cfg["headers"].(map[string]any); ok {
		for k, v := range headers {
			hs := fmt.Sprint(v)
			hs, err = tpl.Render(hs, data)
			if err != nil {
				return fmt.Errorf("http: header %s template: %w", k, err)
			}
			req.Header.Set(k, hs)
		}
	}
	if body != "" && req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := a.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("http: status %d", resp.StatusCode)
	}
	return nil
}
