package phonid_test

import (
	"math/big"
	"testing"

	. "github.com/iilei/phonid/pkg"
)

// TestMaxCapacityConsistentAcrossShuffleRounds verifies that the maximum
// accepted value for encoding is the same regardless of shuffle configuration.
// This test ensures cycle walking maintains consistent capacity limits.
func TestMaxCapacityConsistentAcrossShuffleRounds(t *testing.T) {
	// Create a simple config with known capacity
	// CVCVC with 3 chars each: 3^5 = 243 combinations (0-242)
	baseConfig := &PhonidConfig{
		Patterns: []string{"CVCVC"},
		Placeholders: PlaceholderMap{
			Vowel:     RuneSet{'a', 'o', 'i'},
			Consonant: RuneSet{'b', 'z', 'k'},
		},
	}

	// Test with different shuffle configurations
	testCases := []struct {
		name   string
		rounds int
		seed   uint64
	}{
		{"no shuffle", 0, 0},
		{"shuffle 3 rounds", 3, 12345},
		{"shuffle 6 rounds", 6, 67890},
		{"shuffle 10 rounds", 10, 99999},
	}

	expectedMaxValue := 242 // 243 - 1

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			config := &PhonidConfig{
				Patterns:     baseConfig.Patterns,
				Placeholders: baseConfig.Placeholders,
			}

			// Add shuffle config if rounds > 0
			if tc.rounds > 0 {
				config.Shuffle = &ShuffleConfig{
					BitWidth: 0, // auto-calculate
					Rounds:   tc.rounds,
					Seed:     tc.seed,
				}
			}

			encoder, err := NewPhoneticEncoderSkipPreflight(config)
			if err != nil {
				t.Fatalf("Failed to create encoder: %v", err)
			}

			// Test that max value (242) succeeds
			maxNum := NewPositiveInt(int64(expectedMaxValue))
			word, err := encoder.Encode(maxNum)
			if err != nil {
				t.Errorf("Failed to encode max value %d: %v", expectedMaxValue, err)
			} else {
				t.Logf("Max value %d encoded successfully to: %s", expectedMaxValue, word)
			}

			// Test that max+1 (243) fails
			overMaxNum := NewPositiveInt(int64(expectedMaxValue + 1))
			_, err = encoder.Encode(overMaxNum)
			if err == nil {
				t.Errorf("Expected error encoding %d (beyond capacity), but succeeded", expectedMaxValue+1)
			} else {
				t.Logf("Correctly rejected value %d: %v", expectedMaxValue+1, err)
			}

			// Test additional values beyond capacity to ensure they all fail
			testBeyondCapacity := []int64{243, 244, 250, 255, 256, 1000}
			for _, val := range testBeyondCapacity {
				num := NewPositiveInt(val)
				_, err := encoder.Encode(num)
				if err == nil {
					t.Errorf("Expected error encoding %d (beyond capacity of %d), but succeeded",
						val, expectedMaxValue)
				}
			}
		})
	}
}

// TestShufflerBitWidthVsPatternCapacity demonstrates cycle walking in action:
// When shuffle rounds > 0, the bit width is rounded up (e.g., 8 bits for capacity 243),
// but cycle walking ensures only values 0-242 are produced, not 0-255.
func TestShufflerBitWidthVsPatternCapacity(t *testing.T) {
	config := &PhonidConfig{
		Patterns: []string{"CVCVC"}, // Capacity ~> 3^5 = 243
		Placeholders: PlaceholderMap{
			Vowel:     RuneSet{'a', 'o', 'i'},
			Consonant: RuneSet{'b', 'z', 'k'},
		},
		Shuffle: &ShuffleConfig{
			BitWidth: 0, // Will auto-calculate to 8 bits (since 2^7=128 < 243 < 2^8=256)
			Rounds:   4,
			Seed:     12345,
		},
	}

	encoder, err := NewPhoneticEncoderSkipPreflight(config)
	if err != nil {
		t.Fatalf("Failed to create encoder: %v", err)
	}

	// Get pattern info to verify bit width calculation
	patternInfo := encoder.GetPatternInfo()
	if len(patternInfo) != 1 {
		t.Fatalf("Expected 1 pattern, got %d", len(patternInfo))
	}

	trueCapacity := patternInfo[0].TrueCapacity.BigInt()
	expectedCapacity := big.NewInt(242) // 243 - 1 (max encodable value)

	if trueCapacity.Cmp(expectedCapacity) != 0 {
		t.Fatalf("Expected capacity %s, got %s", expectedCapacity, trueCapacity)
	}

	t.Logf("Pattern capacity: 0-%s (243 total combinations)", trueCapacity)
	t.Logf("Shuffler bit width: 8 bits (256 possible values)")
	t.Logf("Cycle walking ensures: Only values 0-242 are produced after shuffling")

	// Test: values in the "gap" (243-255) should be rejected at INPUT validation
	gapValues := []int64{243, 244, 250, 255}
	for _, val := range gapValues {
		num := NewPositiveInt(val)
		_, err := encoder.Encode(num)
		if err == nil {
			t.Errorf("BUG: Value %d should be rejected (beyond pattern capacity)", val)
		} else {
			t.Logf("✓ Correctly rejected input value %d: %v", val, err)
		}
	}

	// Values within capacity should work - cycle walking handles intermediate values
	validValues := []int64{0, 1, 100, 200, 240, 241, 242}
	for _, val := range validValues {
		num := NewPositiveInt(val)
		word, err := encoder.Encode(num)
		if err != nil {
			t.Errorf("Failed to encode valid value %d: %v", val, err)
			continue
		}
		t.Logf("✓ Successfully encoded valid value %d to: %s", val, word)

		// Verify round-trip
		decoded, err := encoder.Decode(word)
		if err != nil {
			t.Errorf("Failed to decode %s: %v", word, err)
			continue
		}

		decodedVal, _ := decoded.Int64()
		if decodedVal != val {
			t.Errorf("Round-trip failed: encoded %d, decoded %d", val, decodedVal)
		}
	}
}

