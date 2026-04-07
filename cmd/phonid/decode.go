package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var decodeCmd = &cobra.Command{
	Use:   "decode [word]",
	Short: "Decode a phonetic identifier back to its numeric value",
	Args:  cobra.ExactArgs(1),
	RunE:  decodeCommand,
}

func decodeCommand(cmd *cobra.Command, args []string) error {
	config, checks, err := loadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	encoder, err := newCLIEncoder(config, checks)
	if err != nil {
		return fmt.Errorf("failed to create encoder: %w", err)
	}

	result, err := encoder.Decode(args[0])
	if err != nil {
		return fmt.Errorf("failed to decode %q: %w", args[0], err)
	}

	_, _ = fmt.Fprintln(cmd.OutOrStdout(), result.String())
	return nil
}
