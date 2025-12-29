package phonid_test

import (
	"testing"

	. "github.com/iilei/phonid/pkg"
)

func TestPhoneticConfigValidate_Defaults(t *testing.T) {
	pc := &PhonidConfig{
		Patterns: []string{"CLVCV"},
	}

	err := pc.Validate()
	if err != nil {
		t.Errorf("expected default config to be valid, got error: %v", err)
	}
}

func TestPhoneticConfigValidate_EmptyPattern(t *testing.T) {
	pc := &PhonidConfig{
		Patterns: []string{""},
	}

	err := pc.Validate()
	if err == nil {
		t.Error("expected error for empty pattern")
	}
}

func TestPhoneticConfigValidate_InvalidPatternLength(t *testing.T) {
	tests := []struct {
		name    string
		pattern string
		wantErr bool
	}{
		{"length 5", "CVCVC", false},
		{"length 6", "CVVCVC", true},
		{"length 11", "CVCVCVCVCVC", false},
		{"length 12", "CVCVCVCVCVCV", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pc := &PhonidConfig{
				Patterns: []string{tt.pattern},
			}
			err := pc.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("pattern %s: wantErr=%v, got error=%v", tt.pattern, tt.wantErr, err)
			}
		})
	}
}

func TestPhoneticConfigValidate_UndefinedPlaceholder(t *testing.T) {
	pc := &PhonidConfig{
		Patterns: []string{"XVCVC"},
		Placeholders: map[PlaceholderType]RuneSet{
			Vowel:     RuneSet("aei"),
			Consonant: RuneSet("bdk"),
		},
	}

	err := pc.Validate()
	if err == nil {
		t.Error("expected error for undefined placeholder 'X'")
	}
}

func TestPhoneticConfigValidate_MinimumCharacters(t *testing.T) {
	pc := &PhonidConfig{
		Patterns: []string{"CVCVC"},
		Placeholders: map[PlaceholderType]RuneSet{
			Vowel:     RuneSet("a"),  // Only 1 vowel, need at least 2
			Consonant: RuneSet("bd"), // 2 consonants is OK
		},
	}

	err := pc.Validate()
	if err == nil {
		t.Error("expected error for insufficient vowels")
	}
}

func TestPhoneticConfigValidate_DuplicateCharacters(t *testing.T) {
	pc := &PhonidConfig{
		Patterns: []string{"CVCVC"},
		Placeholders: map[PlaceholderType]RuneSet{
			Vowel:     RuneSet("aea"), // Duplicate 'a'
			Consonant: RuneSet("bdk"),
		},
	}

	err := pc.Validate()
	if err == nil {
		t.Error("expected error for duplicate characters in placeholder")
	}
}

func TestPhoneticConfigValidate_OverlappingPlaceholders(t *testing.T) {
	pc := &PhonidConfig{
		Patterns: []string{"CLVCV"},
		Placeholders: map[PlaceholderType]RuneSet{
			Vowel:     RuneSet("aei"),
			Consonant: RuneSet("bdkl"), // 'l' overlaps with L
			Liquid:    RuneSet("lmn"),
		},
	}

	err := pc.Validate()
	if err == nil {
		t.Error("expected error for overlapping placeholders")
	}
}

func TestPhoneticConfigValidate_NoVowelPlaceholder(t *testing.T) {
	pc := &PhonidConfig{
		Patterns: []string{"CLCCC"},
		Placeholders: map[PlaceholderType]RuneSet{
			Consonant: RuneSet("bdkt"),
			Liquid:    RuneSet("lmn"),
		},
	}

	err := pc.Validate()
	if err == nil {
		t.Error("expected error for pattern with no vowel placeholder")
	}
}

func TestPhoneticConfigValidate_InsufficientCombinations(t *testing.T) {
	pc := &PhonidConfig{
		Patterns: []string{"CVCVC"},
		Placeholders: map[PlaceholderType]RuneSet{
			Vowel:     RuneSet("ae"),
			Consonant: RuneSet("bd"),
		},
	}

	err := pc.Validate()
	if err == nil {
		t.Error("expected error for insufficient combinations")
	}
}

