// Package phonid generates phonetic identifiers using configurable patterns.
package phonid

import (
	"errors"
	"fmt"
	"slices"
	"strings"
	"unicode"

	"golang.org/x/text/unicode/norm"
)

const (
	// MinPatternLength is the minimum allowed pattern length.
	MinPatternLength = 3
	// MaxPatternLength is the maximum allowed pattern length.
	MaxPatternLength = 127

	// MinCharsForVowelShort is minimum vowels for short patterns (length 3-4).
	// With 3 consonants and 2 vowels: 3^2 * 2^2 = 36 combinations (length 4), 3^2 * 2^1 = 18 (length 3).
	MinCharsForVowelShort = 2
	// MinCharsForVowelMedium is minimum vowels for medium patterns (length 5-6).
	// With 3 consonants and 3 vowels: 3^3 * 3^2 = 243 combinations (length 5).
	MinCharsForVowelMedium = 3
	// MinCharsForVowelLong is minimum vowels for longer patterns (length >= 7).
	// With 3 consonants and 2 vowels: 3^4 * 2^3 = 648 combinations (> 128).
	MinCharsForVowelLong = 2
	// MinPatternLengthForMediumVowels is the pattern length that requires MinCharsForVowelMedium.
	MinPatternLengthForMediumVowels = 5
	// MinPatternLengthForLongVowels is the pattern length that requires MinCharsForVowelLong.
	MinPatternLengthForLongVowels = 7
	// MinCharsForComplement placeholder type minimal set of runes.
	MinCharsForComplement = 3 // At least one non-vowel category (C, L, N, S, or F) must have this many

	Consonant PlaceholderType = 'C'
	Vowel     PlaceholderType = 'V'
	Liquid    PlaceholderType = 'L'
	Nasal     PlaceholderType = 'N'
	Sibilant  PlaceholderType = 'S'
	Fricative PlaceholderType = 'F'
	CustomX   PlaceholderType = 'X'
	CustomY   PlaceholderType = 'Y'
	CustomZ   PlaceholderType = 'Z'

	// ProQuintPatternShort in accordance with ProQuint-compatible configuration
	// Based on the Proquint specification: https://arxiv.org/html/0901.4016
	// Provides a pre-configured encoder that generates identifiers compatible with
	// the original Proquint library, using the pattern CVCVC-CVCVC to encode 32-bit values.
	ProQuintPatternShort = "CVCVC"

	// ProQuintPattern in accordance with ProQuint-compatible configuration
	// Based on the Proquint specification: https://arxiv.org/html/0901.4016
	// Provides a pre-configured encoder that generates identifiers compatible with
	// the original Proquint library, using the pattern CVCVC-CVCVC to encode 32-bit values.
	ProQuintPattern = "CVCVCXCVCVC"

	ProquintVowels     = "aiou"
	ProquintConsonants = "bdfghjklmnprstvz"
	ProquintDelimiter  = "-"
	// ProQuintBlockBitWidth is the entropy contributed by one CVCVC block
	// with canonical ProQuint alphabets (16 consonants, 4 vowels).
	ProQuintBlockBitWidth = 16
	// ProQuintBitWidth is the bit width for ProQuint encoding (32-bit values).
	ProQuintBitWidth = 32
	// ProQuintSHA256BitWidth is the bit width for the SHA-256-compatible ProQuint preset.
	ProQuintSHA256BitWidth = 256
	// ProQuintSHA256Blocks is the number of CVCVC blocks required for 256 bits.
	ProQuintSHA256Blocks = ProQuintSHA256BitWidth / ProQuintBlockBitWidth
)

