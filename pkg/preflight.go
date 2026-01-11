package phonid

import (
	"errors"
	"fmt"
	"math"
	"math/big"
)

const (
	// Shuffle validation thresholds.
	gapRatioCritical = 0.5  // Above this ratio, reject immediately
	gapRatioWarning  = 0.25 // Between warning and critical, test more samples
	gapRatioNotable  = 0.15 // Below this is acceptable, above shows warning

	// Testing parameters.
	smallCapacityThreshold = 10000 // Test all values if capacity <= this
	largeSampleSize        = 1000  // Number of samples for large capacity
	maxFailuresBeforeStop  = 3     // Stop testing after this many failures

	// Display formatting.
	percentMultiplier = 100.0
)

// ErrNoPreflightChecks is returned when no preflight checks are provided.
var ErrNoPreflightChecks = errors.New("config must include at least one [[preflight]] check\n\n" +
	"Example:\n" +
	"  [[preflight]]\n" +
	"  input = 0\n" +
	"  output = \"babab\"\n\n" +
	"Hint: Run 'phonid preflight --suggest' to generate recommended checks")

// ValidatePreflight checks if preflight tests pass for this encoder
// Performs bidirectional validation: encoding (int->string) and decoding (string->int).
// If shuffling is enabled (rounds > 0), also validates that cycle walking can find
// valid values across the entire capacity range.
func (p *PhoneticEncoder) ValidatePreflight(checks []PreflightCheck) error {
	if len(checks) == 0 {
		return ErrNoPreflightChecks
	}

	for i, check := range checks {
		// Convert TOML input to PositiveInt
		input := check.Input.ToPositiveInt()

		// Test encoding
		encoded, err := p.Encode(input)
		if err != nil {
			return fmt.Errorf("check #%d failed: encoding %s failed: %w", i+1, input.String(), err)
		}
		if encoded != check.Output {
			return fmt.Errorf("check #%d failed: encoding %s produced %q, but expected %q",
				i+1, input.String(), encoded, check.Output)
		}

		// Test decoding (implicit round-trip)
		decoded, err := p.Decode(check.Output)
		if err != nil {
			return fmt.Errorf("check #%d failed: decoding %q failed: %w",
				i+1, check.Output, err)
		}
		if decoded.Cmp(input) != 0 {
			return fmt.Errorf("check #%d failed: decoding %q produced %s, but expected %s",
				i+1, check.Output, decoded.String(), input.String())
		}
	}

	// If shuffling is enabled, validate cycle walking across capacity range
	if p.shuffler != nil && p.config.Shuffle != nil && p.config.Shuffle.Rounds > 0 {
		if err := p.validateShuffleCycleWalking(); err != nil {
			return fmt.Errorf("shuffle validation failed: %w", err)
		}
	}

	return nil
}

// validateShuffleCycleWalking tests that cycle walking can successfully find valid
// values across the entire capacity range. This catches patterns where the gap between
// shuffler domain (power of 2) and pattern capacity is too large, causing frequent
// cycle walking failures.
//
//nolint:gocyclo // Complexity comes from necessary validation logic for different gap ratios
func (p *PhoneticEncoder) validateShuffleCycleWalking() error {
	maxCapacity := p.patternEncoders[len(p.patternEncoders)-1].totalCombinations
	maxUint64 := new(big.Int).SetUint64(math.MaxUint64)

	// Skip validation if capacity exceeds uint64 (shuffling disabled for such patterns)
	if maxCapacity.Cmp(maxUint64) > 0 {
		return nil
	}

	maxCapacityUint64 := maxCapacity.Uint64()
	bitWidth := p.shuffler.bitWidth
	shufflerCapacity := uint64(1) << bitWidth

	// Calculate gap ratio
	gapSize := shufflerCapacity - maxCapacityUint64
	gapRatio := float64(gapSize) / float64(shufflerCapacity)

	// If gap is > 50%, warn immediately - likely to cause failures
	if gapRatio > gapRatioCritical {
		return fmt.Errorf(
			"pattern capacity (%d) is too far from shuffler domain (2^%d = %d).\n"+
				"Gap: %d invalid values (%.1f%% of shuffler domain).\n"+
				"This will cause frequent cycle walking failures.\n"+
				"Solution: Adjust pattern to have capacity closer to a power of 2, or reduce shuffle rounds to 0",
			maxCapacityUint64, bitWidth, shufflerCapacity, gapSize, gapRatio*percentMultiplier,
		)
	}

	// Sample test across the capacity range
	// Test boundary values and representative samples
	var testValues []uint64

	switch {
	case maxCapacityUint64 <= smallCapacityThreshold:
		// If capacity is small enough, test all values
		testValues = make([]uint64, maxCapacityUint64)
		for i := range maxCapacityUint64 {
			testValues[i] = i
		}
	case gapRatio > gapRatioWarning:
		// For problematic gap ratios (25-50%), test more samples
		sampleSize := min(maxCapacityUint64, uint64(largeSampleSize))
		step := maxCapacityUint64 / sampleSize
		// Collect all values first
		var samples []uint64
		for i := uint64(0); i < maxCapacityUint64; i += step {
			samples = append(samples, i)
		}
		// Ensure we test the last value
		if len(samples) == 0 || samples[len(samples)-1] != maxCapacityUint64-1 {
			samples = append(samples, maxCapacityUint64-1)
		}
		testValues = samples
	default:
		// Default sample for acceptable gap ratios
		testValues = []uint64{
			0,
			1,
			maxCapacityUint64 / 4,
			maxCapacityUint64 / 2,
			maxCapacityUint64 * 3 / 4,
			maxCapacityUint64 - 1,
		}
	}

	// Test encoding each value
	failures := 0
	firstFailure := ""

	for _, val := range testValues {
		input := NewPositiveInt(int64(val))
		if val > math.MaxInt64 {
			input = NewPositiveIntFromBig(new(big.Int).SetUint64(val))
		}

		_, err := p.Encode(input)
		if err != nil {
			failures++
			if firstFailure == "" {
				firstFailure = fmt.Sprintf("value %d: %v", val, err)
			}
			// Stop after finding a few failures to avoid excessive testing
			if failures >= maxFailuresBeforeStop {
				break
			}
		}
	}

	if failures > 0 {
		return fmt.Errorf(
			"cycle walking failures detected in shuffle configuration.\n"+
				"Gap ratio: %.1f%% (%d invalid values in shuffler domain).\n"+
				"First failure: %s\n"+
				"Total failures in sample: %d/%d\n"+
				"Solution: Adjust pattern capacity to be closer to a power of 2, or reduce shuffle rounds to 0",
			gapRatio*percentMultiplier, gapSize, firstFailure, failures, len(testValues),
		)
	}

	// If gap ratio is notable (15-25%), issue a warning in error format so user is aware
	if gapRatio > gapRatioNotable {
		return fmt.Errorf(
			"pattern capacity (%d) has a %.1f%% gap from shuffler domain (2^%d = %d).\n"+
				"This may occasionally cause cycle walking failures.\n"+
				"Consider adjusting pattern to have capacity closer to a power of 2 for optimal performance",
			maxCapacityUint64, gapRatio*percentMultiplier, bitWidth, shufflerCapacity,
		)
	}

	return nil
}
