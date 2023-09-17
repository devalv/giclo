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

	if cfg.User == "" {
		return fmt.Errorf("Config `User` value is not set")
	}

	if cfg.Debug {
		log.Debug().Msg("Config is ok")
	}

	return nil
}

type GithubAPIRepoResponse struct {
	DirName  string `json:"full_name"`
	CloneURL string `json:"clone_url"`
}

type ReposToClone struct {
	CloneURL string
	CloneDir string
}
