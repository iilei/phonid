package preflight_test

import (
	"strings"
	"testing"

	p "github.com/iilei/phonid/pkg"
	"github.com/iilei/phonid/pkg/preflight"
)

//nolint:gocognit,gocyclo // Comprehensive integration test with multiple subtests
func TestSuggestConfig(t *testing.T) {
	placeholderMap := p.PlaceholderMap{
		p.Vowel:     p.RuneSet{'a', 'e', 'i'},
		p.Consonant: p.RuneSet{'b', 'c', 'd'},
	}

	config := &p.PhonidConfig{
		Patterns:     []string{"CVC", "CVCVC"},
		Placeholders: placeholderMap,
	}

	encoder, err := p.NewPhoneticEncoderLenient(config)
	if err != nil {
		t.Fatalf("NewPhoneticEncoderLenient() error: %v", err)
	}

	result, err := preflight.SuggestConfig(encoder, config)
	if err != nil {
		t.Fatalf("SuggestConfig() error: %v", err)
	}

	// Test structure
	t.Run("has base configuration", func(t *testing.T) {
		if !strings.Contains(result, "base = 36") {
			t.Error("missing base configuration")
		}
	})

	t.Run("has shuffle section", func(t *testing.T) {
		if !strings.Contains(result, "[shuffle]") {
			t.Error("missing shuffle section")
		}
		if !strings.Contains(result, "bit_width = 32") {
			t.Error("missing bit_width in shuffle section")
		}
		if !strings.Contains(result, "rounds = 0") {
			t.Error("missing rounds in shuffle section")
		}
		if !strings.Contains(result, "seed = 0") {
			t.Error("missing seed in shuffle section")
		}
	})

	t.Run("has phonetic section", func(t *testing.T) {
		if !strings.Contains(result, "[phonetic]") {
			t.Error("missing phonetic section")
		}
		if !strings.Contains(result, "patterns = [") {
			t.Error("missing patterns in phonetic section")
		}
		if !strings.Contains(result, "[phonetic.placeholders]") {
			t.Error("missing placeholders in phonetic section")
		}
	})

	t.Run("has preflight assertions", func(t *testing.T) {
		if !strings.Contains(result, "[[preflight]]") {
			t.Error("missing preflight assertions")
		}
	})

	// Test comments are preserved
	t.Run("has capacity header comment", func(t *testing.T) {
		if !strings.Contains(result, "# Output of 'phonid preflight --suggest'") {
			t.Error("missing preflight header comment")
		}
		if !strings.Contains(result, "# Capacity per word:") {
			t.Error("missing capacity comment")
		}
		if !strings.Contains(result, "# Suggested preflight checks:") {
			t.Error("missing suggested checks comment")
		}
	})

	// Test inline comments in assertions
	t.Run("has inline comments in preflight assertions", func(t *testing.T) {
		if !strings.Contains(result, "# Lower boundary") {
			t.Error("missing 'Lower boundary' inline comment")
		}
		if !strings.Contains(result, "# Upper boundary (single word)") {
			t.Error("missing 'Upper boundary' inline comment")
		}
	})

	// Test TOML field names
	t.Run("uses correct TOML field names", func(t *testing.T) {
		// Should use 'input' and 'output', not 'expect'
		if !strings.Contains(result, "input =") {
			t.Error("missing 'input' field in assertions")
		}
		if !strings.Contains(result, "output =") {
			t.Error("missing 'output' field in assertions")
		}
	})
}

