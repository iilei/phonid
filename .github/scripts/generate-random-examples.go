package main

import (
	"fmt"
	"math/big"
	"os"
	"sort"
	"strings"
	"unicode/utf8"

	phonid "github.com/iilei/phonid/pkg"
	"github.com/iilei/phonid/pkg/preflight"
)

type tableRow struct {
	number   *big.Int // Use big.Int to handle arbitrarily large numbers
	encoding string
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <config-file>\n", os.Args[0])
		os.Exit(1)
	}

	configPath := os.Args[1]

	// Load config using LoadPhonidRCLenient (allows preflight failures)
	cfg, _, err := phonid.LoadPhonidRCLenient(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	// Create encoder
	encoder, err := phonid.NewPhoneticEncoderLenient(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating encoder: %v\n", err)
		os.Exit(1)
	}

	// Generate 9 random examples and collect them
	rows := make([]tableRow, 0, 9)
	for range 9 {
		assertion, err := preflight.GetRandom(encoder)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error generating random: %v\n", err)
			os.Exit(1)
		}
		rows = append(rows, tableRow{
			number:   assertion.Input.Value.BigInt(),
			encoding: assertion.Output,
		})
	}

	// Sort by number ascending
	sort.Slice(rows, func(i, j int) bool {
		return rows[i].number.Cmp(rows[j].number) < 0
	})

	// Calculate column widths
	maxNumberWidth := len("Number:")
	maxEncodingWidth := len("Encoding")
	for _, row := range rows {
		numStr := row.number.String()
		if len(numStr) > maxNumberWidth {
			maxNumberWidth = len(numStr)
		}
		// Add 2 for the backticks around encoding
		encodingLen := utf8.RuneCountInString(row.encoding) + 2
		if encodingLen > maxEncodingWidth {
			maxEncodingWidth = encodingLen
		}
	}

	// Print table header
	fmt.Printf("| %-*s | %-*s |\n", maxNumberWidth, "Number", maxEncodingWidth, "Encoding")
	fmt.Printf("|-%s-:|-%s-|\n", strings.Repeat("-", maxNumberWidth-1), strings.Repeat("-", maxEncodingWidth))

	// Print table rows
	for _, row := range rows {
		fmt.Printf("| %-*s | %-*s |\n", maxNumberWidth, row.number.String(), maxEncodingWidth, fmt.Sprintf("`%s`", row.encoding))
	}
}
