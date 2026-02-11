package config

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/spf13/viper"
)

// Config holds all configuration for the application
type Config struct {
	DataDir      string `mapstructure:"data_dir"`
	DatabasePath string `mapstructure:"database_path"`
	AssetsDir    string `mapstructure:"assets_dir"`
	LogLevel     string `mapstructure:"log_level"`
	IsHeadless   bool   `mapstructure:"is_headless"`
}

// Load loads the configuration from environment and defaults
func Load() (*Config, error) {
	v := viper.New()

	// Set defaults
	v.SetDefault("log_level", "info")
	v.SetDefault("is_headless", false)

	// Determine default data directory based on OS
	dataDir, err := getDefaultDataDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get default data directory: %w", err)
	}

	v.SetDefault("data_dir", dataDir)
	v.SetDefault("database_path", filepath.Join(dataDir, "brique.db"))
	v.SetDefault("assets_dir", filepath.Join(dataDir, "assets"))

	// Allow environment variable overrides with BRIQUE_ prefix
	v.SetEnvPrefix("BRIQUE")
	v.AutomaticEnv()

	// Try to read config file if it exists
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(dataDir)
	v.AddConfigPath(".")

	// Reading the config file is optional
	_ = v.ReadInConfig()

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Ensure directories exist
	if err := ensureDirectories(&cfg); err != nil {
		return nil, fmt.Errorf("failed to create directories: %w", err)
	}

	return &cfg, nil
}

// getDefaultDataDir returns the default data directory based on the OS
func getDefaultDataDir() (string, error) {
	var baseDir string

	switch runtime.GOOS {
	case "windows":
		baseDir = os.Getenv("APPDATA")
		if baseDir == "" {
			return "", fmt.Errorf("APPDATA environment variable not set")
		}
		return filepath.Join(baseDir, "Brique"), nil
	case "darwin":
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(homeDir, "Library", "Application Support", "Brique"), nil
	case "linux":
		// Check if running as service (headless mode)
		if os.Getuid() == 0 {
			return "/var/lib/brique", nil
		}
		// User mode
		configDir := os.Getenv("XDG_CONFIG_HOME")
		if configDir == "" {
			homeDir, err := os.UserHomeDir()
			if err != nil {
				return "", err
			}
			configDir = filepath.Join(homeDir, ".config")
		}
		return filepath.Join(configDir, "brique"), nil
	default:
		return "", fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}
}

// ensureDirectories creates all required directories
func ensureDirectories(cfg *Config) error {
	dirs := []string{
		cfg.DataDir,
		cfg.AssetsDir,
		filepath.Dir(cfg.DatabasePath),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	return nil
}
