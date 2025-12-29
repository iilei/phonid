package phonid

import (
	"encoding/binary"
	"fmt"
	"hash/fnv"
)

type (
	// ShuffleConfig holds Feistel shuffler configuration.
	ShuffleConfig struct {
		BitWidth int    `default:"32"`
		Rounds   int    `default:"0"`
		Seed     uint64 `default:"0"`
	}

	// FeistelShuffler provides bijective integer shuffling using Feistel networks
	// Supports configurable number space size and uses standard Go libraries.
	FeistelShuffler struct {
		rounds    int      // Number of Feistel rounds (3-6 recommended)
		bitWidth  int      // Total bit width of the number space
		halfBits  int      // Bits per half (left/right)
		mask      uint64   // Mask for half-width values
		roundKeys []uint64 // Round keys derived from seed
	}
)

// Validate checks if the shuffle config is valid.
func (sc *ShuffleConfig) Validate() error {
	if sc.BitWidth < 4 || sc.BitWidth > 64 {
		return fmt.Errorf("bit_width must be between 4 and 64, got %d", sc.BitWidth)
	}
	if sc.Rounds < 3 || sc.Rounds > 10 {
		return fmt.Errorf("rounds must be between 3 and 10, got %d", sc.Rounds)
	}
	return nil
}

// NewFeistelShuffler creates a new shuffler for the given bit width
// bitWidth: total bits (8, 16, 32, 64, etc.)
// rounds: number of Feistel rounds (3-6 recommended. "0" will preserve linear order)
// seed: seed value for generating round keys
func NewFeistelShuffler(bitWidth, rounds int, seed uint64) (*FeistelShuffler, error) {
	if bitWidth < 4 || bitWidth > 64 {
		return nil, fmt.Errorf("bitWidth must be between 4 and 64, got %d", bitWidth)
	}
	if rounds < 0 || rounds > 10 {
		return nil, fmt.Errorf("rounds must be between 0 and 10, got %d", rounds)
	}

	halfBits := bitWidth / 2
	mask := (uint64(1) << halfBits) - 1

	// Generate round keys from seed using FNV hash
	roundKeys := make([]uint64, rounds)
	h := fnv.New64a()
	for i := range rounds {
		h.Reset()
		_ = binary.Write(h, binary.LittleEndian, seed)
		// #nosec G115 -- i is bounded by validation (0-10), no overflow possible
		roundIndex := uint64(i)
		_ = binary.Write(h, binary.LittleEndian, roundIndex)
		roundKeys[i] = h.Sum64()
	}

	return &FeistelShuffler{
		rounds:    rounds,
		bitWidth:  bitWidth,
		halfBits:  halfBits,
		mask:      mask,
		roundKeys: roundKeys,
	}, nil
}

// Encode performs bijective shuffling of input value.
func (fs *FeistelShuffler) Encode(input uint64) (uint64, error) {
	// Ensure input fits in our bit width
	if fs.bitWidth == 64 {
		// For 64-bit, all uint64 values are valid
	} else {
		maxValue := uint64(1) << fs.bitWidth
		if input >= maxValue {
			return 0, fmt.Errorf("input %d exceeds bit width %d (max: %d)", input, fs.bitWidth, maxValue-1)
		}
	}

	// Split input into left and right halves
	left := input >> fs.halfBits
	right := input & fs.mask

	// Feistel rounds
	for i := range fs.rounds {
		// Apply round function to right half with round key
		roundOutput := fs.roundFunction(right, fs.roundKeys[i])

		// XOR with left half and swap
		newRight := left ^ roundOutput
		left = right
		right = newRight & fs.mask // Ensure it stays within half-bit width
	}

	// Combine halves back together
	return (left << fs.halfBits) | right, nil
}

// Decode performs bijective reverse shuffling (inverse of Encode).
func (fs *FeistelShuffler) Decode(encoded uint64) (uint64, error) {
	// Ensure encoded value fits in our bit width
	if fs.bitWidth == 64 {
		// For 64-bit, all uint64 values are valid
	} else {
		maxValue := uint64(1) << fs.bitWidth
		if encoded >= maxValue {
			return 0, fmt.Errorf("encoded value %d exceeds bit width %d (max: %d)", encoded, fs.bitWidth, maxValue-1)
		}
	}

	// Split encoded value into left and right halves
	left := encoded >> fs.halfBits
	right := encoded & fs.mask

	// Reverse Feistel rounds (apply in reverse order)
	for i := fs.rounds - 1; i >= 0; i-- {
		// Apply round function to left half with round key
		roundOutput := fs.roundFunction(left, fs.roundKeys[i])

		// XOR with right half and swap
		newLeft := right ^ roundOutput
		right = left
		left = newLeft & fs.mask // Ensure it stays within half-bit width
	}

	// Combine halves back together
	return (left << fs.halfBits) | right, nil
}

// MaxValue returns the maximum value that can be shuffled.
func (fs *FeistelShuffler) MaxValue() uint64 {
	if fs.bitWidth == 64 {
		return ^uint64(0) // All bits set (max uint64)
	}
	return (uint64(1) << fs.bitWidth) - 1
}

// BitWidth returns the configured bit width.
func (fs *FeistelShuffler) BitWidth() int {
	return fs.bitWidth
}

// Rounds returns the number of Feistel rounds.
func (fs *FeistelShuffler) Rounds() int {
	return fs.rounds
}

// roundFunction implements the Feistel round function using FNV hash.
func (fs *FeistelShuffler) roundFunction(input, key uint64) uint64 {
	h := fnv.New64a()
	_ = binary.Write(h, binary.LittleEndian, input)
	_ = binary.Write(h, binary.LittleEndian, key)
	result := h.Sum64()

	// Mask to half-bit width to ensure proper size
	return result & fs.mask
}
