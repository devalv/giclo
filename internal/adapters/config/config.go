package config

import (
	"github.com/ilyakaznacheev/cleanenv"
	"github.com/rs/zerolog/log"

	"giclo/internal/domain/errors"
	"giclo/internal/domain/models"
)

func NewConfig() (*models.Config, error) {
	var cfg models.Config

	err := cleanenv.ReadEnv(&cfg)
	if err != nil {
		log.Fatal().Msgf(errors.ConfigError, err)
	}

	return &cfg, nil
}
