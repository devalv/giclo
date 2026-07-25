package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/ilyakaznacheev/cleanenv"
	"github.com/rs/zerolog/log"
)

const (
	workerSize = 5
	timeout    = 10
	perPage    = 50
	apiURL     = "https://api.github.com/users/%s/starred?page=%d&per_page=%d"
)

type GithubAPIRepoResponse struct {
	DirName  string `json:"full_name"`
	CloneURL string `json:"clone_url"`
}

type ReposToClone struct {
	CloneURL string
	CloneDir string
}

type Config struct {
	Debug      bool   `yaml:"debug"`
	GitHubUser string `yaml:"gh_user"`
	Dir        string `yaml:"dir"`
	Token      string `yaml:"token"`
}

func NewConfig() (*Config, error) {
	var cfgPath string
	flag.StringVar(&cfgPath, "config", "./config.yml", "path to config file")
	flag.Parse()

	s, err := os.Stat(cfgPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	if s.IsDir() {
		return nil, fmt.Errorf("'%s' is a directory, not a file", cfgPath)
	}

	cfg := &Config{
		Debug: false,
		Dir:   ".",
	}

	err = cleanenv.ReadConfig(cfgPath, cfg)
	if err != nil {
		return nil, fmt.Errorf("config read err: %w", err)
	}

	if cfg.GitHubUser == "" {
		return nil, errors.New("Config `gh_user` value is not set")
	}

	if cfg.Debug {
		log.Debug().Msg("Config is ok")
	}

	return cfg, nil
}

// createReposDirectory like 2023-09-10 for repos cloning.
func createReposDirectory(cfg *Config) (string, error) {
	reposPath := filepath.Join(cfg.Dir, time.Now().Format(time.DateOnly))
	if err := os.MkdirAll(reposPath, 0o750); err != nil { //nolint:mnd
		return "", fmt.Errorf("create repos directory: %w", err)
	}

	if cfg.Debug {
		log.Debug().Str("path", reposPath).Msg("created repository directory")
	}

	return reposPath, nil
}

func getAPIResponse(ctx context.Context, cfg *Config, page int) (*http.Response, error) {
	url := fmt.Sprintf(apiURL, cfg.GitHubUser, page, perPage)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-Github-Api-Version", "2022-11-28")

	if cfg.Token != "" {
		req.Header.Set("Authorization", "Bearer "+cfg.Token)
	}

	client := &http.Client{
		Timeout: timeout * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		defer func() {
			if closeErr := resp.Body.Close(); closeErr != nil {
				log.Error().Err(closeErr).Msg("error closing request body")
			}
		}()

		body, _ := io.ReadAll(resp.Body)

		return nil, fmt.Errorf(
			"github api returned status %d\nbody:\n%s",
			resp.StatusCode,
			string(body),
		)
	}

	if cfg.Debug {
		log.Debug().
			Int("page", page).
			Msg("github api request succeeded")
	}

	return resp, nil
}

func getRepos(resp *http.Response) ([]GithubAPIRepoResponse, error) {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body: %w", err)
	}

	var repos []GithubAPIRepoResponse

	if err := json.Unmarshal(body, &repos); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return repos, nil
}

func hasNextPage(resp *http.Response) bool {
	for _, header := range resp.Header.Values("Link") {
		for part := range strings.SplitSeq(header, ",") {
			if strings.Contains(part, `rel="next"`) {
				return true
			}
		}
	}

	return false
}

func getLikedRepos(ctx context.Context, reposPath string, cfg *Config) ([]ReposToClone, error) {
	var reposToClone []ReposToClone

	for page := 1; ; page++ {
		resp, err := getAPIResponse(ctx, cfg, page)
		if err != nil {
			return nil, err
		}

		repos, err := getRepos(resp)
		nextPage := hasNextPage(resp)

		if err := resp.Body.Close(); err != nil {
			log.Error().
				Err(err).
				Msg("failed to close response body")
		}

		if err != nil {
			return nil, err
		}

		if cfg.Debug {
			log.Debug().
				Int("page", page).
				Int("repos", len(repos)).
				Bool("has_next", nextPage).
				Msg("received github page")
		}

		for _, repo := range repos {
			if repo.CloneURL == "" || repo.DirName == "" {
				log.Warn().
					Str("clone_url", repo.CloneURL).
					Str("dir_name", repo.DirName).
					Msg("repository has empty fields")

				continue
			}

			reposToClone = append(reposToClone, ReposToClone{
				CloneURL: repo.CloneURL,
				CloneDir: filepath.Join(reposPath, repo.DirName),
			})
		}

		if !nextPage {
			break
		}
	}

	return reposToClone, nil
}

func cloneRepo(wg *sync.WaitGroup, isDebug bool, repoURL, dirPath string) {
	defer wg.Done()
	defer log.Info().Msgf("Клонирован %s", repoURL)

	if isDebug {
		log.Debug().Msgf("Собираемся клонировать %s в %s", repoURL, dirPath)
	}

	_, err := git.PlainClone(dirPath, false, &git.CloneOptions{
		URL:      repoURL,
		Progress: nil,
	})
	if err != nil {
		log.Warn().Err(err).Msgf("failed to clone repo: %s", repoURL)
	}
}

// clone repos concurrently to a local fs with a WaitGroup.
func cloneRepos(isDebug bool, likedRepos []ReposToClone) {
	// TODO: таймаут на клонирование каждого отдельного репозитория

	for i := 0; i < len(likedRepos); i += workerSize {
		end := min(i+workerSize, len(likedRepos))

		var wg sync.WaitGroup

		for _, repo := range likedRepos[i:end] {
			wg.Add(1)

			go cloneRepo(&wg, isDebug, repo.CloneURL, repo.CloneDir)
		}

		wg.Wait()
	}
}

func run() error {
	cfg, err := NewConfig()
	if err != nil {
		return fmt.Errorf("failed to read config %w", err)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, os.Interrupt, syscall.SIGSEGV)
	defer cancel()

	reposPath, err := createReposDirectory(cfg)
	if err != nil {
		return fmt.Errorf("failed to create repos directory: %w", err)
	}

	likedRepos, err := getLikedRepos(ctx, reposPath, cfg)
	if err != nil {
		return fmt.Errorf("could not get liked repos: %w", err)
	}
	log.Debug().Msgf("liked repos: %v", likedRepos)

	cloneRepos(cfg.Debug, likedRepos)

	return nil
}

func main() {
	if err := run(); err != nil {
		log.Fatal().Err(err).Msg("application failed")
	}
}
