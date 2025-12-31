package preflight

import (
	"bytes"
	"io"

	"github.com/pelletier/go-toml/v2"
)

// TOMLFormatter implements the Formatter interface for TOML output.
type TOMLFormatter struct {
	encoder *toml.Encoder
}

// NewTOMLFormatter provides toml formatting.
func NewTOMLFormatter() Formatter {
	return &TOMLFormatter{
		encoder: newTOMLFormatter(),
	}
}

// Name returns the format name.
func (f *TOMLFormatter) Name() OutputFormat {
	return FormatTOML
}

// Format writes preflight assertions as TOML to the writer.
func (f *TOMLFormatter) Format(w io.Writer, assertions *AssertionTable) error {
	enc := toml.NewEncoder(w)
	enc.SetIndentTables(true)
	return enc.Encode(assertions)
}

// newTOMLFormatter is the internal constructor.
func newTOMLFormatter() *toml.Encoder {
	buf := bytes.Buffer{}
	enc := toml.NewEncoder(&buf)
	enc.SetIndentTables(true)
	return enc
}
