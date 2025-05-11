package config

import (
	"fmt"
	"os"
	"path/filepath"
)

type Config struct {
	DataDir           string
	CachePath         string
	CacheSizeLimitGB  int
	CircuitNum        int
	TorBundledDir     string
}

func Default() *Config {
	home, _ := os.UserHomeDir()
	base := filepath.Join(home, ".torturbo")
	return &Config{
		DataDir:          base,
		CachePath:        filepath.Join(base, "cache.db"),
		CacheSizeLimitGB: 2,
		CircuitNum:       8,
		TorBundledDir:    "third_party/tor",
	}
}

func (c *Config) EnsureDirs() error {
	if err := os.MkdirAll(c.DataDir, 0o755); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(c.CachePath), 0o755); err != nil {
		return err
	}
	return nil
}

func (c *Config) TorBinaryPath() string {
	return filepath.Join(c.TorBundledDir, "bin", "tor")
}

func (c *Config) TorDataDir() string {
	return filepath.Join(c.DataDir, "tor")
}

func (c *Config) String() string {
	return fmt.Sprintf("DataDir=%s Tor=%s Cache=%s", c.DataDir, c.TorBinaryPath(), c.CachePath)
}
