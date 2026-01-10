package phonid

import (
	"errors"
	"fmt"
)

// ErrNoPreflightChecks is returned when no preflight checks are provided.
var ErrNoPreflightChecks = errors.New("config must include at least one [[preflight]] check\n\n" +
	"Example:\n" +
	"  [[preflight]]\n" +
	"  input = 0\n" +
	"  output = \"babab\"\n\n" +
	"Hint: Run 'phonid preflight --suggest' to generate recommended checks")

// ValidatePreflight checks if preflight tests pass for this encoder
// Performs bidirectional validation: encoding (int->string) and decoding (string->int).
func (p *PhoneticEncoder) ValidatePreflight(checks []PreflightCheck) error {
	if len(checks) == 0 {
		return ErrNoPreflightChecks
	}

	for i, check := range checks {
		// Test encoding
		encoded, err := p.Encode(check.Input)
		if err != nil {
			return fmt.Errorf("preflight[%d]: encode(%d) failed: %w", i, check.Input, err)
		}
		if encoded != check.Output {
			return fmt.Errorf("preflight[%d]: encode(%d) = %q, want %q",
				i, check.Input, encoded, check.Output)
		}

		// Test decoding (implicit round-trip)
		decoded, err := p.Decode(check.Output)
		if err != nil {
			return fmt.Errorf("preflight[%d]: decode(%q) failed: %w",
				i, check.Output, err)
		}
		if decoded != int(check.Input) {
			return fmt.Errorf("preflight[%d]: decode(%q) = %d, want %d",
				i, check.Output, decoded, check.Input)
		}
	}

	return nil
}
