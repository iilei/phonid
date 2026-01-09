package preflight_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/iilei/phonid/pkg/preflight"
)

func TestTOMLFormatter_Name(t *testing.T) {
	formatter := preflight.NewTOMLFormatter()
	if got := formatter.Name(); got != preflight.FormatTOML {
		t.Errorf("TOMLFormatter.Name() = %v, want %v", got, preflight.FormatTOML)
	}
}

func TestTOMLFormatter_Format(t *testing.T) {
	assertions := &preflight.AssertionTable{
		{Input: 0, Output: "bac", Comment: "Lower boundary"},
		{Input: 10, Output: "dae", Comment: "Mid-range"},
	}

	formatter := preflight.NewTOMLFormatter()
	var buf bytes.Buffer
	err := formatter.Format(&buf, assertions)
	if err != nil {
		t.Fatalf("TOMLFormatter.Format() error = %v", err)
	}

	result := buf.String()

	// The Format() method encodes the AssertionTable directly, not as [[preflight]]
	// It creates array of tables format
	if !strings.Contains(result, "input = 0") {
		t.Errorf("missing first input value, got:\n%s", result)
	}
	// go-toml uses single quotes for strings, both 'bac' and "bac" are valid TOML
	if !strings.Contains(result, "output = 'bac'") && !strings.Contains(result, "output = \"bac\"") {
		t.Errorf("missing first output value, got:\n%s", result)
	}
	if !strings.Contains(result, "input = 10") {
		t.Errorf("missing second input value, got:\n%s", result)
	}
	if !strings.Contains(result, "output = 'dae'") && !strings.Contains(result, "output = \"dae\"") {
		t.Errorf("missing second output value, got:\n%s", result)
	}
	// Comments should NOT be in Format output (only in SuggestConfig)
	if strings.Contains(result, "# Lower boundary") {
		t.Error("Format() should not include inline comments")
	}
}
