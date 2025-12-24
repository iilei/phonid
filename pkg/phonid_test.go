package phonid

import (
	"testing"
)

func TestPhoneticConfigValidate_Defaults(t *testing.T) {
	pc := &PhonidConfig{
		Pattern: "CLVCV",
	}

	err := pc.Validate(Base36)
	if err != nil {
		t.Errorf("expected default config to be valid, got error: %v", err)
	}
}

func TestPhoneticConfigValidate_EmptyPattern(t *testing.T) {
	pc := &PhonidConfig{
		Pattern: "",
	}

	err := pc.Validate(Base36)
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
		{"length 3", "CVC", true},
		{"length 5", "CVCVC", false},
		{"length 6", "CVVCVC", true},
		{"length 7", "CVCLVCV", false},
		{"length 11", "CVCVCVCVCVC", false},
		{"length 12", "CVCVCVCVCVCV", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pc := &PhonidConfig{
				Pattern: tt.pattern,
			}
			err := pc.Validate(Base36)
			if (err != nil) != tt.wantErr {
				t.Errorf("pattern %s: wantErr=%v, got error=%v", tt.pattern, tt.wantErr, err)
			}
		})
	}
}

func TestPhoneticConfigValidate_UndefinedPlaceholder(t *testing.T) {
	pc := &PhonidConfig{
		Pattern: "XVCVC",
		Placeholders: map[PlaceholderType][]rune{
			Vowel:     {'a', 'e', 'i'},
			Consonant: {'b', 'd', 'k'},
		},
	}

	err := pc.Validate(Base36)
	if err == nil {
		t.Error("expected error for undefined placeholder 'X'")
	}
}

func TestPhoneticConfigValidate_MinimumCharacters(t *testing.T) {
	pc := &PhonidConfig{
		Pattern: "CVCVC",
		Placeholders: map[PlaceholderType][]rune{
			Vowel:     {'a'},      // Only 1 vowel, need at least 2
			Consonant: {'b', 'd'}, // 2 consonants is OK
		},
	}

	err := pc.Validate(Base36)
	if err == nil {
		t.Error("expected error for insufficient vowels")
	}
}

func TestPhoneticConfigValidate_DuplicateCharacters(t *testing.T) {
	pc := &PhonidConfig{
		Pattern: "CVCVC",
		Placeholders: map[PlaceholderType][]rune{
			Vowel:     {'a', 'e', 'a'}, // Duplicate 'a'
			Consonant: {'b', 'd', 'k'},
		},
	}

	err := pc.Validate(Base36)
	if err == nil {
		t.Error("expected error for duplicate characters in placeholder")
	}
}

func TestPhoneticConfigValidate_OverlappingPlaceholders(t *testing.T) {
	pc := &PhonidConfig{
		Pattern: "CLVCV",
		Placeholders: map[PlaceholderType][]rune{
			Vowel:     {'a', 'e', 'i'},
			Consonant: {'b', 'd', 'k', 'l'}, // 'l' overlaps with L
			Liquid:    {'l', 'm', 'n'},
		},
	}

	err := pc.Validate(Base36)
	if err == nil {
		t.Error("expected error for overlapping placeholders")
	}
}

func TestPhoneticConfigValidate_NoVowelPlaceholder(t *testing.T) {
	pc := &PhonidConfig{
		Pattern: "CLCCC",
		Placeholders: map[PlaceholderType][]rune{
			Consonant: {'b', 'd', 'k', 't'},
			Liquid:    {'l', 'm', 'n'},
		},
	}

	err := pc.Validate(Base36)
	if err == nil {
		t.Error("expected error for pattern with no vowel placeholder")
	}
}

func TestPhoneticConfigValidate_InsufficientCombinations(t *testing.T) {
	// With only 2 consonants and 2 vowels in CVCVC pattern:
	// 2^3 * 2^2 = 8 * 4 = 32 combinations (less than 36 needed for Base36)
	pc := &PhonidConfig{
		Pattern: "CVCVC",
		Placeholders: map[PlaceholderType][]rune{
			Vowel:     {'a', 'e'},
			Consonant: {'b', 'd'},
		},
	}

	err := pc.Validate(Base36)
	if err == nil {
		t.Error("expected error for insufficient combinations")
	}
}

func TestPhoneticConfigValidate_SufficientCombinations(t *testing.T) {
	// With 3 consonants and 2 vowels in CVCVC pattern:
	// 3^3 * 2^2 = 27 * 4 = 108 combinations (more than 36 needed)
	pc := &PhonidConfig{
		Pattern: "CVCVC",
		Placeholders: map[PlaceholderType][]rune{
			Vowel:     {'a', 'e'},
			Consonant: {'b', 'd', 'k'},
		},
	}

	err := pc.Validate(Base36)
	if err != nil {
		t.Errorf("expected valid config, got error: %v", err)
	}
}

func TestPhoneticConfigValidate_UnusedPlaceholders(t *testing.T) {
	// Should not validate unused placeholders
	pc := &PhonidConfig{
		Pattern: "CVCVC",
		Placeholders: map[PlaceholderType][]rune{
			Vowel:     {'a', 'e', 'i'},
			Consonant: {'b', 'd', 'k'},
			Liquid:    {'z'}, // Only 1 char but not used in pattern - should be OK
		},
	}

	err := pc.Validate(Base36)
	if err != nil {
		t.Errorf("expected unused placeholders to be ignored, got error: %v", err)
	}
}