func TestPhoneticConfigValidate_SufficientCombinations(t *testing.T) {
	// With 3 consonants and 2 vowels in CVCVC pattern:
	// 3^3 * 2^2 = 27 * 4 = 108 combinations (more than 36 needed)
	pc := &PhonidConfig{
		Patterns: []string{"CVCVC"},
		Placeholders: map[PlaceholderType]RuneSet{
			Vowel:     RuneSet("ae"),
			Consonant: RuneSet("bdk"),
		},
	}

	err := pc.Validate()
	if err != nil {
		t.Errorf("expected valid config, got error: %v", err)
	}
}

func TestPhoneticConfigValidate_UnusedPlaceholders(t *testing.T) {
	// Should not validate unused placeholders
	pc := &PhonidConfig{
		Patterns: []string{"CVCVC"},
		Placeholders: map[PlaceholderType]RuneSet{
			Vowel:     RuneSet("aei"),
			Consonant: RuneSet("bdk"),
			Liquid:    RuneSet("z"), // Only 1 char but not used in pattern - should be OK
		},
	}

	err := pc.Validate()
	if err != nil {
		t.Errorf("expected unused placeholders to be ignored, got error: %v", err)
	}
}

func TestPhoneticConfigValidate_MisconfiguredComplement(t *testing.T) {
	// Should not validate unused placeholders
	pc := &PhonidConfig{
		Patterns: []string{"VCVCV"},
		Placeholders: map[PlaceholderType]RuneSet{
			Vowel:     RuneSet("aei"),
			Consonant: RuneSet("b"),
			Fricative: RuneSet("fgt"), // minimum length but no pattern hit
		},
	}

	err := pc.Validate()
	if err == nil {
		t.Error("expected error for misconfigured complement")
	}
}

func TestPhoneticConfigValidate_LiquidsPattern(t *testing.T) {
	pc := &PhonidConfig{
		Patterns: []string{"CLVCLVC"},
		Placeholders: map[PlaceholderType]RuneSet{
			Consonant: RuneSet("bdkt"),
			Liquid:    RuneSet("lr"),
			Vowel:     RuneSet("aei"),
		},
	}

	err := pc.Validate()
	if err != nil {
		t.Errorf("expected liquids pattern to work, got error: %v", err)
	}
}

func TestValidate_DuplicateDetection(t *testing.T) {
	tests := []struct {
		name        string
		placeholder RuneSet
		wantErr     bool
	}{
		{"no duplicates", RuneSet("bdk"), false},
		{"has duplicate 'b'", RuneSet("bdb"), true},
		{"has duplicate 'z'", RuneSet("xyzz"), true},
		{"has duplicate 'k'", RuneSet("bdktk"), true},
		{"empty causes min-size error", RuneSet(""), true},
		{"single causes min-size error", RuneSet("b"), true},
		{"multiple duplicates", RuneSet("bbddkk"), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pc := &PhonidConfig{
				Patterns: []string{"CVCVC"},
				Placeholders: map[PlaceholderType]RuneSet{
					Vowel:     RuneSet("aei"),
					Consonant: tt.placeholder,
				},
			}
			err := pc.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("config with %q: wantErr=%v, got error=%v", tt.placeholder, tt.wantErr, err)
			}
		})
	}
}

