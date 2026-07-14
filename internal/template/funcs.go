package template

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	texttemplate "text/template"
)

var eventTitles = map[string]string{
	"email.failed":   "Email Failed",
	"email.bounced":  "Email Bounced",
	"email.clicked":  "Email Clicked",
	"email.opened":   "Email Opened",
	"email.received": "Email Received",
}

var eventColors = map[string]int{
	"email.failed":   15158332,
	"email.bounced":  16753920,
	"email.clicked":  3447003,
	"email.opened":   5763719,
	"email.received": 3066993,
}

func FuncMap() texttemplate.FuncMap {
	return texttemplate.FuncMap{
		"default": func(v any, def string) string {
			s := stringify(v)
			if s == "" || s == "<nil>" || s == "None" {
				return def
			}
			return s
		},
		"join": func(v any, sep string) string {
			switch t := v.(type) {
			case []any:
				parts := make([]string, 0, len(t))
				for _, x := range t {
					parts = append(parts, stringify(x))
				}
				if len(parts) == 0 {
					return "N/A"
				}
				return strings.Join(parts, sep)
			case []string:
				if len(t) == 0 {
					return "N/A"
				}
				return strings.Join(t, sep)
			case string:
				if t == "" {
					return "N/A"
				}
				return t
			case nil:
				return "N/A"
			default:
				return stringify(v)
			}
		},
		"json": func(v any) string {
			if v == nil {
				return "None"
			}
			b, err := json.Marshal(v)
			if err != nil {
				return "None"
			}
			return string(b)
		},
		"eventTitle": func(event string) string {
			if t, ok := eventTitles[event]; ok {
				return t
			}
			return "Resend Event"
		},
		"eventColor": func(event string) int {
			if c, ok := eventColors[event]; ok {
				return c
			}
			return 9807270
		},
	}
}

func stringify(v any) string {
	if v == nil {
		return ""
	}
	switch t := v.(type) {
	case string:
		return t
	case float64:
		if t == float64(int64(t)) {
			return fmt.Sprintf("%d", int64(t))
		}
		return fmt.Sprintf("%v", t)
	default:
		return fmt.Sprintf("%v", t)
	}
}

// Render applies Go text/template with helpers against data.
func Render(tmpl string, data any) (string, error) {
	t, err := texttemplate.New("").Funcs(FuncMap()).Option("missingkey=zero").Parse(tmpl)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// Context builds the template root for an event map.
func Context(event string, createdAt string, data map[string]any, raw string) map[string]any {
	if data == nil {
		data = map[string]any{}
	}
	return map[string]any{
		"event":      event,
		"created_at": createdAt,
		"data":       data,
		"raw":        raw,
	}
}
