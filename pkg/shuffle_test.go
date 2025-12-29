package phonid

import (
	"math"
	"testing"
)

//gocognit:ignore
func TestFeistelShufflerBasicBijection(t *testing.T) {
	testCases := []struct {
		bitWidth int
		rounds   int
		seed     uint64
	}{
		{8, 4, 12345},
		{16, 4, 67890},
		{32, 6, 54321},
		{64, 6, 98765},
	}

	for _, tc := range testCases {
		t.Run("", func(t *testing.T) {
			shuffler, _ := NewFeistelShuffler(tc.bitWidth, tc.rounds, tc.seed)

			maxValue := shuffler.MaxValue()
			var testValues []uint64

			testValues = []uint64{0, 1, 42}

			if maxValue >= 255 {
				testValues = append(testValues, 255)
			}
			if maxValue >= 1000 {
				testValues = append(testValues, 1000)
			}
			if maxValue >= 65535 {
				testValues = append(testValues, 65535)
			}

			// Werte nahe dem Maximum
			testValues = append(testValues, maxValue/4, maxValue/2, maxValue-1)

			for _, original := range testValues {
				if original > maxValue {
					continue
				}

				// Encode
				encoded, _ := shuffler.Encode(original)

				if encoded > maxValue {
					t.Errorf("Encoded value %d exceeds max value %d", encoded, maxValue)
				}

				// Decode
				decoded, _ := shuffler.Decode(encoded)

				if decoded != original {
					t.Errorf("Bijection failed: original=%d, encoded=%d, decoded=%d (bitWidth=%d)",
						original, encoded, decoded, tc.bitWidth)
				}

				if original > 1 && encoded == original {
					t.Logf(
						"Warning: Encoding didn't change value %d (might be rare edge case)",
						original,
					)
				}
			}
		})
	}
}

func TestFeistelShufflerCompleteBijection(t *testing.T) {
	shuffler, _ := NewFeistelShuffler(8, 4, 42)
	maxValue := shuffler.MaxValue() // 255

	used := make(map[uint64]bool)

	for i := uint64(0); i <= maxValue; i++ {
		encoded, _ := shuffler.Encode(i)

		if used[encoded] {
			t.Errorf("Collision detected: value %d produces duplicate encoded value %d", i, encoded)
		}
		used[encoded] = true

		// Decode-Test
		decoded, _ := shuffler.Decode(encoded)
		if decoded != i {
			t.Errorf("Bijection failed for %d: encoded=%d, decoded=%d", i, encoded, decoded)
		}
	}

	if len(used) != 256 {
		t.Errorf("Expected 256 unique encoded values, got %d", len(used))
	}
}

func TestFeistelShufflerDifferentSeeds(t *testing.T) {
	seed1 := uint64(11111)
	seed2 := uint64(22222)

	shuffler1, _ := NewFeistelShuffler(16, 4, seed1)
	shuffler2, _ := NewFeistelShuffler(16, 4, seed2)

	testValue := uint64(12345)

	encoded1, _ := shuffler1.Encode(testValue)
	encoded2, _ := shuffler2.Encode(testValue)

	if encoded1 == encoded2 {
		t.Errorf("Different seeds produced same result: %d", encoded1)
	}

	decoded1, _ := shuffler1.Decode(encoded1)
	if decoded1 != testValue {
		t.Errorf("Shuffler1 failed to decode correctly")
	}
	decoded2, _ := shuffler2.Decode(encoded2)
	if decoded2 != testValue {
		t.Errorf("Shuffler2 failed to decode correctly")
	}

	crossDecoded1, _ := shuffler2.Decode(encoded1)
	crossDecoded2, _ := shuffler1.Decode(encoded2)

	if crossDecoded1 == testValue {
		t.Errorf("Cross-decoding incorrectly succeeded (seed2 decoded seed1's value)")
	}
	if crossDecoded2 == testValue {
		t.Errorf("Cross-decoding incorrectly succeeded (seed1 decoded seed2's value)")
	}
}

