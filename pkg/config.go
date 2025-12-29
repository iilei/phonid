package phonid

import (
	"errors"
	"fmt"

	"github.com/creasty/defaults"
)

// Config holds the configuration for phonetic ID generation.
type (
	Config struct {
		// ID format settings
		Phonetic *PhonidConfig `default:"{}"`

		// Feistel shuffler settings
		Shuffle *ShuffleConfig `default:"{}"`

		// Optional: Expected BitWidth for preflight assertion
		// If set, Validate() will fail if calculated BitWidth doesn't match
		ExpectedBitWidth int
	}

	// ConfigOption is a functional option for configuring Config.
	ConfigOption func(*Config)
)

// NewConfig returns a Config with sensible defaults applied.
// BitWidth is auto-calculated during Validate() based on phonetic patterns.
func NewConfig() (*Config, error) {
	cfg := &Config{}
	if err := defaults.Set(cfg); err != nil {
		return nil, fmt.Errorf("failed to set defaults: %w", err)
	}
	return cfg, nil
}

// NewConfigWithOptions returns a Config with defaults, then applies the provided options.
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

// Validate checks if the config values are valid and auto-calculates BitWidth.
func (c *Config) Validate() error {
	// Ensure required fields are initialized
	if c.Shuffle == nil {
		return errors.New("shuffle config is required")
	}
	if c.Phonetic == nil {
		return errors.New("phonetic config is required")
	}

	// Validate phonetic config first
	if err := c.Phonetic.Validate(); err != nil {
		return fmt.Errorf("phonetic config invalid: %w", err)
	}

	// Create encoder to determine optimal bit width
	encoder, err := NewPhoneticEncoder(c.Phonetic)
	if err != nil {
		return fmt.Errorf("failed to create encoder: %w", err)
	}

	if len(encoder.patternEncoders) == 0 {
		return errors.New("no valid patterns configured")
	}

	// Auto-calculate BitWidth from largest pattern's capacity
	largestPattern := encoder.patternEncoders[len(encoder.patternEncoders)-1]
	c.Shuffle.BitWidth = calculateRequiredBitWidth(int(largestPattern.totalCombinations))

	// Preflight assertion: check if BitWidth matches expected value
	if c.ExpectedBitWidth > 0 && c.Shuffle.BitWidth != c.ExpectedBitWidth {
		return fmt.Errorf(
			"preflight assertion failed: calculated BitWidth is %d, but expected %d\n"+
				"This indicates a breaking change in the phonetic configuration.\n"+
				"Update ExpectedBitWidth to %d if this change is intentional",
			c.Shuffle.BitWidth,
			c.ExpectedBitWidth,
			c.Shuffle.BitWidth,
		)
	}

	// Validate shuffle config after BitWidth is set
	if err := c.Shuffle.Validate(); err != nil {
		return fmt.Errorf("shuffle config invalid: %w", err)
	}

	return nil
}

// WithRounds sets the number of Feistel rounds.
func WithRounds(rounds int) ConfigOption {
	return func(c *Config) {
		if c.Shuffle == nil {
			c.Shuffle = &ShuffleConfig{}
		}
		c.Shuffle.Rounds = rounds
	}
}

// WithSeed sets the seed value.
func WithSeed(seed uint64) ConfigOption {
	return func(c *Config) {
		if c.Shuffle == nil {
			c.Shuffle = &ShuffleConfig{}
		}
		c.Shuffle.Seed = seed
	}
}

// WithShuffle sets the shuffle configuration.
func WithShuffle(shuffle *ShuffleConfig) ConfigOption {
	return func(c *Config) {
		c.Shuffle = shuffle
	}
}

// WithPhonetic sets the phonetic configuration.
func WithPhonetic(phonetic *PhonidConfig) ConfigOption {
	return func(c *Config) {
		c.Phonetic = phonetic
	}
}

// WithExpectedBitWidth sets the expected bit width for preflight assertion.
// If the calculated BitWidth doesn't match, Validate() will fail.
// This helps catch breaking changes in phonetic configuration.
func WithExpectedBitWidth(bitWidth int) ConfigOption {
	return func(c *Config) {
		c.ExpectedBitWidth = bitWidth
	}
}

// calculateRequiredBitWidth returns the minimum bit width needed to represent totalCombinations.
func calculateRequiredBitWidth(totalCombinations int) int {
	if totalCombinations <= 1 {
		return 1
	}
	// Calculate ceil(log2(totalCombinations))
	bitWidth := 0
	value := totalCombinations - 1
	for value > 0 {
		bitWidth++
		value >>= 1
	}
	return bitWidth
}