// TestRoundTripWithShuffleConsistency verifies that encoding and decoding
// work correctly across all values in the pattern capacity with cycle walking enabled.
func TestRoundTripWithShuffleConsistency(t *testing.T) {
	config := &PhonidConfig{
		Patterns: []string{"CVCVC"},
		Placeholders: PlaceholderMap{
			Vowel:     RuneSet{'a', 'o', 'i'},
			Consonant: RuneSet{'b', 'z', 'k'},
		},
		Shuffle: &ShuffleConfig{
			Rounds: 4,
			Seed:   12345,
		},
	}

	encoder, err := NewPhoneticEncoderSkipPreflight(config)
	if err != nil {
		t.Fatalf("Failed to create encoder: %v", err)
	}

	// Test a sample of values across the entire range
	testValues := []int64{
		0, 1, 2, 3, // Start
		100, 120, 150, // Middle-low
		200, 220, 240, 241, 242, // Near end (including edge cases)
	}

	for _, val := range testValues {
		num := NewPositiveInt(val)
		word, err := encoder.Encode(num)
		if err != nil {
			t.Errorf("Failed to encode %d: %v", val, err)
			continue
		}

		decoded, err := encoder.Decode(word)
		if err != nil {
			t.Errorf("Failed to decode %s (from %d): %v", word, val, err)
			continue
		}

		decodedVal, _ := decoded.Int64()
		if decodedVal != val {
			t.Errorf("Round-trip inconsistency: %d -> %s -> %d", val, word, decodedVal)
		} else {
			t.Logf("✓ Round-trip success: %d -> %s -> %d", val, word, decodedVal)
		}
	}

	// Verify that 243 and above still fail at input validation
	_, err = encoder.Encode(NewPositiveInt(243))
	if err == nil {
		t.Error("Value 243 should be rejected but was accepted")
	} else {
		t.Logf("✓ Value 243 correctly rejected: %v", err)
	}
}

// TestCycleWalkingEdgeCases tests specific edge cases for cycle walking.
func TestCycleWalkingEdgeCases(t *testing.T) {
	config := &PhonidConfig{
		Patterns: []string{"CVCVC"},
		Placeholders: PlaceholderMap{
			Vowel:     RuneSet{'a', 'o', 'i'},
			Consonant: RuneSet{'b', 'z', 'k'},
		},
		Shuffle: &ShuffleConfig{
			Rounds: 6, // More rounds to test cycle walking behavior
			Seed:   999,
		},
	}

	encoder, err := NewPhoneticEncoderSkipPreflight(config)
	if err != nil {
		t.Fatalf("Failed to create encoder: %v", err)
	}

	// Test boundary values that are most likely to trigger cycle walking
	boundaryValues := []int64{
		0,   // Minimum
		1,   // Minimum + 1
		128, // 2^7 (bit width boundary)
		242, // Maximum (243 - 1)
		241, // Maximum - 1
		127, // 2^7 - 1
	}

	for _, val := range boundaryValues {
		num := NewPositiveInt(val)

		// Encode
		word, err := encoder.Encode(num)
		if err != nil {
			t.Errorf("Failed to encode boundary value %d: %v", val, err)
			continue
		}

		// Decode
		decoded, err := encoder.Decode(word)
		if err != nil {
			t.Errorf("Failed to decode %s (from boundary value %d): %v", word, val, err)
			continue
		}

		decodedVal, _ := decoded.Int64()
		if decodedVal != val {
			t.Errorf("Boundary value round-trip failed: %d -> %s -> %d", val, word, decodedVal)
		} else {
			t.Logf("✓ Boundary value %d: %s (round-trip OK)", val, word)
		}
	}
}
