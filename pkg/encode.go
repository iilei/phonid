package phonid

import (
	"errors"
	"fmt"
	"math"
	"math/big"
	"slices"
	"strings"
)

type (
	// PhoneticEncoder handles encoding/decoding between numbers and phonetic words.
	PhoneticEncoder struct {
		config          *PhonidConfig
		patternEncoders []*PatternEncoder // ordered by totalCombinations ascending
	}

	// PatternEncoder represents a single pattern configuration.
	PatternEncoder struct {
		pattern           string
		positions         []Position
		totalCombinations *big.Int // Exact capacity, may exceed int
		length            int      // Number of positions/characters in the pattern
	}

	// Position represents one character position in the pattern.
	Position struct {
		placeholder string
		chars       []rune
		base        int
	}
)

// NewPhoneticEncoder creates an encoder with validated config and preflight checks.
// This is the standard way to create an encoder - preflight validation is required.
func NewPhoneticEncoder(config *PhonidConfig, checks []PreflightCheck) (*PhoneticEncoder, error) {
	encoder, err := newPhoneticEncoderValidated(config)
	if err != nil {
		return nil, err
	}

	if err := encoder.ValidatePreflight(checks); err != nil {
		return nil, fmt.Errorf("preflight validation failed: %w", err)
	}

	return encoder, nil
}

// NewPhoneticEncoderSkipPreflight creates an encoder without preflight validation.
// Only use this when preflight checks are genuinely unavailable (e.g., testing scenarios).
func NewPhoneticEncoderSkipPreflight(config *PhonidConfig) (*PhoneticEncoder, error) {
	return newPhoneticEncoderValidated(config)
}

// newPhoneticEncoderValidated creates an encoder with a validated config (internal).
func newPhoneticEncoderValidated(config *PhonidConfig) (*PhoneticEncoder, error) {
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return newPhoneticEncoder(config)
}

// NewPhoneticEncoderLenient creates an encoder with minimal validation.
// Used exclusively by 'phonid preflight --suggest' command to allow
// generating preflight-check suggestions even with incomplete configs.
// If config is nil, uses DefaultConfig (ProQuint-compatible encoding).
func NewPhoneticEncoderLenient(config *PhonidConfig) (*PhoneticEncoder, error) {
	// Skip preflight validation but still validate config structure
	if config == nil {
		config = &DefaultConfig
	}

	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return newPhoneticEncoder(config)
}

// buildPatternEncoder creates a PatternEncoder from a pattern string and placeholders.
func buildPatternEncoder(pattern string, placeholders PlaceholderMap) (*PatternEncoder, error) {
	if pattern == "" {
		return nil, errors.New("pattern cannot be empty")
	}

	positions := make([]Position, 0, len(pattern))
	totalCombinations := big.NewInt(1)

	// Parse each character in the pattern
	for i, char := range pattern {
		placeholderType := PlaceholderType(char)

		// Look up the character set for this placeholder
		chars, exists := placeholders[placeholderType]
		if !exists {
			return nil, fmt.Errorf(
				"placeholder '%c' at position %d not found in placeholders",
				char,
				i,
			)
		}

		if len(chars) == 0 {
			return nil, fmt.Errorf("placeholder '%c' has empty character set", char)
		}

		// Create position
		position := Position{
			placeholder: string(char),
			chars:       []rune(chars),
			base:        len(chars),
		}

		positions = append(positions, position)

		// Multiply using big.Int (no overflow possible)
		baseBig := big.NewInt(int64(position.base))
		totalCombinations = totalCombinations.Mul(totalCombinations, baseBig)
	}

	return &PatternEncoder{
		pattern:           pattern,
		positions:         positions,
		totalCombinations: totalCombinations,
		length:            len(positions),
	}, nil
}

// newPhoneticEncoder is the internal constructor (assumes valid config).
func newPhoneticEncoder(config *PhonidConfig) (*PhoneticEncoder, error) {
	patternEncoders := make([]*PatternEncoder, 0, len(config.Patterns))

	for _, pattern := range config.Patterns {
		encoder, err := buildPatternEncoder(pattern, config.Placeholders)
		if err != nil {
			return nil, err
		}
		patternEncoders = append(patternEncoders, encoder)
	}

	// Sort by totalCombinations
	for i := 0; i < len(patternEncoders); i++ {
		for j := i + 1; j < len(patternEncoders); j++ {
			if patternEncoders[i].totalCombinations.Cmp(patternEncoders[j].totalCombinations) > 0 {
				patternEncoders[i], patternEncoders[j] = patternEncoders[j], patternEncoders[i]
			}
		}
	}

	// Check for duplicate totalCombinations
	for i := range len(patternEncoders) - 1 {
		if patternEncoders[i].totalCombinations.Cmp(patternEncoders[i+1].totalCombinations) == 0 {
			return nil, fmt.Errorf(
				"duplicate total combinations: patterns '%s' and '%s' both produce %s combinations",
				patternEncoders[i].pattern,
				patternEncoders[i+1].pattern,
				patternEncoders[i].totalCombinations.String(),
			)
		}
	}

	return &PhoneticEncoder{
		config:          config,
		patternEncoders: patternEncoders,
	}, nil
}

