package preflight_test

import (
	"bytes"
	"strings"
	"testing"

	p "github.com/iilei/phonid/pkg"
	"github.com/iilei/phonid/pkg/preflight"
)

func TestGoFormatter_Name(t *testing.T) {
	formatter := preflight.NewGoFormatter()
	if got := formatter.Name(); got != preflight.FormatGo {
		t.Errorf("GoFormatter.Name() = %v, want %v", got, preflight.FormatGo)
	}
}

func TestGoFormatter_Format(t *testing.T) {
	assertions := &preflight.AssertionTable{
		{
			Input:   &p.TomlPositiveInt{Value: p.NewPositiveInt(0)},
			Output:  "bac",
			Comment: "Lower boundary",
		},
		{
			Input:   &p.TomlPositiveInt{Value: p.NewPositiveInt(10)},
			Output:  "dae",
			Comment: "Mid-range",
		},
	}

	formatter := preflight.NewGoFormatter()
	var buf bytes.Buffer
	err := formatter.Format(&buf, assertions)
	if err != nil {
		t.Fatalf("GoFormatter.Format() error = %v", err)
	}

	result := buf.String()

	// Check for expected Go code elements
	expectedStrings := []string{
		"package main",
		"import",
		"github.com/iilei/phonid/pkg",
		"func GetPhonidPreflightAssertions()",
		"preflight.AssertionTable",
		"pkg.NewPositiveInt(0)",
		"pkg.NewPositiveInt(10)",
		`Output: "bac"`,
		`Output: "dae"`,
		`Comment: "Lower boundary"`,
		`Comment: "Mid-range"`,
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(result, expected) {
			t.Errorf("GoFormatter.Format() result missing expected string: %q\nGot:\n%s", expected, result)
		}
	}
}

func TestGoFormatter_Format_EmptyAssertions(t *testing.T) {
	formatter := preflight.NewGoFormatter()
	var buf bytes.Buffer

	// Test with nil
	err := formatter.Format(&buf, nil)
	if err == nil {
		t.Error("GoFormatter.Format() with nil assertions should return error")
	}

	// Test with empty table
	emptyTable := &preflight.AssertionTable{}
	buf.Reset()
	err = formatter.Format(&buf, emptyTable)
	if err == nil {
		t.Error("GoFormatter.Format() with empty assertions should return error")
	}
}

func TestGoFormatter_Format_WithoutComment(t *testing.T) {
	assertions := &preflight.AssertionTable{
		{
			Input:  &p.TomlPositiveInt{Value: p.NewPositiveInt(42)},
			Output: "test",
		},
	}

	formatter := preflight.NewGoFormatter()
	var buf bytes.Buffer
	err := formatter.Format(&buf, assertions)
	if err != nil {
		t.Fatalf("GoFormatter.Format() error = %v", err)
	}

	result := buf.String()

	if !strings.Contains(result, "pkg.NewPositiveInt(42)") {
		t.Errorf("missing input value in output")
	}
	if !strings.Contains(result, `Output: "test"`) {
		t.Errorf("missing output value in output")
	}
	// Should not contain Comment field when comment is empty
	if strings.Contains(result, `Comment: ""`) {
		t.Errorf("should not include empty Comment field")
	}
}
