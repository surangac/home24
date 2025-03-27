package config

import (
	"fmt"
	"os"

	"home24/internal/analyzer"

	"gopkg.in/yaml.v3"
)

// Config holds all application configuration
type Config struct {
	Server   ServerConfig            `yaml:"server"`
	Analyzer analyzer.AnalyzerConfig `yaml:"analyzer"`
}

// ServerConfig holds server-specific configuration
type ServerConfig struct {
	Port         string `yaml:"port"`
	ReadTimeout  string `yaml:"readTimeout"`
	WriteTimeout string `yaml:"writeTimeout"`
	IdleTimeout  string `yaml:"idleTimeout"`
}

// LoadConfig loads configuration from a YAML file
func LoadConfig(configPath string) (*Config, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("error parsing config file: %w", err)
	}

	// Set defaults if not specified
	if config.Server.Port == "" {
		config.Server.Port = "8080"
	}
	if config.Server.ReadTimeout == "" {
		config.Server.ReadTimeout = "10s"
	}
	if config.Server.WriteTimeout == "" {
		config.Server.WriteTimeout = "30s"
	}
	if config.Server.IdleTimeout == "" {
		config.Server.IdleTimeout = "120s"
	}

	// If analyzer config is empty, use defaults
	if config.Analyzer.Timeout == 0 {
		config.Analyzer = analyzer.DefaultConfig()
	}

	return &config, nil
}
