package phonid_test

import (
	"reflect"
	"testing"

	. "github.com/iilei/phonid/pkg"
)

func TestNewConfig(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		{"sensible defaults", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewConfig()
			if (err != nil) != tt.wantErr {
				t.Errorf("NewConfig(): wantErr=%v, got error=%v", tt.wantErr, err)
			}
			if got.Phonetic == nil {
				t.Error("NewConfig().Phonetic is nil")
			}
			if got.Shuffle == nil {
				t.Error("NewConfig().Shuffle is nil")
			}
			// BitWidth is 0 until Validate() is called
			if got.Shuffle.BitWidth != 0 {
				t.Errorf("NewConfig().Shuffle.BitWidth should be 0 before Validate(), got %d", got.Shuffle.BitWidth)
			}
			// After validation, BitWidth should be auto-calculated
			if err := got.Validate(); err != nil {
				t.Errorf("Validate() error = %v", err)
			}
			if got.Shuffle.BitWidth == 0 {
				t.Error("After Validate(), Shuffle.BitWidth should be auto-calculated, got 0")
			}
		})
	}
}

func TestConfig_Validate(t *testing.T) {
	type fields struct {
		Phonetic *PhonidConfig
		Shuffle  *ShuffleConfig
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "valid config",
			fields: fields{
				Phonetic: &PhonidConfig{
					Patterns: []string{"CXVXC"},
					Placeholders: PlaceholderMap{
						Consonant: RuneSet("bcd"),
						Vowel:     RuneSet("ae"),
						CustomX:   RuneSet("."),
					},
				},
				Shuffle: &ShuffleConfig{
					Seed:   0,
					Rounds: 0,
				},
			},
			wantErr: false,
		},
		{
			name: "nil fields",
			fields: fields{
				Phonetic: nil,
				Shuffle:  nil,
			},
			wantErr: true,
		},
		{
			name: "invalid PhonidConfig",
			fields: fields{
				Phonetic: &PhonidConfig{
					Patterns: []string{"CVC"},
					Placeholders: PlaceholderMap{
						Consonant: RuneSet("b"),
						Vowel:     RuneSet("a"),
					},
				},
				Shuffle: &ShuffleConfig{
					Seed:   0,
					Rounds: 0,
				},
			},
			wantErr: true,
		},
		{
			name: "valid PhonidConfig",
			fields: fields{
				Phonetic: &PhonidConfig{
					Patterns: []string{"CVC"},
					Placeholders: PlaceholderMap{
						Consonant: RuneSet("bcdx"),
						Vowel:     RuneSet("ae"),
					},
				},
				Shuffle: &ShuffleConfig{
					Seed:   0,
					Rounds: 0,
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Config{
				Phonetic: tt.fields.Phonetic,
				Shuffle:  tt.fields.Shuffle,
			}
			if err := c.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("Config.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNewConfigWithOptions(t *testing.T) {
	type args struct {
		opts []ConfigOption
	}
	tests := []struct {
		name    string
		args    args
		want    *Config
		wantErr bool
	}{
		{
			name: "with custom options",
			args: args{
				opts: []ConfigOption{
					WithPhonetic(&PhonidConfig{
						Patterns: []string{"CVC"},
						Placeholders: PlaceholderMap{
							Consonant: RuneSet("bcdx"),
							Vowel:     RuneSet("ae"),
						},
					}),
					WithSeed(12345),
					WithRounds(3),
				},
			},
			want: &Config{
				Phonetic: &PhonidConfig{
					Patterns: []string{"CVC"},
					Placeholders: PlaceholderMap{
						Consonant: RuneSet("bcdx"),
						Vowel:     RuneSet("ae"),
					},
				},
				Shuffle: &ShuffleConfig{
					BitWidth: 5, // Auto-calculated: 4*2*4 = 32 combinations, needs 5 bits
					Seed:     12345,
					Rounds:   3,
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewConfigWithOptions(tt.args.opts...)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewConfigWithOptions() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got.Phonetic, tt.want.Phonetic) {
				t.Errorf("NewConfigWithOptions().Phonetic = %v, want %v", got.Phonetic, tt.want.Phonetic)
			}
			if !reflect.DeepEqual(got.Shuffle, tt.want.Shuffle) {
				t.Errorf("NewConfigWithOptions().Shuffle = %v, want %v", got.Shuffle, tt.want.Shuffle)
			}
		})
	}
}

func TestConfig_PreflightAssertion(t *testing.T) {
	tests := []struct {
		name              string
		phonetic          *PhonidConfig
		expectedBitWidth  int
		wantErr           bool
		wantCalculatedBit int
	}{
		{
			name: "matching expected bit width",
			phonetic: &PhonidConfig{
				Patterns: []string{"CVC"},
				Placeholders: PlaceholderMap{
					Consonant: RuneSet("bcdx"),
					Vowel:     RuneSet("ae"),
				},
			},
			expectedBitWidth:  5, // 4*2*4 = 32, needs 5 bits
			wantCalculatedBit: 5,
			wantErr:           false,
		},
		{
			name: "mismatched expected bit width",
			phonetic: &PhonidConfig{
				Patterns: []string{"CVC"},
				Placeholders: PlaceholderMap{
					Consonant: RuneSet("bcdx"),
					Vowel:     RuneSet("ae"),
				},
			},
			expectedBitWidth:  6, // Wrong expectation
			wantCalculatedBit: 5,
			wantErr:           true,
		},
		{
			name: "no expected bit width (skip assertion)",
			phonetic: &PhonidConfig{
				Patterns: []string{"CVC"},
				Placeholders: PlaceholderMap{
					Consonant: RuneSet("bcdx"),
					Vowel:     RuneSet("ae"),
				},
			},
			expectedBitWidth:  0, // No assertion
			wantCalculatedBit: 5,
			wantErr:           false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := NewConfigWithOptions(
				WithPhonetic(tt.phonetic),
				WithExpectedBitWidth(tt.expectedBitWidth),
			)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewConfigWithOptions() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && cfg.Shuffle.BitWidth != tt.wantCalculatedBit {
				t.Errorf("Calculated BitWidth = %d, want %d", cfg.Shuffle.BitWidth, tt.wantCalculatedBit)
			}
		})
	}
}
