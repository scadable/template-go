package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMustLoadDefaults(t *testing.T) {
	// Ensure env vars are unset
	if err := os.Unsetenv("LISTEN_ADDR"); err != nil {
		t.Fatalf("failed to unset LISTEN_ADDR: %v", err)
	}
	if err := os.Unsetenv("OTEL_EXPORTER"); err != nil {
		t.Fatalf("failed to unset OTEL_EXPORTER: %v", err)
	}
	if err := os.Unsetenv("OTEL_SERVICE_NAME"); err != nil {
		t.Fatalf("failed to unset OTEL_SERVICE_NAME: %v", err)
	}

	cfg := MustLoad()

	assert.Equal(t, ":8080", cfg.ListenAddr)
	assert.Equal(t, "otlp", cfg.OTELExporter)
	assert.Equal(t, "template-go", cfg.OTELServiceName)
}

func TestMustLoadWithEnvOverrides(t *testing.T) {
	t.Setenv("LISTEN_ADDR", "127.0.0.1:9090")
	t.Setenv("OTEL_EXPORTER", "prometheus")
	t.Setenv("OTEL_SERVICE_NAME", "custom-service")

	cfg := MustLoad()

	assert.Equal(t, "127.0.0.1:9090", cfg.ListenAddr)
	assert.Equal(t, "prometheus", cfg.OTELExporter)
	assert.Equal(t, "custom-service", cfg.OTELServiceName)
}
