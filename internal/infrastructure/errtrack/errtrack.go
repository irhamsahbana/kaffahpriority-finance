// Package errtrack reports errors and panics to Sentry (or any Sentry-compatible
// backend such as GlitchTip, by pointing the DSN there).
// Every exported function is a no-op until Init is called with a non-empty DSN,
// so environments without a DSN behave exactly as before.
package errtrack

import (
	"context"
	"fmt"
	"runtime/debug"
	"strings"
	"sync/atomic"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/rs/zerolog/log"
)

// SkipField marks a zerolog event that must not be collected by LogWriter,
// used when the same failure is already reported explicitly (e.g. panics).
const SkipField = "errtrack_skip"

const (
	pkgPath      = "codebase-app/internal/infrastructure/errtrack"
	flushTimeout = 2 * time.Second
)

var enabled atomic.Bool

type Config struct {
	DSN         string
	Environment string
	Release     string
	SampleRate  float64
}

// Init configures the Sentry client. An empty DSN keeps error tracking disabled.
func Init(cfg Config) error {
	if cfg.DSN == "" {
		return nil
	}

	err := sentry.Init(sentry.ClientOptions{
		Dsn:         cfg.DSN,
		Environment: cfg.Environment,
		// Empty release lets the SDK derive it from the VCS revision embedded at build time.
		Release:          cfg.Release,
		SampleRate:       cfg.SampleRate,
		AttachStacktrace: true,
		// Finance data: never send IPs, cookies, headers or bodies.
		SendDefaultPII: false,
		BeforeSend:     scrubEvent,
	})
	if err != nil {
		return fmt.Errorf("init sentry: %w", err)
	}

	enabled.Store(true)
	return nil
}

func Enabled() bool {
	return enabled.Load()
}

// Flush waits for queued events to be delivered; call it before the process exits.
func Flush() {
	if !enabled.Load() {
		return
	}
	sentry.Flush(flushTimeout)
}

// CaptureError reports err with a stack trace pointing at the caller.
func CaptureError(ctx context.Context, err error, tags map[string]string) {
	if !enabled.Load() || err == nil {
		return
	}

	event := newExceptionEvent(sentry.LevelError, fmt.Sprintf("%T", err), err.Error(), currentStacktrace())
	capture(ctx, event, tags)
}

// CaptureRecovered reports a value returned by recover(). Call it from the deferred
// function so the stack trace still contains the frame that panicked.
func CaptureRecovered(ctx context.Context, recovered any, tags map[string]string) {
	if !enabled.Load() || recovered == nil {
		return
	}

	value := fmt.Sprint(recovered)
	if err, ok := recovered.(error); ok {
		value = err.Error()
	}

	event := newExceptionEvent(sentry.LevelFatal, "panic", value, currentStacktrace())
	handled := false
	event.Exception[0].Mechanism = &sentry.Mechanism{Type: "recover", Handled: &handled}
	capture(ctx, event, tags)
}

// SafeGo runs fn in a new goroutine and recovers any panic, so a failing background
// task (e.g. writing an activity log) is reported instead of crashing the process.
func SafeGo(ctx context.Context, fn func()) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Ctx(ctx).Error().
					Bool(SkipField, true).
					Interface("panic", r).
					Str("stack", string(debug.Stack())).
					Msg("recovered from panic in background goroutine")
				CaptureRecovered(ctx, r, map[string]string{"goroutine": "background"})
			}
		}()
		fn()
	}()
}

func capture(ctx context.Context, event *sentry.Event, tags map[string]string) {
	hub := sentry.CurrentHub().Clone()
	hub.WithScope(func(scope *sentry.Scope) {
		if requestID := requestIDFromContext(ctx); requestID != "" {
			scope.SetTag("request_id", requestID)
		}
		scope.SetTags(tags)
		hub.CaptureEvent(event)
	})
}

func newExceptionEvent(level sentry.Level, excType, value string, st *sentry.Stacktrace) *sentry.Event {
	event := sentry.NewEvent()
	event.Level = level
	event.Exception = []sentry.Exception{{
		Type:       excType,
		Value:      value,
		Stacktrace: st,
	}}
	return event
}

// currentStacktrace returns the caller's stack without frames from this package
// and zerolog, so Sentry groups issues by application code instead of plumbing.
func currentStacktrace() *sentry.Stacktrace {
	st := sentry.NewStacktrace()
	if st == nil {
		return nil
	}

	// Frames are ordered oldest first; trim plumbing from the newest end only.
	frames := st.Frames
	for len(frames) > 0 {
		frame := frames[len(frames)-1]
		ownFrame := frame.Module == pkgPath && !strings.HasSuffix(frame.AbsPath, "_test.go")
		if !ownFrame && !strings.HasPrefix(frame.Module, "github.com/rs/zerolog") {
			break
		}
		frames = frames[:len(frames)-1]
	}
	st.Frames = frames
	return st
}

func requestIDFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	requestID, _ := ctx.Value("request_id").(string)
	return requestID
}

// scrubEvent is a last line of defense in case an integration attaches request data.
func scrubEvent(event *sentry.Event, _ *sentry.EventHint) *sentry.Event {
	if event.Request != nil {
		event.Request.Headers = nil
		event.Request.Cookies = ""
		event.Request.Data = ""
		event.Request.QueryString = ""
		event.Request.Env = nil
	}
	event.User.IPAddress = ""
	event.User.Email = ""
	return event
}
