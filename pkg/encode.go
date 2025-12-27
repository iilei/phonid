package phonid

import (
	"fmt"
	"strings"
)

// PhoneticEncoder handles encoding/decoding between numbers and phonetic words
type PhoneticEncoder struct {
	config   *PhonidConfig
	patterns []*PatternEncoder // ordered by totalCombinations ascending
}

// PatternEncoder represents a single pattern configuration
type PatternEncoder struct {
	pattern           string
	positions         []Position
	totalCombinations PositiveInt
	length            int // Number of positions/characters in the pattern
}

// Position represents one character position in the pattern
type Position struct {
	placeholder string
	chars       []rune
	base        int
}

// NewPhoneticEncoder creates an encoder with a validated config
func NewPhoneticEncoder(config *PhonidConfig) (*PhoneticEncoder, error) {
	// Validate first
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return newPhoneticEncoder(config)
}

// buildPatternEncoder creates a PatternEncoder from a pattern string and placeholders
func buildPatternEncoder(pattern string, placeholders PlaceholderMap) (*PatternEncoder, error) {
	if pattern == "" {
		return nil, fmt.Errorf("pattern cannot be empty")
	}

	positions := make([]Position, 0, len(pattern))
	totalCombinations := 1

	// Parse each character in the pattern
	for i, char := range pattern {
		placeholderType := PlaceholderType(char)

		// Look up the character set for this placeholder
		chars, exists := placeholders[placeholderType]
		if !exists {
			return nil, fmt.Errorf("placeholder '%c' at position %d not found in placeholders", char, i)
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
		totalCombinations *= position.base
	}

	return &PatternEncoder{
		pattern:           pattern,
		positions:         positions,
		totalCombinations: PositiveInt(totalCombinations),
		length:            len(positions),
	}, nil
}

// newPhoneticEncoder is the internal constructor (assumes valid config)
func newPhoneticEncoder(config *PhonidConfig) (*PhoneticEncoder, error) {
	patterns := make([]*PatternEncoder, 0, len(config.Patterns))

	for _, pattern := range config.Patterns {
		encoder, err := buildPatternEncoder(pattern, config.Placeholders)
		if err != nil {
			return nil, err
		}
		patterns = append(patterns, encoder)
	}

	// Sort by totalCombinations
	for i := 0; i < len(patterns); i++ {
		for j := i + 1; j < len(patterns); j++ {
			if patterns[i].totalCombinations > patterns[j].totalCombinations {
				patterns[i], patterns[j] = patterns[j], patterns[i]
			}
		}
	}

	return &PhoneticEncoder{
		config:   config,
		patterns: patterns,
	}, nil
}

// Encode converts a number to a phonetic word, automatically selecting the best pattern
func (e *PhoneticEncoder) Encode(number PositiveInt) (string, error) {
	if number < 0 {
		return "", fmt.Errorf("number must be non-negative, got %d", number)
	}

	// Find the smallest pattern that can encode this number
	for _, pattern := range e.patterns {
		if number < pattern.totalCombinations {
			return pattern.Encode(number)
		}
	}

	// Number too large for any pattern
	return "", fmt.Errorf("number %d exceeds capacity of largest pattern (max: %d)",
		number, e.patterns[len(e.patterns)-1].totalCombinations-1)
}

func (e *PhoneticEncoder) Decode(word string) (int, error) {
	wordRunes := []rune(word)

	// Try to match pattern by length
	for _, pattern := range e.patterns {
		if len(wordRunes) == pattern.length {
			return pattern.Decode(word)
		}
	}

	return 0, fmt.Errorf("word length %d doesn't match any pattern", len(wordRunes))
}

// Encode converts a number to a phonetic word
func (e *PatternEncoder) Encode(number PositiveInt) (string, error) {
	if number >= e.totalCombinations {
		return "", fmt.Errorf("number %d exceeds maximum %d", number, e.totalCombinations-1)
	}

	var result strings.Builder
	remaining := int(number)

	// Convert to mixed-radix representation (right-to-left)
	for i := len(e.positions) - 1; i >= 0; i-- {
		position := e.positions[i]
		charIndex := remaining % position.base
		remaining /= position.base

		result.WriteRune(position.chars[charIndex])
	}

	// Reverse the string since we built it backwards
	word := result.String()
	return reverseString(word), nil
}

// Decode converts a phonetic word back to a number
func (e *PatternEncoder) Decode(word string) (int, error) {
	runes := []rune(word)
	if len(runes) != len(e.positions) {
		return 0, fmt.Errorf("word length %d doesn't match pattern length %d", len(runes), len(e.positions))
	}

	var result int

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
			return 0, fmt.Errorf("character '%c' at position %d is not valid for placeholder '%s'", r, i, position.placeholder)
		}

		// Add to result using positional notation
		multiplier := int(1)
		for j := i + 1; j < len(e.positions); j++ {
			multiplier *= e.positions[j].base
		}

		result += charIndex * multiplier
	}

	return result, nil
}

// MaxValue returns the maximum number that can be encoded
func (e *PatternEncoder) MaxValue() int {
	return int(e.totalCombinations) - 1
}

// reverseString reverses a string
func reverseString(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}