var (
	// AllowedVowels defines the permitted vowel characters.
	AllowedVowels = map[rune]bool{
		'a': true, 'e': true, 'i': true, 'o': true, 'u': true, 'y': true,
		'A': true, 'E': true, 'I': true, 'O': true, 'U': true, 'Y': true,
	}

	// AllowedPatternLengths is deprecated. Pattern lengths are now validated as: MinPatternLength <= length <= MaxPatternLength.
	// This array is kept for reference only and shows the previously restricted prime-number lengths.
	//
	// Deprecated: Use isAllowedLength() function instead. Prime-number constraint has been removed.
	AllowedPatternLengths = []int{5, 7, 11, 13, 23, 29, 31, 37, 41, 43, 47}

	// AllowedPlaceholders defines the valid placeholder identifiers.
	AllowedPlaceholders = map[PlaceholderType]string{
		Consonant: "Consonant", // Hard consonants: b,c,d,f,g,h,j,k,p,q,s,t,v,w,x,z
		Vowel:     "Vowel",     // Pure vowels: a,e,i,o,u
		Liquid:    "Liquid",    // Liquid consonants: l,m,n,r
		Nasal:     "Nasal",     // Nasal sounds: m,n (or use IPA: ŋ for ng)
		Sibilant:  "Sibilant",  // Hissing sounds: s,z (or use IPA: ʃ,ʒ for sh,zh)
		Fricative: "Fricative", // Friction sounds: f,v (or use IPA: θ,ð for th,dh)
		CustomX:   "User-defined category 1",
		CustomY:   "User-defined category 2",
		CustomZ:   "User-defined category 3",
	}
	ProQuintPlaceholders = PlaceholderMap{
		Vowel:     []rune(ProquintVowels),
		Consonant: []rune(ProquintConsonants),
		CustomX:   []rune(ProquintDelimiter),
	}

	// ProQuintConfig provides Proquint-compatible encoding
	// See: https://arxiv.org/html/0901.4016
	ProQuintConfig = PhonidConfig{
		Patterns:     []string{ProQuintPattern},
		Placeholders: ProQuintPlaceholders,
	}

	// ProQuintTinyConfig keeps canonical ProQuint alphabets but uses a single short pattern.
	ProQuintTinyConfig = PhonidConfig{
		Patterns:     []string{ProQuintPatternShort},
		Placeholders: ProQuintPlaceholders,
	}

	// ProQuintSHA256Pattern encodes exactly 256 bits with 16 CVCVC blocks.
	// Capacity: (16^3 * 4^2)^16 = 2^256.
	ProQuintSHA256Pattern = strings.TrimSuffix(strings.Repeat(ProQuintPatternShort+"X", ProQuintSHA256Blocks), "X")

	// ProQuintSHA256Config uses canonical ProQuint alphabets with exact SHA-256 capacity.
	ProQuintSHA256Config = PhonidConfig{
		Patterns:     []string{ProQuintSHA256Pattern},
		Placeholders: ProQuintPlaceholders,
	}

	// ComplementPlaceholders lists all non-vowel phonetic categories.
	ComplementPlaceholders = []PlaceholderType{
		Consonant,
		Liquid,
		Nasal,
		Sibilant,
		Fricative,
	}

	// DefaultPlaceholders provides sensible defaults for common phonetic categories.
	DefaultPlaceholders = ProQuintPlaceholders

	DefaultPatterns = []string{ProQuintPatternShort, ProQuintPattern}

	DefaultConfig = ProQuintConfig
)

type (
	// PlaceholderType represents a valid phonetic placeholder identifier.

	PlaceholderType rune
	PlaceholderMap  map[PlaceholderType]RuneSet

	// RuneSet is a slice of runes that can be unmarshaled from a string.
	// This allows TOML configs to use simple strings like C = "bcdfg" instead of arrays.
	RuneSet []rune

	// PhonidConfig holds phonetic pattern configuration.
	//
	// Custom categories (X, Y, Z) can be used for domain-specific sounds:
	//
	//	config := PhonidConfig{
	//	    Patterns: []string{"CXVC"},  // Mix custom with built-in
	//	    Placeholders: PlaceholderMap{
	//	        Consonant: RuneSet("bcd"),
	//	        Vowel: RuneSet("ae"),
	//	        CustomX: RuneSet("ŋ"),  // Velar nasal
	//	    },
	//	}
	PhonidConfig struct {
		Patterns     []string       // e.g., "CVCVC", "CLVCV", "VCCVL" // Each character becomes a placeholder key
		Placeholders PlaceholderMap // Maps placeholder to character set, e.g., {"C": "bcdfg", "V": "aeiou"}
	}
)

// UnmarshalText implements encoding.TextUnmarshaler for TOML/JSON unmarshaling.
func (rs *RuneSet) UnmarshalText(text []byte) error {
	*rs = []rune(string(text))
	return nil
}

