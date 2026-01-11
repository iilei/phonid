package phonid_test

import (
	"testing"

	. "github.com/iilei/phonid/pkg"
	"github.com/iilei/phonid/pkg/preflight"
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
						Vowel:     RuneSet("aei"),
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
						Vowel:     RuneSet("aei"),
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
						Output: "babib",
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

func TestPhoneticEncoder_ValidateShuffleCycleWalking(t *testing.T) {
	tests := []struct {
		name        string
		config      *PhonidConfig
		wantErr     bool
		errContains string
	}{
		{
			name: "actually good pattern - very small capacity",
			config: &PhonidConfig{
				Patterns: []string{"CVCVC"}, // 5 chars
				Placeholders: map[PlaceholderType]RuneSet{
					Consonant: RuneSet("bptk"), // 4 chars
					Vowel:     RuneSet("aeo"),  // 3 chars
				},
				// Capacity: 4^3 × 3^2 = 64 × 9 = 576
				// Bit width: will be rounded to 16 (2^16 = 65,536)
				// Gap: 64,960 (99.1% - still bad due to rounding!)
				Shuffle: &ShuffleConfig{
					Rounds: 3,
					Seed:   12345,
				},
			},
			wantErr:     true,
			errContains: "pattern capacity",
		},
		{
			name: "problematic pattern - high gap ratio",
			config: &PhonidConfig{
				Patterns: []string{"CXVCVCVCCVCXX"}, // 13 chars
				Placeholders: map[PlaceholderType]RuneSet{
					Consonant: RuneSet("mbkftsn"), // 7 chars
					Vowel:     RuneSet("eua"),     // 3 chars
					CustomX:   RuneSet("o"),       // 1 char
				},
				// Capacity: 7×1×3×7×3×7×3×7×7×3×7×1×1 = 9,529,569
				// Bit width: rounds to 32 (2^32 = 4,294,967,296)
				// Gap: 4,285,437,727 (99.8% - very bad!)
				Shuffle: &ShuffleConfig{
					Rounds: 3,
					Seed:   12345,
				},
			},
			wantErr:     true,
			errContains: "pattern capacity",
		},
		{
			name: "no shuffle - should pass",
			config: &PhonidConfig{
				Patterns: []string{"CXVCVCVCCVCXX"},
				Placeholders: map[PlaceholderType]RuneSet{
					Consonant: RuneSet("mbkftsn"),
					Vowel:     RuneSet("eua"),
					CustomX:   RuneSet("o"),
				},
				Shuffle: &ShuffleConfig{
					Rounds: 0, // No shuffling
				},
			},
			wantErr: false,
		},
		{
			name: "nil shuffle config - should pass",
			config: &PhonidConfig{
				Patterns: []string{"CVCVC"},
				Placeholders: map[PlaceholderType]RuneSet{
					Consonant: RuneSet("bdf"),
					Vowel:     RuneSet("aei"),
				},
				Shuffle: nil,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Generate preflight checks
			encoder, err := NewPhoneticEncoderLenient(tt.config)
			if err != nil {
				t.Fatalf("Failed to create encoder: %v", err)
			}

			// Generate suggestions for this config
			suggestions, err := preflight.GenerateSuggestions(encoder)
			if err != nil {
				t.Fatalf("Failed to generate suggestions: %v", err)
			}

			// Convert suggestions to PreflightCheck
			checks := make([]PreflightCheck, len(suggestions))
			for i, s := range suggestions {
				checks[i] = PreflightCheck{
					Input:  s.Input,
					Output: s.Output,
				}
			}

			// Validate with preflight checks (includes shuffle validation)
			err = encoder.ValidatePreflight(checks)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ValidatePreflight() expected error but got none")
				} else if tt.errContains != "" && !contains(err.Error(), tt.errContains) {
					t.Errorf("ValidatePreflight() error = %v, want error containing %q", err, tt.errContains)
				}
			} else {
				if err != nil {
					t.Errorf("ValidatePreflight() unexpected error = %v", err)
				}
			}
		})
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) &&
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
			containsHelper(s, substr)))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
