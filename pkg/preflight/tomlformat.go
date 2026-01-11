package preflight

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/pelletier/go-toml/v2"

	phonid "github.com/iilei/phonid/pkg"
)

const (
	// Default configuration values for TOML generation.
	defaultRounds = 0
	defaultSeed   = 0

	// String processing constants.
	keyValueParts = 2 // Expected parts when splitting "key = value"

	// Formatting constants.
	inputValuePadding = 13 // Padding width for input values in inline comments
)

type (
	// TOMLConfig represents the full .phonidrc TOML structure.
	TOMLConfig struct {
		Shuffle   *ShuffleConfig `toml:"shuffle,omitempty"` // Omitted if capacity > uint64 max
		Phonetic  PhoneticConfig `toml:"phonetic"`
		Preflight []Assertion    `toml:"preflight"`
	}

	// ShuffleConfig represents shuffle configuration in TOML.
	// BitWidth is now auto-calculated from pattern capacity.
	ShuffleConfig struct {
		Rounds int `toml:"rounds"`
		Seed   int `toml:"seed"`
	}

	// PhoneticConfig represents phonetic configuration in TOML.
	PhoneticConfig struct {
		Patterns     []string          `toml:"patterns"`
		Placeholders map[string]string `toml:"placeholders"`
	}

	// TOMLFormatter implements the Formatter interface for TOML output.
	TOMLFormatter struct {
		encoder *toml.Encoder
	}
)

// NewTOMLFormatter provides toml formatting.
func NewTOMLFormatter() Formatter {
	return &TOMLFormatter{
		encoder: newTOMLFormatter(),
	}
}

// Name returns the format name.
func (f *TOMLFormatter) Name() OutputFormat {
	return FormatTOML
}

// Format writes preflight assertions as TOML to the writer.
func (f *TOMLFormatter) Format(w io.Writer, assertions *AssertionTable) error {
	enc := toml.NewEncoder(w)
	enc.SetIndentTables(true)
	return enc.Encode(assertions)
}

// newTOMLFormatter is the internal constructor.
func newTOMLFormatter() *toml.Encoder {
	buf := bytes.Buffer{}
	enc := toml.NewEncoder(&buf)
	enc.SetIndentTables(true)
	return enc
}

// SuggestConfig generates a complete TOML configuration string with preflight suggestions.
// It includes the base config, shuffle settings, phonetic patterns, and suggested test assertions
// with inline comments preserved.
// If encoder is nil, creates a default encoder. If config is nil, uses DefaultConfig.
func SuggestConfig(encoder *phonid.PhoneticEncoder, config *phonid.PhonidConfig) (string, error) {
	if config == nil {
		config = &phonid.DefaultConfig
	}
	if encoder == nil {
		var err error
		encoder, err = phonid.NewPhoneticEncoderLenient(config)
		if err != nil {
			return "", fmt.Errorf("failed to create default encoder: %w", err)
		}
	}

	// Generate suggestions
	assertions, err := GenerateSuggestions(encoder)
	if err != nil {
		return "", fmt.Errorf("failed to generate suggestions: %w", err)
	}

	// Build full config (pass encoder for capacity checking)
	tomlConfig := buildTOMLConfig(config, assertions, encoder)
	capacityInt := encoder.GetSmallestPatternCapacity()
	if capacityInt < 0 {
		return "", errors.New("encoder capacity cannot be negative")
	}

	// Generate base TOML
	var buf bytes.Buffer
	enc := toml.NewEncoder(&buf)
	enc.SetIndentTables(true)

	if err := enc.Encode(tomlConfig); err != nil {
		return "", fmt.Errorf("failed to encode TOML: %w", err)
	}

	capacity := uint64(capacityInt)
	// Post-process to add inline comments
	result := addInlineComments(buf.String(), assertions, capacity)
	return result, nil
}

