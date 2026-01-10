package phonid

import (
	"bytes"
	"errors"
	"fmt"
	"math"
	"math/big"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/pelletier/go-toml/v2"
)

const (
	RcFileName      = ".phonidrc"
	RcFileOptSuffix = ".toml"
	decimalBase     = 10 // Base for decimal number parsing
)

type (
	// PositiveInt represents a non-negative integer that can be either native int64 or big.Int.
	// This allows seamless handling of both regular integers and arbitrarily large numbers
	// with automatic optimization for the common case of small numbers.
	PositiveInt interface {
		// Int64 returns the value as int64 if it fits, otherwise returns (0, false)
		Int64() (int64, bool)

		// BigInt returns the value as *big.Int (always succeeds)
		BigInt() *big.Int

		// Cmp compares with another PositiveInt (-1 if less, 0 if equal, 1 if greater)
		Cmp(other PositiveInt) int

		// String returns decimal string representation
		String() string

		// Sign returns -1, 0, or 1 (should always be >= 0 for PositiveInt)
		Sign() int

		// Validate checks that the value is non-negative
		Validate() error
	}

	// positiveIntNative is the fast-path implementation for values that fit in int64.
	positiveIntNative int64

	// positiveIntBig is the slow-path implementation for arbitrarily large values.
	positiveIntBig struct {
		value *big.Int
	}

	// TOMLConfig represents the top-level TOML structure.
	TOMLConfig struct {
		Base      *TomlPositiveInt  `toml:"base,omitempty"`
		Shuffle   TOMLShuffleConfig `toml:"shuffle,omitempty"`
		Phonetic  TOMLPhonidConfig  `toml:"phonetic,omitempty"`
		Preflight []PreflightCheck  `toml:"preflight"` // Required - no omitempty
	}

	// PreflightCheck represents a single input->output verification.
	PreflightCheck struct {
		Input  *TomlPositiveInt `toml:"input"`
		Output string           `toml:"output"`
	}

	// TOMLShuffleConfig represents shuffle configuration.
	// BitWidth is calculated automatically from pattern capacity.
	TOMLShuffleConfig struct {
		Rounds *TomlPositiveInt `toml:"rounds,omitempty"`
		Seed   *TomlPositiveInt `toml:"seed,omitempty"`
	}

	// TOMLPhonidConfig represents the phonetic configuration.
	TOMLPhonidConfig struct {
		Patterns     []string          `toml:"patterns,omitempty"`
		Placeholders map[string]string `toml:"placeholders,omitempty"`
	}

	// TomlPositiveInt is a wrapper for TOML unmarshaling that stores the actual PositiveInt.
	TomlPositiveInt struct {
		Value PositiveInt
	}
)

// ========== positiveIntNative implementation ==========

func (n positiveIntNative) Int64() (int64, bool) {
	return int64(n), true
}

func (n positiveIntNative) BigInt() *big.Int {
	return big.NewInt(int64(n))
}

func (n positiveIntNative) Cmp(other PositiveInt) int {
	if v, ok := other.Int64(); ok {
		a, b := int64(n), v
		if a < b {
			return -1
		} else if a > b {
			return 1
		}
		return 0
	}
	// other is big.Int, compare as big.Int
	return n.BigInt().Cmp(other.BigInt())
}

func (n positiveIntNative) String() string {
	return strconv.FormatInt(int64(n), 10)
}

func (n positiveIntNative) Sign() int {
	if n < 0 {
		return -1
	} else if n > 0 {
		return 1
	}
	return 0
}

func (n positiveIntNative) Validate() error {
	if n < 0 {
		return fmt.Errorf("value must be non-negative, got %d", n)
	}
	return nil
}

// ========== positiveIntBig implementation ==========

func (n *positiveIntBig) Int64() (int64, bool) {
	if n.value.IsInt64() {
		return n.value.Int64(), true
	}
	return 0, false
}

func (n *positiveIntBig) BigInt() *big.Int {
	return new(big.Int).Set(n.value)
}

func (n *positiveIntBig) Cmp(other PositiveInt) int {
	return n.value.Cmp(other.BigInt())
}

func (n *positiveIntBig) String() string {
	return n.value.String()
}

func (n *positiveIntBig) Sign() int {
	return n.value.Sign()
}

func (n *positiveIntBig) Validate() error {
	if n.value.Sign() < 0 {
		return fmt.Errorf("value must be non-negative, got %s", n.value.String())
	}
	return nil
}

// ========== PositiveInt constructors ==========

// NewPositiveInt creates a PositiveInt from an int64.
// Panics if the value is negative (use with validated input).
func NewPositiveInt(n int64) PositiveInt {
	if n < 0 {
		panic(fmt.Sprintf("PositiveInt must be non-negative, got %d", n))
	}
	return positiveIntNative(n)
}

