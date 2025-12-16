package pkg_test

import (
	"codebase-app/pkg"
	"context"
	"testing"
	"time"
)

func TestGenerateChildContext(t *testing.T) {
	t.Run("sets deadline close to timeout", func(t *testing.T) {
		timeout := 100 * time.Millisecond
		start := time.Now()
		ctx := pkg.GenerateChildContext(context.Background(), timeout)

		deadline, ok := ctx.Deadline()
		if !ok {
			t.Fatalf("expected deadline to be set")
		}

		delta := deadline.Sub(start)
		if delta < timeout-30*time.Millisecond || delta > timeout+150*time.Millisecond {
			t.Fatalf("deadline delta out of range: got %v, want around %v", delta, timeout)
		}
	})

	t.Run("detached from parent cancel and times out independently", func(t *testing.T) {
		parent, cancel := context.WithCancel(context.Background())
		timeout := 50 * time.Millisecond
		ctx := pkg.GenerateChildContext(parent, timeout)

		cancel() // cancel parent immediately
		time.Sleep(10 * time.Millisecond)

		if err := ctx.Err(); err != nil && err != context.DeadlineExceeded {
			t.Fatalf("child context canceled unexpectedly by parent: %v", err)
		}

		select {
		case <-ctx.Done():
			if ctx.Err() != context.DeadlineExceeded {
				t.Fatalf("expected DeadlineExceeded, got %v", ctx.Err())
			}
		case <-time.After(200 * time.Millisecond):
			t.Fatalf("child context did not timeout within expected duration")
		}
	})

	t.Run("propagates values from parent", func(t *testing.T) {
		type keyType struct{}
		key := keyType{}
		parent := context.WithValue(context.Background(), key, "val")
		ctx := pkg.GenerateChildContext(parent, 10*time.Millisecond)

		if got := ctx.Value(key); got != "val" {
			t.Fatalf("expected value 'val', got %v", got)
		}
	})
}
