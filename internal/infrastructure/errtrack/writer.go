package errtrack

import (
	"encoding/json"

	"github.com/getsentry/sentry-go"
	"github.com/rs/zerolog"
)

type logWriter struct{}

// LogWriter returns a zerolog writer that turns error-level logs written during an
// HTTP request into pending Sentry events (see EndRequest). Zerolog writes
// synchronously, so the stack trace captured here points at the logging call site.
func LogWriter() zerolog.LevelWriter {
	return logWriter{}
}

func (logWriter) Write(p []byte) (int, error) {
	return len(p), nil
}

// WriteLevel never returns an error so error tracking can't break application logging.
func (logWriter) WriteLevel(level zerolog.Level, p []byte) (int, error) {
	if !enabled.Load() || level < zerolog.ErrorLevel || level > zerolog.PanicLevel {
		return len(p), nil
	}

	var entry struct {
		Message   string `json:"message"`
		Error     string `json:"error"`
		RequestID string `json:"request_id"`
		Caller    string `json:"caller"`
		Skip      bool   `json:"errtrack_skip"`
	}
	if err := json.Unmarshal(p, &entry); err != nil || entry.Skip {
		return len(p), nil
	}
	if _, ok := loadPending(entry.RequestID); !ok {
		return len(p), nil
	}

	value := entry.Error
	if value == "" {
		value = entry.Message
	}
	excType := entry.Message
	if excType == "" {
		excType = "error log"
	}

	event := newExceptionEvent(sentry.LevelError, excType, value, currentStacktrace())
	if entry.Caller != "" {
		event.Extra["caller"] = entry.Caller
	}
	bufferEvent(entry.RequestID, event)

	return len(p), nil
}
