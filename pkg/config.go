package phonid

import (
	_ "embed"
	"errors"
	"fmt"
	"hash/fnv"
	"os"
	"strings"

	"github.com/pelletier/go-toml/v2"
)

//go:embed data/base_config.toml
var baseConfig string

type Config struct {
	Seed      string `toml:"seed"`
	Rounds    int8   `toml:"rounds"`
	SeedValue uint64 `toml:"-"`
}

func LoadConfig(configPath string) (*Config, error) {
	doc := baseConfig

	if configPath != "" {
		override, err := os.ReadFile(configPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
		doc = string(override)
	}

	r := strings.NewReader(doc)
	d := toml.NewDecoder(r)
	d.DisallowUnknownFields()

	cfg := &Config{}
	err := d.Decode(cfg)

	if err != nil {
		var details *toml.StrictMissingError
		if errors.As(err, &details) {
			fmt.Println(details.String())
		}
		return nil, err
	}

	cfg.SeedValue = hashSeed(cfg.Seed)
	return cfg, nil
}

func hashSeed(s string) uint64 {
	h := fnv.New64a()
	h.Write([]byte(s))
	return h.Sum64()
}
