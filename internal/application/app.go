package application

import (
	"context"

	"github.com/rs/zerolog/log"

	"giclo/internal/domain/models"
)

type Application struct {
	cfg *models.Config
}

func NewApplication(cfg *models.Config) *Application {
	app := &Application{cfg: cfg}
	return app
}

func (app *Application) Start(ctx context.Context) {
	if app.cfg.Debug {
		log.Debug().Msgf("Config is: `%v`", app.cfg)
		log.Debug().Msgf("Context: `%v`", ctx)
	}
	log.Debug().Msg("Starting application")
}

func (app *Application) Stop(ctx context.Context) {
	if app.cfg.Debug {
		log.Debug().Msgf("Config is: `%v`", app.cfg)
		log.Debug().Msgf("Context: `%v`", ctx)
	}
	log.Debug().Msgf("Application stopped: `%v`", ctx)
}
