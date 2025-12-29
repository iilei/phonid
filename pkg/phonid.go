package phonid

import (
	"fmt"
	"math"
	"strings"
	"unicode"

	"golang.org/x/text/unicode/norm"
)

const (
	// Minimum requirements per placeholder type
	MinCharsForVowel      = 2
	MinCharsForComplement = 3 // At least one non-vowel category (C, L, N, S, or F) must have this many

	// Valid placeholder types
	Consonant PlaceholderType = 'C'
	Vowel     PlaceholderType = 'V'
	Liquid    PlaceholderType = 'L'
	Nasal     PlaceholderType = 'N'
	Sibilant  PlaceholderType = 'S'
	Fricative PlaceholderType = 'F'
	CustomX   PlaceholderType = 'X'
	CustomY   PlaceholderType = 'Y'
	CustomZ   PlaceholderType = 'Z'

	// ProQuint-compatible configuration
	// Based on the Proquint specification: https://arxiv.org/html/0901.4016
	// Provides a pre-configured encoder that generates identifiers compatible with
	// the original Proquint library, using the pattern CVCVC-CVCVC to encode 32-bit values.
	ProQuintPattern = "CVCVCXCVCVC"
)

var (
	// AllowedVowels defines the permitted vowel characters
	AllowedVowels = map[rune]bool{
		'a': true, 'e': true, 'i': true, 'o': true, 'u': true, 'y': true,
		'A': true, 'E': true, 'I': true, 'O': true, 'U': true, 'Y': true,
	}

	// AllowedPatternLengths defines the permitted pattern lengths
	AllowedPatternLengths = []int{3, 5, 7, 11, 23}

	// AllowedPlaceholders defines the valid placeholder identifiers
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
		Vowel:     []rune("aiou"),
		Consonant: []rune("bdfghjklmnprstvz"),
		CustomX:   RuneSet{'-'},
	}

	// ProQuintConfig provides Proquint-compatible encoding
	// See: https://arxiv.org/html/0901.4016
	ProQuintConfig = PhonidConfig{
		Patterns:     []string{ProQuintPattern},
		Placeholders: ProQuintPlaceholders,
	}

	// ComplementPlaceholders lists all non-vowel phonetic categories
	ComplementPlaceholders = []PlaceholderType{
		Consonant,
		Liquid,
		Nasal,
		Sibilant,
		Fricative,
	}

	// DefaultPlaceholders provides sensible defaults for common phonetic categories
	DefaultPlaceholders = map[PlaceholderType]RuneSet{
		Consonant: RuneSet("bcdfghjkpqstvwxz"),
		Liquid:    RuneSet("lmnr"),
		Vowel:     RuneSet("aeiou"),
		// Note: Sibilant, Fricative, and Nasal can be customized by users
		// to include IPA symbols (ʃ,ʒ,θ,ð,ŋ) for more precise phonetic representation
	}

	DefaultPatterns = []string{
		"CVC",
		"VCCVC",
		"CVCVCVC",
		"CVCVCVCVCVC",
	}
)

type (
	// PlaceholderType represents a valid phonetic placeholder identifier

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

func validatePattern(pattern string, placeholders PlaceholderMap) error {
	// Count occurrences of each placeholder in pattern
	placeholderCounts := make(map[PlaceholderType]int) // Change key type
	hasMinimalComplement := false

	for _, r := range pattern {
		placeholder := PlaceholderType(r) // Convert rune to PlaceholderType
		if _, exists := placeholders[placeholder]; !exists {
			return fmt.Errorf("pattern contains '%c' but no character set defined for it", r)
		}
		placeholderCounts[placeholder]++
	}

	// Validate each placeholder's character set
	for placeholder, chars := range placeholders {
		// Only validate placeholders actually used in pattern
		if placeholderCounts[placeholder] == 0 {
			continue
		}

		// Check if this is a complement (non-vowel) phonetic category
		if isComplementPlaceholder(placeholder) && len(chars) >= MinCharsForComplement {
			hasMinimalComplement = true
		}

		// Vowel must always meet minimum pronouncability and information density requirements
		if placeholder == Vowel && len(chars) < MinCharsForVowel {
			return fmt.Errorf("vowel placeholder needs at least %d characters, got %d",
				MinCharsForVowel, len(chars))
		}

		if hasDuplicates(chars) {
			return fmt.Errorf("placeholder '%c' contains duplicate characters", placeholder)
		}

		// Special validation for vowel placeholder
		if placeholder == Vowel {

			if len(chars) == 0 {
				return fmt.Errorf(
					"vowel placeholder '%c' must have at least one character",
					placeholder,
				)
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
		}
	}

	// Require at least one vowel placeholder for pronounceability
	hasVowel := false
	for placeholder := range placeholderCounts {
		if placeholder == Vowel {
			hasVowel = true
			break
		}
	}
	if !hasVowel {
		return fmt.Errorf(
			"pattern must contain at least one vowel placeholder ('%c': %s)",
			Vowel,
			AllowedPlaceholders[Vowel],
		)
	}

	// Require at least one complement category with sufficient variety
	if !hasMinimalComplement {
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

	// Check for overlaps between all placeholder character sets
	allPlaceholders := make([]PlaceholderType, 0, len(placeholderCounts))
	for p := range placeholderCounts {
		allPlaceholders = append(allPlaceholders, p)
	}

	// Compare each unique pair (triangular matrix pattern: inner loop starts one step ahead)
	for i := 0; i < len(allPlaceholders); i++ {
		for j := i + 1; j < len(allPlaceholders); j++ {
			p1, p2 := allPlaceholders[i], allPlaceholders[j]
			if hasOverlap(placeholders[p1], placeholders[p2]) {
				return fmt.Errorf("placeholders '%c' and '%c' have overlapping characters", p1, p2)
			}
		}
	}

	// Calculate total combinations
	combinations := 1
	for placeholder, count := range placeholderCounts {
		chars := placeholders[placeholder]
		combinations *= int(math.Pow(float64(len(chars)), float64(count)))
	}

	return nil
}

// Validate checks if the phonetic config is valid
func (pc *PhonidConfig) Validate() error {
	// Apply defaults if not provided
	if len(pc.Patterns) == 0 {
		pc.Patterns = DefaultPatterns
	}
	if len(pc.Placeholders) == 0 {
		pc.Placeholders = DefaultPlaceholders
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
				"pattern length %d is not allowed (must be one of %v)",
				patternLen,
				AllowedPatternLengths,
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

// isComplementPlaceholder checks if a placeholder is a non-vowel phonetic category
func isComplementPlaceholder(p PlaceholderType) bool {
	for _, complement := range ComplementPlaceholders {
		if p == complement {
			return true
		}
	}
	return false
}

// hasDuplicates checks if a rune slice contains duplicates
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

// hasOverlap checks if two rune slices have any common elements
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

// isAllowedLength checks if a length is in the allowed lengths list
func isAllowedLength(length int) bool {
	for _, allowed := range AllowedPatternLengths {
		if length == allowed {
			return true
		}
	}
	return false
}

// isVowelBase checks if a rune is a vowel, stripping diacritics
// Supports characters like ü, ä, ö, é, è which normalize to base vowels
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
