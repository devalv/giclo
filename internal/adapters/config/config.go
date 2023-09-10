package config

import (
	"fmt"

	"github.com/ilyakaznacheev/cleanenv"

	"giclo/internal/domain/errors"
	"giclo/internal/domain/models"
)

func NewConfig(cfgPath string) (*models.Config, error) {
	var cfg models.Config

	err := cleanenv.ReadConfig("config.yml", &cfg)
	if err != nil {
		return nil, fmt.Errorf(errors.ConfigError, err)
	}

	return &cfg, nil
}
