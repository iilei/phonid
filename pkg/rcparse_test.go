package phonid

import (
	"slices"
	"testing"
)

var config = `
base = 36

[shuffle]
bit_width = 32
rounds    = 0
seed      = 0

[phonetic]
patterns = ["CVC", "CVCVC", "CVCVCVC", "CVCVCVCVCVC"]

[phonetic.placeholders]
C = "bcdfghjkpqstvwxz"
L = "lmnr"
V = "aeiou"
# # TOML supports Unicode escape sequences - useful for IPA symbols!
S = "\u0283\u0292" # ʃʒ (sh, zh sounds)
F = "\u03B8\u00F0" # θð (th sounds: voiceless, voiced)

# # Output of 'phonid preflight --suggest'
# # Capacity per word: 62_500 combinations (0-62_499)
# #
# # Suggested preflight checks:
#
# [[preflight]]
# input = 0            # Lower boundary
# expect = "babab"
#
# [[preflight]]
# input = 31_249       # Mid-range (single word)
# expect = "kuduk"
#
# [[preflight]]
# input = 62_499       # Upper boundary (single word)
# expect = "zuzuz"
#
# [[preflight]]
# input = 62_500       # Multi-word encoding begins
# expect = "babab cabab"
#
# [[preflight]]
# input = 624_999      # Larger multi-word example
# expect = "bavab babab"
`

// Note: Sibilant, Fricative, and Nasal can be customized by users
// to include IPA symbols (ʃ,ʒ,θ,ð,ŋ) for more precise phonetic representation.
func TestParsePhonidRCTOMLConfig(t *testing.T) {
	fullConfig := `
base = 36

[shuffle]
bit_width = 32
rounds = 3
seed = 12345

[phonetic]
patterns = ["CVC", "CVCVC"]

[phonetic.placeholders]
C = "bcdfg"
V = "aei"

[[preflight]]
input = 0
output = "babab"

[[preflight]]
input = 100
output = "kodak"
`

	got, preflight, err := ParsePhonidRC(fullConfig)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}

	// Test Base field (note: base is not yet used in PhonidConfig struct, so we can't verify it directly)
	// This test ensures it parses without error

	// Test Shuffle configuration (note: shuffle is not yet in PhonidConfig, but parsing should succeed)
	// The TOML parser validates these fields exist and are properly formatted

	// Test Phonetic patterns
	wantPatterns := []string{"CVC", "CVCVC"}
	if !slices.Equal(got.Patterns, wantPatterns) {
		t.Errorf("Patterns = %v, want %v", got.Patterns, wantPatterns)
	}

	// Test Phonetic placeholders
	if string(got.Placeholders[Consonant]) != "bcdfg" {
		t.Errorf("Consonant = %v, want bcdfg", string(got.Placeholders[Consonant]))
	}
	if string(got.Placeholders[Vowel]) != "aei" {
		t.Errorf("Vowel = %v, want aei", string(got.Placeholders[Vowel]))
	}

	// Test Preflight checks
	if len(preflight) != 2 {
		t.Errorf("got %d preflight checks, want 2", len(preflight))
		return
	}

	if preflight[0].Input != 0 || preflight[0].Output != "babab" {
		t.Errorf("preflight[0] = {Input: %d, Output: %s}, want {Input: 0, Output: babab}",
			preflight[0].Input, preflight[0].Output)
	}

	if preflight[1].Input != 100 || preflight[1].Output != "kodak" {
		t.Errorf("preflight[1] = {Input: %d, Output: %s}, want {Input: 100, Output: kodak}",
			preflight[1].Input, preflight[1].Output)
	}
}

