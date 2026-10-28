package middleware

import (
	"codebase-app/internal/infrastructure/errtrack"
	"codebase-app/pkg/response"
	"errors"
	"net/http/httptest"
	"testing"

	"github.com/getsentry/sentry-go"
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func newTestApp(t *testing.T) (*fiber.App, *sentry.MockTransport) {
	t.Helper()

	if err := errtrack.Init(errtrack.Config{DSN: "https://public@example.com/1"}); err != nil {
		t.Fatalf("init errtrack: %v", err)
	}
	transport := &sentry.MockTransport{}
	client, err := sentry.NewClient(sentry.ClientOptions{Dsn: "https://public@example.com/1", Transport: transport})
	if err != nil {
		t.Fatalf("new client: %v", err)
	}
	sentry.CurrentHub().BindClient(client)

	logger := zerolog.New(errtrack.LogWriter())
	app := fiber.New()
	app.Use(RequestID, WithAppLogger(logger), ErrorCapture(), Recover())

	app.Get("/fail", func(c *fiber.Ctx) error {
		log.Ctx(c.UserContext()).Error().Err(errors.New("pq: connection refused")).Msg("failed to get summaries")
		return c.Status(fiber.StatusInternalServerError).JSON(response.Error("Permintaan anda gagal diproses"))
	})
	app.Get("/invalid", func(c *fiber.Ctx) error {
		log.Ctx(c.UserContext()).Error().Err(errors.New("not found")).Msg("failed to get template")
		return c.Status(fiber.StatusNotFound).JSON(response.Error("not found"))
	})
	app.Get("/panic", func(c *fiber.Ctx) error {
		panic("nil map write")
	})

	return app, transport
}

func TestErrorCapture(t *testing.T) {
	tests := []struct {
		path       string
		wantStatus int
		wantEvents int
		wantType   string
		wantValue  string
	}{
		{path: "/fail", wantStatus: 500, wantEvents: 1, wantType: "failed to get summaries", wantValue: "pq: connection refused"},
		{path: "/invalid", wantStatus: 404, wantEvents: 0},
		{path: "/panic", wantStatus: 500, wantEvents: 1, wantType: "panic", wantValue: "nil map write"},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			app, transport := newTestApp(t)

			resp, err := app.Test(httptest.NewRequest("GET", tt.path, nil))
			if err != nil {
				t.Fatalf("request: %v", err)
			}
			if resp.StatusCode != tt.wantStatus {
				t.Fatalf("status = %d, want %d", resp.StatusCode, tt.wantStatus)
			}

			events := transport.Events()
			if len(events) != tt.wantEvents {
				t.Fatalf("got %d events, want %d", len(events), tt.wantEvents)
			}
			if tt.wantEvents == 0 {
				return
			}

			exc := events[0].Exception[0]
			if exc.Type != tt.wantType || exc.Value != tt.wantValue {
				t.Errorf("exception = %q: %q, want %q: %q", exc.Type, exc.Value, tt.wantType, tt.wantValue)
			}
			if events[0].Tags["request_id"] == "" {
				t.Error("missing request_id tag")
			}
		})
	}
}
