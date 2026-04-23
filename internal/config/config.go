package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Model struct {
		Default  string `yaml:"default"`
		Fallback string `yaml:"fallback"`
		// Ollama/local provider settings
		APIType        string `yaml:"api_type"`
		Host           string `yaml:"host"`
		ModelID        string `yaml:"model_id"`
		TimeoutSeconds int    `yaml:"timeout_seconds"`
		UseCLI         bool   `yaml:"use_cli"`
	} `yaml:"model"`
	Memory struct {
		Type string `yaml:"type"`
		Path string `yaml:"path"`
	} `yaml:"memory"`
	Security struct {
		RequireApproval bool `yaml:"require_approval"`
		AllowShell      bool `yaml:"allow_shell"`
	} `yaml:"security"`
	Workspace struct {
		Root string `yaml:"root"`
	} `yaml:"workspace"`
}

func Load(path string) (*Config, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(content, &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	applyDefaults(&cfg)
	return &cfg, nil
}

func applyDefaults(cfg *Config) {
	if cfg.Model.Default == "" {
		cfg.Model.Default = "openai"
	}
	if cfg.Model.Fallback == "" {
		cfg.Model.Fallback = cfg.Model.Default
	}
	if cfg.Memory.Type == "" {
		cfg.Memory.Type = "sqlite"
	}
	if cfg.Memory.Path == "" {
		cfg.Memory.Path = "./data/memory.db"
	}
	if cfg.Workspace.Root == "" {
		cfg.Workspace.Root = "."
	}
}