func TestFeistelShufflerRoundsEffect(t *testing.T) {
	seed := uint64(99999)
	testValue := uint64(12345)

	shuffler3, _ := NewFeistelShuffler(32, 3, seed)
	shuffler6, _ := NewFeistelShuffler(32, 6, seed)

	encoded3, _ := shuffler3.Encode(testValue)
	encoded6, _ := shuffler6.Encode(testValue)

	if encoded3 == encoded6 {
		t.Errorf("Different round counts produced same result: %d", encoded3)
	}

	crossDecoded3, _ := shuffler3.Decode(encoded3)
	crossDecoded6, _ := shuffler6.Decode(encoded6)

	if crossDecoded3 != testValue {
		t.Errorf("3-round shuffler failed to decode correctly")
	}
	if crossDecoded6 != testValue {
		t.Errorf("6-round shuffler failed to decode correctly")
	}
}

func TestFeistelShufflerInvalidInputs(t *testing.T) {
	_, err := NewFeistelShuffler(2, 4, 12345)
	if err == nil {
		t.Errorf("Expected error for invalid bitWidth")
	}
}

func TestFeistelShufflerInvalidInputs2(t *testing.T) {
	_, err := NewFeistelShuffler(32, 15, 12345)
	if err == nil {
		t.Errorf("Expected error for too many rounds")
	}
}

func TestFeistelShufflerValueTooLarge(t *testing.T) {
	shuffler, _ := NewFeistelShuffler(8, 4, 12345)

	_, err := shuffler.Encode(256)
	if err == nil {
		t.Errorf("Expected error for input exceeding bit width")
	}
}

func TestFeistelShufflerProperties(t *testing.T) {
	shuffler, _ := NewFeistelShuffler(16, 5, 54321)

	if shuffler.BitWidth() != 16 {
		t.Errorf("Expected bit width 16, got %d", shuffler.BitWidth())
	}

	if shuffler.Rounds() != 5 {
		t.Errorf("Expected 5 rounds, got %d", shuffler.Rounds())
	}

	expectedMaxValue := uint64((1 << 16) - 1)
	if shuffler.MaxValue() != expectedMaxValue {
		t.Errorf("Expected max value %d, got %d", expectedMaxValue, shuffler.MaxValue())
	}
}

func BenchmarkFeistelShufflerEncode(b *testing.B) {
	shuffler, _ := NewFeistelShuffler(32, 6, 123456)
	testValue := uint64(987654)

	for b.Loop() {
		_, _ = shuffler.Encode(testValue)
	}
}

func BenchmarkFeistelShufflerDecode(b *testing.B) {
	shuffler, _ := NewFeistelShuffler(32, 6, 123456)
	testValue := uint64(987654)
	encoded, _ := shuffler.Encode(testValue)

	for b.Loop() {
		_, _ = shuffler.Decode(encoded)
	}
}

func TestCrossPlatformConsistency(t *testing.T) {
	shuffler, _ := NewFeistelShuffler(64, 4, 12345)

	testCases := []struct {
		input, encoded uint64
	}{
		{42, 16609768389896683095},
		{1337, 937660670618793403},
		{0, 16615908886813486803},
		{math.MaxUint64, 15298063205617206831},
	}

	for _, tc := range testCases {
		input, encoded := tc.input, tc.encoded

		actual, _ := shuffler.Encode(input)
		if actual != encoded {
			t.Errorf("Cross-platform inconsistency: input=%d, expected=%d, got=%d",
				input, encoded, actual)
		}

		reversed, _ := shuffler.Decode(actual)
		if reversed != input {
			t.Errorf("Bijection failed: input=%d, encoded=%d, decoded=%d",
				input, actual, reversed)
		}
	}
}

func TestNoRoundsShuffling(t *testing.T) {
	shuffler, _ := NewFeistelShuffler(64, 0, 12345)

	testCases := []struct {
		input, encoded uint64
	}{
		{42, 42},
		{1337, 1337},
		{0, 0},
		{math.MaxUint64, math.MaxUint64},
	}

	for _, tc := range testCases {
		input, encoded := tc.input, tc.encoded

		actual, _ := shuffler.Encode(input)
		if actual != encoded {
			t.Errorf("Cross-platform inconsistency: input=%d, expected=%d, got=%d",
				input, encoded, actual)
		}

		reversed, _ := shuffler.Decode(actual)
		if reversed != input {
			t.Errorf("Bijection failed: input=%d, encoded=%d, decoded=%d",
				input, actual, reversed)
		}
	}
}
