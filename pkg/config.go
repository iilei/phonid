package phonid

import (
	"fmt"

	"github.com/creasty/defaults"
)

// Config holds the configuration for phonetic ID generation
type (
	Config struct {
		// ID format settings
		Phonetic *PhonidConfig `default:"{}"`

		// Feistel shuffler settings
		Shuffle *ShuffleConfig `default:"{}"`
	}

	// ConfigOption is a functional option for configuring Config
	ConfigOption func(*Config)
)

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

// Validate checks if the config values are valid
func (c *Config) Validate() error {
	// Validate shuffle config
	if c.Shuffle != nil {
		if err := c.Shuffle.Validate(); err != nil {
			return fmt.Errorf("shuffle config invalid: %w", err)
		}
	}

	// Validate phonetic config if provided
	if c.Phonetic != nil {
		if err := c.Phonetic.Validate(); err != nil {
			return fmt.Errorf("phonetic config invalid: %w", err)
		}
	}

	return nil
}

// WithBitWidth sets the bit width
func WithBitWidth(bitWidth int) ConfigOption {
	return func(c *Config) {
		if c.Shuffle == nil {
			c.Shuffle = &ShuffleConfig{}
		}
		c.Shuffle.BitWidth = bitWidth
	}
}

// WithRounds sets the number of Feistel rounds
func WithRounds(rounds int) ConfigOption {
	return func(c *Config) {
		if c.Shuffle == nil {
			c.Shuffle = &ShuffleConfig{}
		}
		c.Shuffle.Rounds = rounds
	}
}

// WithSeed sets the seed value
func WithSeed(seed uint64) ConfigOption {
	return func(c *Config) {
		if c.Shuffle == nil {
			c.Shuffle = &ShuffleConfig{}
		}
		c.Shuffle.Seed = seed
	}
}

// WithShuffle sets the shuffle configuration
func WithShuffle(shuffle *ShuffleConfig) ConfigOption {
	return func(c *Config) {
		c.Shuffle = shuffle
	}
}

// WithPhonetic sets the phonetic configuration
func WithPhonetic(phonetic *PhonidConfig) ConfigOption {
	return func(c *Config) {
		c.Phonetic = phonetic
	}
}
