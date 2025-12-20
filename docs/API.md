package phonid // import "github.com/iilei/phonid/pkg"


TYPES

type Config struct {
	Seed      string `toml:"seed"`
	Rounds    int8   `toml:"rounds"`
	SeedValue uint64 `toml:"-"`
}

func LoadConfig(configPath string) (*Config, error)

