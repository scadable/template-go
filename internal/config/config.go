package config

import "os"

// Config holds basic runtime configuration.
type Config struct {
	ListenAddr      string
	OTELExporter    string
	OTELServiceName string
}

// MustLoad loads configuration from environment variables or defaults.
func MustLoad() Config {
	return Config{
		ListenAddr:      getenv("LISTEN_ADDR", ":8080"),
		OTELExporter:    getenv("OTEL_EXPORTER", "otlp"),
		OTELServiceName: getenv("OTEL_SERVICE_NAME", "template-go"),
	}
}

// getenv retrieves an environment variable or returns a fallback value.
func getenv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
