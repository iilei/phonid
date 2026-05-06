package main

import (
	"bytes"
	"math/big"
	"strings"
	"testing"
)

const (
	testArgPreset = "--preset"
	testArgConfig = "--config"
	testArgDecode = "decode"

	testEncoded1337 = "babab-bihun"
)

func TestEncodeCommand_DefaultDecimalInput(t *testing.T) {
	stdout := &bytes.Buffer{}
	resetCLIState(t)
	rootCmd.SetOut(stdout)

	rootCmd.SetArgs([]string{"1328"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("rootCmd.Execute() error = %v", err)
	}

	if got := strings.TrimSpace(stdout.String()); got != "babab-bihub" {
		t.Fatalf("encode output = %q, want %q", got, "babab-bihub")
	}
}

func TestEncodeCommand_AcceptsLowercaseHexPrefix(t *testing.T) {
	stdout := &bytes.Buffer{}
	resetCLIState(t)
	rootCmd.SetOut(stdout)

	rootCmd.SetArgs([]string{"0x70000539"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("rootCmd.Execute() error = %v", err)
	}

	if got := strings.TrimSpace(stdout.String()); got != "labab-bihun" {
		t.Fatalf("encode output = %q, want %q", got, "labab-bihun")
	}
}

func TestEncodeCommand_AcceptsUppercaseHexPrefix(t *testing.T) {
	stdout := &bytes.Buffer{}
	resetCLIState(t)
	rootCmd.SetOut(stdout)

	rootCmd.SetArgs([]string{"0X532"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("rootCmd.Execute() error = %v", err)
	}

	if got := strings.TrimSpace(stdout.String()); got != "babab-bihuf" {
		t.Fatalf("encode output = %q, want %q", got, "babab-bihuf")
	}
}

func TestEncodeCommand_AcceptsLowercaseBinaryPrefix(t *testing.T) {
	stdout := &bytes.Buffer{}
	resetCLIState(t)
	rootCmd.SetOut(stdout)

	rootCmd.SetArgs([]string{"0o2471"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("rootCmd.Execute() error = %v", err)
	}

	if got := strings.TrimSpace(stdout.String()); got != testEncoded1337 {
		t.Fatalf("encode output = %q, want %q", got, testEncoded1337)
	}
}

func TestEncodeCommand_AcceptsLowercaseOctalPrefix(t *testing.T) {
	stdout := &bytes.Buffer{}
	resetCLIState(t)
	rootCmd.SetOut(stdout)

	rootCmd.SetArgs([]string{"0o52"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("rootCmd.Execute() error = %v", err)
	}

	if got := strings.TrimSpace(stdout.String()); got != "babab-babop" {
		t.Fatalf("encode output = %q, want %q", got, "babab-babop")
	}
}

func TestEncodeCommand_InvalidHexInput(t *testing.T) {
	stdout := &bytes.Buffer{}
	resetCLIState(t)
	rootCmd.SetOut(stdout)

	rootCmd.SetArgs([]string{"0xZZ"})
	err := rootCmd.Execute()
	if err == nil {
		t.Fatal("rootCmd.Execute() error = nil, want error")
	}
	if !strings.Contains(err.Error(), "invalid number: 0xZZ") {
		t.Fatalf("error = %q, want invalid number context", err)
	}
	if got := strings.TrimSpace(stdout.String()); got != "" {
		t.Fatalf("stdout = %q, want empty", got)
	}
}

func TestEncodeCommand_MissingHexDigits(t *testing.T) {
	stdout := &bytes.Buffer{}
	resetCLIState(t)
	rootCmd.SetOut(stdout)

	rootCmd.SetArgs([]string{"0x"})
	err := rootCmd.Execute()
	if err == nil {
		t.Fatal("rootCmd.Execute() error = nil, want error")
	}
	if !strings.Contains(err.Error(), "invalid number: 0x") {
		t.Fatalf("error = %q, want invalid number context", err)
	}
	if got := strings.TrimSpace(stdout.String()); got != "" {
		t.Fatalf("stdout = %q, want empty", got)
	}
}

func TestEncodeCommand_InvalidBinaryInput(t *testing.T) {
	stdout := &bytes.Buffer{}
	resetCLIState(t)
	rootCmd.SetOut(stdout)

	rootCmd.SetArgs([]string{"0b102"})
	err := rootCmd.Execute()
	if err == nil {
		t.Fatal("rootCmd.Execute() error = nil, want error")
	}
	if !strings.Contains(err.Error(), "invalid number: 0b102") {
		t.Fatalf("error = %q, want invalid number context", err)
	}
	if got := strings.TrimSpace(stdout.String()); got != "" {
		t.Fatalf("stdout = %q, want empty", got)
	}
}

func TestEncodeCommand_InvalidOctalInput(t *testing.T) {
	stdout := &bytes.Buffer{}
	resetCLIState(t)
	rootCmd.SetOut(stdout)

	rootCmd.SetArgs([]string{"0o89"})
	err := rootCmd.Execute()
	if err == nil {
		t.Fatal("rootCmd.Execute() error = nil, want error")
	}
	if !strings.Contains(err.Error(), "invalid number: 0o89") {
		t.Fatalf("error = %q, want invalid number context", err)
	}
	if got := strings.TrimSpace(stdout.String()); got != "" {
		t.Fatalf("stdout = %q, want empty", got)
	}
}

func TestEncodeCommand_PresetProquintMatchesStandard(t *testing.T) {
	stdout := &bytes.Buffer{}
	resetCLIState(t)
	rootCmd.SetOut(stdout)

	rootCmd.SetArgs([]string{testArgPreset, presetProQuint, "100"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("rootCmd.Execute() error = %v", err)
	}

	if got := strings.TrimSpace(stdout.String()); got != "babab-badoh" {
		t.Fatalf("encode output = %q, want %q", got, "babab-badoh")
	}
}

func TestEncodeCommand_PresetProquintTinyRejectsOverflow(t *testing.T) {
	stdout := &bytes.Buffer{}
	resetCLIState(t)
	rootCmd.SetOut(stdout)

	rootCmd.SetArgs([]string{testArgPreset, presetProQuintTiny, "70000"})
	err := rootCmd.Execute()
	if err == nil {
		t.Fatal("rootCmd.Execute() error = nil, want error")
	}
	if !strings.Contains(err.Error(), "exceeds largest pattern capacity") {
		t.Fatalf("error = %q, want capacity context", err)
	}
	if got := strings.TrimSpace(stdout.String()); got != "" {
		t.Fatalf("stdout = %q, want empty", got)
	}
}

func TestEncodeCommand_InvalidPreset(t *testing.T) {
	stdout := &bytes.Buffer{}
	resetCLIState(t)
	rootCmd.SetOut(stdout)

	rootCmd.SetArgs([]string{testArgPreset, "unknown-preset", "1337"})
	err := rootCmd.Execute()
	if err == nil {
		t.Fatal("rootCmd.Execute() error = nil, want error")
	}
	if !strings.Contains(err.Error(), "unknown preset") {
		t.Fatalf("error = %q, want unknown preset context", err)
	}
	if got := strings.TrimSpace(stdout.String()); got != "" {
		t.Fatalf("stdout = %q, want empty", got)
	}
}

func TestEncodeCommand_PresetProquintSHA256RoundTripMaxValue(t *testing.T) {
	stdout := &bytes.Buffer{}
	resetCLIState(t)
	rootCmd.SetOut(stdout)

	maxHex := "0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"
	rootCmd.SetArgs([]string{testArgPreset, presetProQuintSHA, maxHex})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("rootCmd.Execute() error = %v", err)
	}

	encoded := strings.TrimSpace(stdout.String())
	if encoded == "" {
		t.Fatal("encode output is empty")
	}
	if strings.Count(encoded, "-") != 15 {
		t.Fatalf("encoded separator count = %d, want 15", strings.Count(encoded, "-"))
	}

	decodedOut := &bytes.Buffer{}
	resetCLIState(t)
	rootCmd.SetOut(decodedOut)
	rootCmd.SetArgs([]string{testArgPreset, presetProQuintSHA, testArgDecode, encoded})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("decode round-trip error = %v", err)
	}

	decoded := strings.TrimSpace(decodedOut.String())
	want := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 256), big.NewInt(1)).String()
	if decoded != want {
		t.Fatalf("decoded value = %q, want %q", decoded, want)
	}
}

func TestEncodeCommand_PresetProquintSHA256RejectsOverflow(t *testing.T) {
	stdout := &bytes.Buffer{}
	resetCLIState(t)
	rootCmd.SetOut(stdout)

	overflow := new(big.Int).Lsh(big.NewInt(1), 256).String()
	rootCmd.SetArgs([]string{testArgPreset, presetProQuintSHA, overflow})
	err := rootCmd.Execute()
	if err == nil {
		t.Fatal("rootCmd.Execute() error = nil, want error")
	}
	if !strings.Contains(err.Error(), "exceeds largest pattern capacity") {
		t.Fatalf("error = %q, want capacity context", err)
	}
	if got := strings.TrimSpace(stdout.String()); got != "" {
		t.Fatalf("stdout = %q, want empty", got)
	}
}

func TestEncodeCommand_ConfigAndPresetMutuallyExclusive(t *testing.T) {
	stdout := &bytes.Buffer{}
	resetCLIState(t)
	rootCmd.SetOut(stdout)

	configPath := "my-config"
	rootCmd.SetArgs([]string{testArgConfig, configPath, testArgPreset, presetProQuint, "1337"})
	err := rootCmd.Execute()
	if err == nil {
		t.Fatal("rootCmd.Execute() error = nil, want error")
	}
	if !strings.Contains(err.Error(), "config") || !strings.Contains(err.Error(), "preset") {
		t.Fatalf("error = %q, want config/preset exclusivity context", err)
	}
	if got := strings.TrimSpace(stdout.String()); got != "" {
		t.Fatalf("stdout = %q, want empty", got)
	}
}
