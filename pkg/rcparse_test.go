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
