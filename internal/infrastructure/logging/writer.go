package logging

import (
	"context"
	"encoding/json"
	"io"
	"time"

	"go.opentelemetry.io/otel/log"
)

type OtelWriter struct {
	Next   io.Writer
	Logger log.Logger
}

func (w *OtelWriter) Write(p []byte) (n int, err error) {
	// Write to the next writer (console/file) first
	if w.Next != nil {
		n, err = w.Next.Write(p)
		if err != nil {
			return n, err
		}
	} else {
		n = len(p)
	}

	// Parse JSON log
	var data map[string]interface{}
	if err := json.Unmarshal(p, &data); err != nil {
		// If not JSON, send as plain text body
		r := log.Record{}
		r.SetTimestamp(time.Now())
		r.SetBody(log.StringValue(string(p)))
		w.Logger.Emit(context.Background(), r)
		return n, nil
	}

	r := log.Record{}

	// Handle Timestamp
	if tStr, ok := data["time"].(string); ok {
		if t, err := time.Parse(time.RFC3339, tStr); err == nil {
			r.SetTimestamp(t)
		} else {
			r.SetTimestamp(time.Now())
		}
	} else {
		r.SetTimestamp(time.Now())
	}

	// Handle Message
	if msg, ok := data["message"].(string); ok {
		r.SetBody(log.StringValue(msg))
	}

	// Handle Level
	if lvl, ok := data["level"].(string); ok {
		r.SetSeverity(toSeverity(lvl))
		r.SetSeverityText(lvl)
	}

	// Handle Attributes
	for k, v := range data {
		if k == "time" || k == "message" || k == "level" {
			continue
		}

		// Simple conversion for now
		switch val := v.(type) {
		case string:
			r.AddAttributes(log.String(k, val))
		case float64:
			r.AddAttributes(log.Float64(k, val))
		case bool:
			r.AddAttributes(log.Bool(k, val))
		default:
			if jsonBytes, err := json.Marshal(val); err == nil {
				r.AddAttributes(log.String(k, string(jsonBytes)))
			} else {
				r.AddAttributes(log.String(k, "unsupported_type"))
			}
		}
	}

	w.Logger.Emit(context.Background(), r)

	return n, nil
}

func toSeverity(level string) log.Severity {
	switch level {
	case "trace":
		return log.SeverityTrace
	case "debug":
		return log.SeverityDebug
	case "info":
		return log.SeverityInfo
	case "warn":
		return log.SeverityWarn
	case "error":
		return log.SeverityError
	case "fatal":
		return log.SeverityFatal
	case "panic":
		return log.SeverityFatal
	default:
		return log.SeverityInfo
	}
}
