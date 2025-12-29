package phonid

import (
	"bytes"
	"fmt"
	"os"

	"github.com/pelletier/go-toml/v2"
)

type (
	// PositiveInt represents a non-negative integer
	PositiveInt int
	// TOMLConfig represents the top-level TOML structure
	TOMLConfig struct {
		Base      PositiveInt       `toml:"base,omitempty"`
		Shuffle   TOMLShuffleConfig `toml:"shuffle,omitempty"`
		Phonetic  TOMLPhonidConfig  `toml:"phonetic,omitempty"`
		Preflight []PreflightCheck  `toml:"preflight"` // Required - no omitempty
	}

	// PreflightCheck represents a single input->output verification
	PreflightCheck struct {
		Input  PositiveInt `toml:"input"`
		Output string      `toml:"output"`
	}

	// TOMLShuffleConfig represents shuffle configuration
	TOMLShuffleConfig struct {
		BitWidth PositiveInt `toml:"bit_width,omitempty"`
		Rounds   PositiveInt `toml:"rounds,omitempty"`
		Seed     PositiveInt `toml:"seed,omitempty"`
	}

	// TOMLPhonidConfig represents the phonetic configuration
	TOMLPhonidConfig struct {
		Patterns     []string          `toml:"patterns,omitempty"`
		Placeholders map[string]string `toml:"placeholders,omitempty"`
	}
)

func (p PositiveInt) Validate() error {
	if p < 0 {
		return fmt.Errorf("value must be non-negative, got %d", p)
	}
	return nil
}

// LoadPhonidRC loads and validates a PhonidConfig from a phonidrc file with strict preflight validation
func LoadPhonidRC(filepath string) (*PhonidConfig, []PreflightCheck, error) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, make([]PreflightCheck, 0), fmt.Errorf("failed to read %s: %w", filepath, err)
	}

	return ParsePhonidRC(string(data))
}

// LoadPhonidRCLenient loads a PhonidConfig without requiring preflight checks
// Used exclusively by 'phonid preflight --suggest' command
func LoadPhonidRCLenient(filepath string) (*PhonidConfig, []PreflightCheck, error) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, make([]PreflightCheck, 0), fmt.Errorf("failed to read %s: %w", filepath, err)
	}

	return ParsePhonidRCLenient(string(data))
}

// ParsePhonidRC parses TOML content requiring preflight checks
// Used exclusively by 'phonid preflight --suggest' command
func ParsePhonidRC(content string) (*PhonidConfig, []PreflightCheck, error) {
	return parsePhonidRCInternal(content, false)
}

// ParsePhonidRCLenient parses TOML content without requiring preflight checks
// Used exclusively by 'phonid preflight --suggest' command
func ParsePhonidRCLenient(content string) (*PhonidConfig, []PreflightCheck, error) {
	return parsePhonidRCInternal(content, true)
}

// ParsePhonidRC parses TOML content into a validated PhonidConfig using strict mode
func parsePhonidRCInternal(content string, lenitent bool) (*PhonidConfig, []PreflightCheck, error) {
	var tomlConfig TOMLConfig

	// Create decoder with strict mode enabled
	decoder := toml.NewDecoder(bytes.NewReader([]byte(content)))
	decoder.DisallowUnknownFields() // Strict mode - reject unknown fields
	preflight := make([]PreflightCheck, 0)

	if err := decoder.Decode(&tomlConfig); err != nil {
		// pelletier/go-toml v2 provides contextualized error messages
		return nil, preflight, fmt.Errorf("failed to parse TOML config: %w", err)
	}

	// Require at least one preflight check
	if len(tomlConfig.Preflight) == 0 && !lenitent {
		return nil, make([]PreflightCheck, 0), fmt.Errorf(
			"config must include at least one [[preflight]] check\n\n" +
				"Example:\n" +
				"  [[preflight]]\n" +
				"  input = 0\n" +
				"  output = \"babab\"\n\n" +
				"Hint: Run 'phonid preflight --suggest' to generate recommended checks")
	}
	preflight = tomlConfig.Preflight

	// Validate PositiveInt fields
	if err := tomlConfig.Base.Validate(); err != nil {
		return nil, preflight, fmt.Errorf("invalid base: %w", err)
	}

	// Convert TOML structure to PhonidConfig
	config := &PhonidConfig{
		Patterns: tomlConfig.Phonetic.Patterns,
	}

	// Convert string-based placeholders to PlaceholderType-based
	if tomlConfig.Phonetic.Placeholders != nil {
		config.Placeholders = make(map[PlaceholderType]RuneSet)

		for keyStr, stringChars := range tomlConfig.Phonetic.Placeholders {
			// Validate placeholder key - convert to runes first for proper UTF-8 handling
			keyRunes := []rune(keyStr)
			if len(keyRunes) != 1 {
				return nil, preflight, fmt.Errorf(
					"placeholder key '%s' must be single character",
					keyStr,
				)
			}

			placeholderType := PlaceholderType(keyRunes[0])

			// Validate placeholder type is allowed
			if _, isAllowed := AllowedPlaceholders[placeholderType]; !isAllowed {
				return nil, preflight, fmt.Errorf(
					"placeholder '%c' is not allowed. Valid placeholders: %v",
					placeholderType,
					getValidPlaceholderKeys(),
				)
			}

			// Convert string to RuneSet (simple conversion)
			config.Placeholders[placeholderType] = RuneSet(stringChars)
		}
	} else {
		// Use defaults if no placeholders specified
		config.Placeholders = DefaultPlaceholders
	}
	return config, preflight, nil
}

// ValidatePhonidRC validates a PhonidConfig loaded from RC file with base encoding
func ValidatePhonidRC(config *PhonidConfig) error {
	if config == nil {
		return fmt.Errorf("config cannot be nil")
	}

	return config.Validate()
}

// getValidPlaceholderKeys returns a slice of valid placeholder characters for error messages
func getValidPlaceholderKeys() []string {
	keys := make([]string, 0, len(AllowedPlaceholders))
	for key := range AllowedPlaceholders {
		keys = append(keys, string(key))
	}
	return keys
}

// // LoadAndValidatePhonidRC is a convenience function that loads and validates in one step
// func LoadAndValidatePhonidRC(filepath string, base BaseEncoding) (*PhonidConfig, []PreflightCheck, error) {
// 	config, preflight, err := LoadPhonidRC(filepath)
// 	if err != nil {
// 		return nil, preflight, err
// 	}

// 	if err := ValidatePhonidRC(config, base); err != nil {
// 		return nil, preflight, fmt.Errorf("invalid config in %s: %w", filepath, err)
// 	}

// 	return config, preflight, nil
// }

// // LoadAndValidatePhonidRCLenient is a convenience function that loads and validates without requiring preflight checks
// // Used exclusively by 'phonid preflight --suggest' command
// func LoadAndValidatePhonidRCLenient(filepath string, base BaseEncoding) (*PhonidConfig, []PreflightCheck, error) {
// 	config, preflight, err := LoadPhonidRCLenient(filepath)
// 	if err != nil {
// 		return nil, preflight, err
// 	}

// 	if err := ValidatePhonidRC(config, base); err != nil {
// 		return nil, preflight, fmt.Errorf("invalid config in %s: %w", filepath, err)
// 	}

// 	return config, preflight, nil
// }
