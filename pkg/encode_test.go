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
			number:  PositiveInt(0),
			want:    "bab",
			wantErr: false,
		},
		{
			config:  *configB,
			name:    "encode zero",
			number:  PositiveInt(0),
			want:    "a" + Air + Air + Air + Air, // a🜁🜁🜁🜁
			wantErr: false,
		},
		{
			config:  *configB,
			name:    "encode 7999",
			number:  PositiveInt(7916),
			want:    "o" + Water + Fire + Regulus + HighVoltageSign, // o🜄🜂🜲⚡
			wantErr: false,
		},
		{
			config:  *configA,
			name:    "encode 1",
			number:  PositiveInt(1),
			want:    "baz",
			wantErr: false,
		},
		{
			config:  *configA,
			name:    "encode small number",
			number:  PositiveInt(5),
			want:    "bok",
			wantErr: false,
		},
		{
			config:  *configA,
			name:    "encode max value for pattern",
			number:  PositiveInt(26), // 3*3*3 - 1
			want:    "kik",
			wantErr: false,
		},
		{
			config:  *configA,
			name:    "encode number beyond max",
			number:  PositiveInt(27),
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		encoder, err := NewPhoneticEncoder(&tt.config)
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
