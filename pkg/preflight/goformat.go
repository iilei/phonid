package preflight

import (
	"errors"
	"fmt"
	"io"
	"strings"
)

// GoFormatter implements the Formatter interface for Go code output.
// Note: This formatter only outputs preflight assertions, not configuration.
// For full configuration including shuffle settings, use the TOML formatter.
//
// Shuffle Limitations:
//
//	Patterns with capacity > 18,446,744,073,709,551,615 (uint64 max) cannot be shuffled.
//	When using such patterns, shuffle configuration should be omitted entirely.
type GoFormatter struct{}

// NewGoFormatter creates a new Go code formatter.
func NewGoFormatter() Formatter {
	return &GoFormatter{}
}

// Name returns the format name.
func (f *GoFormatter) Name() OutputFormat {
	return FormatGo
}

// Format writes preflight assertions as Go code to the writer.
func (f *GoFormatter) Format(w io.Writer, assertions *AssertionTable) error {
	if assertions == nil || len(*assertions) == 0 {
		return errors.New("no assertions provided")
	}

	var sb strings.Builder

	// Package declaration
	sb.WriteString("package main\n\n")

	// Imports
	sb.WriteString("import (\n")
	sb.WriteString("\t\"github.com/iilei/phonid/pkg\"\n")
	sb.WriteString(")\n\n")

	// Function to get preflight assertions
	sb.WriteString("// GetPhonidPreflightAssertions returns the generated test assertions.\n")
	sb.WriteString("func GetPhonidPreflightAssertions() preflight.AssertionTable {\n")
	sb.WriteString("\treturn preflight.AssertionTable{\n")

	// Generate each assertion
	for _, assertion := range *assertions {
		sb.WriteString("\t\t{\n")

		// Input field
		if assertion.Input != nil {
			inputStr := assertion.Input.Value.String()
			sb.WriteString(
				fmt.Sprintf("\t\t\tInput:  &pkg.TomlPositiveInt{Value: pkg.NewPositiveInt(%s)},\n", inputStr),
			)
		}

		// Output field
		sb.WriteString(fmt.Sprintf("\t\t\tOutput: %q,\n", assertion.Output))

		// Comment field (if present)
		if assertion.Comment != "" {
			sb.WriteString(fmt.Sprintf("\t\t\tComment: %q,\n", assertion.Comment))
		}

		sb.WriteString("\t\t},\n")
	}

	sb.WriteString("\t}\n")
	sb.WriteString("}\n")

	_, err := w.Write([]byte(sb.String()))
	return err
}
