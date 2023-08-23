package forex

import "time"

type Client interface {
	GetCurrentFxRate(fromCurrency, toCurrency string) (float64, error)
	GetMaxRate() time.Duration
}