// Validate checks if the phonetic config is valid.
func (pc *PhonidConfig) Validate() error {
	// Apply defaults only if both patterns and placeholders are empty
	// This ensures consistency between patterns and placeholders
	switch {
	case len(pc.Patterns) == 0 && len(pc.Placeholders) == 0:
		pc.Patterns = DefaultPatterns
		pc.Placeholders = DefaultPlaceholders
	case len(pc.Patterns) == 0:
		// If only patterns are empty but placeholders are provided,
		// we cannot safely apply default patterns as they might require
		// placeholders not in the provided map
		return errors.New("patterns cannot be empty when placeholders are provided")
	case len(pc.Placeholders) == 0:
		// If only placeholders are empty but patterns are provided,
		// require explicit placeholder definitions to avoid ambiguity
		return errors.New("placeholders cannot be empty when patterns are provided")
	}

	patterns := pc.Patterns
	patternLengths := make(map[int]struct{})

	// ensure lengths allow 1:1 mapping with patterns
	for _, p := range patterns {
		patternLen := len(p)
		if _, exists := patternLengths[patternLen]; exists {
			return fmt.Errorf("duplicate pattern length %d found", patternLen)
		}
		if !isAllowedLength(patternLen) {
			return fmt.Errorf(
				"pattern length %d is not allowed (must be between %d and %d)",
				patternLen,
				MinPatternLength,
				MaxPatternLength,
			)
		}

		// Validate individual pattern
		if err := validatePattern(p, pc.Placeholders); err != nil {
			return fmt.Errorf("pattern '%s': %w", p, err)
		}
		patternLengths[patternLen] = struct{}{}
	}

	return nil
}

func validatePattern(pattern string, placeholders PlaceholderMap) error {
	placeholderCounts, err := countPlaceholders(pattern, placeholders)
	if err != nil {
		return err
	}

	if err := validatePlaceholderSets(placeholderCounts, placeholders, len(pattern)); err != nil {
		return err
	}

	if err := validatePatternRequirements(placeholderCounts, placeholders); err != nil {
		return err
	}

	if err := validateNoOverlaps(placeholderCounts, placeholders); err != nil {
		return err
	}

	return nil
}

// countPlaceholders counts occurrences of each placeholder in the pattern.
func countPlaceholders(pattern string, placeholders PlaceholderMap) (map[PlaceholderType]int, error) {
	counts := make(map[PlaceholderType]int)

	for _, r := range pattern {
		placeholder := PlaceholderType(r)
		if _, exists := placeholders[placeholder]; !exists {
			return nil, fmt.Errorf("pattern contains '%c' but no character set defined for it", r)
		}
		counts[placeholder]++
	}

	return counts, nil
}

// validatePlaceholderSets validates each placeholder's character set.
func validatePlaceholderSets(counts map[PlaceholderType]int, placeholders PlaceholderMap, patternLength int) error {
	for placeholder, chars := range placeholders {
		// Only validate placeholders actually used in pattern
		if counts[placeholder] == 0 {
			continue
		}

		if hasDuplicates([]rune(chars)) {
			return fmt.Errorf("placeholder '%c' contains duplicate characters", placeholder)
		}

		if err := validateVowelSet(placeholder, chars); err != nil {
			return err
		}

		if err := validateMinimumSize(placeholder, chars, patternLength); err != nil {
			return err
		}
	}
	return nil
}

// validateVowelSet validates vowel placeholder character sets.
func validateVowelSet(placeholder PlaceholderType, chars RuneSet) error {
	if placeholder != Vowel {
		return nil
	}

	if len(chars) == 0 {
		return fmt.Errorf("vowel placeholder '%c' must have at least one character", placeholder)
	}

	for _, char := range chars {
		if !isVowelBase(char) {
			return fmt.Errorf(
				"vowel placeholder '%c' contains invalid vowel '%c' (allowed: a,e,i,o,u,y and their diacritical variants)",
				placeholder,
				char,
			)
		}
	}
	return nil
}

