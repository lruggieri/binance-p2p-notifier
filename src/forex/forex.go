package forex

import (
	"context"
	"time"
)

type Client interface {
	GetCurrentFxRate(ctx context.Context, fromCurrency, toCurrency string) (float64, error)
	GetMaxRate() time.Duration
}