// Encode converts a number to a phonetic word, automatically selecting the best pattern.
// The number can be either a native int64 or a big.Int, with automatic optimization.
func (e *PhoneticEncoder) Encode(number PositiveInt) (string, error) {
	if err := number.Validate(); err != nil {
		return "", err
	}

	// Validate input against pattern capacity
	maxCapacity := e.patternEncoders[len(e.patternEncoders)-1].totalCombinations
	if number.BigInt().Cmp(maxCapacity) >= 0 {
		maxValue := new(big.Int).Sub(maxCapacity, big.NewInt(1))
		var displayMax string
		if maxValue.Cmp(big.NewInt(int64(math.MaxInt))) > 0 {
			displayMax = fmt.Sprintf(">%d", math.MaxInt)
		} else {
			displayMax = maxValue.String()
		}
		return "", fmt.Errorf("number %s exceeds largest pattern capacity (max value: %s)",
			number.String(), displayMax)
	}

	// Find the smallest pattern that can encode this number
	for _, pattern := range e.patternEncoders {
		if number.BigInt().Cmp(pattern.totalCombinations) < 0 {
			return pattern.Encode(number)
		}
	}

	// This should never be reached due to input validation,
	// but keep as a safety net for defensive programming
	maxValue := new(big.Int).Sub(maxCapacity, big.NewInt(1))
	var displayMax string
	if maxValue.Cmp(big.NewInt(int64(math.MaxInt))) > 0 {
		displayMax = fmt.Sprintf(">%d", math.MaxInt)
	} else {
		displayMax = maxValue.String()
	}
	return "", fmt.Errorf("number %s exceeds largest pattern capacity (max value: %s)",
		number.String(), displayMax)
}

func (e *PhoneticEncoder) Decode(word string) (PositiveInt, error) {
	wordRunes := []rune(word)

	// Try to match pattern by length
	for _, pattern := range e.patternEncoders {
		if len(wordRunes) != pattern.length {
			continue
		}

		return pattern.Decode(word)
	}

	return nil, fmt.Errorf("word length %d doesn't match any pattern", len(wordRunes))
}

// Encode converts a number to a phonetic word.
// Uses optimized int64 arithmetic for small numbers, big.Int for large numbers.
func (e *PatternEncoder) Encode(number PositiveInt) (string, error) {
	// Check if number exceeds pattern capacity
	if number.BigInt().Cmp(e.totalCombinations) >= 0 {
		// For display, handle capacity > MaxInt
		var displayMax string
		maxInt := big.NewInt(int64(math.MaxInt))
		if e.totalCombinations.Cmp(maxInt) > 0 {
			displayMax = fmt.Sprintf(">%d", math.MaxInt)
		} else {
			// Capacity fits in int, show exact value
			maxVal := new(big.Int).Sub(e.totalCombinations, big.NewInt(1))
			displayMax = maxVal.String()
		}
		return "", fmt.Errorf("number %s exceeds maximum %s", number.String(), displayMax)
	}

	// Fast path: use int64 arithmetic if number fits
	if val, ok := number.Int64(); ok {
		return e.encodeInt64(val), nil
	}

	// Slow path: use big.Int arithmetic
	return e.encodeBigInt(number.BigInt()), nil
}

// Decode converts a phonetic word back to a number.
// Returns PositiveInt which can represent both small and large values.
func (e *PatternEncoder) Decode(word string) (PositiveInt, error) {
	runes := []rune(word)
	if len(runes) != len(e.positions) {
		return nil, fmt.Errorf(
			"word length %d doesn't match pattern length %d",
			len(runes),
			len(e.positions),
		)
	}

	// Check if result might exceed int64
	needsBigInt := e.totalCombinations.Cmp(big.NewInt(math.MaxInt64)) > 0

	if needsBigInt {
		return e.decodeBigInt(runes)
	}
	return e.decodeInt64(runes)
}

// MaxValue returns the maximum number that can be encoded.
// If capacity exceeds math.MaxInt, returns MaxInt.
func (e *PatternEncoder) MaxValue() int {
	maxInt := big.NewInt(int64(math.MaxInt))
	if e.totalCombinations.Cmp(maxInt) > 0 {
		return math.MaxInt
	}
	// totalCombinations fits in int, return exact value - 1
	maxVal := new(big.Int).Sub(e.totalCombinations, big.NewInt(1))
	return int(maxVal.Int64())
}

// GetSmallestPatternCapacity returns the maximum value that can be encoded
// with the smallest (first) pattern. This is useful for generating preflight suggestions.
// If the capacity exceeds math.MaxInt, returns MaxInt.
func (e *PhoneticEncoder) GetSmallestPatternCapacity() int {
	if len(e.patternEncoders) == 0 {
		return 0
	}
	return e.patternEncoders[0].MaxValue()
}

