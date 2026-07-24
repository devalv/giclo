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
	"regexp"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/ilyakaznacheev/cleanenv"
	"github.com/rs/zerolog/log"
)

type Config struct {
	Debug bool   `yaml:"debug"`
	User  string `yaml:"user"`
	Dir   string `yaml:"dir"`
	Token string `yaml:"token"`
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

	if cfg.User == "" {
		return nil, errors.New("Config `User` value is not set")
	}

	if cfg.Debug {
		log.Debug().Msg("Config is ok")
	}

	return cfg, nil
}

type GithubAPIRepoResponse struct {
	DirName  string `json:"full_name"`
	CloneURL string `json:"clone_url"`
}

type ReposToClone struct {
	CloneURL string
	CloneDir string
}

// createReposDirectory like 2023-09-10 for repos cloning.
func createReposDirectory(cfg *Config) (string, error) {
	currentTime := time.Now().Format(time.DateOnly)
	reposPath := filepath.Join(cfg.Dir, currentTime)
	if _, err := os.Stat(reposPath); os.IsNotExist(err) {
		if err := os.Mkdir(reposPath, os.ModeDir); err != nil {
			return "", fmt.Errorf("failed to create new repos directory: %w", err)
		}
	}

	if cfg.Debug {
		log.Debug().Msgf("Created directory %s", reposPath)
	}

	return reposPath, nil
}

// get API response.
func getAPIResponse(ctx context.Context, cfg *Config, page, perPage int) (*http.Response, error) {
	requestURL := fmt.Sprintf("https://api.github.com/users/%s/starred?page=%d&per_page=%d", cfg.User, page, perPage)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("error making http request: %w", err)
	}

	req.Header.Set("Accept", "application/vnd.github.v3+json")
	if cfg.Token != "" {
		req.Header.Set("Authorization", "token "+cfg.Token)
	}

	client := http.Client{
		Timeout: 30 * time.Second, //nolint:mnd
	}

	// TODO: extra timeout for req
	res, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending http request: %w", err)
	}

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad response status code: %d", res.StatusCode)
	}

	if cfg.Debug {
		log.Debug().Msg("Got GithubAPI response.")
	}

	// TODO: не возвращать res, чтобы удобнее работать с defer
	return res, nil
}

// extract repo links from a response.
func getRepos(resp *http.Response) (*[]GithubAPIRepoResponse, error) {
	resBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("could not read response body: %w", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			log.Error().Err(closeErr).Msg("error closing request body")
		}
	}()

	var result []GithubAPIRepoResponse
	if err := json.Unmarshal(resBody, &result); err != nil {
		return nil, fmt.Errorf("could not unmarshal JSON: %w", err)
	}

	return &result, nil
}

// get pages count liked by a user.
func getTotalPages(ctx context.Context, cfg *Config) (int, error) {
	resp, err := getAPIResponse(ctx, cfg, 1, 50) //nolint:mnd
	if err != nil {
		return 0, err
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			log.Error().Err(closeErr).Msg("error closing request body")
		}
	}()

	totalPages := 0
	// called only once per app execution
	pat := regexp.MustCompile(`([?|&]page=\d+)`)

	for _, link := range resp.Header.Values("Link") {
		for part := range strings.SplitSeq(link, ", ") {
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
		return totalPages, errors.New("could not extract total pages")
	}

	return totalPages, nil
}

// get a list of repositories liked by a user.
func getLikedRepos(ctx context.Context, reposPath string, cfg *Config) (*[]ReposToClone, error) {
	totalPages, err := getTotalPages(ctx, cfg)
	if cfg.Debug {
		log.Debug().Msgf("Total pages %d", totalPages)
	}

	if err != nil {
		return nil, err
	}

	var reposToClone []ReposToClone

	for i := 1; i <= totalPages; i++ {
		// чтобы не нарваться на бан - получаем список последовательно
		resp, err := getAPIResponse(ctx, cfg, i, 50) //nolint:mnd
		defer func() {
			if closeErr := resp.Body.Close(); closeErr != nil {
				log.Error().Err(closeErr).Msg("error closing request body")
			}
		}()
		if err != nil {
			log.Warn().Err(err).Msg("failed to get api response")

			continue
		}

		result, err := getRepos(resp)
		if err != nil {
			log.Warn().Err(err).Msgf("failed to get liked repos response")

			continue
		}

		for _, repo := range *result {
			if repo.CloneURL == "" || repo.DirName == "" {
				log.Warn().Msgf("Can`t get url or name %s|%s", repo.CloneURL, repo.DirName)

				continue
			}
			repoDir := fmt.Sprintf("%s/%s", reposPath, repo.DirName)
			reposToClone = append(reposToClone, ReposToClone{CloneURL: repo.CloneURL, CloneDir: repoDir})
		}
	}

	return &reposToClone, nil
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
func cloneRepos(isDebug bool, likedRepos *[]ReposToClone) {
	// TODO: таймаут на клонирование каждого отдельного репозитория

	var lastEl int
	for i := 0; i < len(*likedRepos); i += 5 {
		tmpRepos := *likedRepos
		var waitGroup sync.WaitGroup

		lastEl = min(i+5, len(*likedRepos)) //nolint:mnd

		for _, repo := range tmpRepos[i:lastEl] {
			waitGroup.Add(1)
			go cloneRepo(&waitGroup, isDebug, repo.CloneURL, repo.CloneDir)
		}

		waitGroup.Wait()
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

	// TODO: old
	likedRepos, err := getLikedRepos(ctx, reposPath, cfg)
	if err != nil {
		return fmt.Errorf("could not get liked repos: %w", err)
	}

	cloneRepos(cfg.Debug, likedRepos)

	return nil
}

func main() {
	if err := run(); err != nil {
		log.Fatal().Err(err).Msg("application failed")
	}
}
