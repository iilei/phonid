package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestDigestCommand_FromHexInput(t *testing.T) {
	stdout := &bytes.Buffer{}
	resetCLIState(t)
	rootCmd.SetOut(stdout)

	rootCmd.SetArgs([]string{"digest", "0x186A"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("rootCmd.Execute() error = %v", err)
	}

	if got := strings.TrimSpace(stdout.String()); got != "babab-dodop" {
		t.Fatalf("digest output = %q, want %q", got, "babab-dodop")
	}
}

func TestDigestCommand_FromLargeHexInputUsesModulo(t *testing.T) {
	stdout := &bytes.Buffer{}
	resetCLIState(t)
	rootCmd.SetOut(stdout)

	rootCmd.SetArgs([]string{"digest", "0x100000539"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("rootCmd.Execute() error = %v", err)
	}

	if got := strings.TrimSpace(stdout.String()); got != "babab-bihun" {
		t.Fatalf("digest output = %q, want %q", got, "babab-bihun")
	}
}

func TestDigestCommand_FromTextIsDeterministic(t *testing.T) {
	runDigest := func(t *testing.T, input string) string {
		t.Helper()

		stdout := &bytes.Buffer{}
		resetCLIState(t)
		rootCmd.SetOut(stdout)

		rootCmd.SetArgs([]string{"digest", input})
		if err := rootCmd.Execute(); err != nil {
			t.Fatalf("rootCmd.Execute() error = %v", err)
		}

		return strings.TrimSpace(stdout.String())
	}

	first := runDigest(t, "hello world")
	second := runDigest(t, "hello world")

	if first == "" {
		t.Fatal("digest output is empty")
	}
	if first != second {
		t.Fatalf("digest output is not deterministic: %q != %q", first, second)
	}
}

func TestDigestCommand_InvalidHexInput(t *testing.T) {
	stdout := &bytes.Buffer{}
	resetCLIState(t)
	rootCmd.SetOut(stdout)

	rootCmd.SetArgs([]string{"digest", "0xZZ"})
	err := rootCmd.Execute()
	if err == nil {
		t.Fatal("rootCmd.Execute() error = nil, want error")
	}
	if !strings.Contains(err.Error(), "invalid digest input") {
		t.Fatalf("error = %q, want digest input context", err)
	}
	if got := strings.TrimSpace(stdout.String()); got != "" {
		t.Fatalf("stdout = %q, want empty", got)
	}
}
