package framework

import (
	"context"
	"fmt"
	"time"
)

// Check checks the health of all configured services plus any custom health checkers
// registered via WithHealthChecker.
func (b *Bootstrap) Health(ctx context.Context) map[string]interface{} {
	health := make(map[string]interface{})

	if b.svc != nil && b.svc.database != nil {
		if err := b.svc.database.Ping(); err != nil {
			health["database"] = err
		} else {
			health["database"] = "OK"
		}
	}

	if b.svc != nil && b.svc.redisCache != nil {
		ctxTimeout, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
		if err := b.svc.redisCache.Ping(ctxTimeout); err != nil {
			health["redis"] = err
		} else {
			health["redis"] = "OK"
		}
	}

	if b.svc != nil && b.svc.rabbitmq != nil {
		if err := b.svc.rabbitmq.Health(); err != nil {
			health["rabbitmq"] = err
		} else {
			health["rabbitmq"] = "OK"
		}
	}

	if b.svc != nil && b.svc.kafka != nil {
		if b.svc.kafka.IsClosed() {
			health["kafka"] = fmt.Errorf("client closed")
		} else {
			health["kafka"] = "OK"
		}
	}

	for _, entry := range b.healthCheckers {
		if err := entry.fn(ctx); err != nil {
			health[entry.name] = err
		} else {
			health[entry.name] = "OK"
		}
	}

	return health
}
