package main

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"math/big"
	"strings"

	"github.com/spf13/cobra"

	phonid "github.com/iilei/phonid/pkg"
)

var digestCmd = &cobra.Command{
	Use:   "digest [value-or-text]",
	Short: "Generate a one-way phonetic digest",
	Long: `Generate a one-way phonetic digest.

Input rules:
  - If input starts with 0x/0X, it is treated as a hexadecimal number.
  - Otherwise input is treated as raw text and hashed with SHA-256.

The resulting value is reduced modulo active config capacity before encoding.
Digest output is intentionally not reversible.`,
	Args: cobra.ExactArgs(1),
	RunE: digestCommand,
}

func digestCommand(cmd *cobra.Command, args []string) error {
	config, checks, err := loadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	encoder, err := newCLIEncoder(config, checks)
	if err != nil {
		return fmt.Errorf("failed to create encoder: %w", err)
	}

	inputValue, err := digestInputToBigInt(args[0])
	if err != nil {
		return fmt.Errorf("invalid digest input: %w", err)
	}

	modBase, err := digestModuloBase(encoder)
	if err != nil {
		return err
	}

	reduced := new(big.Int).Mod(inputValue, modBase)
	result, err := encoder.Encode(phonid.NewPositiveIntFromBig(reduced))
	if err != nil {
		return fmt.Errorf("failed to generate digest: %w", err)
	}

	_, _ = fmt.Fprintln(cmd.OutOrStdout(), result)
	return nil
}

func digestInputToBigInt(input string) (*big.Int, error) {
	if strings.HasPrefix(input, "0x") || strings.HasPrefix(input, "0X") {
		hexDigits := input[2:]
		if hexDigits == "" {
			return nil, errors.New("hex value is empty")
		}

		parsed := new(big.Int)
		if _, ok := parsed.SetString(hexDigits, cliHexBase); !ok {
			return nil, fmt.Errorf("invalid hexadecimal value: %s", input)
		}

		return parsed, nil
	}

	hash := sha256.Sum256([]byte(input))
	return new(big.Int).SetBytes(hash[:]), nil
}

func digestModuloBase(encoder *phonid.PhoneticEncoder) (*big.Int, error) {
	info := encoder.GetPatternInfo()
	if len(info) == 0 {
		return nil, errors.New("failed to generate digest: no patterns available")
	}

	largest := info[len(info)-1]
	base := new(big.Int).Add(largest.TrueCapacity.BigInt(), big.NewInt(1))
	if base.Sign() <= 0 {
		return nil, errors.New("failed to generate digest: invalid pattern capacity")
	}

	return base, nil
}
