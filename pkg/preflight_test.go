package phonid_test

import (
	"testing"

	. "github.com/iilei/phonid/pkg"
)

func TestPhoneticEncoder_ValidatePreflight(t *testing.T) {
	type fields struct {
		config *PhonidConfig
		// patternEncoders []*PatternEncoder
	}
	type args struct {
		checks []PreflightCheck
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "basic test",
			fields: fields{
				config: &PhonidConfig{
					Patterns: []string{"CVC"},
					Placeholders: map[PlaceholderType]RuneSet{
						Vowel:     RuneSet("ae"),
						Consonant: RuneSet("bdf"),
					},
				},
			},
			args: args{
				checks: []PreflightCheck{
					{
						Input:  PositiveInt(0),
						Output: "bab",
					},
					{
						Input:  PositiveInt(2),
						Output: "baf",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "err test",
			fields: fields{
				config: &PhonidConfig{
					Patterns: []string{"CVC"},
					Placeholders: map[PlaceholderType]RuneSet{
						Vowel:     RuneSet("ae"),
						Consonant: RuneSet("bdf"),
					},
				},
			},
			args: args{
				checks: []PreflightCheck{
					{
						Input:  PositiveInt(0),
						Output: "bab",
					},
					{
						Input:  PositiveInt(2),
						Output: "bad",
					},
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := NewPhoneticEncoderWithPreflight(tt.fields.config, tt.args.checks)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewPhoneticEncoderWithPreflight() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err == nil && p == nil {
				t.Error("NewPhoneticEncoderWithPreflight() returned nil encoder with no error")
			}
		})
	}
}
