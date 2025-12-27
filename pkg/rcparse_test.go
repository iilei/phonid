package phonid

import (
	"testing"
)

// Note: Sibilant, Fricative, and Nasal can be customized by users
// to include IPA symbols (ʃ,ʒ,θ,ð,ŋ) for more precise phonetic representation
func TestParsePhonidRCEscapeChars(t *testing.T) {
	config := `

patterns = [
"CVC",
"CVCVC",
"CVCVCVC",
"CVCVCVCVCVC"
]

[placeholders]
C = "bcdfghjkpqstvwxz"
L = "lmnr"
V = "aeiou"
# TOML supports Unicode escape sequences - useful for IPA symbols!
S = "\u0283\u0292"  # ʃʒ (sh, zh sounds)
F = "\u03B8\u00F0"  # θð (th sounds: voiceless, voiced)

# Output of 'phonid preflight --suggest'
# Capacity per word: 62,500 combinations (0-62,499)
#
# Suggested preflight checks:
#
# [[preflight]]
# input = 0           # Lower boundary
# expect = "babab"
#
# [[preflight]]
# input = 31249       # Mid-range (single word)
# expect = "kuduk"
#
# [[preflight]]
# input = 62499       # Upper boundary (single word)
# expect = "zuzuz"
#
# [[preflight]]
# input = 62500       # Multi-word encoding begins
# expect = "babab cabab"
#
# [[preflight]]
# input = 624999      # Larger multi-word example
# expect = "bavab babab"
`

	got, err := ParsePhonidRC(config)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}

	if got.Pattern != "CLVCV" {
		t.Errorf("Pattern = %v, want CLVCV", got.Pattern)
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
