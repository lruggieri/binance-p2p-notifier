package config

const (
	// DefaultMaxSurplusPercentage : P2P ads that are less than this percentage higher that the fx price will be
	// considered valid
	DefaultMaxSurplusPercentage = 1
)

type Config struct {
	BlackList struct {
		Line []string `json:"line"`
		Bank []string `json:"bank"`
	} `json:"blackList"`
	// MaxSurplusPercentage :
	MaxSurplusPercentage float64 `json:"maxSurplusPercentage"`
	TargetCurrency       string  `json:"targetCurrency"`
}

// SetDefault : set default fields if necessary
func (c *Config) SetDefault() {
	if c.MaxSurplusPercentage == 0 {
		c.MaxSurplusPercentage = DefaultMaxSurplusPercentage
	}

	if c.TargetCurrency == "" {
		c.TargetCurrency = "JPY"
	}
}
