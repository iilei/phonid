package main

import (
	"bytes"
	"strings"
	"testing"
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

func TestEncodeCommand_PresetProquintMatchesStandard(t *testing.T) {
	stdout := &bytes.Buffer{}
	resetCLIState(t)
	rootCmd.SetOut(stdout)

	rootCmd.SetArgs([]string{"--preset", "proquint", "100"})
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

	rootCmd.SetArgs([]string{"--preset", "proquint-tiny", "70000"})
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

	rootCmd.SetArgs([]string{"--preset", "unknown-preset", "1337"})
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

func TestEncodeCommand_ConfigAndPresetMutuallyExclusive(t *testing.T) {
	stdout := &bytes.Buffer{}
	resetCLIState(t)
	rootCmd.SetOut(stdout)

	configPath := "my-config"
	rootCmd.SetArgs([]string{"--config", configPath, "--preset", "proquint", "1337"})
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
