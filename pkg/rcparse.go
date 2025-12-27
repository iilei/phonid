package phonid

import (
	"bytes"
	"fmt"
	"os"

	"github.com/pelletier/go-toml/v2"
)

// TOMLPhonidConfig represents the TOML file structure with strict validation
type TOMLPhonidConfig struct {
	Patterns     []string          `toml:"patterns"`
	Placeholders map[string]string `toml:"placeholders,omitempty"`
}

// LoadPhonidRC loads and validates a PhonidConfig from a phonidrc file
func LoadPhonidRC(filepath string) (*PhonidConfig, error) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %w", filepath, err)
	}

	return ParsePhonidRC(string(data))
}

// ParsePhonidRC parses TOML content into a validated PhonidConfig using strict mode
func ParsePhonidRC(content string) (*PhonidConfig, error) {
	var tomlConfig TOMLPhonidConfig

	// Create decoder with strict mode enabled
	decoder := toml.NewDecoder(bytes.NewReader([]byte(content)))
	decoder.DisallowUnknownFields() // Strict mode - reject unknown fields

	if err := decoder.Decode(&tomlConfig); err != nil {
		// pelletier/go-toml v2 provides contextualized error messages
		return nil, fmt.Errorf("failed to parse TOML config: %w", err)
	}

	// Convert TOML structure to PhonidConfig
	config := &PhonidConfig{
		Patterns: tomlConfig.Patterns,
	}

	// Convert string-based placeholders to PlaceholderType-based
	if tomlConfig.Placeholders != nil {
		config.Placeholders = make(map[PlaceholderType]RuneSet)

		for keyStr, stringChars := range tomlConfig.Placeholders {
			// Validate placeholder key - convert to runes first for proper UTF-8 handling
			keyRunes := []rune(keyStr)
			if len(keyRunes) != 1 {
				return nil, fmt.Errorf("placeholder key '%s' must be single character", keyStr)
			}

			placeholderType := PlaceholderType(keyRunes[0])

			// Validate placeholder type is allowed
			if _, isAllowed := AllowedPlaceholders[placeholderType]; !isAllowed {
				return nil, fmt.Errorf("placeholder '%c' is not allowed. Valid placeholders: %v",
					placeholderType, getValidPlaceholderKeys())
			}

			// Convert string to RuneSet (simple conversion)
			config.Placeholders[placeholderType] = RuneSet(stringChars)
		}
	} else {
		// Use defaults if no placeholders specified
		config.Placeholders = DefaultPlaceholders
	}
	return config, nil
}

// ValidatePhonidRC validates a PhonidConfig loaded from RC file with base encoding
func ValidatePhonidRC(config *PhonidConfig, base BaseEncoding) error {
	if config == nil {
		return fmt.Errorf("config cannot be nil")
	}

	return config.Validate(base)
}

// LoadAndValidatePhonidRC is a convenience function that loads and validates in one step
func LoadAndValidatePhonidRC(filepath string, base BaseEncoding) (*PhonidConfig, error) {
	config, err := LoadPhonidRC(filepath)
	if err != nil {
		return nil, err
	}

	if err := ValidatePhonidRC(config, base); err != nil {
		return nil, fmt.Errorf("invalid config in %s: %w", filepath, err)
	}

	return config, nil
}

// getValidPlaceholderKeys returns a slice of valid placeholder characters for error messages
func getValidPlaceholderKeys() []string {
	keys := make([]string, 0, len(AllowedPlaceholders))
	for key := range AllowedPlaceholders {
		keys = append(keys, string(key))
	}
	return keys
}
