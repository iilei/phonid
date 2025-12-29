package phonid

import (
	"testing"
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
			name: "ok config, no patterns",
			args: args{
				config: &PhonidConfig{Patterns: []string{}, Placeholders: placeholderCustomOK},
			},
			want:    &PhoneticEncoder{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewPhoneticEncoder(tt.args.config)
			if (err != nil) != tt.wantErr {
				t.Fatalf("NewPhoneticEncoder() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if got == nil {
				t.Error("NewPhoneticEncoder() = nil, want *PhoneticEncoder")
			}
		})
	}
}

func Test_buildPatternEncoder(t *testing.T) {
	type args struct {
		pattern      string
		placeholders PlaceholderMap
	}
	tests := []struct {
		name                  string
		args                  args
		want                  *PatternEncoder
		wantErr               bool
		wantTotalCombinations int
	}{
		{
			name: "pattern Encoder built with minimal config",
			args: args{
				pattern: "CVC",
				placeholders: PlaceholderMap{
					Vowel:     RuneSet{'o', 'e', 'a'},
					Consonant: RuneSet{'z', 'b', 'k'},
				},
			},
			want:                  &PatternEncoder{},
			wantErr:               false,
			wantTotalCombinations: 27,
		},
		{
			name: "pattern Encoder built regardless of vowel presence",
			args: args{
				pattern: "CCF",
				placeholders: PlaceholderMap{
					Fricative: RuneSet{'f'},
					Consonant: RuneSet{'z', 'b', 'k'},
				},
			},
			want:                  &PatternEncoder{},
			wantErr:               false,
			wantTotalCombinations: 9,
		},
		{
			name: "total Combinations corresponds with proquint",
			args: args{
				pattern:      ProQuintConfig.Patterns[0],
				placeholders: ProQuintConfig.Placeholders,
			},
			want:                  &PatternEncoder{},
			wantErr:               false,
			wantTotalCombinations: 4294967295 + 1, // Correct: this is the COUNT of combinations
		},
		{
			name: "total combinations unaffected by single-option placeholder positions",
			args: args{
				pattern:      "CVCVCCVCVC",
				placeholders: ProQuintConfig.Placeholders,
			},
			want:                  &PatternEncoder{},
			wantErr:               false,
			wantTotalCombinations: 4294967295 + 1, // same as proquint
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := buildPatternEncoder(tt.args.pattern, tt.args.placeholders)
			if (err != nil) != tt.wantErr {
				t.Fatalf("buildPatternEncoder() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if got.totalCombinations != PositiveInt(tt.wantTotalCombinations) {
				t.Errorf(
					"totalCombinations expected to be %d, got %d",
					tt.wantTotalCombinations,
					got.totalCombinations,
				)
			}
		})
	}
}

