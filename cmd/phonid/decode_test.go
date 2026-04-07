package main

import (
	"bytes"
	"io"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/spf13/viper"
)

func TestDecodeCommand_DefaultConfig(t *testing.T) {
	stdout := &bytes.Buffer{}
	resetCLIState(t)
	rootCmd.SetOut(stdout)

	rootCmd.SetArgs([]string{"decode", "babab-bihun"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("rootCmd.Execute() error = %v", err)
	}

	if got := strings.TrimSpace(stdout.String()); got != "1337" {
		t.Fatalf("decode output = %q, want %q", got, "1337")
	}
}

func TestDecodeCommand_WithExplicitConfig(t *testing.T) {
	stdout := &bytes.Buffer{}
	resetCLIState(t)
	rootCmd.SetOut(stdout)

	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("failed to determine test file path")
	}
	configPath := filepath.Join(filepath.Dir(thisFile), "..", "..", "public_presets", ".proquint.phonidrc.toml")
	rootCmd.SetArgs([]string{"--config", configPath, "decode", "babab-bihun"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("rootCmd.Execute() error = %v", err)
	}

	if got := strings.TrimSpace(stdout.String()); got != "1337" {
		t.Fatalf("decode output = %q, want %q", got, "1337")
	}
}

func TestDecodeCommand_InvalidWord(t *testing.T) {
	stdout := &bytes.Buffer{}
	resetCLIState(t)
	rootCmd.SetOut(stdout)

	rootCmd.SetArgs([]string{"decode", "invalid"})
	err := rootCmd.Execute()
	if err == nil {
		t.Fatal("rootCmd.Execute() error = nil, want error")
	}
	if !strings.Contains(err.Error(), "failed to decode \"invalid\"") {
		t.Fatalf("error = %q, want decode context", err)
	}
	if got := strings.TrimSpace(stdout.String()); got != "" {
		t.Fatalf("stdout = %q, want empty", got)
	}
}

func resetCLIState(t *testing.T) {
	t.Helper()

	viper.Reset()
	rootCmd.SetArgs(nil)
	rootCmd.SetOut(io.Discard)
	rootCmd.SetErr(io.Discard)
	rootCmd.SilenceErrors = true
	rootCmd.SilenceUsage = true
	cfgFile = ""
	presetName = ""

	if f := rootCmd.PersistentFlags().Lookup("config"); f != nil {
		f.Changed = false
	}
	if f := rootCmd.PersistentFlags().Lookup("preset"); f != nil {
		f.Changed = false
	}
}
