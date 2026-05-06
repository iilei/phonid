package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestPreflightCommand_RejectsPreset(t *testing.T) {
	stdout := &bytes.Buffer{}
	resetCLIState(t)
	rootCmd.SetOut(stdout)

	rootCmd.SetArgs([]string{testArgPreset, presetProQuintTiny, "preflight", "--suggest"})
	err := rootCmd.Execute()
	if err == nil {
		t.Fatal("rootCmd.Execute() error = nil, want error")
	}
	if !strings.Contains(err.Error(), "--preset is not supported with preflight") {
		t.Fatalf("error = %q, want preflight preset rejection", err)
	}
	if got := strings.TrimSpace(stdout.String()); got != "" {
		t.Fatalf("stdout = %q, want empty", got)
	}
}