func Test_newPhoneticEncoder(t *testing.T) {
	placeholders := PlaceholderMap{
		Vowel:     RuneSet{'a', 'e', 'o'},
		Consonant: RuneSet{'b', 'd', 'k'},
	}

	type args struct {
		config *PhonidConfig
	}
	tests := []struct {
		name                  string
		args                  args
		wantPatternCount      int
		wantTotalCombinations []PositiveInt // ordered by size
		wantErr               bool
	}{
		{
			name: "single pattern",
			args: args{config: &PhonidConfig{
				Patterns:     []string{"CVC"},
				Placeholders: placeholders,
			}},
			wantPatternCount:      1,
			wantTotalCombinations: []PositiveInt{27}, // 3*3*3
			wantErr:               false,
		},
		{
			name: "multiple patterns sorted by totalCombinations",
			args: args{config: &PhonidConfig{
				// VCVVCVV ~> 4*2*4*4*2*4*4 = 4096
				//   CXXXC ~>     2*1*1*1*2 = 4
				//     VCV ~>         4*2*4 = 32

				Patterns: []string{"VCV", "CXXXC", "VCVVCVV"}, // unsorted input with
				Placeholders: PlaceholderMap{
					Vowel:     RuneSet{'a', 'e', 'o', 'i'},
					Consonant: RuneSet{'b', 'd'},
					CustomX:   RuneSet{'g'}, // Single character = multiplier of 1
				},
			}},
			wantPatternCount:      3,
			wantTotalCombinations: []PositiveInt{4, 32, 4096},
			wantErr:               false,
		},
		{
			name: "empty patterns",
			args: args{config: &PhonidConfig{
				Patterns:     []string{},
				Placeholders: placeholders,
			}},
			wantPatternCount:      0,
			wantTotalCombinations: []PositiveInt{},
			wantErr:               false,
		},
		{
			name: "different pattern lengths with identical total combinations",
			args: args{config: &PhonidConfig{
				Patterns: []string{
					"CVC",   // ~> 3*3*3 = 27
					"CVCXX", // ~> 3*3*3*1*1 = 27
				},
				Placeholders: PlaceholderMap{
					Vowel:     RuneSet{'a', 'e', 'o'},
					Consonant: RuneSet{'b', 'd', 'k'},
					CustomX: RuneSet{
						'g',
					}, // Single character = multiplier of 1, doesn't affect outcome
				},
			}},
			wantPatternCount:      2,
			wantTotalCombinations: nil, // Both patterns produce 27 combinations
			wantErr:               true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := newPhoneticEncoder(tt.args.config)
			if (err != nil) != tt.wantErr {
				t.Fatalf("newPhoneticEncoder() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}

			// Check pattern count
			if len(got.patternEncoders) != tt.wantPatternCount {
				t.Errorf(
					"patternEncoders length = %d, want %d",
					len(got.patternEncoders),
					tt.wantPatternCount,
				)
			}

			// Compare totalCombinations
			for i, encoder := range got.patternEncoders {
				if i >= len(tt.wantTotalCombinations) {
					t.Errorf("unexpected extra pattern encoder at index %d", i)
					continue
				}
				if encoder.totalCombinations != tt.wantTotalCombinations[i] {
					t.Errorf("patternEncoders[%d].totalCombinations = %d, want %d",
						i, encoder.totalCombinations, tt.wantTotalCombinations[i])
				}
			}
		})
	}
}

func TestPhoneticEncoder_Encode(t *testing.T) {
	// Create a simple config for testing
	simpleConfig := &PhonidConfig{
		Patterns: []string{"CVC"},
		Placeholders: PlaceholderMap{
			Vowel:     RuneSet{'a', 'o', 'i'},
			Consonant: RuneSet{'b', 'z', 'k'},
		},
	}

	// Create encoder once
	encoder, err := NewPhoneticEncoder(simpleConfig)
	if err != nil {
		t.Fatalf("failed to create encoder: %v", err)
	}

	tests := []struct {
		name    string
		number  PositiveInt
		want    string
		wantErr bool
	}{
		{
			name:    "encode zero",
			number:  PositiveInt(0),
			want:    "bab",
			wantErr: false,
		},
		{
			name:    "encode one",
			number:  PositiveInt(1),
			want:    "baz",
			wantErr: false,
		},
		{
			name:    "encode small number",
			number:  PositiveInt(5),
			want:    "bok",
			wantErr: false,
		},
		{
			name:    "encode max value for pattern",
			number:  PositiveInt(26), // 3*3*3 - 1
			want:    "kik",
			wantErr: false,
		},
		{
			name:    "encode number beyond max",
			number:  PositiveInt(27),
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
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
	encoder, err := NewPhoneticEncoder(simpleConfig)
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
			if got != tt.want {
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

	encoder, err := NewPhoneticEncoder(simpleConfig)
	if err != nil {
		t.Fatalf("failed to create encoder: %v", err)
	}

	// Test every valid number in the range
	maxValue := 27 // 3*3*3
	for i := range maxValue {
		num := PositiveInt(i)

		// Encode
		word, err := encoder.Encode(num)
		if err != nil {
			t.Fatalf("Encode(%d) failed: %v", num, err)
		}

		// Decode
		decoded, err := encoder.Decode(word)
		if err != nil {
			t.Fatalf("Decode(%s) failed: %v", word, err)
		}

		// Verify round-trip
		if decoded != int(num) {
			t.Errorf("Round-trip failed: %d -> %s -> %d", num, word, decoded)
		}
	}
}

func TestPatternEncoder_Encode(t *testing.T) {
	type fields struct {
		pattern           string
		positions         []Position
		totalCombinations PositiveInt
		length            int
	}
	type args struct {
		number PositiveInt
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    string
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &PatternEncoder{
				pattern:           tt.fields.pattern,
				positions:         tt.fields.positions,
				totalCombinations: tt.fields.totalCombinations,
				length:            tt.fields.length,
			}
			got, err := e.Encode(tt.args.number)
			if (err != nil) != tt.wantErr {
				t.Fatalf("PatternEncoder.Encode() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if got != tt.want {
				t.Errorf("PatternEncoder.Encode() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPatternEncoder_Decode(t *testing.T) {
	// Build a simple pattern encoder directly
	encoder, err := buildPatternEncoder("CVC", PlaceholderMap{
		Vowel:     RuneSet{'a', 'o', 'i'},
		Consonant: RuneSet{'b', 'z', 'k'},
	})
	if err != nil {
		t.Fatalf("failed to build pattern encoder: %v", err)
	}

	tests := []struct {
		name    string
		word    string
		want    int
		wantErr bool
	}{
		{
			name:    "decode first value",
			word:    "bab",
			want:    0,
			wantErr: false,
		},
		{
			name:    "decode last value",
			word:    "kik",
			want:    26,
			wantErr: false,
		},
		{
			name:    "wrong length",
			word:    "ba",
			want:    0,
			wantErr: true,
		},
		{
			name:    "invalid character",
			word:    "bax",
			want:    0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := encoder.Decode(tt.word)
			if (err != nil) != tt.wantErr {
				t.Fatalf("PatternEncoder.Decode() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if got != tt.want {
				t.Errorf("PatternEncoder.Decode() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPatternEncoder_MaxValue(t *testing.T) {
	placeholders := PlaceholderMap{
		Vowel:     RuneSet{'a', 'o', 'i'},
		Consonant: RuneSet{'b', 'z', 'k'},
	}

	tests := []struct {
		name    string
		pattern string
		want    int
	}{
		{
			name:    "VC pattern - 3x3",
			pattern: "VC",
			want:    8, // 9 combinations - 1
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encoder, err := buildPatternEncoder(tt.pattern, placeholders)
			if err != nil {
				t.Fatalf("buildPatternEncoder() failed: %v", err)
			}
			if got := encoder.MaxValue(); got != tt.want {
				t.Errorf("PatternEncoder.MaxValue() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_reverseString(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "unicode support",
			args: args{s: "\u26a1\u2728\U0001b66d\u0061\u0062\u0063"},
			want: "\u0063\u0062\u0061\U0001b66d\u2728\u26a1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := reverseString(tt.args.s); got != tt.want {
				t.Errorf("reverseString() = %q, want %q", got, tt.want)
			}
		})
	}
}
