package phonid

import (
	"testing"
)

func TestConfigHasSeed(t *testing.T) {
	cfg, err := LoadConfig("")
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	if cfg.Seed == "" {
		t.Error("Config.Seed should not be empty")
	}
}
