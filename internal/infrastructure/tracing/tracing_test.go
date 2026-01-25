package tracing

import (
	"codebase-app/internal/infrastructure/config"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseHeaders(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected map[string]string
	}{
		{
			name:     "empty string",
			input:    "",
			expected: map[string]string{},
		},
		{
			name:  "single header",
			input: "api-key=12345",
			expected: map[string]string{
				"api-key": "12345",
			},
		},
		{
			name:  "multiple headers",
			input: "api-key=12345,custom-header=value",
			expected: map[string]string{
				"api-key":       "12345",
				"custom-header": "value",
			},
		},
		{
			name:  "headers with spaces",
			input: " api-key = 12345 , custom-header = value ",
			expected: map[string]string{
				"api-key":       "12345",
				"custom-header": "value",
			},
		},
		{
			name:  "header with equals sign in value",
			input: "auth=Bearer token=123",
			expected: map[string]string{
				"auth": "Bearer token=123",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseHeaders(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestInitTracerConfig(t *testing.T) {
	// Setup mock config
	config.Envs = &config.Config{}
	config.Envs.Instrumentation.OtlpEndpoint = "localhost:4317"
	config.Envs.Instrumentation.OtlpHeaders = "api-key=test-key"
	config.Envs.Instrumentation.OtlpInsecure = true // Default for test to avoid connecting

	// We can't easily assert that the exporter has the headers without inspecting internal state,
	// but we can at least ensure InitTracer doesn't panic and returns a provider.
	
	cfg := &Config{
		AppName:    "test-app",
		AppVersion: "1.0.0",
		AppEnv:     "test",
		LogWriter:  nil,
	}

	tp, err := InitTracer(cfg)
	assert.NoError(t, err)
	assert.NotNil(t, tp)
}
