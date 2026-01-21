package preflight

import (
	"crypto/rand"
	"errors"
	"fmt"
	"math"
	"math/big"

	phonid "github.com/iilei/phonid/pkg"
)

const (
	// midRangePercentage is the divisor for calculating mid-range value.
	midRangePercentage  = 2
	LocalhostDecimal    = 2130706433
	maxInt64            = math.MaxInt
	ConfigHeaderComment = `
Pattern Requirements:
  - Pattern Length: Must be 3 or more characters (e.g., CVC, CCVC, CVCVC)
    (no duplicates by length allowed)
  - Vowels (V): Length 3-4 needs ≥2 vowels; length 5-6 needs ≥3 vowels; length 7+ needs ≥2 vowels
  - Consonants (C): Need ≥3 characters (e.g., "bzk" or "bdfghjklmnprstvz")

Available Placeholders:
  C = Consonants, V = Vowels, L = Liquids (l,m,n,r)
  N = Nasals, S = Sibilants, F = Fricatives
  X/Y/Z = Custom categories`
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

// GetRandom generates a random assertion within the encoder's capacity.
// It determines the maximum encodable number from the encoder's patterns
// and returns an Assertion with a random input and its encoded output.
// The Comment field is left empty.
// If encoder is nil, creates a default encoder using ProQuint configuration.
func GetRandom(encoder *phonid.PhoneticEncoder) (Assertion, error) {
	if encoder == nil {
		var err error
		encoder, err = phonid.NewPhoneticEncoderLenient(nil)
		if err != nil {
			return Assertion{}, fmt.Errorf("failed to create default encoder: %w", err)
		}
	}

	// Get pattern information to determine max capacity
	patterns := encoder.GetPatternInfo()
	if len(patterns) == 0 {
		return Assertion{}, errors.New("encoder has no patterns")
	}

	// Get the maximum capacity from the last pattern
	lastPattern := patterns[len(patterns)-1]
	var maxValue phonid.PositiveInt

	if lastPattern.ExceedsInt64 {
		// Use true capacity for large patterns
		maxValue = lastPattern.TrueCapacity
	} else {
		// Use int-based capacity - 1 for standard patterns
		if lastPattern.Capacity <= 0 {
			return Assertion{}, fmt.Errorf("invalid capacity: %d", lastPattern.Capacity)
		}
		maxValue = phonid.NewPositiveInt(int64(lastPattern.Capacity - 1))
	}

	// Generate random number between 0 and maxValue (inclusive)
	randomValue, err := generateRandomPositiveInt(maxValue)
	if err != nil {
		return Assertion{}, fmt.Errorf("failed to generate random value: %w", err)
	}

	// Encode the random value
	output, err := encoder.Encode(randomValue)
	if err != nil {
		return Assertion{}, fmt.Errorf("failed to encode random value %s: %w", randomValue.String(), err)
	}

	return Assertion{
		Input:   &phonid.TomlPositiveInt{Value: randomValue},
		Output:  output,
		Comment: "",
	}, nil
}

// generateRandomPositiveInt generates a cryptographically secure random PositiveInt
// between 0 and maxValue (inclusive).
func generateRandomPositiveInt(maxValue phonid.PositiveInt) (phonid.PositiveInt, error) {
	// Convert maxValue to big.Int for crypto/rand
	maxBigInt := maxValue.BigInt()

	// crypto/rand.Int generates [0, maxValue), so we need to add 1 to include maxValue
	maxPlusOne := new(big.Int).Add(maxBigInt, big.NewInt(1))

	// Generate random big.Int
	randomBig, err := rand.Int(rand.Reader, maxPlusOne)
	if err != nil {
		return nil, fmt.Errorf("failed to generate random number: %w", err)
	}

	// Convert back to PositiveInt
	return phonid.NewPositiveIntFromBig(randomBig), nil
}