// validateMinimumSize checks minimum character requirements for placeholders.
// Vowel minimums depend on pattern length to ensure sufficient shuffle capacity:
//   - Length 3-4: needs 2 vowels (with 3 consonants: 18-36 combinations)
//   - Length 5-6: needs 3 vowels (with 3 consonants: 243 combinations)
//   - Length 7+:  needs 2 vowels (with 3 consonants: 3^4 * 2^3 = 648 combinations)
func validateMinimumSize(placeholder PlaceholderType, chars RuneSet, patternLength int) error {
	if placeholder != Vowel {
		return nil
	}

	var minVowels int
	switch {
	case patternLength >= MinPatternLengthForLongVowels:
		minVowels = MinCharsForVowelLong
	case patternLength >= MinPatternLengthForMediumVowels:
		minVowels = MinCharsForVowelMedium
	default: // Length 3-4
		minVowels = MinCharsForVowelShort
	}

	if len(chars) < minVowels {
		return fmt.Errorf("vowel placeholder needs at least %d characters for pattern length %d, got %d",
			minVowels, patternLength, len(chars))
	}
	return nil
}

// validatePatternRequirements checks pattern has required placeholder types.
func validatePatternRequirements(counts map[PlaceholderType]int, placeholders PlaceholderMap) error {
	if err := requireVowel(counts); err != nil {
		return err
	}

	if err := requireMinimalComplement(counts, placeholders); err != nil {
		return err
	}

	return nil
}

// requireVowel ensures pattern contains at least one vowel.
func requireVowel(counts map[PlaceholderType]int) error {
	if counts[Vowel] > 0 {
		return nil
	}

	return fmt.Errorf(
		"pattern must contain at least one vowel placeholder ('%c': %s)",
		Vowel,
		AllowedPlaceholders[Vowel],
	)
}

// requireMinimalComplement ensures sufficient non-vowel variety.
func requireMinimalComplement(counts map[PlaceholderType]int, placeholders PlaceholderMap) error {
	for placeholder := range counts {
		if isComplementPlaceholder(placeholder) &&
			len(placeholders[placeholder]) >= MinCharsForComplement {
			return nil
		}
	}

	complementNames := make([]string, len(ComplementPlaceholders))
	for i, complement := range ComplementPlaceholders {
		complementNames[i] = string(complement)
	}

	return fmt.Errorf(
		"pattern must use at least one complement placeholder (%s) with at least %d characters",
		strings.Join(complementNames, ", "),
		MinCharsForComplement,
	)
}

// validateNoOverlaps checks for character overlap between placeholders.
func validateNoOverlaps(counts map[PlaceholderType]int, placeholders PlaceholderMap) error {
	allPlaceholders := make([]PlaceholderType, 0, len(counts))
	for p := range counts {
		allPlaceholders = append(allPlaceholders, p)
	}

	for i := range allPlaceholders {
		for j := i + 1; j < len(allPlaceholders); j++ {
			p1, p2 := allPlaceholders[i], allPlaceholders[j]
			if hasOverlap(placeholders[p1], placeholders[p2]) {
				return fmt.Errorf("placeholders '%c' and '%c' have overlapping characters", p1, p2)
			}
		}
	}

	return nil
}

// isComplementPlaceholder checks if a placeholder is a non-vowel phonetic category.
func isComplementPlaceholder(p PlaceholderType) bool {
	return slices.Contains(ComplementPlaceholders, p)
}

// hasDuplicates checks if a rune slice contains duplicates.
func hasDuplicates(runes []rune) bool {
	seen := make(map[rune]bool)
	for _, r := range runes {
		if seen[r] {
			return true
		}
		seen[r] = true
	}
	return false
}

// hasOverlap checks if two rune slices have any common elements.
func hasOverlap(a, b []rune) bool {
	set := make(map[rune]bool)
	for _, r := range a {
		set[r] = true
	}
	for _, r := range b {
		if set[r] {
			return true
		}
	}
	return false
}

// isAllowedLength checks if a length is within the allowed range.
func isAllowedLength(length int) bool {
	return length >= MinPatternLength && length <= MaxPatternLength
}

// isVowelBase checks if a rune is a vowel, stripping diacritics
// Supports characters like ü, ä, ö, é, è which normalize to base vowels.
func isVowelBase(r rune) bool {
	// First check if it's directly in allowed vowels
	if AllowedVowels[r] {
		return true
	}

	// Normalize to decomposed form (NFD) and check base character
	normalized := norm.NFD.String(string(r))

	// Get base character (first rune before combining marks)
	for _, char := range normalized {
		// Skip combining diacritical marks
		if !unicode.Is(unicode.Mn, char) { // Mn = Nonspacing Mark (diacritics)
			return AllowedVowels[char]
		}
	}
	return false
}
