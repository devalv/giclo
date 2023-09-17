package application

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/go-git/go-git/v5"
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

// get API response
func getAPIResponse(ctx context.Context, cfg *models.Config, page, perPage int) (*http.Response, error) {
	requestURL := fmt.Sprintf("https://api.github.com/users/%s/starred?page=%d&per_page=%d", cfg.User, page, perPage)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf(errors.APIRequestCreateError, err)
	}

	req.Header.Set("accept", "application/vnd.github.v3+json")
	if cfg.Token != "" {
		req.Header.Set("Authorization", fmt.Sprintf("token %s", cfg.Token))
	}

	client := http.Client{
		Timeout: 30 * time.Second,
	}

	res, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf(errors.APIRequestSendError, err)
	}

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf(errors.APIResponseBadStatusError)
	}

	if cfg.Debug {
		log.Debug().Msg("Got GithubAPI response.")
	}

	return res, nil
}

// extract repo links from a response
func getRepos(resp *http.Response) (*[]models.GithubAPIRepoResponse, error) {
	resBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf(errors.ReadResponseBodyError, err)
	}
	defer resp.Body.Close()

	var result []models.GithubAPIRepoResponse
	if err := json.Unmarshal(resBody, &result); err != nil {
		return nil, fmt.Errorf(errors.ResponseUnmarshalError, err)
	}

	return &result, nil
}

// get pages count liked by a user
func getTotalPages(ctx context.Context, cfg *models.Config) (int, error) {
	resp, err := getAPIResponse(ctx, cfg, 1, 10)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	totalPages := 0
	// called only once per app execution
	pat := regexp.MustCompile(`([?|&]page=\d+)`)

	for _, link := range resp.Header.Values("link") {
		for _, part := range strings.Split(link, ", ") {
			if strings.Contains(part, `rel="last"`) {
				lastPageSub := pat.FindString(part)
				if lastPageSub != "" {
					totalPages, err = strconv.Atoi(strings.Split(lastPageSub, "=")[1])
					if err != nil {
						totalPages = 0
						break
					}
				}
			}
		}
	}

	if totalPages == 0 {
		return totalPages, fmt.Errorf(errors.APIResponseTotalPagesError)
	}

	return totalPages, nil
}

// get a list of repositories liked by a user
func getLikedRepos(ctx context.Context, reposPath string, cfg *models.Config) (*[]models.ReposToClone, error) {
	totalPages, err := getTotalPages(ctx, cfg)
	if cfg.Debug {
		log.Debug().Msgf("Total pages %d", totalPages)
	}

	if err != nil {
		return nil, err
	}

	var reposToClone []models.ReposToClone

	for i := 1; i <= totalPages; i++ {
		// TODO: goroutines
		resp, err := getAPIResponse(ctx, cfg, i, 10)
		if err != nil {
			log.Warn().Err(err)
			continue
		}

		result, err := getRepos(resp)
		if err != nil {
			log.Warn().Err(err)
			continue
		}

		for _, repo := range *result {
			if repo.CloneURL == "" || repo.DirName == "" {
				log.Warn().Msgf("Can`t get url or name %s|%s", repo.CloneURL, repo.DirName)
				continue
			}
			repoDir := fmt.Sprintf("%s/%s", reposPath, repo.DirName)
			reposToClone = append(reposToClone, models.ReposToClone{CloneURL: repo.CloneURL, CloneDir: repoDir})
		}
	}

	return &reposToClone, nil
}

// clone repo to a local fs
func cloneRepo(repoURL, dirPath string) error {
	_, err := git.PlainClone(dirPath, false, &git.CloneOptions{
		URL:      repoURL,
		Progress: os.Stdout,
	})

	return err
}

func (app *Application) Start(ctx context.Context) {
	if app.cfg.Debug {
		log.Debug().Msgf("Config is: `%v`", app.cfg)
	}
	log.Info().Msg("Starting application.")

	reposPath, err := createReposDirectory(app.cfg)
	if err != nil {
		log.Fatal().Err(err).Msgf(errors.CreateDirectoryError, err)
	}

	likedRepos, err := getLikedRepos(ctx, reposPath, app.cfg)
	if err != nil {
		log.Fatal().Err(err).Msgf(errors.APILikedResponseError, err)
	}

	// TODO: явно нужна горутина
	for _, repo := range *likedRepos {
		if app.cfg.Debug {
			log.Debug().Msgf("Собираемся клонировать %s в %s", repo.CloneURL, repo.CloneDir)
		}
		err := cloneRepo(repo.CloneURL, repo.CloneDir)
		if err != nil {
			log.Warn().Err(err)
		}
	}

	app.Stop(ctx)
}

func (app *Application) Stop(ctx context.Context) {
	if app.cfg.Debug {
		log.Debug().Msgf("Config is: `%v`, ctx is: `%v`.", app.cfg, ctx)
	}
	log.Info().Msg("Application stopped.")
	os.Exit(0)
}