// FormatTOMLWithSuggestions generates a complete TOML configuration string with the provided suggestions.
// Similar to SuggestConfig but uses pre-generated suggestions instead of generating new ones.
func FormatTOMLWithSuggestions(
	encoder *phonid.PhoneticEncoder,
	config *phonid.PhonidConfig,
	suggestions *AssertionTable,
) (string, error) {
	if config == nil {
		config = &phonid.DefaultConfig
	}
	if encoder == nil {
		var err error
		encoder, err = phonid.NewPhoneticEncoderLenient(config)
		if err != nil {
			return "", fmt.Errorf("failed to create default encoder: %w", err)
		}
	}

	// Build full config (pass encoder for capacity checking)
	tomlConfig := buildTOMLConfig(config, *suggestions, encoder)
	capacityInt := encoder.GetSmallestPatternCapacity()
	if capacityInt < 0 {
		return "", errors.New("encoder capacity cannot be negative")
	}

	// Generate base TOML
	var buf bytes.Buffer
	enc := toml.NewEncoder(&buf)
	enc.SetIndentTables(true)

	if err := enc.Encode(tomlConfig); err != nil {
		return "", fmt.Errorf("failed to encode TOML: %w", err)
	}

	capacity := uint64(capacityInt)
	// Post-process to add inline comments
	result := addInlineComments(buf.String(), *suggestions, capacity)
	return result, nil
}

// buildTOMLConfig constructs a TOMLConfig from PhonidConfig and assertions.
// Uses shuffle settings from config if present, otherwise uses defaults.
// Omits shuffle config entirely if pattern capacity exceeds uint64 max (shuffling not supported).
func buildTOMLConfig(
	config *phonid.PhonidConfig,
	assertions AssertionTable,
	encoder *phonid.PhoneticEncoder,
) *TOMLConfig {
	placeholders := make(map[string]string)
	for k, v := range config.Placeholders {
		placeholders[string(k)] = string(v)
	}

	// Check if shuffling is possible (capacity must fit in uint64)
	var shuffleConfig *ShuffleConfig
	patterns := encoder.GetPatternInfo()
	if len(patterns) > 0 {
		largestPattern := patterns[len(patterns)-1]
		// Only include shuffle config if capacity fits in uint64 range
		if !largestPattern.ExceedsInt64 {
			// Extract shuffle config from input or use defaults
			shuffleConfig = &ShuffleConfig{
				Rounds: defaultRounds,
				Seed:   defaultSeed,
			}

			// If config has shuffle settings, use them
			if config.Shuffle != nil {
				shuffleConfig.Rounds = config.Shuffle.Rounds
				//#nosec G115 -- Seed is validated by ShuffleConfig.Validate() to be <= MaxSafeSeed (int64 max)
				shuffleConfig.Seed = int(config.Shuffle.Seed)
			}
		}
		// If ExceedsInt64, shuffleConfig remains nil and will be omitted from TOML
	}

	return &TOMLConfig{
		Shuffle: shuffleConfig,
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

	// Add header comment at the very top
	fmt.Fprintf(&result, "# Output of 'phonid preflight --suggest'\n")
	fmt.Fprintf(&result, "#\n")
	fmt.Fprintf(&result, "# Capacity per word: %s combinations (0-%d)\n", formatCapacity(capacity), capacity)
	fmt.Fprintf(&result, "#\n")

	// Add ConfigHeaderComment as TOML comments
	for line := range strings.SplitSeq(strings.Trim(ConfigHeaderComment, "\n"), "\n") {
		if line == "" {
			fmt.Fprintf(&result, "#\n")
		} else {
			fmt.Fprintf(&result, "# %s\n", line)
		}
	}
	fmt.Fprintf(&result, "\n")

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

	fmt.Fprintf(&result, "# Suggested preflight checks:\n\n")
	// Process preflight section with inline comments
	assertionIdx := 0
	for i := preflightStart; i < len(lines); i++ {
		line := lines[i]
		trimmed := strings.TrimSpace(line)

		// Add inline comment to input lines
		if strings.HasPrefix(trimmed, "input =") {
			if assertionIdx >= len(assertions) {
				result.WriteString(line)
				continue
			}
			// Extract the input line and add comment
			parts := strings.SplitN(line, "=", keyValueParts)
			if len(parts) == keyValueParts {
				value := strings.TrimSpace(parts[1])
				indent := strings.Repeat(" ", len(line)-len(trimmed))
				// Format: input = VALUE # Comment
				fmt.Fprintf(&result, "%sinput = %-*s # %s\n",
					indent, inputValuePadding, value, assertions[assertionIdx].Comment)
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

// formatCapacity formats a capacity number with underscores as thousands separators.
func formatCapacity(n uint64) string {
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
