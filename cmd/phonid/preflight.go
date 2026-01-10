package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	phonid "github.com/iilei/phonid/pkg"
	"github.com/iilei/phonid/pkg/preflight"
)

var (
	suggestFlag      bool
	noBoundariesFlag bool
	formatFlag       string

	preflightCmd = &cobra.Command{
		Use:   "preflight [numbers...]",
		Short: "Generate or validate preflight checks",
		Long: `Generate suggested preflight checks or validate existing ones.

Examples:
  phonid preflight --suggest                     # Generate boundary checks (0, 50%, max)
  phonid preflight --suggest 42 1337             # Add custom value checks + boundaries
  phonid preflight --suggest --no-boundaries 42  # Custom values only
  phonid preflight --suggest --format toml       # Output as TOML (default)
  phonid preflight --suggest --format go         # Output as Go test code`,
		RunE: preflightCommand,
	}
)

func init() {
	preflightCmd.Flags().BoolVar(&suggestFlag, "suggest", false, "generate suggested preflight checks")
	preflightCmd.Flags().
		BoolVar(&noBoundariesFlag, "no-boundaries", false, "exclude boundary checks (only with custom values)")
	preflightCmd.Flags().StringVar(&formatFlag, "format", "toml", "output format: toml, go")
}

func preflightCommand(cmd *cobra.Command, args []string) error {
	if !suggestFlag {
		return errors.New("only --suggest mode is currently supported")
	}

	// Parse custom values from args
	customValues := make([]phonid.PositiveInt, 0, len(args))
	for _, arg := range args {
		num, err := phonid.ParsePositiveInt(arg)
		if err != nil {
			return fmt.Errorf("invalid number: %s", arg)
		}
		customValues = append(customValues, num)
	}

	// Load config (lenient mode for suggestion generation)
	config, _, err := loadConfigLenient()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Create encoder
	encoder, err := phonid.NewPhoneticEncoderLenient(config)
	if err != nil {
		return fmt.Errorf("failed to create encoder: %w", err)
	}

	// Determine whether to include boundaries
	includeBoundaries := !noBoundariesFlag || len(customValues) == 0

	// Generate suggestions
	var suggestions preflight.AssertionTable
	if len(customValues) > 0 || !includeBoundaries {
		suggestions, err = preflight.GenerateSuggestionsWithCustom(encoder, customValues, includeBoundaries)
	} else {
		suggestions, err = preflight.GenerateSuggestions(encoder)
	}
	if err != nil {
		return fmt.Errorf("failed to generate suggestions: %w", err)
	}

	// Format output
	if formatFlag == "toml" {
		// For TOML format, generate full config with the suggestions
		output, err := preflight.FormatTOMLWithSuggestions(encoder, config, &suggestions)
		if err != nil {
			return fmt.Errorf("failed to generate config: %w", err)
		}
		fmt.Print(output)
		return nil
	}

	// For other formats, use the formatter registry
	registry := preflight.NewFormatterRegistry()
	formatter, err := registry.Get(preflight.OutputFormat(formatFlag))
	if err != nil {
		return fmt.Errorf("invalid format: %w", err)
	}

	if err := formatter.Format(os.Stdout, &suggestions); err != nil {
		return fmt.Errorf("failed to format output: %w", err)
	}

	return nil
}

func loadConfigLenient() (*phonid.PhonidConfig, []phonid.PreflightCheck, error) {
	configFile := cfgFile
	if configFile == "" {
		configFile = viper.ConfigFileUsed()
	}

	if configFile != "" {
		absPath, err := filepath.Abs(configFile)
		if err != nil {
			return nil, nil, err
		}
		config, checks, err := phonid.LoadPhonidRCLenient(absPath)
		if err != nil {
			return nil, nil, err
		}

		// If config is effectively empty (no patterns), use defaults
		if len(config.Patterns) == 0 {
			return &phonid.DefaultConfig, nil, nil
		}

		return config, checks, nil
	}

	// Use default config
	return &phonid.DefaultConfig, nil, nil
}