// GetPatternInfo returns information about all patterns for generating suggestions.
// Returns a slice with pattern details including true mathematical capacity.
// Each pattern can encode numbers from 0 to its TrueCapacity-1.
// Capacity field is capped at math.MaxInt for compatibility, while TrueCapacity shows the real limit.
func (e *PhoneticEncoder) GetPatternInfo() []struct {
	Pattern      string
	Capacity     int
	TrueCapacity PositiveInt
	ExceedsInt64 bool
} {
	result := make([]struct {
		Pattern      string
		Capacity     int
		TrueCapacity PositiveInt
		ExceedsInt64 bool
	}, 0, len(e.patternEncoders))

	maxInt := big.NewInt(int64(math.MaxInt))
	for _, pe := range e.patternEncoders {
		exceedsInt64 := pe.totalCombinations.Cmp(maxInt) > 0

		// Capacity capped at MaxInt for compatibility
		capacity := math.MaxInt
		if !exceedsInt64 {
			capacity = int(pe.totalCombinations.Int64())
		}

		// TrueCapacity is actual max value (totalCombinations - 1)
		trueMax := new(big.Int).Sub(pe.totalCombinations, big.NewInt(1))
		trueCapacity := NewPositiveIntFromBig(trueMax)

		result = append(result, struct {
			Pattern      string
			Capacity     int
			TrueCapacity PositiveInt
			ExceedsInt64 bool
		}{
			Pattern:      pe.pattern,
			Capacity:     capacity,
			TrueCapacity: trueCapacity,
			ExceedsInt64: exceedsInt64,
		})
	}

	return result
}

// encodeInt64 is the fast path for encoding int64 values.
func (e *PatternEncoder) encodeInt64(number int64) string {
	var result strings.Builder
	remaining := number

	// Convert to mixed-radix representation (right-to-left)
	for i := len(e.positions) - 1; i >= 0; i-- {
		position := e.positions[i]
		charIndex := remaining % int64(position.base)
		remaining /= int64(position.base)

		result.WriteRune(position.chars[charIndex])
	}

	// Reverse the string since we built it backwards
	word := result.String()
	return reverseString(word)
}

// encodeBigInt is the slow path for encoding arbitrarily large big.Int values.
func (e *PatternEncoder) encodeBigInt(number *big.Int) string {
	var result strings.Builder
	remaining := new(big.Int).Set(number)
	base := new(big.Int)
	charIndex := new(big.Int)

	// Convert to mixed-radix representation (right-to-left)
	for i := len(e.positions) - 1; i >= 0; i-- {
		position := e.positions[i]
		base.SetInt64(int64(position.base))

		charIndex.Mod(remaining, base)
		remaining.Div(remaining, base)

		result.WriteRune(position.chars[charIndex.Int64()])
	}

	// Reverse the string since we built it backwards
	word := result.String()
	return reverseString(word)
}

// decodeInt64 is the fast path for decoding to int64.
func (e *PatternEncoder) decodeInt64(runes []rune) (PositiveInt, error) {
	var result int64

	for i, r := range runes {
		position := e.positions[i]

		// Find character index in this position's alphabet
		charIndex := -1
		for idx, char := range position.chars {
			if char == r {
				charIndex = idx
				break
			}
		}

		if charIndex == -1 {
			return nil, fmt.Errorf(
				"character '%c' at position %d is not valid for placeholder '%s'",
				r,
				i,
				position.placeholder,
			)
		}

		// Add to result using positional notation
		multiplier := int64(1)
		for j := i + 1; j < len(e.positions); j++ {
			multiplier *= int64(e.positions[j].base)
		}

		result += int64(charIndex) * multiplier
	}

	return NewPositiveInt(result), nil
}

// decodeBigInt is the slow path for decoding to big.Int.
func (e *PatternEncoder) decodeBigInt(runes []rune) (PositiveInt, error) {
	result := new(big.Int)
	multiplier := new(big.Int)
	charValue := new(big.Int)
	temp := new(big.Int)

	for i, r := range runes {
		position := e.positions[i]

		// Find character index in this position's alphabet
		charIndex := -1
		for idx, char := range position.chars {
			if char == r {
				charIndex = idx
				break
			}
		}

		if charIndex == -1 {
			return nil, fmt.Errorf(
				"character '%c' at position %d is not valid for placeholder '%s'",
				r,
				i,
				position.placeholder,
			)
		}

		// Calculate multiplier for this position
		multiplier.SetInt64(1)
		for j := i + 1; j < len(e.positions); j++ {
			temp.SetInt64(int64(e.positions[j].base))
			multiplier.Mul(multiplier, temp)
		}

		// Add charIndex * multiplier to result
		charValue.SetInt64(int64(charIndex))
		temp.Mul(charValue, multiplier)
		result.Add(result, temp)
	}

	return NewPositiveIntFromBig(result), nil
}

// reverseString reverses a string.
func reverseString(s string) string {
	runes := []rune(s)
	slices.Reverse(runes)
	return string(runes)
}
