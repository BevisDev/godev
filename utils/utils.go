package utils

import (
	"context"
	"github.com/BevisDev/godev/consts"
	"github.com/BevisDev/godev/utils/random"
	"strings"
	"time"
)

func GetState(ctx context.Context) string {
	if ctx == nil {
		return random.RandUUID()
	}
	state, ok := ctx.Value(consts.State).(string)
	if !ok {
		state = random.RandUUID()
	}
	return state
}

func CreateCtx() context.Context {
	return context.WithValue(context.Background(), consts.State, random.RandUUID())
}

func CreateCtxTimeout(ctx context.Context, timeoutSec int) (context.Context, context.CancelFunc) {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithTimeout(ctx, time.Duration(timeoutSec)*time.Second)
}

func CreateCtxCancel(ctx context.Context) (context.Context, context.CancelFunc) {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithCancel(ctx)
}

func ContainsIgnoreCase(s, substr string) bool {
	return strings.Contains(
		strings.ToLower(s),
		strings.ToLower(substr),
	)
}
