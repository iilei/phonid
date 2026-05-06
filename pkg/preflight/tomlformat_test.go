package preflight_test

import (
	"bytes"
	"strings"
	"testing"

	p "github.com/iilei/phonid/pkg"
	"github.com/iilei/phonid/pkg/preflight"
)

const (
	tomlFmtOutputBac = "bac"
	tomlFmtOutputDae = "dae"
	tomlFmtLower     = "Lower boundary"
	tomlFmtMid       = "Mid-range"
)

func TestTOMLFormatter_Name(t *testing.T) {
	formatter := preflight.NewTOMLFormatter()
	if got := formatter.Name(); got != preflight.FormatTOML {
		t.Errorf("TOMLFormatter.Name() = %v, want %v", got, preflight.FormatTOML)
	}
}

func TestTOMLFormatter_Format(t *testing.T) {
	assertions := &preflight.AssertionTable{
		{Input: &p.TomlPositiveInt{Value: p.NewPositiveInt(0)}, Output: tomlFmtOutputBac, Comment: tomlFmtLower},
		{Input: &p.TomlPositiveInt{Value: p.NewPositiveInt(10)}, Output: tomlFmtOutputDae, Comment: tomlFmtMid},
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
	// Input values are now marshaled as strings via TomlPositiveInt.MarshalText
	if !strings.Contains(result, "input = '0'") && !strings.Contains(result, "input = \"0\"") {
		t.Errorf("missing first input value, got:\n%s", result)
	}
	// go-toml uses single quotes for strings, both 'bac' and "bac" are valid TOML
	if !strings.Contains(result, "output = '"+tomlFmtOutputBac+"'") &&
		!strings.Contains(result, "output = \""+tomlFmtOutputBac+"\"") {
		t.Errorf("missing first output value, got:\n%s", result)
	}
	if !strings.Contains(result, "input = '10'") && !strings.Contains(result, "input = \"10\"") {
		t.Errorf("missing second input value, got:\n%s", result)
	}
	if !strings.Contains(result, "output = '"+tomlFmtOutputDae+"'") &&
		!strings.Contains(result, "output = \""+tomlFmtOutputDae+"\"") {
		t.Errorf("missing second output value, got:\n%s", result)
	}
	// Comments should NOT be in Format output (only in SuggestConfig)
	if strings.Contains(result, "# "+tomlFmtLower) {
		t.Error("Format() should not include inline comments")
	}
}
