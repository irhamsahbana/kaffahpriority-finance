package pkg

import (
	"context"
	"time"
)

func GenerateChildContext(ctx context.Context, timeout time.Duration) context.Context {
	detached := context.WithoutCancel(ctx)
	bgCtx, _ := context.WithTimeout(detached, timeout)
	return bgCtx
}
