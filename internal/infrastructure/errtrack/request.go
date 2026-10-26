package errtrack

import (
	"fmt"
	"sync"
	"time"

	"github.com/getsentry/sentry-go"
)

// maxEventsPerRequest bounds memory when a request logs errors in a loop.
const maxEventsPerRequest = 10

// Handlers answer 5xx with c.Status(...).JSON(...) and return nil, so the original
// error is only visible in the error logs written while the request runs. Those
// logs are buffered per request ID and sent only if the response turns out to be 5xx.
var inflight sync.Map // request ID -> *pendingRequest

type pendingRequest struct {
	mu       sync.Mutex
	events   []*sentry.Event
	reported bool
}

// RequestInfo describes a finished HTTP request.
type RequestInfo struct {
	RequestID string
	Method    string
	Route     string
	Path      string
	Status    int
	UserID    string
	// Err is the error returned by the handler chain, if any.
	Err error
}

// BeginRequest starts buffering error logs tagged with requestID.
func BeginRequest(requestID string) {
	if !enabled.Load() || requestID == "" {
		return
	}
	inflight.Store(requestID, &pendingRequest{})
}

// MarkReported prevents EndRequest from sending another event for a request whose
// failure was already captured explicitly (e.g. a recovered panic).
func MarkReported(requestID string) {
	if p, ok := loadPending(requestID); ok {
		p.mu.Lock()
		p.reported = true
		p.mu.Unlock()
	}
}

// EndRequest stops buffering and, for 5xx responses, sends the first buffered error
// (the root cause) with the remaining ones attached as breadcrumbs.
func EndRequest(info RequestInfo) {
	if !enabled.Load() || info.RequestID == "" {
		return
	}

	v, ok := inflight.LoadAndDelete(info.RequestID)
	if info.Status < 500 {
		return
	}

	var events []*sentry.Event
	if ok {
		p := v.(*pendingRequest)
		p.mu.Lock()
		reported := p.reported
		events = p.events
		p.mu.Unlock()
		if reported {
			return
		}
	}

	var event *sentry.Event
	switch {
	case len(events) > 0:
		event = events[0]
		for _, e := range events[1:] {
			event.Breadcrumbs = append(event.Breadcrumbs, &sentry.Breadcrumb{
				Category:  "log",
				Level:     sentry.LevelError,
				Message:   e.Exception[0].Type + ": " + e.Exception[0].Value,
				Timestamp: e.Timestamp,
			})
		}
		if info.Err != nil {
			event.Extra["handler_error"] = info.Err.Error()
		}
	case info.Err != nil:
		// No stack trace: it would only point at this middleware.
		event = newExceptionEvent(sentry.LevelError, fmt.Sprintf("%T", info.Err), info.Err.Error(), nil)
	default:
		event = sentry.NewEvent()
		event.Level = sentry.LevelError
		event.Message = fmt.Sprintf("HTTP %d %s %s", info.Status, info.Method, info.Route)
	}

	event.Transaction = info.Method + " " + info.Route
	event.Request = &sentry.Request{Method: info.Method, URL: info.Path}

	hub := sentry.CurrentHub().Clone()
	hub.WithScope(func(scope *sentry.Scope) {
		scope.SetTags(map[string]string{
			"request_id":       info.RequestID,
			"http.method":      info.Method,
			"http.route":       info.Route,
			"http.status_code": fmt.Sprint(info.Status),
		})
		if info.UserID != "" {
			scope.SetUser(sentry.User{ID: info.UserID})
		}
		hub.CaptureEvent(event)
	})
}

func bufferEvent(requestID string, event *sentry.Event) {
	p, ok := loadPending(requestID)
	if !ok {
		// Not inside a tracked request (startup, consumer, or a background goroutine
		// that outlived its request): only explicit captures are reported there.
		return
	}

	p.mu.Lock()
	defer p.mu.Unlock()
	if len(p.events) < maxEventsPerRequest {
		event.Timestamp = time.Now()
		p.events = append(p.events, event)
	}
}

func loadPending(requestID string) (*pendingRequest, bool) {
	if requestID == "" {
		return nil, false
	}
	v, ok := inflight.Load(requestID)
	if !ok {
		return nil, false
	}
	return v.(*pendingRequest), true
}
