package config

import (
	"testing"

	"github.com/spf13/viper"
)

func TestNewLogger(t *testing.T) {
	v := viper.New()
	v.Set("SERVICE_NAME", "test-service")
	v.Set("APP_ENV", "test")
	v.Set("LOG_LEVEL", "debug")

	log, err := NewLogger(v)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if log == nil {
		t.Fatal("expected non-nil logger")
	}

	// Test fallback levels
	for _, lvl := range []string{"info", "warn", "error", "unknown"} {
		v.Set("LOG_LEVEL", lvl)
		l, err := NewLogger(v)
		if err != nil || l == nil {
			t.Fatalf("failed to create logger for level %s: %v", lvl, err)
		}
	}
}
