package phonid

import (
	"reflect"
	"testing"
)

func TestNewConfig(t *testing.T) {
	cfg := &Config{
		Phonetic: &PhonidConfig{}, // defaults are applied when validatePattern is called
		Shuffle:  &ShuffleConfig{32, 0, 0},
	}

	tests := []struct {
		name    string
		want    Config
		wantErr bool
	}{
		{"sensible defaults", *cfg, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewConfig()
			if (err != nil) != tt.wantErr {
				t.Errorf("NewConfig(): wantErr=%v, got error=%v", tt.wantErr, err)
			}
			if !reflect.DeepEqual(got.Phonetic, tt.want.Phonetic) {
				t.Errorf(
					"NewConfigWithOptions().Phonetic = %v, want %v",
					got.Phonetic,
					tt.want.Phonetic,
				)
			}
			if !reflect.DeepEqual(got.Shuffle, tt.want.Shuffle) {
				t.Errorf(
					"NewConfigWithOptions().Shuffle = %v, want %v",
					got.Shuffle,
					tt.want.Shuffle,
				)
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
			name: "non-shuffle",
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
					BitWidth: 32,
					Seed:     0,
					Rounds:   0,
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
					BitWidth: 32,
					Seed:     0,
					Rounds:   0,
				},
			},
			wantErr: true,
		},
		{
			name: "invalid ShuffleConfig",
			fields: fields{
				Phonetic: &PhonidConfig{
					Patterns: []string{"CVC"},
					Placeholders: PlaceholderMap{
						Consonant: RuneSet("bcdx"),
						Vowel:     RuneSet("ae"),
					},
				},
				Shuffle: &ShuffleConfig{
					BitWidth: 32,
					Seed:     0,
					Rounds:   0,
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
					WithBitWidth(16),
					WithSeed(12345),
					WithRounds(3),
					WithPhonetic(&PhonidConfig{
						Patterns: []string{"CVC"},
						Placeholders: PlaceholderMap{
							Consonant: RuneSet("bcdx"),
							Vowel:     RuneSet("ae"),
						},
					}),
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
					BitWidth: 16,
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