// NewPositiveIntFromBig creates a PositiveInt from a *big.Int.
// Automatically optimizes to native int64 if the value fits.
// Panics if the value is negative (use with validated input).
func NewPositiveIntFromBig(n *big.Int) PositiveInt {
	if n.Sign() < 0 {
		panic("PositiveInt must be non-negative, got " + n.String())
	}

	// Optimize: use native int64 if it fits
	if n.IsInt64() {
		return positiveIntNative(n.Int64())
	}

	return &positiveIntBig{value: new(big.Int).Set(n)}
}

// ParsePositiveInt parses a decimal string into a PositiveInt.
// Automatically uses the most efficient representation.
func ParsePositiveInt(s string) (PositiveInt, error) {
	// Try parsing as int64 first (fast path)
	if n, err := strconv.ParseInt(s, 10, 64); err == nil {
		if n < 0 {
			return nil, fmt.Errorf("number must be non-negative, got %s", s)
		}
		return positiveIntNative(n), nil
	}

	// Fall back to big.Int
	bigN := new(big.Int)
	if _, ok := bigN.SetString(s, decimalBase); !ok {
		return nil, fmt.Errorf("invalid number: %s", s)
	}

	if bigN.Sign() < 0 {
		return nil, fmt.Errorf("number must be non-negative, got %s", s)
	}

	return NewPositiveIntFromBig(bigN), nil
}

// ========== TOML unmarshaling support ==========

func (t *TomlPositiveInt) UnmarshalText(data []byte) error {
	s := string(data)
	val, err := ParsePositiveInt(s)
	if err != nil {
		return err
	}
	t.Value = val
	return nil
}

func (t *TomlPositiveInt) MarshalText() ([]byte, error) {
	if t == nil || t.Value == nil {
		return []byte("0"), nil
	}
	return []byte(t.Value.String()), nil
}

// ToPositiveInt converts the TOML wrapper to PositiveInt.
func (t *TomlPositiveInt) ToPositiveInt() PositiveInt {
	if t == nil || t.Value == nil {
		return NewPositiveInt(0)
	}
	return t.Value
}

// LoadPhonidRC loads and validates a PhonidConfig from a phonidrc file with strict preflight validation.
func LoadPhonidRC(fp string) (*PhonidConfig, []PreflightCheck, error) {
	data, err := readConfigFile(fp)
	if err != nil {
		return nil, nil, err
	}

	return ParsePhonidRC(string(data))
}

// LoadPhonidRCLenient loads a PhonidConfig without requiring preflight checks
// Used exclusively by 'phonid preflight --suggest' command.
func LoadPhonidRCLenient(fp string) (*PhonidConfig, []PreflightCheck, error) {
	data, err := readConfigFile(fp)
	if err != nil {
		return nil, nil, err
	}

	return ParsePhonidRCLenient(string(data))
}

// ParsePhonidRC parses TOML content requiring preflight checks.
// Returns config and preflight checks for validation with NewPhoneticEncoder.
func ParsePhonidRC(content string) (*PhonidConfig, []PreflightCheck, error) {
	return parsePhonidRCInternal(content, false)
}

// ParsePhonidRCLenient parses TOML content without requiring preflight checks.
// Used exclusively by 'phonid preflight --suggest' command.
func ParsePhonidRCLenient(content string) (*PhonidConfig, []PreflightCheck, error) {
	return parsePhonidRCInternal(content, true)
}

