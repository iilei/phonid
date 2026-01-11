package preflight

import (
	"errors"
	"fmt"
	"math"

	phonid "github.com/iilei/phonid/pkg"
)

const (
	// midRangePercentage is the divisor for calculating mid-range value.
	midRangePercentage  = 2
	LocalhostDecimal    = 2130706433
	maxInt64            = math.MaxInt
	ConfigHeaderComment = `
Pattern Requirements:
  - Pattern Length: Must be 5, 7, 11, 13, 23, 29, 31, 37, 41, 43, or 47 characters
    (e.g., CVCVC, CVCVCVC)
  - Vowels (V): Length 5 needs ≥3 vowels; length 7+ needs ≥2 vowels
  - Consonants (C): Need ≥3 characters (e.g., "bzk" or "bdfghjklmnprstvz")
  - Shuffling: Ensure 128+ combinations for shuffle rounds ≥ 1 (automatically validated)

Available Placeholders:
  C = Consonants, V = Vowels, L = Liquids (l,m,n,r)
  N = Nasals, S = Sibilants, F = Fricatives
  X/Y/Z = Custom categories

Shuffle Limitations:
  Patterns with capacity > 18,446,744,073,709,551,615 (uint64 max) cannot be shuffled.
  Shuffling will be automatically disabled for such patterns, and values will be
  encoded in sequential order without permutation.`
)

type (
	// Assertion represents a single suggested preflight check.
	Assertion struct {
		Input   *phonid.TomlPositiveInt `toml:"input"`
		Output  string                  `toml:"output"`
		Comment string                  `toml:"-"` // e.g., "Lower boundary", "Mid-range", etc.
	}
	// AssertionTable represents a collection of preflight check assertions.
	AssertionTable []Assertion
)

// GenerateSuggestions creates preflight check suggestions for an encoder.
// It generates boundary values and representative test points across the encoding space.
// If encoder is nil, creates a default encoder using ProQuint configuration.
func GenerateSuggestions(encoder *phonid.PhoneticEncoder) (AssertionTable, error) {
	return GenerateSuggestionsWithCustom(encoder, nil, true)
}

// GenerateSuggestionsWithCustom creates test assertions for boundaries and/or custom values.
// If includeBoundaries is true, generates lower boundary, mid-range, and upper boundary checks
// for each pattern in the configuration.
// Custom values are validated against encoder capacity and added with descriptive comments.
// If encoder is nil, creates a default encoder using ProQuint configuration.
func GenerateSuggestionsWithCustom(
	encoder *phonid.PhoneticEncoder,
	customValues []phonid.PositiveInt,
	includeBoundaries bool,
) (AssertionTable, error) {
	if encoder == nil {
		var err error
		encoder, err = phonid.NewPhoneticEncoderLenient(nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create default encoder: %w", err)
		}
	}

	// Get pattern information
	patterns := encoder.GetPatternInfo()
	if len(patterns) == 0 {
		return nil, errors.New("encoder has no patterns")
	}

	suggestions := AssertionTable{}

	// Add boundary checks for each pattern if requested
	if includeBoundaries {
		if err := addBoundarySuggestions(&suggestions, encoder, patterns); err != nil {
			return nil, err
		}
	}

	// Add custom and localhost checks
	maxCapacity := patterns[len(patterns)-1].Capacity - 1
	if err := addCustomSuggestions(&suggestions, encoder, customValues, maxCapacity); err != nil {
		return nil, err
	}

	return suggestions, nil
}

// addBoundarySuggestions generates lower, mid-range, and upper boundary checks for each pattern.
// For patterns exceeding int64 capacity, generates both int64 boundaries and true capacity boundaries.
func addBoundarySuggestions(
	suggestions *AssertionTable,
	encoder *phonid.PhoneticEncoder,
	patterns []struct {
		Pattern      string
		Capacity     int
		TrueCapacity phonid.PositiveInt
		ExceedsInt64 bool
	},
) error {
	for _, pattern := range patterns {
		if err := validatePatternCapacity(pattern); err != nil {
			return err
		}

		if pattern.ExceedsInt64 {
			if err := addLargePatternBoundaries(suggestions, encoder, pattern); err != nil {
				return err
			}
		} else {
			if err := addStandardPatternBoundaries(suggestions, encoder, pattern); err != nil {
				return err
			}
		}
	}
	return nil
}

