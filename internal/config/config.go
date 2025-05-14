package config

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type Config struct {
	DataDir          string // 數據存儲目錄
	CachePath        string // 緩存文件路徑
	CacheSizeLimitGB int    // 緩存大小限制 (GB)
	CircuitNum       int    // 最大 Circuit 數量
	TorBundledDir    string // Tor 二進制文件目錄
	once           sync.Once
	cachedTorPath  string
	cachedTorError error
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

// 在 TorBundledDir 底下遞迴搜尋 tor 或 tor.exe 執行檔並返回路徑
func (c *Config) searchTorExecutable() (string, error) {
	var found string
	err := filepath.WalkDir(c.TorBundledDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			// 如果遍歷失敗，返回錯誤終止
			return err
		}
		if d.IsDir() {
			return nil
		}
		name := strings.ToLower(d.Name())
		if name == "tor" || name == "tor.exe" {
			found = path
			// 用 io.EOF 作為結束 WalkDir 的信號
			return io.EOF
		}
		return nil
	})
	if err != nil && err != io.EOF {
		return "", err
	}
	return found, nil
}

// 快取 Tor 執行檔路徑
func (c *Config) TorBinaryPath() (string, error) {
	c.once.Do(func() {
		// 先嘗試預設 bin/tor
		def := filepath.Join(c.TorBundledDir, "bin", "tor")
		if info, err := os.Stat(def); err == nil && !info.IsDir() {
			c.cachedTorPath = def
			return
		}
		// 否則遞迴搜尋
		if path, err := c.searchTorExecutable(); err != nil {
			// 若搜尋出現錯誤，回退到預設
			c.cachedTorError = err
			c.cachedTorPath = def
		} else if path != "" {
			c.cachedTorPath = path
		} else {
			// 未找到，使用預設
			c.cachedTorPath = def
		}
	})
	return c.cachedTorPath, c.cachedTorError
}

func (c *Config) TorDataDir() string {
	return filepath.Join(c.DataDir, "tor")
}

func (c *Config) String() string {
	torPath, err := c.TorBinaryPath()
	if err != nil {
		torPath = "ERROR:" + err.Error()
	}
	return fmt.Sprintf("DataDir=%s Tor=%s Cache=%s", c.DataDir, torPath, c.CachePath)
}