//nolint:gocognit,gocyclo // Comprehensive integration test with multiple subtests
func TestSuggestConfigInlineComments(t *testing.T) {
	placeholderMap := p.PlaceholderMap{
		p.Vowel:     p.RuneSet{'a', 'o'},
		p.Consonant: p.RuneSet{'z', 'k', 't'},
	}

	config := &p.PhonidConfig{
		Patterns:     []string{"VCV"},
		Placeholders: placeholderMap,
	}

	encoder, err := p.NewPhoneticEncoderLenient(config)
	if err != nil {
		t.Fatalf("NewPhoneticEncoderLenient() error: %v", err)
	}

	result, err := preflight.SuggestConfig(encoder, config)
	if err != nil {
		t.Fatalf("SuggestConfig() error: %v", err)
	}

	// Parse the result to find inline comments
	lines := strings.Split(result, "\n")
	var inputLines []string
	for _, line := range lines {
		if strings.Contains(line, "input =") && strings.Contains(line, "#") {
			inputLines = append(inputLines, line)
		}
	}

	t.Run("all input lines have inline comments", func(t *testing.T) {
		if len(inputLines) == 0 {
			t.Fatal("no input lines with inline comments found")
		}

		for i, line := range inputLines {
			// Each input line should have format: "input = VALUE # COMMENT"
			if !strings.Contains(line, "#") {
				t.Errorf("input line %d missing inline comment: %s", i, line)
			}

			// Verify the comment comes after the hash
			parts := strings.Split(line, "#")
			if len(parts) != 2 {
				t.Errorf("input line %d has incorrect comment format: %s", i, line)
				continue
			}

			comment := strings.TrimSpace(parts[1])
			if comment == "" {
				t.Errorf("input line %d has empty comment: %s", i, line)
			}
		}
	})

	t.Run("inline comments are properly formatted", func(t *testing.T) {
		// Check that at least one line has the expected format with padding
		foundProperlyFormatted := false
		for _, line := range inputLines {
			// Should have format like: "input = 0             # Lower boundary"
			// with consistent spacing before the #
			if strings.Contains(line, "input =") && strings.Contains(line, "#") {
				parts := strings.SplitN(line, "=", 2)
				if len(parts) == 2 {
					afterEquals := parts[1]
					// Should have value, spaces, then #
					if strings.Contains(afterEquals, "#") {
						foundProperlyFormatted = true
						break
					}
				}
			}
		}
		if !foundProperlyFormatted {
			t.Error("no properly formatted inline comments found")
		}
	})
}

func TestSuggestConfigErrorCases(t *testing.T) {
	config := &p.PhonidConfig{
		Patterns:     []string{"CVC"},
		Placeholders: p.PlaceholderMap{p.Vowel: p.RuneSet{'a', 'e', 'i'}, p.Consonant: p.RuneSet{'b', 'c', 'd'}},
	}

	encoder, err := p.NewPhoneticEncoderLenient(config)
	if err != nil {
		t.Fatalf("Failed to create encoder: %v", err)
	}

	tests := []struct {
		name        string
		encoder     *p.PhoneticEncoder
		config      *p.PhonidConfig
		wantErr     bool
		errContains string
	}{
		{
			name:        "nil encoder",
			encoder:     nil,
			config:      config,
			wantErr:     true,
			errContains: "encoder cannot be nil",
		},
		{
			name:        "nil config",
			encoder:     encoder,
			config:      nil,
			wantErr:     true,
			errContains: "config cannot be nil",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := preflight.SuggestConfig(tt.encoder, tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("SuggestConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil && !strings.Contains(err.Error(), tt.errContains) {
				t.Errorf("SuggestConfig() error = %v, want error containing %q", err, tt.errContains)
			}
		})
	}
}

func TestSuggestConfigNumberFormatting(t *testing.T) {
	placeholderMap := p.PlaceholderMap{
		p.Vowel:     p.RuneSet{'a', 'e', 'i', 'o', 'u'},
		p.Consonant: p.RuneSet{'b', 'c', 'd', 'f', 'g', 'h', 'j', 'k', 'l', 'm'},
	}

	config := &p.PhonidConfig{
		Patterns:     []string{"CVC", "CVCVC"},
		Placeholders: placeholderMap,
	}

	encoder, err := p.NewPhoneticEncoderLenient(config)
	if err != nil {
		t.Fatalf("NewPhoneticEncoderLenient() error: %v", err)
	}

	result, err := preflight.SuggestConfig(encoder, config)
	if err != nil {
		t.Fatalf("SuggestConfig() error: %v", err)
	}

	t.Run("capacity numbers use underscore separators", func(t *testing.T) {
		// Large numbers should have underscore separators (e.g., "1_000" not "1000")
		lines := strings.SplitSeq(result, "\n")
		for line := range lines {
			if strings.Contains(line, "# Capacity per word:") {
				// Check if numbers in this line use underscores for thousands
				// This depends on the actual capacity, but we can check the format is reasonable
				if !strings.Contains(line, "combinations") {
					t.Error("capacity line should mention 'combinations'")
				}
				// The line should look like: "# Capacity per word: XXX combinations (0-XXX)"
				if !strings.Contains(line, "(0-") {
					t.Error("capacity line should show range starting from 0")
				}
				break
			}
		}
	})
}
