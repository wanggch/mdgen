package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type Config struct {
	ListenAddr  string `json:"listen_addr"`
	OutputDir   string `json:"output_dir"`
	APIKey      string `json:"api_key"`
	LogLevel    string `json:"log_level"`
	MaxBodySize int64  `json:"max_body_size"`
}

const (
	defaultListenAddr  = "127.0.0.1:27123"
	defaultOutputDir   = "~/Documents/notes"
	defaultLogLevel    = "info"
	defaultMaxBodySize = int64(1048576)
)

func Load(configPath string) (Config, error) {
	cfg := Config{
		ListenAddr:  defaultListenAddr,
		OutputDir:   defaultOutputDir,
		APIKey:      "",
		LogLevel:    defaultLogLevel,
		MaxBodySize: defaultMaxBodySize,
	}

	var filePath string
	if configPath != "" {
		filePath = configPath
	} else {
		if home, err := os.UserHomeDir(); err == nil {
			candidate := filepath.Join(home, ".config", "mdgen", "config.yaml")
			if fileExists(candidate) {
				filePath = candidate
			}
		}
		if filePath == "" && fileExists("config.yaml") {
			filePath = "config.yaml"
		}
		if filePath == "" && fileExists("config.json") {
			filePath = "config.json"
		}
	}

	if filePath != "" {
		if err := loadFromFile(filePath, &cfg); err != nil {
			return Config{}, err
		}
	}

	applyEnvOverrides(&cfg)

	cfg.OutputDir = expandHome(cfg.OutputDir)
	if cfg.ListenAddr == "" {
		cfg.ListenAddr = defaultListenAddr
	}
	if cfg.LogLevel == "" {
		cfg.LogLevel = defaultLogLevel
	}
	if cfg.MaxBodySize <= 0 {
		cfg.MaxBodySize = defaultMaxBodySize
	}

	return cfg, nil
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func loadFromFile(path string, cfg *Config) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read config: %w", err)
	}
	switch strings.ToLower(filepath.Ext(path)) {
	case ".json":
		if err := json.Unmarshal(data, cfg); err != nil {
			return fmt.Errorf("parse json: %w", err)
		}
	case ".yaml", ".yml":
		if err := parseSimpleYAML(data, cfg); err != nil {
			return fmt.Errorf("parse yaml: %w", err)
		}
	default:
		// try json
		if err := json.Unmarshal(data, cfg); err != nil {
			return fmt.Errorf("parse config: %w", err)
		}
	}
	return nil
}

// parseSimpleYAML parses a very small subset of YAML used by our flat config.
func parseSimpleYAML(data []byte, cfg *Config) error {
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			return errors.New("invalid yaml line: " + line)
		}
		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])
		val = strings.Trim(val, "\"'")
		switch key {
		case "listen_addr":
			cfg.ListenAddr = val
		case "output_dir":
			cfg.OutputDir = val
		case "api_key":
			cfg.APIKey = val
		case "log_level":
			cfg.LogLevel = val
		case "max_body_size":
			if val == "" {
				continue
			}
			v, err := strconv.ParseInt(val, 10, 64)
			if err != nil {
				return err
			}
			cfg.MaxBodySize = v
		}
	}
	return nil
}

func applyEnvOverrides(cfg *Config) {
	if v, ok := os.LookupEnv("MDGEN_LISTEN_ADDR"); ok {
		cfg.ListenAddr = v
	}
	if v, ok := os.LookupEnv("MDGEN_OUTPUT_DIR"); ok {
		cfg.OutputDir = v
	}
	if v, ok := os.LookupEnv("MDGEN_API_KEY"); ok {
		cfg.APIKey = v
	}
	if v, ok := os.LookupEnv("MDGEN_LOG_LEVEL"); ok {
		cfg.LogLevel = v
	}
	if v, ok := os.LookupEnv("MDGEN_MAX_BODY_SIZE"); ok {
		if val, err := strconv.ParseInt(v, 10, 64); err == nil {
			cfg.MaxBodySize = val
		}
	}
}

func expandHome(path string) string {
	if path == "" || !strings.HasPrefix(path, "~") {
		return path
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return path
	}
	return filepath.Join(home, strings.TrimPrefix(path, "~"))
}
