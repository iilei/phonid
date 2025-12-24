package phonid

import (
	"fmt"
	"strings"
)

// PhoneticEncoder handles encoding/decoding between numbers and phonetic words
type PhoneticEncoder struct {
	config *PhonidConfig

	// Derived from config for fast encoding/decoding
	positions         []Position
	totalCombinations uint64
}

// Position represents one character position in the pattern
type Position struct {
	placeholder string
	chars       []rune
	base        int
}

// NewPhoneticEncoder creates an encoder for the given config
func NewPhoneticEncoder(config *PhonidConfig) (*PhoneticEncoder, error) {
	if config == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	// Config should already be validated with proper defaults
	placeholders := config.Placeholders

	// Build position array
	positions := make([]Position, len(config.Pattern))
	totalCombinations := uint64(1)

	for i, r := range config.Pattern {
		placeholder := PlaceholderType(r) // Convert rune to PlaceholderType
		chars, exists := placeholders[placeholder]
		if !exists {
			return nil, fmt.Errorf("no character set for placeholder '%c'", r)
		}

		positions[i] = Position{
			placeholder: string(r), // Keep as string for display
			chars:       chars,
			base:        len(chars),
		}

		totalCombinations *= uint64(len(chars))
	}

	return &PhoneticEncoder{
		config:            config,
		positions:         positions,
		totalCombinations: totalCombinations,
	}, nil
}

// Encode converts a number to a phonetic word
func (e *PhoneticEncoder) Encode(number uint64) (string, error) {
	if number >= e.totalCombinations {
		return "", fmt.Errorf("number %d exceeds maximum %d", number, e.totalCombinations-1)
	}

	var result strings.Builder
	remaining := number

	// Convert to mixed-radix representation (right-to-left)
	for i := len(e.positions) - 1; i >= 0; i-- {
		position := e.positions[i]
		charIndex := remaining % uint64(position.base)
		remaining /= uint64(position.base)

		result.WriteRune(position.chars[charIndex])
	}

	// Reverse the string since we built it backwards
	word := result.String()
	return reverseString(word), nil
}

// Decode converts a phonetic word back to a number
func (e *PhoneticEncoder) Decode(word string) (uint64, error) {
	runes := []rune(word)
	if len(runes) != len(e.positions) {
		return 0, fmt.Errorf("word length %d doesn't match pattern length %d", len(runes), len(e.positions))
	}

	var result uint64

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
		multiplier := uint64(1)
		for j := i + 1; j < len(e.positions); j++ {
			multiplier *= uint64(e.positions[j].base)
		}

		result += uint64(charIndex) * multiplier
	}

	return result, nil
}

// MaxValue returns the maximum number that can be encoded
func (e *PhoneticEncoder) MaxValue() uint64 {
	return e.totalCombinations - 1
}

// reverseString reverses a string
func reverseString(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}
