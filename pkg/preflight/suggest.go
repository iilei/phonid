// Package preflight represents preflight checks and code generation.
package preflight

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"
	"strings"

	phonid "github.com/iilei/phonid/pkg"
	"github.com/pelletier/go-toml/v2"
)

const (
	// Default configuration values for TOML generation.
	defaultBase     = 36
	defaultBitWidth = 32
	defaultRounds   = 0
	defaultSeed     = 0

	// String processing constants.
	keyValueParts = 2 // Expected parts when splitting "key = value"
)

type (
	// Assertion represents a single suggested preflight check.
	Assertion struct {
		Input   phonid.PositiveInt `toml:"input"`
		Output  string             `toml:"output"`
		Comment string             `toml:"-"` // e.g., "Lower boundary", "Mid-range", etc.
	}
	// AssertionTable represents a collection of preflight check assertions.
	AssertionTable []Assertion

	// TOMLConfig represents the full .phonidrc TOML structure.
	TOMLConfig struct {
		Base      int            `toml:"base"`
		Shuffle   ShuffleConfig  `toml:"shuffle"`
		Phonetic  PhoneticConfig `toml:"phonetic"`
		Preflight []Assertion    `toml:"preflight"`
	}

	// ShuffleConfig represents shuffle configuration in TOML.
	ShuffleConfig struct {
		BitWidth int `toml:"bit_width"`
		Rounds   int `toml:"rounds"`
		Seed     int `toml:"seed"`
	}

	// PhoneticConfig represents phonetic configuration in TOML.
	PhoneticConfig struct {
		Patterns     []string          `toml:"patterns"`
		Placeholders map[string]string `toml:"placeholders"`
	}
)

// GenerateSuggestions creates preflight check suggestions for an encoder.
// It generates boundary values and representative test points across the encoding space.
func GenerateSuggestions(encoder *phonid.PhoneticEncoder) (AssertionTable, error) {
	if encoder == nil {
		return nil, errors.New("encoder cannot be nil")
	}

	// Get capacity from smallest pattern (first pattern encoder)
	maxValue := encoder.GetSmallestPatternCapacity()
	if maxValue == 0 {
		return nil, errors.New("encoder has zero capacity")
	}

	suggestions := AssertionTable{}

	// 1. Lower boundary (0)
	if err := addSuggestion(&suggestions, encoder, 0, "Lower boundary"); err != nil {
		return nil, err
	}

	// 2. Upper boundary (single word)
	if err := addSuggestion(&suggestions, encoder, phonid.PositiveInt(maxValue), "Upper boundary (single word)"); err != nil {
		return nil, err
	}

	return suggestions, nil
}

// addSuggestion is a helper to encode and add a suggestion point.
func addSuggestion(
	suggestions *AssertionTable,
	encoder *phonid.PhoneticEncoder,
	input phonid.PositiveInt,
	comment string,
) error {
	output, err := encoder.Encode(input)
	if err != nil {
		return fmt.Errorf("failed to encode %d: %w", input, err)
	}
	assertion := Assertion{
		Input:   input,
		Output:  output,
		Comment: comment,
	}

	*suggestions = append(*suggestions, assertion)

	return nil
}

// SuggestConfig generates a complete TOML configuration string with preflight suggestions.
// It includes the base config, shuffle settings, phonetic patterns, and suggested test assertions
// with inline comments preserved.
func SuggestConfig(encoder *phonid.PhoneticEncoder, config *phonid.PhonidConfig) (string, error) {
	if encoder == nil {
		return "", errors.New("encoder cannot be nil")
	}
	if config == nil {
		return "", errors.New("config cannot be nil")
	}

	// Generate suggestions
	assertions, err := GenerateSuggestions(encoder)
	if err != nil {
		return "", fmt.Errorf("failed to generate suggestions: %w", err)
	}

	// Build full config
	tomlConfig := buildTOMLConfig(config, assertions)
	capacityInt := encoder.GetSmallestPatternCapacity()
	if capacityInt < 0 {
		return "", errors.New("encoder capacity cannot be negative")
	}
	capacity := uint64(capacityInt)

	// Generate base TOML
	var buf bytes.Buffer
	enc := toml.NewEncoder(&buf)
	enc.SetIndentTables(true)

	if err := enc.Encode(tomlConfig); err != nil {
		return "", fmt.Errorf("failed to encode TOML: %w", err)
	}

	// Post-process to add inline comments
	result := addInlineComments(buf.String(), assertions, capacity)
	return result, nil
}

// buildTOMLConfig constructs a TOMLConfig from PhonidConfig and assertions.
func buildTOMLConfig(config *phonid.PhonidConfig, assertions AssertionTable) *TOMLConfig {
	placeholders := make(map[string]string)
	for k, v := range config.Placeholders {
		placeholders[string(k)] = string(v)
	}

	return &TOMLConfig{
		Base: defaultBase,
		Shuffle: ShuffleConfig{
			BitWidth: defaultBitWidth,
			Rounds:   defaultRounds,
			Seed:     defaultSeed,
		},
		Phonetic: PhoneticConfig{
			Patterns:     config.Patterns,
			Placeholders: placeholders,
		},
		Preflight: assertions,
	}
}

// addInlineComments post-processes TOML output to add inline comments to assertions.
func addInlineComments(tomlStr string, assertions AssertionTable, capacity uint64) string {
	lines := strings.Split(tomlStr, "\n")
	var result strings.Builder

	// Find where preflight section starts
	preflightStart := -1
	for i, line := range lines {
		if strings.Contains(line, "[[preflight]]") {
			preflightStart = i
			break
		}
	}

	// Write everything before preflight section
	if preflightStart > 0 {
		for i := range preflightStart {
			result.WriteString(lines[i])
			result.WriteString("\n")
		}
	}

	// Add header comment before preflight section
	fmt.Fprintf(&result, "# Output of 'phonid preflight --suggest'\n")
	fmt.Fprintf(&result, "# Capacity per word: %s combinations (0-%s)\n",
		formatNumber(capacity), formatNumber(capacity-1))
	fmt.Fprintf(&result, "#\n")
	fmt.Fprintf(&result, "# Suggested preflight checks:\n\n")

	// Process preflight section with inline comments
	assertionIdx := 0
	for i := preflightStart; i < len(lines); i++ {
		line := lines[i]
		trimmed := strings.TrimSpace(line)

		// Add inline comment to input lines
		if strings.HasPrefix(trimmed, "input =") && assertionIdx < len(assertions) {
			// Extract the input line and add comment
			parts := strings.SplitN(line, "=", keyValueParts)
			if len(parts) == keyValueParts {
				value := strings.TrimSpace(parts[1])
				indent := strings.Repeat(" ", len(line)-len(trimmed))
				// Format: input = VALUE # Comment
				fmt.Fprintf(&result, "%sinput = %-13s # %s\n",
					indent, value, assertions[assertionIdx].Comment)
				assertionIdx++
				continue
			}
		}

		result.WriteString(line)
		if i < len(lines)-1 {
			result.WriteString("\n")
		}
	}

	return result.String()
}

// formatNumber formats a number with underscores as thousands separators.
func formatNumber(n uint64) string {
	s := strconv.FormatUint(n, 10)
	var result []rune
	for i, c := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			result = append(result, '_')
		}
		result = append(result, c)
	}
	return string(result)
}
