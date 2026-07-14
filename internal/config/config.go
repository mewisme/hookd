package config

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"

	"hookd/internal/plugin"
)

type Config struct {
	Triggers []Trigger `yaml:"triggers"`
}

type Trigger struct {
	ID          string `yaml:"id"`
	Type        string `yaml:"type"`
	Path        string `yaml:"path"`
	Secret      string `yaml:"secret"`
	Header      string `yaml:"header,omitempty"` // generic plugin
	FailOnError bool   `yaml:"fail_on_error"`
	Hooks       []Hook `yaml:"hooks"`
}

type Hook struct {
	Event   string           `yaml:"event"`
	Actions []map[string]any `yaml:"actions"`
}

var envRe = regexp.MustCompile(`\$\{([A-Za-z_][A-Za-z0-9_]*)\}`)

func expandEnv(s string) string {
	return envRe.ReplaceAllStringFunc(s, func(m string) string {
		key := envRe.FindStringSubmatch(m)[1]
		return os.Getenv(key)
	})
}

func expandAny(v any) any {
	switch t := v.(type) {
	case string:
		return expandEnv(t)
	case map[string]any:
		out := make(map[string]any, len(t))
		for k, vv := range t {
			out[k] = expandAny(vv)
		}
		return out
	case map[any]any:
		out := make(map[string]any, len(t))
		for k, vv := range t {
			out[fmt.Sprint(k)] = expandAny(vv)
		}
		return out
	case []any:
		out := make([]any, len(t))
		for i, vv := range t {
			out[i] = expandAny(vv)
		}
		return out
	default:
		return v
	}
}

func Load(path string) (*Config, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	expanded := expandEnv(string(b))
	var cfg Config
	if err := yaml.Unmarshal([]byte(expanded), &cfg); err != nil {
		return nil, err
	}
	for i := range cfg.Triggers {
		t := &cfg.Triggers[i]
		if t.ID == "" {
			return nil, fmt.Errorf("trigger[%d]: id required", i)
		}
		if t.Type == "" {
			return nil, fmt.Errorf("trigger %q: type required", t.ID)
		}
		if t.Path == "" {
			return nil, fmt.Errorf("trigger %q: path required", t.ID)
		}
		if !strings.HasPrefix(t.Path, "/") {
			t.Path = "/" + t.Path
		}
		if _, err := plugin.GetTrigger(t.Type); err != nil {
			return nil, fmt.Errorf("trigger %q: %w", t.ID, err)
		}
		for j, h := range t.Hooks {
			if h.Event == "" {
				return nil, fmt.Errorf("trigger %q hook[%d]: event required", t.ID, j)
			}
			for k, a := range h.Actions {
				typ, _ := a["type"].(string)
				if typ == "" {
					return nil, fmt.Errorf("trigger %q hook[%d] action[%d]: type required", t.ID, j, k)
				}
				if _, err := plugin.GetAction(typ); err != nil {
					return nil, fmt.Errorf("trigger %q hook[%d] action[%d]: %w", t.ID, j, k, err)
				}
				t.Hooks[j].Actions[k] = expandAny(a).(map[string]any)
			}
		}
	}
	return &cfg, nil
}
