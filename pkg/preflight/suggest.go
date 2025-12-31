package preflight

import (
	"errors"
	"fmt"

	phonid "github.com/iilei/phonid/pkg"
)

type (
	// Assertion represents a single suggested preflight check.
	Assertion struct {
		Input   phonid.PositiveInt
		Output  string
		Comment string // e.g., "Lower boundary", "Mid-range", etc.
	}
	// AssertionTable represents a collection of preflight check assertions.
	AssertionTable []Assertion
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
