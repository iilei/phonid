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
					Patterns: []string{"CVCVC"},
					Placeholders: map[PlaceholderType]RuneSet{
						Vowel:     RuneSet("ae"),
						Consonant: RuneSet("bdf"),
					},
				},
			},
			args: args{
				checks: []PreflightCheck{
					{
						Input:  &TomlPositiveInt{Value: NewPositiveInt(0)},
						Output: "babab",
					},
					{
						Input:  &TomlPositiveInt{Value: NewPositiveInt(2)},
						Output: "babaf",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "err test",
			fields: fields{
				config: &PhonidConfig{
					Patterns: []string{"CVCVC"},
					Placeholders: map[PlaceholderType]RuneSet{
						Vowel:     RuneSet("ae"),
						Consonant: RuneSet("bdf"),
					},
				},
			},
			args: args{
				checks: []PreflightCheck{
					{
						Input:  &TomlPositiveInt{Value: NewPositiveInt(0)},
						Output: "babab",
					},
					{
						Input:  &TomlPositiveInt{Value: NewPositiveInt(2)},
						Output: "wrong", // This should cause an error
					},
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := NewPhoneticEncoder(tt.fields.config, tt.args.checks)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewPhoneticEncoder() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err == nil && p == nil {
				t.Error("NewPhoneticEncoder() returned nil encoder with no error")
			}
		})
	}
}
