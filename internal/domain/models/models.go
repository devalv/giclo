package models

import (
	"fmt"
	"os"

	"github.com/rs/zerolog/log"
)

type Config struct {
	Debug    bool   `yaml:"debug" env:"DEBUG"`
	User     string `yaml:"user" env:"USER" env-default:"user"`
	Dir      string `yaml:"dir" env:"DIR" env-default:"."`
	Token    string `yaml:"token" env:"TOKEN"`
	Compress bool   `yaml:"compress" env:"COMPRESS"`
}

func (cfg *Config) Check() error {
	// directory can be read and exists
	s, err := os.Stat(cfg.Dir)
	if err != nil {
		return err
	}

	if !s.IsDir() {
		return fmt.Errorf("'%s' is a file, not a directory", cfg.Dir)
	}

	if cfg.Debug {
		log.Debug().Msg("Config is ok")
	}

	return nil
}