// addLargePatternBoundaries adds boundary suggestions for patterns exceeding int64 capacity.
func addLargePatternBoundaries(
	suggestions *AssertionTable,
	encoder *phonid.PhoneticEncoder,
	pattern struct {
		Pattern      string
		Capacity     int
		TrueCapacity phonid.PositiveInt
		ExceedsInt64 bool
	},
) error {
	// Lower boundary with capacity note
	lowerComment := fmt.Sprintf("Pattern %s capacity: %s - Lower boundary",
		pattern.Pattern, formatLargeNumber(pattern.TrueCapacity))
	if err := addPatternBoundary(
		suggestions,
		encoder,
		phonid.NewPositiveInt(0),
		pattern.Pattern,
		lowerComment,
	); err != nil {
		return err
	}

	// Add int64 mid-range
	midInt64 := phonid.NewPositiveInt(int64(maxInt64 / midRangePercentage))
	if err := addPatternBoundary(suggestions, encoder, midInt64, pattern.Pattern, "Mid-range of int64"); err != nil {
		return err
	}

	// Add int64 upper boundary
	maxInt64Val := phonid.NewPositiveInt(maxInt64)
	if err := addPatternBoundary(
		suggestions,
		encoder,
		maxInt64Val,
		pattern.Pattern,
		"Upper boundary of int64",
	); err != nil {
		return err
	}

	// Add true capacity upper boundary
	return addPatternBoundary(
		suggestions,
		encoder,
		pattern.TrueCapacity,
		pattern.Pattern,
		"Upper boundary - true capacity",
	)
}

// addStandardPatternBoundaries adds boundary suggestions for patterns within int64 capacity.
func addStandardPatternBoundaries(
	suggestions *AssertionTable,
	encoder *phonid.PhoneticEncoder,
	pattern struct {
		Pattern      string
		Capacity     int
		TrueCapacity phonid.PositiveInt
		ExceedsInt64 bool
	},
) error {
	// Lower boundary
	if err := addPatternBoundary(
		suggestions,
		encoder,
		phonid.NewPositiveInt(0),
		pattern.Pattern,
		"Lower boundary",
	); err != nil {
		return err
	}

	maxValue := pattern.Capacity - 1
	if maxValue < 0 {
		return fmt.Errorf("pattern %s capacity overflow", pattern.Pattern)
	}

	// Add mid-range (if maxValue > 0)
	if maxValue > 0 {
		midValue := maxValue / midRangePercentage
		if err := addPatternBoundary(
			suggestions,
			encoder,
			phonid.NewPositiveInt(int64(midValue)),
			pattern.Pattern,
			"Mid-range",
		); err != nil {
			return err
		}
	}

	// Add upper boundary
	return addPatternBoundary(
		suggestions,
		encoder,
		phonid.NewPositiveInt(int64(maxValue)),
		pattern.Pattern,
		"Upper boundary",
	)
}

// addCustomSuggestions adds custom value checks and localhost check if applicable.
func addCustomSuggestions(
	suggestions *AssertionTable,
	encoder *phonid.PhoneticEncoder,
	customValues []phonid.PositiveInt,
	maxCapacity int,
) error {
	for _, value := range customValues {
		// Check if value exceeds capacity
		valInt64, ok := value.Int64()
		if !ok || valInt64 > int64(maxCapacity) {
			return fmt.Errorf("value %s exceeds encoder capacity (max: %d)", value.String(), maxCapacity)
		}
		comment := "Custom check for " + value.String()
		if err := addSuggestion(suggestions, encoder, value, comment); err != nil {
			return err
		}
	}

	// Add localhost check if it fits
	if LocalhostDecimal <= maxCapacity {
		if err := addSuggestion(
			suggestions,
			encoder,
			phonid.NewPositiveInt(LocalhostDecimal),
			"localhost IP address (127.0.0.1 = 2130706433)",
		); err != nil {
			return err
		}
	}

	return nil
}

// validatePatternCapacity ensures the pattern has a valid capacity.
func validatePatternCapacity(pattern struct {
	Pattern      string
	Capacity     int
	TrueCapacity phonid.PositiveInt
	ExceedsInt64 bool
},
) error {
	if pattern.Capacity <= 0 {
		return fmt.Errorf("pattern %s has invalid capacity %d", pattern.Pattern, pattern.Capacity)
	}
	return nil
}

// addPatternBoundary adds a boundary suggestion with a pattern-specific comment.
func addPatternBoundary(
	suggestions *AssertionTable,
	encoder *phonid.PhoneticEncoder,
	value phonid.PositiveInt,
	patternName string,
	boundaryType string,
) error {
	comment := fmt.Sprintf("%s (%s)", boundaryType, patternName)
	return addSuggestion(suggestions, encoder, value, comment)
}

// formatLargeNumber formats a PositiveInt with underscore separators for readability.
func formatLargeNumber(n phonid.PositiveInt) string {
	str := n.String()
	// Add underscores every 3 digits from right to left
	var result []rune
	for i, r := range str {
		if i > 0 && (len(str)-i)%3 == 0 {
			result = append(result, '_')
		}
		result = append(result, r)
	}
	return string(result)
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
		Input:   &phonid.TomlPositiveInt{Value: input},
		Output:  output,
		Comment: comment,
	}

	*suggestions = append(*suggestions, assertion)

	return nil
}
