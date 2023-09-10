package application

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/rs/zerolog/log"

	"giclo/internal/domain/errors"
	"giclo/internal/domain/models"
)

type Application struct {
	cfg *models.Config
}

func NewApplication(cfg *models.Config) *Application {
	app := &Application{cfg: cfg}
	return app
}

// create directory like 2023-09-10 17:45:13 for repos cloning
func createReposDirectory(cfg *models.Config) (string, error) {
	currentTime := time.Now().Format(time.DateTime)
	reposPath := filepath.Join(cfg.Dir, currentTime)
	if _, err := os.Stat(reposPath); os.IsNotExist(err) {
		err := os.Mkdir(reposPath, os.ModePerm)
		if err != nil {
			return "", err
		}
	}

	if cfg.Debug {
		log.Debug().Msgf("Created directory %s", reposPath)
	}

	return reposPath, nil
}

func (app *Application) Start(ctx context.Context) {
	if app.cfg.Debug {
		log.Debug().Msgf("Config is: `%v`", app.cfg)
		log.Debug().Msgf("Context: `%v`", ctx)
	}
	log.Debug().Msg("Starting application")

	reposPath, isOk := createReposDirectory(app.cfg)
	if isOk != nil {
		log.Fatal().Err(isOk).Msgf(errors.CreateDirectoryError, isOk)
	}

	fmt.Println(reposPath)

	// TODO: получить список лайканых реп
	// TODO: выполнить клонирование
	// TODO: выполнить архивацию
}

func (app *Application) Stop(ctx context.Context) {
	if app.cfg.Debug {
		log.Debug().Msgf("Config is: `%v`", app.cfg)
		log.Debug().Msgf("Context: `%v`", ctx)
	}
	log.Debug().Msgf("Application stopped: `%v`", ctx)
}
