package logx_test

import (
	"testing"

	"hookd/internal/logx"
)

func TestNewLevels(t *testing.T) {
	for _, level := range []string{"debug", "info", "warn", "warning", "error", ""} {
		if logx.New(level) == nil {
			t.Fatalf("nil logger for %q", level)
		}
	}
}
