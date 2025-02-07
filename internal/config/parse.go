package config

import (
	"fmt"
	"github.com/spf13/viper"
	"github.com/dndev-xx/go-ninja-chat/internal/validator"
)

func ParseAndValidate(filename string) (*Config, error) {
	viper.SetConfigFile(filename)

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("ошибка чтения конфигурационного файла: %w", err)
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("ошибка разбора конфигурации: %w", err)
	}

	if err := validator.Validator.Struct(config); err!= nil {
        return nil, fmt.Errorf("ошибка валидации конфигурации: %w", err)
    }

	return &config, nil
}
