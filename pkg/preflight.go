package phonid

import "fmt"

// ValidatePreflight checks if preflight tests pass for this encoder
// Performs bidirectional validation: encoding (int->string) and decoding (string->int)
func (p *PhoneticEncoder) ValidatePreflight(checks []PreflightCheck) error {
	if len(checks) == 0 {
		return fmt.Errorf("at least one preflight check is required")
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
