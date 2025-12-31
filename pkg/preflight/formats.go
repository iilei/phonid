// Package preflight represents preflight checks and code generation.
package preflight

import (
	"fmt"
	"io"
)

const (
	FormatTOML OutputFormat = "toml"
	FormatGo   OutputFormat = "go"
)

type (
	// OutputFormat represents a supported output format.
	OutputFormat string

	// Formatter handles rendering suggestions in a specific format.
	Formatter interface {
		Format(w io.Writer, suggestions *AssertionTable) error
		Name() OutputFormat
	}

	// FormatterRegistry manages available formatters.
	FormatterRegistry struct {
		formatters map[OutputFormat]Formatter
	}
)

// NewFormatterRegistry creates a registry with all built-in formatters.
func NewFormatterRegistry() *FormatterRegistry {
	registry := &FormatterRegistry{
		formatters: make(map[OutputFormat]Formatter),
	}
	registry.Register(NewTOMLFormatter())
	return registry
}

// Register adds a formatter to the registry.
func (r *FormatterRegistry) Register(formatter Formatter) {
	r.formatters[formatter.Name()] = formatter
}

// Get retrieves a formatter by name.
func (r *FormatterRegistry) Get(format OutputFormat) (Formatter, error) {
	formatter, exists := r.formatters[format]
	if !exists {
		return nil, fmt.Errorf("unsupported format: %s (available: toml, go)", format)
	}
	return formatter, nil
}

// AvailableFormats returns a list of all registered format names.
func (r *FormatterRegistry) AvailableFormats() []OutputFormat {
	formats := make([]OutputFormat, 0, len(r.formatters))
	for name := range r.formatters {
		formats = append(formats, name)
	}
	return formats
}
