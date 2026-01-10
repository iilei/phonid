package phonid_test

import (
	"testing"

	. "github.com/iilei/phonid/pkg"
)

// see https://www.fileformat.info/info/unicode/block/alchemical_symbols/list.htm for more.
const (
	Air             = "\U0001F701" // 🜁 ALCHEMICAL SYMBOL FOR AIR
	Fire            = "\U0001F702" // 🜂 ALCHEMICAL SYMBOL FOR FIRE
	Earth           = "\U0001F703" // 🜃 ALCHEMICAL SYMBOL FOR EARTH
	Water           = "\U0001F704" // 🜄 ALCHEMICAL SYMBOL FOR WATER
	Aqua            = "\U0001F709" // 🜉 ALCHEMICAL SYMBOL FOR AQUA VITAE-2
	Regulus         = "\U0001F732" // 🜲 ALCHEMICAL SYMBOL FOR REGULUS
	HighVoltageSign = "\u26a1"     // ⚡ HIGH VOLTAGE SIGN
	Sparkles        = "\u2728"     // ✨ SPARKLES
)

func TestNewPhoneticEncoder(t *testing.T) {
	placeholderMapFewComposites := PlaceholderMap{Vowel: RuneSet{'a', 'e'}, Consonant: RuneSet{'z'}}
	placeholderMapNoVowels := PlaceholderMap{Vowel: RuneSet{}, Consonant: RuneSet{'z'}}
	placeholderCustomOK := PlaceholderMap{
		Vowel:     RuneSet{'a', 'e', 'o'},
		Consonant: RuneSet{'z', 'b', 'k'},
	}

	type args struct {
		config *PhonidConfig
	}
	tests := []struct {
		name    string
		args    args
		want    *PhoneticEncoder
		wantErr bool
	}{
		{
			name:    "nil config",
			args:    args{config: &PhonidConfig{}},
			want:    &PhoneticEncoder{},
			wantErr: false,
		},
		{
			name: "bad config: too few composites",
			args: args{
				config: &PhonidConfig{
					Patterns:     []string{"VCV"},
					Placeholders: placeholderMapFewComposites,
				},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "bad config: no Vowels",
			args: args{
				config: &PhonidConfig{
					Patterns:     []string{"VCV"},
					Placeholders: placeholderMapNoVowels,
				},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "custom config ok",
			args: args{
				config: &PhonidConfig{Patterns: []string{"VCV"}, Placeholders: placeholderCustomOK},
			},
			want:    &PhoneticEncoder{},
			wantErr: false,
		},
		{
			name: "ok config, bad patterns",
			args: args{
				config: &PhonidConfig{Patterns: []string{"V"}, Placeholders: placeholderCustomOK},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "empty patterns with custom placeholders should error",
			args: args{
				config: &PhonidConfig{Patterns: []string{}, Placeholders: placeholderCustomOK},
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewPhoneticEncoderSkipPreflight(tt.args.config)
			if (err != nil) != tt.wantErr {
				t.Fatalf("NewPhoneticEncoderSkipPreflight() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if got == nil {
				t.Error("NewPhoneticEncoderSkipPreflight() = nil, want *PhoneticEncoder")
			}
		})
	}
}

func TestPhoneticEncoder_Encode(t *testing.T) {
	// Create a simple config for testing
	configA := &PhonidConfig{
		Patterns: []string{"CVC"},
		Placeholders: PlaceholderMap{
			Vowel:     RuneSet{'a', 'o', 'i'},
			Consonant: []rune("bzk"),
		},
	}
	configB := &PhonidConfig{
		Patterns: []string{"VCCCC"},
		Placeholders: PlaceholderMap{
			Vowel:     RuneSet{'a', 'o', 'i', 'e', 'u'},
			Consonant: []rune(Air + Aqua + Earth + Fire + HighVoltageSign + Regulus + Sparkles + Water),
		},
	}

	tests := []struct {
		config  PhonidConfig
		name    string
		number  PositiveInt
		want    string
		wantErr bool
	}{
		{
			config:  *configA,
			name:    "encode zero",
			number:  NewPositiveInt(0),
			want:    "bab",
			wantErr: false,
		},
		{
			config:  *configB,
			name:    "encode zero",
			number:  NewPositiveInt(0),
			want:    "a" + Air + Air + Air + Air, // a🜁🜁🜁🜁
			wantErr: false,
		},
		{
			config:  *configB,
			name:    "encode 7999",
			number:  NewPositiveInt(7916),
			want:    "o" + Water + Fire + Regulus + HighVoltageSign, // o🜄🜂🜲⚡
			wantErr: false,
		},
		{
			config:  *configA,
			name:    "encode 1",
			number:  NewPositiveInt(1),
			want:    "baz",
			wantErr: false,
		},
		{
			config:  *configA,
			name:    "encode small number",
			number:  NewPositiveInt(5),
			want:    "bok",
			wantErr: false,
		},
		{
			config:  *configA,
			name:    "encode max value for pattern",
			number:  NewPositiveInt(26), // 3*3*3 - 1
			want:    "kik",
			wantErr: false,
		},
		{
			config:  *configA,
			name:    "encode number beyond max",
			number:  NewPositiveInt(27),
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		encoder, err := NewPhoneticEncoderSkipPreflight(&tt.config)
		if err != nil {
			t.Fatalf("failed to create encoder: %v", err)
		}

		t.Run(tt.name, func(t *testing.T) {
			got, err := encoder.Encode(tt.number)
			if (err != nil) != tt.wantErr {
				t.Fatalf("PhoneticEncoder.Encode() error = %v, wantErr %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("PhoneticEncoder.Encode() = %v, want %v", got, tt.want)
			}
			if (err != nil) != tt.wantErr {
				t.Fatalf("PhoneticEncoder.Encode() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPhoneticEncoder_Decode(t *testing.T) {
	// Create a simple config for testing
	simpleConfig := &PhonidConfig{
		Patterns: []string{"CVC"},
		Placeholders: PlaceholderMap{
			Vowel:     RuneSet{'a', 'o', 'i'},
			Consonant: RuneSet{'b', 'z', 'k'},
		},
	}

	// Create encoder once
	encoder, err := NewPhoneticEncoderSkipPreflight(simpleConfig)
	if err != nil {
		t.Fatalf("failed to create encoder: %v", err)
	}

	tests := []struct {
		name    string
		word    string
		want    int
		wantErr bool
	}{
		{
			name:    "decode zero",
			word:    "bab",
			want:    0,
			wantErr: false,
		},
		{
			name:    "decode one",
			word:    "baz",
			want:    1,
			wantErr: false,
		},
		{
			name:    "decode small number",
			word:    "bok",
			want:    5,
			wantErr: false,
		},
		{
			name:    "decode max value for pattern",
			word:    "kik",
			want:    26, // 3*3*3 - 1
			wantErr: false,
		},
		{
			name:    "decode invalid word - wrong length",
			word:    "ba",
			want:    0,
			wantErr: true,
		},
		{
			name:    "decode invalid word - invalid character",
			word:    "bax", // 'x' not in consonant set
			want:    0,
			wantErr: true,
		},
		{
			name:    "decode invalid word - wrong placeholder",
			word:    "bbb", // vowel position has consonant
			want:    0,
			wantErr: true,
		},
		{
			name:    "decode empty string",
			word:    "",
			want:    0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := encoder.Decode(tt.word)
			if (err != nil) != tt.wantErr {
				t.Fatalf("PhoneticEncoder.Decode() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			gotVal, ok := got.Int64()
			if !ok {
				t.Fatalf("too large")
			}
			if int(gotVal) != tt.want {
				t.Errorf("PhoneticEncoder.Decode() = %v, want %v", got, tt.want)
			}
		})
	}
}

// Round-trip test to verify Encode/Decode are inverses.
func TestPhoneticEncoder_RoundTrip(t *testing.T) {
	simpleConfig := &PhonidConfig{
		Patterns: []string{"CVC"},
		Placeholders: PlaceholderMap{
			Vowel:     RuneSet{'a', 'o', 'i'},
			Consonant: RuneSet{'b', 'z', 'k'},
		},
	}

	encoder, err := NewPhoneticEncoderSkipPreflight(simpleConfig)
	if err != nil {
		t.Fatalf("failed to create encoder: %v", err)
	}

	// Test every valid number in the range
	maxValue := 27 // 3*3*3
	for i := range maxValue {
		num := NewPositiveInt(int64(i))

		// Encode
		word, err := encoder.Encode(num)
		if err != nil {
			t.Fatalf("Encode(%s) failed: %v", num.String(), err)
		}

		// Decode
		decoded, err := encoder.Decode(word)
		if err != nil {
			t.Fatalf("Decode(%s) failed: %v", word, err)
		}

		// Verify round-trip
		if decoded.Cmp(num) != 0 {
			t.Errorf("Round-trip failed: %s -> %s -> %s", num.String(), word, decoded.String())
		}
	}
}

func TestPhoneticEncoder_LargePatternCapacity(t *testing.T) {
	// Test pattern with capacity exceeding math.MaxInt
	// CVCVCXCVCVCXCVCVCXCVCVC has 23 positions:
	// - 19 C positions (16 choices each) = 16^19
	// - 3 X positions (1 choice each) = 1^3
	// - 4 V positions (4 choices each) = 4^4
	// Total = 16^19 * 4^4 * 1^3 ≈ 2^76 * 2^8 = 2^84 (much larger than math.MaxInt = 2^63-1)
	largePatternConfig := &PhonidConfig{
		Patterns: []string{"CVCVCXCVCVCXCVCVCXCVCVC"},
		Placeholders: PlaceholderMap{
			Vowel:     RuneSet{'a', 'i', 'o', 'u'},
			Consonant: RuneSet{'b', 'd', 'f', 'g', 'h', 'j', 'k', 'l', 'm', 'n', 'p', 'r', 's', 't', 'v', 'z'},
			CustomX:   RuneSet{'-'},
		},
	}

	encoder, err := NewPhoneticEncoderSkipPreflight(largePatternConfig)
	if err != nil {
		t.Fatalf("failed to create encoder with large pattern: %v", err)
	}

	// Test encoding small values (should work fine)
	testCases := []struct {
		input    PositiveInt
		expected string
	}{
		{NewPositiveInt(0), "babab-babab-babab-babab"},
		{NewPositiveInt(1), "babab-babab-babab-babad"},
		{NewPositiveInt(255), "babab-babab-babab-baguz"},
		{NewPositiveInt(1000000), "babab-babab-babaz-hanab"},
	}

	for _, tc := range testCases {
		result, err := encoder.Encode(tc.input)
		if err != nil {
			t.Errorf("Encode(%d) failed: %v", tc.input, err)
		}
		if result != tc.expected {
			t.Errorf("Encode(%d) = %q, want %q", tc.input, result, tc.expected)
		}

		// Verify round-trip
		decoded, err := encoder.Decode(result)
		if err != nil {
			t.Errorf("Decode(%q) failed: %v", result, err)
		}
		if decoded.Cmp(tc.input) != 0 {
			t.Errorf("Round-trip failed: %s -> %s -> %s", tc.input.String(), result, decoded.String())
		}
	}

	// Verify capacity reporting (should report MaxInt since actual capacity > MaxInt)
	capacity := encoder.GetSmallestPatternCapacity()
	// math.MaxInt for 64-bit systems
	const expectedMaxInt = 9223372036854775807
	if capacity != expectedMaxInt {
		t.Errorf("GetSmallestPatternCapacity() = %d, want %d (MaxInt)", capacity, expectedMaxInt)
	}
}

// TestShufflingWithUint64Max tests that shuffling works with uint64 maximum value.
// This ensures BigInt support in the shuffler for values > MaxInt64.
func TestShufflingWithUint64Max(t *testing.T) {
	// Config from the example - supports uint64 max
	config := &PhonidConfig{
		Patterns: []string{"VCV", "CVCCV", "CVCVCXCVCVC", "CVCVCXCVCVCXCVCVCXCVCVC"},
		Placeholders: PlaceholderMap{
			Consonant: RuneSet("bdfghjklmnprstvz"),
			Vowel:     RuneSet("aiou"),
			CustomX:   RuneSet("-"),
		},
		Shuffle: &ShuffleConfig{
			Rounds: 2,
			Seed:   0,
		},
	}

	encoder, err := NewPhoneticEncoderSkipPreflight(config)
	if err != nil {
		t.Fatalf("failed to create encoder: %v", err)
	}

	// Test uint64 max value: 18446744073709551615
	maxUint64, err := ParsePositiveInt("18446744073709551615")
	if err != nil {
		t.Fatalf("failed to parse uint64 max: %v", err)
	}

	// Encode with shuffling
	encoded, err := encoder.Encode(maxUint64)
	if err != nil {
		t.Fatalf("failed to encode uint64 max with shuffling: %v", err)
	}

	if encoded == "" {
		t.Error("encoded result is empty")
	}

	// Decode and verify round-trip
	decoded, err := encoder.Decode(encoded)
	if err != nil {
		t.Fatalf("failed to decode: %v", err)
	}

	if decoded.Cmp(maxUint64) != 0 {
		t.Errorf("round-trip failed: expected %s, got %s", maxUint64.String(), decoded.String())
	}

	t.Logf("Successfully encoded and decoded uint64 max (%s) with shuffling to: %s", maxUint64.String(), encoded)
}

// TestShufflingWithLargeNumbers tests shuffling with various large numbers.
func TestShufflingWithLargeNumbers(t *testing.T) {
	config := &PhonidConfig{
		Patterns: []string{"CVCVCXCVCVCXCVCVCXCVCVC"},
		Placeholders: PlaceholderMap{
			Consonant: RuneSet("bdfghjklmnprstvz"),
			Vowel:     RuneSet("aiou"),
			CustomX:   RuneSet("-"),
		},
		Shuffle: &ShuffleConfig{
			Rounds: 3,
			Seed:   12345,
		},
	}

	encoder, err := NewPhoneticEncoderSkipPreflight(config)
	if err != nil {
		t.Fatalf("failed to create encoder: %v", err)
	}

	testCases := []string{
		"9223372036854775807",  // MaxInt64
		"9223372036854775808",  // MaxInt64 + 1
		"18446744073709551614", // MaxUint64 - 1
		"18446744073709551615", // MaxUint64
	}

	for _, numStr := range testCases {
		num, err := ParsePositiveInt(numStr)
		if err != nil {
			t.Fatalf("failed to parse %s: %v", numStr, err)
		}

		// Encode with shuffling
		encoded, err := encoder.Encode(num)
		if err != nil {
			t.Errorf("failed to encode %s: %v", numStr, err)
			continue
		}

		// Decode and verify
		decoded, err := encoder.Decode(encoded)
		if err != nil {
			t.Errorf("failed to decode %s: %v", encoded, err)
			continue
		}

		if decoded.Cmp(num) != 0 {
			t.Errorf("round-trip failed for %s: got %s", numStr, decoded.String())
		}
	}
}