func TestParsePhonidRCEscapeChars(t *testing.T) {
	got, _, err := ParsePhonidRCLenient(config)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}

	// ...existing code...
	want := []string{"CVC", "CVCVC", "CVCVCVC", "CVCVCVCVCVC"}

	if !slices.Equal(got.Patterns, want) {
		t.Errorf("Pattern = %v, want %v", got.Patterns, want)
	}

	// Check placeholders were parsed correctly
	if string(got.Placeholders[Consonant]) != "bcdfghjkpqstvwxz" {
		t.Errorf("Consonant = %v, want bcdfghjkpqstvwxz", string(got.Placeholders[Consonant]))
	}
	if string(got.Placeholders[Liquid]) != "lmnr" {
		t.Errorf("Liquid = %v, want lmnr", string(got.Placeholders[Liquid]))
	}
	if string(got.Placeholders[Vowel]) != "aeiou" {
		t.Errorf("Vowel = %v, want aeiou", string(got.Placeholders[Vowel]))
	}
	// Check Unicode escapes were properly decoded
	if string(got.Placeholders[Sibilant]) != "\u0283\u0292" {
		t.Errorf("Sibilant = %v, want \u0283\u0292", string(got.Placeholders[Sibilant]))
	}
	if string(got.Placeholders[Fricative]) != "\u03B8\u00F0" {
		t.Errorf("Fricative = %v, want \u03B8\u00F0", string(got.Placeholders[Fricative]))
	}
}

func TestIsValidPhonidRCFilename(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		want     bool
	}{
		// Valid: exact matches
		{
			name:     "exact .phonidrc",
			filename: ".phonidrc",
			want:     true,
		},
		{
			name:     "exact .phonidrc.toml",
			filename: ".phonidrc.toml",
			want:     true,
		},

		// Valid: prefixed patterns
		{
			name:     "prefixed .dev.phonidrc",
			filename: ".dev.phonidrc",
			want:     true,
		},
		{
			name:     "prefixed .prod.phonidrc",
			filename: ".prod.phonidrc",
			want:     true,
		},
		{
			name:     "prefixed .test.phonidrc.toml",
			filename: ".test.phonidrc.toml",
			want:     true,
		},
		{
			name:     "prefixed .staging.phonidrc.toml",
			filename: ".staging.phonidrc.toml",
			want:     true,
		},

		// Invalid: no leading dot
		{
			name:     "no leading dot",
			filename: "phonidrc",
			want:     false,
		},
		{
			name:     "no leading dot with toml",
			filename: "phonidrc.toml",
			want:     false,
		},

		// Invalid: doesn't end with .phonidrc
		{
			name:     "wrong suffix",
			filename: ".phonid",
			want:     false,
		},
		{
			name:     "wrong suffix with toml",
			filename: ".phonid.toml",
			want:     false,
		},

		// Invalid: empty prefix (double dot)
		{
			name:     "empty prefix",
			filename: "..phonidrc",
			want:     false,
		},
		{
			name:     "empty prefix with toml",
			filename: "..phonidrc.toml",
			want:     false,
		},

		// Invalid: prefix contains dot
		{
			name:     "prefix with dot",
			filename: ".my.config.phonidrc",
			want:     false,
		},
		{
			name:     "prefix with multiple dots",
			filename: ".a.b.c.phonidrc",
			want:     false,
		},

		// Invalid: prefix contains path separators
		{
			name:     "prefix with forward slash",
			filename: ".my/path.phonidrc",
			want:     false,
		},
		{
			name:     "prefix with backslash",
			filename: ".my\\path.phonidrc",
			want:     false,
		},

		// Invalid: prefix contains colon
		{
			name:     "prefix with colon",
			filename: ".my:config.phonidrc",
			want:     false,
		},

		// Invalid: wrong extension
		{
			name:     "wrong extension .txt",
			filename: ".phonidrc.txt",
			want:     false,
		},
		{
			name:     "wrong extension .yaml",
			filename: ".phonidrc.yaml",
			want:     false,
		},
		{
			name:     "prefixed with wrong extension",
			filename: ".dev.phonidrc.json",
			want:     false,
		},

		// Invalid: various malformed patterns
		{
			name:     "only dot prefix",
			filename: ".phonidrc.",
			want:     false,
		},
		{
			name:     "random file",
			filename: "config.toml",
			want:     false,
		},
		{
			name:     "contains phonidrc but wrong pattern",
			filename: "my.phonidrc.conf",
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isValidPhonidRCFilename(tt.filename)
			if got != tt.want {
				t.Errorf("isValidPhonidRCFilename(%q) = %v, want %v", tt.filename, got, tt.want)
			}
		})
	}
}
