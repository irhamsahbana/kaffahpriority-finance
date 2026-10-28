package errtrack

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/rs/zerolog"
)

func setup(t *testing.T) (*sentry.MockTransport, zerolog.Logger) {
	t.Helper()

	transport := &sentry.MockTransport{}
	if err := sentry.Init(sentry.ClientOptions{
		Dsn:        "https://public@example.com/1",
		Transport:  transport,
		BeforeSend: scrubEvent,
	}); err != nil {
		t.Fatalf("init sentry: %v", err)
	}
	enabled.Store(true)
	t.Cleanup(func() { enabled.Store(false) })

	return transport, zerolog.New(LogWriter())
}

func TestEndRequest(t *testing.T) {
	tests := []struct {
		name       string
		status     int
		logs       int
		reported   bool
		wantEvents int
		wantValue  string
	}{
		{name: "5xx sends logged root cause", status: 500, logs: 2, wantEvents: 1, wantValue: "db down 0"},
		{name: "4xx drops logged errors", status: 400, logs: 1, wantEvents: 0},
		{name: "5xx already reported is skipped", status: 500, logs: 1, reported: true, wantEvents: 0},
		{name: "5xx without logs sends generic event", status: 503, wantEvents: 1},
	}

	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			transport, logger := setup(t)
			requestID := "req-" + string(rune('a'+i))

			BeginRequest(requestID)
			reqLogger := logger.With().Str("request_id", requestID).Logger()
			for n := 0; n < tt.logs; n++ {
				reqLogger.Error().Err(errors.New("db down " + string(rune('0'+n)))).Msg("failed to get summaries")
			}
			if tt.reported {
				MarkReported(requestID)
			}
			EndRequest(RequestInfo{RequestID: requestID, Method: "GET", Route: "/reports", Path: "/reports", Status: tt.status})

			events := transport.Events()
			if len(events) != tt.wantEvents {
				t.Fatalf("got %d events, want %d", len(events), tt.wantEvents)
			}
			if tt.wantEvents == 0 {
				return
			}

			event := events[0]
			if event.Tags["request_id"] != requestID {
				t.Errorf("request_id tag = %q, want %q", event.Tags["request_id"], requestID)
			}
			if tt.wantValue != "" {
				if got := event.Exception[0].Value; got != tt.wantValue {
					t.Errorf("exception value = %q, want %q", got, tt.wantValue)
				}
				if len(event.Breadcrumbs) != tt.logs-1 {
					t.Errorf("got %d breadcrumbs, want %d", len(event.Breadcrumbs), tt.logs-1)
				}
				frames := event.Exception[0].Stacktrace.Frames
				if top := frames[len(frames)-1]; !strings.HasSuffix(top.Function, "TestEndRequest.func1") {
					t.Errorf("top frame = %s, want the logging call site", top.Function)
				}
			}
			if _, stillTracked := loadPending(requestID); stillTracked {
				t.Error("request still buffered after EndRequest")
			}
		})
	}
}

func TestLogWriterIgnoresUntrackedAndSkippedLogs(t *testing.T) {
	transport, logger := setup(t)

	BeginRequest("req-skip")
	logger.Error().Str("request_id", "req-skip").Bool(SkipField, true).Msg("already reported")
	logger.Error().Str("request_id", "req-unknown").Msg("outside a tracked request")
	logger.Warn().Str("request_id", "req-skip").Msg("below error level")
	EndRequest(RequestInfo{RequestID: "req-skip", Status: 500})

	events := transport.Events()
	if len(events) != 1 || events[0].Message == "" {
		t.Fatalf("want only the generic 5xx event, got %d events", len(events))
	}
}

func TestSafeGoRecoversPanic(t *testing.T) {
	transport, _ := setup(t)

	done := make(chan struct{})
	SafeGo(context.Background(), func() {
		defer close(done)
		panic("boom")
	})
	<-done

	deadline := time.Now().Add(time.Second)
	for len(transport.Events()) == 0 && time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
	}

	events := transport.Events()
	if len(events) != 1 || events[0].Exception[0].Value != "boom" {
		t.Fatalf("want one panic event with value boom, got %+v", events)
	}
}

func TestDisabledIsNoop(t *testing.T) {
	transport, logger := setup(t)
	enabled.Store(false)

	BeginRequest("req-off")
	logger.Error().Str("request_id", "req-off").Msg("ignored")
	EndRequest(RequestInfo{RequestID: "req-off", Status: 500})
	CaptureError(context.Background(), errors.New("ignored"), nil)

	if n := len(transport.Events()); n != 0 {
		t.Fatalf("got %d events while disabled", n)
	}
}
