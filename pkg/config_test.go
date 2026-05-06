package phonid_test

import (
	"reflect"
	"testing"

	. "github.com/iilei/phonid/pkg"
)

const cfgPatternCVCVC = "CVCVC"

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
			// After validation, config should be valid
			if err := got.Validate(); err != nil {
				t.Errorf("Validate() error = %v", err)
			}
		})
	}
}

func TestConfig_Validate(t *testing.T) {
	type fields struct {
		Phonetic *PhonidConfig
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
						Vowel:     RuneSet("aei"),
						CustomX:   RuneSet("."),
					},
				},
			},
			wantErr: false,
		},
		{
			name: "nil fields",
			fields: fields{
				Phonetic: nil,
			},
			wantErr: true,
		},
		{
			name: "invalid PhonidConfig",
			fields: fields{
				Phonetic: &PhonidConfig{
					Patterns: []string{cfgPatternCVCVC},
					Placeholders: PlaceholderMap{
						Consonant: RuneSet("bcdx"),
						Vowel:     RuneSet("ae"), // Only 2 vowels, needs 3 for length 5
					},
				},
			},
			wantErr: true,
		},
		{
			name: "valid PhonidConfig",
			fields: fields{
				Phonetic: &PhonidConfig{
					Patterns: []string{cfgPatternCVCVC},
					Placeholders: PlaceholderMap{
						Consonant: RuneSet("bcdx"),
						Vowel:     RuneSet("aei"),
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Config{
				Phonetic: tt.fields.Phonetic,
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
						Patterns: []string{cfgPatternCVCVC},
						Placeholders: PlaceholderMap{
							Consonant: RuneSet("bcdx"),
							Vowel:     RuneSet("aei"),
						},
					}),
				},
			},
			want: &Config{
				Phonetic: &PhonidConfig{
					Patterns: []string{cfgPatternCVCVC},
					Placeholders: PlaceholderMap{
						Consonant: RuneSet("bcdx"),
						Vowel:     RuneSet("aei"),
					},
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
		})
	}
}
