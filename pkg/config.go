package phonid

import (
	"fmt"

	"github.com/creasty/defaults"
)

type BaseEncoding int

const (
	Base36 BaseEncoding = 36 // case-insensitive (0-9, a-z)
	Base62 BaseEncoding = 62 // case-sensitive (0-9, a-z, A-Z)
)

// Config holds the configuration for phonetic ID generation
type Config struct {
	// Feistel shuffler settings
	BitWidth int    `default:"32"`
	Rounds   int    `default:"4"`
	Seed     uint64 `default:"0"`

	// ID format settings
	Prefix string
	Base   BaseEncoding `default:"36"`
}

// NewConfig returns a Config with sensible defaults applied
func NewConfig() (*Config, error) {
	cfg := &Config{}
	if err := defaults.Set(cfg); err != nil {
		return nil, fmt.Errorf("failed to set defaults: %w", err)
	}
	return cfg, nil
}

// NewConfigWithOptions returns a Config with defaults, then applies the provided options
func NewConfigWithOptions(opts ...ConfigOption) (*Config, error) {
	cfg, err := NewConfig()
	if err != nil {
		return nil, err
	}

	for _, opt := range opts {
		opt(cfg)
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// ConfigOption is a functional option for configuring Config
type ConfigOption func(*Config)

// WithBitWidth sets the bit width
func WithBitWidth(bitWidth int) ConfigOption {
	return func(c *Config) {
		c.BitWidth = bitWidth
	}
}

// WithRounds sets the number of Feistel rounds
func WithRounds(rounds int) ConfigOption {
	return func(c *Config) {
		c.Rounds = rounds
	}
}

// WithSeed sets the seed value
func WithSeed(seed uint64) ConfigOption {
	return func(c *Config) {
		c.Seed = seed
	}
}

// WithPrefix sets the ID prefix
func WithPrefix(prefix string) ConfigOption {
	return func(c *Config) {
		c.Prefix = prefix
	}
}

// WithBase sets the base encoding (36 or 62)
func WithBase(base int) ConfigOption {
	return func(c *Config) {
		c.Base = BaseEncoding(base)
	}
}

// Validate checks if the config values are valid
func (c *Config) Validate() error {
	if c.BitWidth < 4 || c.BitWidth > 64 {
		return fmt.Errorf("bit_width must be between 4 and 64, got %d", c.BitWidth)
	}
	if c.Rounds < 3 || c.Rounds > 10 {
		return fmt.Errorf("rounds must be between 3 and 10, got %d", c.Rounds)
	}
	if c.Base != Base36 && c.Base != Base62 {
		return fmt.Errorf("base must be Base36 (36) or Base62 (62), got %d", c.Base)
	}
	return nil
}
