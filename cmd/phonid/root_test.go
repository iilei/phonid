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