// parsePhonidRCInternal parses TOML content into a PhonidConfig.
// Only handles TOML structure parsing and conversion - does not validate behavior.
// Validation is delegated to NewPhoneticEncoder* constructors.
func parsePhonidRCInternal(content string, lenient bool) (*PhonidConfig, []PreflightCheck, error) {
	var tomlConfig TOMLConfig

	// Create decoder with strict mode enabled
	decoder := toml.NewDecoder(bytes.NewReader([]byte(content)))
	decoder.DisallowUnknownFields() // Strict mode - reject unknown fields
	preflight := make([]PreflightCheck, 0)

	if err := decoder.Decode(&tomlConfig); err != nil {
		// pelletier/go-toml v2 provides contextualized error messages
		return nil, preflight, fmt.Errorf("failed to parse TOML config: %w", err)
	}

	// Require at least one preflight check in strict mode
	if len(tomlConfig.Preflight) == 0 && !lenient {
		return nil, preflight, ErrNoPreflightChecks
	}
	preflight = tomlConfig.Preflight

	// Validate Base field if present
	if tomlConfig.Base != nil {
		if err := tomlConfig.Base.ToPositiveInt().Validate(); err != nil {
			return nil, preflight, fmt.Errorf("invalid base: %w", err)
		}
	}

	// Convert TOML structure to PhonidConfig
	config := &PhonidConfig{
		Patterns: tomlConfig.Phonetic.Patterns,
	}

	// Helper to convert TomlPositiveInt to int
	toInt := func(t *TomlPositiveInt) int {
		if t == nil {
			return 0
		}
		val := t.ToPositiveInt()
		if v, ok := val.Int64(); ok {
			return int(v)
		}
		// If it doesn't fit in int, this is a configuration error
		// but we'll let validation catch it later
		return math.MaxInt
	}

	// Helper to convert TomlPositiveInt to uint64
	toUint64 := func(t *TomlPositiveInt) uint64 {
		if t == nil {
			return 0
		}
		val := t.ToPositiveInt()
		if v, ok := val.Int64(); ok {
			//nolint:gosec // G115: Validated non-negative by PositiveInt type
			return uint64(v)
		}
		return math.MaxUint64
	}

	// Convert shuffle config if present
	rounds := toInt(tomlConfig.Shuffle.Rounds)
	seed := toUint64(tomlConfig.Shuffle.Seed)

	if rounds > 0 || seed > 0 {
		config.Shuffle = &ShuffleConfig{
			BitWidth: 0, // Will be calculated after patterns are built
			Rounds:   rounds,
			Seed:     seed,
		}
	}

	// Convert placeholders
	if err := convertPlaceholders(&tomlConfig, config); err != nil {
		return nil, preflight, err
	}

	return config, preflight, nil
}

// convertPlaceholders validates and converts string-based placeholders to PlaceholderType-based.
func convertPlaceholders(tomlConfig *TOMLConfig, config *PhonidConfig) error {
	if tomlConfig.Phonetic.Placeholders != nil {
		config.Placeholders = make(map[PlaceholderType]RuneSet)

		for keyStr, stringChars := range tomlConfig.Phonetic.Placeholders {
			// Validate placeholder key - convert to runes first for proper UTF-8 handling
			keyRunes := []rune(keyStr)
			if len(keyRunes) != 1 {
				return fmt.Errorf(
					"placeholder key '%s' must be single character",
					keyStr,
				)
			}

			placeholderType := PlaceholderType(keyRunes[0])

			// Validate placeholder type is allowed
			if _, isAllowed := AllowedPlaceholders[placeholderType]; !isAllowed {
				return fmt.Errorf(
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
	return nil
}

// ValidatePhonidRC validates a PhonidConfig loaded from RC file with base encoding.
func ValidatePhonidRC(config *PhonidConfig) error {
	if config == nil {
		return errors.New("config cannot be nil")
	}

	return config.Validate()
}

// getValidPlaceholderKeys returns a slice of valid placeholder characters for error messages.
func getValidPlaceholderKeys() []string {
	keys := make([]string, 0, len(AllowedPlaceholders))
	for key := range AllowedPlaceholders {
		keys = append(keys, string(key))
	}
	return keys
}

func readConfigFile(fp string) ([]byte, error) {
	if err := validateConfigPath(fp); err != nil {
		return nil, err
	}
	// #nosec G304 -- filepath is user-provided config path, validated by caller
	data, err := os.ReadFile(fp)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %w", fp, err)
	}
	return data, nil
}

// validateConfigPath checks if the file path is safe to read.
func validateConfigPath(path string) error {
	// Clean path to prevent directory traversal
	cleaned := filepath.Clean(path)

	// Prevent path traversal attacks
	if strings.Contains(cleaned, "..") {
		return errors.New("invalid path: directory traversal not allowed")
	}

	// Validate filename pattern: .phonidrc[.toml] or .*.phonidrc[.toml]
	base := filepath.Base(cleaned)
	if !IsValidPhonidRCFilename(base) {
		return fmt.Errorf("invalid filename: must be '.phonidrc[.toml]' or '.<prefix>.phonidrc[.toml]', got '%s'", base)
	}

	return nil
}

// IsValidPhonidRCFilename checks if filename matches .phonidrc or .<prefix>.phonidrc pattern,
// optionally with a .toml extension.
func IsValidPhonidRCFilename(filename string) bool {
	// Exact match: .phonidrc
	filenameStripped, _ := strings.CutSuffix(filename, RcFileOptSuffix)
	if filenameStripped == RcFileName {
		return true
	}

	// Pattern match: .*.phonidrc
	if strings.HasPrefix(filenameStripped, ".") && strings.HasSuffix(filenameStripped, RcFileName) {
		// Extract prefix between first dot and .phonidrc
		prefix := strings.TrimPrefix(filenameStripped, ".")
		prefix = strings.TrimSuffix(prefix, RcFileName)

		// Ensure prefix is non-empty and doesn't contain path separators or dots
		if prefix != "" && !strings.ContainsAny(prefix, "/\\.:") {
			return true
		}
	}

	return false
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
