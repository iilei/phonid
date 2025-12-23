package phonid

import (
	"fmt"
	"math"
	"unicode"

	"golang.org/x/text/unicode/norm"
)

// Minimum requirements per placeholder type
const (
	MinCharsPerPlaceholder = 2
)

// AllowedVowels defines the permitted vowel characters
var AllowedVowels = map[rune]bool{
	'a': true, 'e': true, 'i': true, 'o': true, 'u': true, 'y': true,
	'A': true, 'E': true, 'I': true, 'O': true, 'U': true, 'Y': true,
}

// AllowedPatternLengths defines the permitted pattern lengths
var AllowedPatternLengths = []int{5, 7, 11}

// DefaultPlaceholders provides sensible defaults for common phonetic categories
var DefaultPlaceholders = map[string][]rune{
	"C": {'b', 'c', 'd', 'f', 'g', 'h', 'j', 'k', 'p', 'q', 's', 't', 'v', 'w', 'x', 'z'}, // consonants
	"L": {'l', 'm', 'n', 'r'},                                                             // liquids
	"V": {'a', 'e', 'i', 'o', 'u'},                                                        // vowels
}

// PhoneticConfig holds phonetic pattern configuration
type PhoneticConfig struct {
	Pattern      string            `default:"CLVCV"` // e.g., "CVCVC", "CLVCV", "VCCVL"
	Placeholders map[string][]rune // Maps placeholder to character set, e.g., {"C": [b,d,k], "V": [a,e]}
}

// Validate checks if the phonetic config is valid
func (pc *PhoneticConfig) Validate(base BaseEncoding) error {
	pattern := pc.Pattern
	if pattern == "" {
		return fmt.Errorf("pattern cannot be empty")
	}

	// Validate pattern format
	if !isAllowedLength(len(pattern)) {
		return fmt.Errorf("pattern length %d is not allowed (must be one of %v)", len(pattern), AllowedPatternLengths)
	}

	// Use defaults if placeholders not provided
	placeholders := pc.Placeholders
	if placeholders == nil {
		placeholders = DefaultPlaceholders
	}

	// Count occurrences of each placeholder in pattern
	placeholderCounts := make(map[string]int)
	for _, r := range pattern {
		placeholder := string(r)
		if _, exists := placeholders[placeholder]; !exists {
			return fmt.Errorf("pattern contains '%s' but no character set defined for it", placeholder)
		}
		placeholderCounts[placeholder]++
	}

	// Validate each placeholder's character set
	for placeholder, chars := range placeholders {
		// Only validate placeholders actually used in pattern
		if placeholderCounts[placeholder] == 0 {
			continue
		}

		if len(chars) < MinCharsPerPlaceholder {
			return fmt.Errorf("placeholder '%s' needs at least %d characters, got %d",
				placeholder, MinCharsPerPlaceholder, len(chars))
		}

		if hasDuplicates(chars) {
			return fmt.Errorf("placeholder '%s' contains duplicate characters", placeholder)
		}

		// Special validation for vowel placeholder
		if placeholder == "V" || placeholder == "v" {
			if len(chars) == 0 {
				return fmt.Errorf("vowel placeholder '%s' must have at least one character", placeholder)
			}
			for _, char := range chars {
				if !isVowelBase(char) {
					return fmt.Errorf("vowel placeholder '%s' contains invalid vowel '%c' (allowed: a,e,i,o,u,y and their diacritical variants)", placeholder, char)
				}
			}
		}
	}

	// Check for overlaps between all placeholder character sets
	allPlaceholders := make([]string, 0, len(placeholderCounts))
	for p := range placeholderCounts {
		allPlaceholders = append(allPlaceholders, p)
	}

	for i := 0; i < len(allPlaceholders); i++ {
		for j := i + 1; j < len(allPlaceholders); j++ {
			p1, p2 := allPlaceholders[i], allPlaceholders[j]
			if hasOverlap(placeholders[p1], placeholders[p2]) {
				return fmt.Errorf("placeholders '%s' and '%s' have overlapping characters", p1, p2)
			}
		}
	}

	// Require at least one vowel-like placeholder for pronounceability
	hasVowel := false
	for placeholder := range placeholderCounts {
		if placeholder == "V" || placeholder == "v" {
			hasVowel = true
			break
		}
	}
	if !hasVowel {
		return fmt.Errorf("pattern must contain at least one vowel placeholder ('V')")
	}

	// Calculate total combinations
	combinations := 1
	for placeholder, count := range placeholderCounts {
		chars := placeholders[placeholder]
		combinations *= int(math.Pow(float64(len(chars)), float64(count)))
	}

	if combinations < int(base) {
		return fmt.Errorf("pattern '%s' produces only %d combinations (need at least %d for base %d)",
			pattern, combinations, base, base)
	}

	return nil
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
