package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/rs/zerolog/log"

	"giclo/internal/adapters/config"
	"giclo/internal/application"
	"giclo/internal/domain/errors"
)

func ParseFlags() (cfgPath string, err error) {
	flag.StringVar(&cfgPath, "config", "./config.yml", "path to config file")
	flag.Parse()

	if err := ValidateConfigPath(cfgPath); err != nil {
		return "", err
	}

	return cfgPath, nil
}

func ValidateConfigPath(path string) error {
	s, err := os.Stat(path)
	if err != nil {
		return err
	}

	if s.IsDir() {
		return fmt.Errorf("'%s' is a directory, not a file", path)
	}

	return nil
}

func main() {
	cfgPath, err := ParseFlags()
	if err != nil {
		log.Fatal().Err(err).Msg(errors.ConfigError)
	}

	cfg, err := config.NewConfig(cfgPath)
	if err != nil {
		log.Fatal().Err(err).Msg(errors.ConfigError)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, os.Interrupt, syscall.SIGSEGV)
	defer cancel()

	app := application.NewApplication(cfg)
	go app.Start(ctx)
	<-ctx.Done()

	app.Stop(ctx)
}
