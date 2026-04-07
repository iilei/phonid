package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	phonid "github.com/iilei/phonid/pkg"
)

var (
	cfgFile string
	rootCmd = &cobra.Command{
		Use:   "phonid [number]",
		Short: "Phonetic identifier generator",
		Long: `Phonid generates pronounceable identifiers from numbers using configurable
phonetic patterns and optional Feistel shuffling.`,
		Args: cobra.MaximumNArgs(1),
		RunE: encodeCommand,
	}
)

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default: search for .phonidrc)")
	rootCmd.AddCommand(decodeCmd)
	rootCmd.AddCommand(preflightCmd)
}

func initConfig() {
	if cfgFile != "" {
		// Use config file from flag
		viper.SetConfigFile(cfgFile)
	} else {
		// Search for config in current directory and home directory
		cwd, err := os.Getwd()
		if err == nil {
			viper.AddConfigPath(cwd)
		}
		home, err := os.UserHomeDir()
		if err == nil {
			viper.AddConfigPath(home)
		}
		viper.SetConfigType("toml")
		viper.SetConfigName(".phonidrc")
	}

	// Read config if it exists (ignore error if not found)
	_ = viper.ReadInConfig()
}

func encodeCommand(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return errors.New("number argument required\n\nUsage: phonid [number]\n  or:  phonid preflight --suggest")
	}

	// Parse input number (supports arbitrarily large numbers)
	numStr := args[0]
	number, err := phonid.ParsePositiveInt(numStr)
	if err != nil {
		return fmt.Errorf("invalid number: %s (must be non-negative integer)", numStr)
	}

	// Load config
	config, checks, err := loadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Use strict preflight validation for explicit config files, but allow
	// the built-in default config to work even when no preflight table exists.
	encoder, err := newCLIEncoder(config, checks)
	if err != nil {
		return fmt.Errorf("failed to create encoder: %w", err)
	}

	// Encode the number
	result, err := encoder.Encode(number)
	if err != nil {
		return fmt.Errorf("failed to encode %s: %w", number.String(), err)
	}

	_, _ = fmt.Fprintln(cmd.OutOrStdout(), result)
	return nil
}

func loadConfig() (*phonid.PhonidConfig, []phonid.PreflightCheck, error) {
	configFile := viper.ConfigFileUsed()
	if configFile == "" {
		// No config file found, use defaults with no preflight checks
		return &phonid.DefaultConfig, []phonid.PreflightCheck{}, nil
	}

	// Load from file
	absPath, err := filepath.Abs(configFile)
	if err != nil {
		return nil, nil, err
	}
	return phonid.LoadPhonidRC(absPath)
}

func newCLIEncoder(config *phonid.PhonidConfig, checks []phonid.PreflightCheck) (*phonid.PhoneticEncoder, error) {
	if viper.ConfigFileUsed() == "" {
		return phonid.NewPhoneticEncoderSkipPreflight(config)
	}

	return phonid.NewPhoneticEncoder(config, checks)
}
