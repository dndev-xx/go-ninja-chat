package config

import (
	"fmt"

	"github.com/BurntSushi/toml"

	"github.com/dndev-xx/go-ninja-chat/internal/validator"
)

func ParseAndValidate(filename string) (*Config, error) {
	var cfg *Config
	if _, err := toml.DecodeFile(filename, &cfg); err != nil {
		return nil, fmt.Errorf("decode file: %v", err)
	}

	if err := validator.Validator.Struct(cfg); err != nil {
		return nil, fmt.Errorf("validate: %v", err)
	}
	return cfg, nil
}
