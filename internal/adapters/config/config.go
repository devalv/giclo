package config

import (
	"github.com/ilyakaznacheev/cleanenv"

	"giclo/internal/domain/models"
)

func NewConfig(cfgPath string) (*models.Config, error) {
	var cfg models.Config

	err := cleanenv.ReadConfig("config.yml", &cfg)
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}
