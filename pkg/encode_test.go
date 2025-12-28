package phonid

import (
	"testing"
)

func TestNewPhoneticEncoder(t *testing.T) {
	placeholderMapFewComposites := PlaceholderMap{Vowel: RuneSet{'a', 'e'}, Consonant: RuneSet{'z'}}
	placeholderMapNoVowels := PlaceholderMap{Vowel: RuneSet{}, Consonant: RuneSet{'z'}}
	placeholderCustomOK := PlaceholderMap{Vowel: RuneSet{'a', 'e', 'o'}, Consonant: RuneSet{'z', 'b', 'k'}}

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
			name:    "bad config: too few composites",
			args:    args{config: &PhonidConfig{Patterns: []string{"VCV"}, Placeholders: placeholderMapFewComposites}},
			want:    nil,
			wantErr: true,
		},
		{
			name:    "bad config: no Vowels",
			args:    args{config: &PhonidConfig{Patterns: []string{"VCV"}, Placeholders: placeholderMapNoVowels}},
			want:    nil,
			wantErr: true,
		},
		{
			name:    "custom config ok",
			args:    args{config: &PhonidConfig{Patterns: []string{"VCV"}, Placeholders: placeholderCustomOK}},
			want:    &PhoneticEncoder{},
			wantErr: false,
		},
		{
			name:    "ok config, bad patterns",
			args:    args{config: &PhonidConfig{Patterns: []string{"V"}, Placeholders: placeholderCustomOK}},
			want:    nil,
			wantErr: true,
		},
		{
			name:    "ok config, no patterns",
			args:    args{config: &PhonidConfig{Patterns: []string{}, Placeholders: placeholderCustomOK}},
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
			name:                  "pattern Encoder built with minimal config",
			args:                  args{pattern: "CVC", placeholders: PlaceholderMap{Vowel: RuneSet{'o', 'e', 'a'}, Consonant: RuneSet{'z', 'b', 'k'}}},
			want:                  &PatternEncoder{},
			wantErr:               false,
			wantTotalCombinations: 27,
		},
		{
			name:                  "pattern Encoder built regardless of vowel presence",
			args:                  args{pattern: "CCF", placeholders: PlaceholderMap{Fricative: RuneSet{'f'}, Consonant: RuneSet{'z', 'b', 'k'}}},
			want:                  &PatternEncoder{},
			wantErr:               false,
			wantTotalCombinations: 9,
		},
		{
			name:                  "total Combinations corresponds with proquint",
			args:                  args{pattern: ProQuintConfig.Patterns[0], placeholders: ProQuintConfig.Placeholders},
			want:                  &PatternEncoder{},
			wantErr:               false,
			wantTotalCombinations: 4294967295 + 1, // Correct: this is the COUNT of combinations
		},
		{
			name:                  "total combinations unaffected by single-option placeholder positions",
			args:                  args{pattern: "CVCVCCVCVC", placeholders: ProQuintConfig.Placeholders},
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
				t.Errorf("totalCombinations expected to be %d, got %d", tt.wantTotalCombinations, got.totalCombinations)
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
					CustomX:   RuneSet{'g'}, // Single character = multiplier of 1, doesn't affect outcome
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
				t.Errorf("patternEncoders length = %d, want %d", len(got.patternEncoders), tt.wantPatternCount)
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
	type fields struct {
		config          *PhonidConfig
		patternEncoders []*PatternEncoder
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
			e := &PhoneticEncoder{
				config:          tt.fields.config,
				patternEncoders: tt.fields.patternEncoders,
			}
			got, err := e.Encode(tt.args.number)
			if (err != nil) != tt.wantErr {
				t.Fatalf("PhoneticEncoder.Encode() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if got != tt.want {
				t.Errorf("PhoneticEncoder.Encode() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPhoneticEncoder_Decode(t *testing.T) {
	type fields struct {
		config          *PhonidConfig
		patternEncoders []*PatternEncoder
	}
	type args struct {
		word string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    int
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &PhoneticEncoder{
				config:          tt.fields.config,
				patternEncoders: tt.fields.patternEncoders,
			}
			got, err := e.Decode(tt.args.word)
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
	type fields struct {
		pattern           string
		positions         []Position
		totalCombinations PositiveInt
		length            int
	}
	type args struct {
		word string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    int
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
			got, err := e.Decode(tt.args.word)
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
	type fields struct {
		pattern           string
		positions         []Position
		totalCombinations PositiveInt
		length            int
	}
	tests := []struct {
		name   string
		fields fields
		want   int
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
			if got := e.MaxValue(); got != tt.want {
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
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := reverseString(tt.args.s); got != tt.want {
				t.Errorf("reverseString() = %v, want %v", got, tt.want)
			}
		})
	}
}