func TestPhoneticConfigValidate_LiquidsPattern(t *testing.T) {
	pc := &PhonidConfig{
		Pattern: "CLVCLVC",
		Placeholders: map[PlaceholderType][]rune{
			Consonant: {'b', 'd', 'k', 't'},
			Liquid:    {'l', 'r'},
			Vowel:     {'a', 'e', 'i'},
		},
	}

	err := pc.Validate(Base36)
	if err != nil {
		t.Errorf("expected liquids pattern to work, got error: %v", err)
	}
}

func TestHasDuplicates(t *testing.T) {
	tests := []struct {
		name  string
		runes []rune
		want  bool
	}{
		{"no duplicates", []rune{'a', 'b', 'c'}, false},
		{"has duplicate", []rune{'a', 'b', 'a'}, true},
		{"empty", []rune{}, false},
		{"single", []rune{'a'}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := hasDuplicates(tt.runes)
			if got != tt.want {
				t.Errorf("hasDuplicates(%v) = %v, want %v", tt.runes, got, tt.want)
			}
		})
	}
}

func TestHasOverlap(t *testing.T) {
	tests := []struct {
		name string
		a    []rune
		b    []rune
		want bool
	}{
		{"no overlap", []rune{'a', 'b'}, []rune{'c', 'd'}, false},
		{"has overlap", []rune{'a', 'b'}, []rune{'b', 'c'}, true},
		{"empty a", []rune{}, []rune{'a', 'b'}, false},
		{"empty b", []rune{'a', 'b'}, []rune{}, false},
		{"both empty", []rune{}, []rune{}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := hasOverlap(tt.a, tt.b)
			if got != tt.want {
				t.Errorf("hasOverlap(%v, %v) = %v, want %v", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

func TestIsAllowedLength(t *testing.T) {
	tests := []struct {
		length int
		want   bool
	}{
		{3, false},
		{5, true},
		{6, false},
		{7, true},
		{11, true},
		{12, false},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			got := isAllowedLength(tt.length)
			if got != tt.want {
				t.Errorf("isAllowedLength(%d) = %v, want %v", tt.length, got, tt.want)
			}
		})
	}
}

func TestIsVowelBase(t *testing.T) {
	tests := []struct {
		name string
		char rune
		want bool
	}{
		// Basic vowels
		{"lowercase a", 'a', true},
		{"lowercase e", 'e', true},
		{"lowercase i", 'i', true},
		{"lowercase o", 'o', true},
		{"lowercase u", 'u', true},
		{"lowercase y", 'y', true},
		{"uppercase A", 'A', true},
		{"uppercase E", 'E', true},

		// German umlauts
		{"lowercase a-umlaut", '\u00E4', true},
		{"lowercase o-umlaut", '\u00F6', true},
		{"lowercase u-umlaut", '\u00FC', true},
		{"uppercase A-umlaut", '\u00C4', true},
		{"uppercase O-umlaut", '\u00D6', true},
		{"uppercase U-umlaut", '\u00DC', true},

		// French accents
		{"e-acute", '\u00E9', true},
		{"e-grave", '\u00E8', true},
		{"e-circumflex", '\u00EA', true},
		{"e-diaeresis", '\u00EB', true},

		// Other vowels with diacritics
		{"a-acute", '\u00E1', true},
		{"i-acute", '\u00ED', true},
		{"o-acute", '\u00F3', true},
		{"u-acute", '\u00FA', true},
		{"n-tilde", '\u00F1', false}, // Not a vowel

		// Consonants
		{"b", 'b', false},
		{"k", 'k', false},
		{"z", 'z', false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isVowelBase(tt.char)
			if got != tt.want {
				t.Errorf("isVowelBase('%c') = %v, want %v", tt.char, got, tt.want)
			}
		})
	}
}

func TestPhoneticConfigValidate_VowelsWithDiacritics(t *testing.T) {
	pc := &PhonidConfig{
		Pattern: "CVCVC",
		Placeholders: map[PlaceholderType][]rune{
			Vowel:     {'\u00E4', '\u00F6', '\u00FC'}, // German umlauts (a-umlaut, o-umlaut, u-umlaut)
			Consonant: {'b', 'd', 'k'},
		},
	}

	err := pc.Validate(Base36)
	if err != nil {
		t.Errorf("expected vowels with diacritics to be valid, got error: %v", err)
	}
}

func TestPhoneticConfigValidate_MixedVowels(t *testing.T) {
	pc := &PhonidConfig{
		Pattern: "CVCVC",
		Placeholders: map[PlaceholderType][]rune{
			Vowel:     {'a', '\u00E9', '\u00F6'}, // Mix of plain and diacritics (a, e-acute, o-umlaut)
			Consonant: {'b', 'd', 'k'},
		},
	}

	err := pc.Validate(Base36)
	if err != nil {
		t.Errorf("expected mixed vowels to be valid, got error: %v", err)
	}
}

func TestPhoneticConfigValidate_InvalidVowelWithDiacritic(t *testing.T) {
	pc := &PhonidConfig{
		Pattern: "CVCVC",
		Placeholders: map[PlaceholderType][]rune{
			Vowel:     {'a', 'e', '\u00F1'}, // n-tilde is not a vowel
			Consonant: {'b', 'd', 'k'},
		},
	}

	err := pc.Validate(Base36)
	if err == nil {
		t.Error("expected error for invalid vowel 'n-tilde'")
	}
}