func TestValidate_OverlapDetection(t *testing.T) {
	tests := []struct {
		name      string
		consonant RuneSet
		liquid    RuneSet
		wantErr   bool
	}{
		{"no overlap", RuneSet("bdk"), RuneSet("lmn"), false},
		{"has overlap 'b'", RuneSet("bdk"), RuneSet("blm"), true},
		{"has overlap 'l'", RuneSet("bdkl"), RuneSet("lmn"), true},
		{"multiple overlaps", RuneSet("bdl"), RuneSet("lmb"), true},
		{"empty liquid", RuneSet("bdk"), RuneSet(""), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pc := &PhonidConfig{
				Patterns: []string{"CLVCV"},
				Placeholders: map[PlaceholderType]RuneSet{
					Vowel:     RuneSet("aei"),
					Consonant: tt.consonant,
					Liquid:    tt.liquid,
				},
			}
			err := pc.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("config C=%q L=%q: wantErr=%v, got error=%v",
					tt.consonant, tt.liquid, tt.wantErr, err)
			}
		})
	}
}

func TestValidate_VowelValidation(t *testing.T) {
	tests := []struct {
		name    string
		vowels  RuneSet
		wantErr bool
	}{
		// Basic vowels
		{"lowercase vowels", RuneSet("aei"), false},
		{"uppercase vowels", RuneSet("AEI"), false},
		{"mixed case vowels", RuneSet("aEi"), false},
		{"vowel y", RuneSet("aey"), false},

		// German umlauts
		{"lowercase umlauts", RuneSet("äöü"), false},
		{"uppercase umlauts", RuneSet("ÄÖÜ"), false},
		{"mixed umlauts", RuneSet("aöü"), false},

		// French accents
		{"e-acute", RuneSet("éèê"), false},
		{"e-diaeresis", RuneSet("aëi"), false},

		// Other vowels with diacritics
		{"accented vowels", RuneSet("áíó"), false},
		{"mixed plain and accented", RuneSet("aéö"), false},

		// Invalid cases - consonants
		{"consonant b", RuneSet("aeb"), true},
		{"consonant k", RuneSet("aek"), true},
		{"consonant z", RuneSet("aez"), true},

		// Invalid cases - other characters
		{"n-tilde not vowel", RuneSet("aeñ"), true},
		{"digit not vowel", RuneSet("ae1"), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pc := &PhonidConfig{
				Patterns: []string{"CVCVC"},
				Placeholders: map[PlaceholderType]RuneSet{
					Vowel:     tt.vowels,
					Consonant: RuneSet("bdk"),
				},
			}
			err := pc.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("config with vowels %q: wantErr=%v, got error=%v",
					tt.vowels, tt.wantErr, err)
			}
		})
	}
}

func TestPhoneticConfigValidate_VowelsWithDiacritics(t *testing.T) {
	pc := &PhonidConfig{
		Patterns: []string{"CVCVC"},
		Placeholders: map[PlaceholderType]RuneSet{
			Vowel:     RuneSet("äöü"), // German umlauts (a-umlaut, o-umlaut, u-umlaut)
			Consonant: RuneSet("bdk"),
		},
	}

	err := pc.Validate()
	if err != nil {
		t.Errorf("expected vowels with diacritics to be valid, got error: %v", err)
	}
}

func TestPhoneticConfigValidate_MixedVowels(t *testing.T) {
	pc := &PhonidConfig{
		Patterns: []string{"CVCVC"},
		Placeholders: map[PlaceholderType]RuneSet{
			Vowel:     RuneSet("aéö"), // Mix of plain and diacritics (a, e-acute, o-umlaut)
			Consonant: RuneSet("bdk"),
		},
	}

	err := pc.Validate()
	if err != nil {
		t.Errorf("expected mixed vowels to be valid, got error: %v", err)
	}
}

func TestPhoneticConfigValidate_InvalidVowelWithDiacritic(t *testing.T) {
	pc := &PhonidConfig{
		Patterns: []string{"CVCVC"},
		Placeholders: map[PlaceholderType]RuneSet{
			Vowel:     RuneSet("aeñ"), // n-tilde is not a vowel
			Consonant: RuneSet("bdk"),
		},
	}

	err := pc.Validate()
	if err == nil {
		t.Error("expected error for invalid vowel 'n-tilde'")
	}
}
