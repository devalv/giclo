package application

import (
	"context"

	"github.com/rs/zerolog/log"

	"giclo/internal/adapters/config"
)

func Start(ctx context.Context) {
	log.Debug().Msgf("Read: `%v`", ctx)
	cfg, _ := config.NewConfig()
	log.Debug().Msgf("Config is: `%v`", cfg)
	log.Debug().Msg("Starting application")
}

func Stop(ctx context.Context) {
	log.Debug().Msgf("Application stopped: `%v`", ctx)
}
