package plugin

import (
	"context"
	"fmt"
	"net/http"
	"sort"
	"sync"
)

// Event is the normalized payload passed to actions after a trigger parses a webhook.
type Event struct {
	Event     string         `json:"event"`
	CreatedAt string         `json:"created_at"`
	Data      map[string]any `json:"data"`
	Raw       string         `json:"raw"`
}

// Trigger verifies and parses an inbound webhook.
type Trigger interface {
	Type() string
	Verify(headers http.Header, body []byte, secret string) error
	Parse(body []byte) (event string, data map[string]any, err error)
	// DeliveryID returns a dedup key from headers (empty = skip dedup).
	DeliveryID(headers http.Header) string
}

// Action runs one outbound step.
type Action interface {
	Type() string
	Run(ctx context.Context, cfg map[string]any, event Event) error
}

var (
	triggersMu sync.RWMutex
	triggers   = map[string]Trigger{}
	actionsMu  sync.RWMutex
	actions    = map[string]Action{}
)

func RegisterTrigger(t Trigger) {
	triggersMu.Lock()
	defer triggersMu.Unlock()
	triggers[t.Type()] = t
}

func RegisterAction(a Action) {
	actionsMu.Lock()
	defer actionsMu.Unlock()
	actions[a.Type()] = a
}

func GetTrigger(typ string) (Trigger, error) {
	triggersMu.RLock()
	defer triggersMu.RUnlock()
	t, ok := triggers[typ]
	if !ok {
		return nil, fmt.Errorf("unknown trigger type %q", typ)
	}
	return t, nil
}

func GetAction(typ string) (Action, error) {
	actionsMu.RLock()
	defer actionsMu.RUnlock()
	a, ok := actions[typ]
	if !ok {
		return nil, fmt.Errorf("unknown action type %q", typ)
	}
	return a, nil
}

// TriggerTypes returns sorted registered trigger type names.
func TriggerTypes() []string {
	triggersMu.RLock()
	defer triggersMu.RUnlock()
	out := make([]string, 0, len(triggers))
	for k := range triggers {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// ActionTypes returns sorted registered action type names.
func ActionTypes() []string {
	actionsMu.RLock()
	defer actionsMu.RUnlock()
	out := make([]string, 0, len(actions))
	for k := range actions {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
