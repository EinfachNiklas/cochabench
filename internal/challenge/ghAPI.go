package challenge

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"

	"github.com/EinfachNiklas/cochabench/internal/config"
)

type ghAsset struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	URL  string `json:"url"`
}

type ghRelease struct {
	TagName string    `json:"tag_name"`
	Assets  []ghAsset `json:"assets"`
}

func githubGet(url string, downloadingFile bool) (*http.Response, error) {
	GITHUB_TOKEN := os.Getenv("GITHUB_TOKEN")

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("Could not create HTTP request: %w", err)
	}

	if downloadingFile {
		req.Header.Set("Accept", "application/octet-stream")
	}

	client := &http.Client{
		CheckRedirect: func(r *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	res, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}

	if res.StatusCode == 302 || res.StatusCode == 301 {
		location := res.Header.Get("Location")
		res, err = http.Get(location)
		if err != nil {
			return nil, fmt.Errorf("Could not follow download redirect: %w", err)
		}
	}

	// If 401/403 and token exists, retry with token
	if (res.StatusCode == 401 || res.StatusCode == 403) && len(GITHUB_TOKEN) != 0 {
		req, err = http.NewRequest("GET", url, nil)
		if err != nil {
			return nil, fmt.Errorf("Could not create HTTP request: %w", err)
		}

		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", GITHUB_TOKEN))

		if downloadingFile {
			req.Header.Set("Accept", "application/octet-stream")
		}

		res, err = client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("HTTP request failed: %w", err)
		}

		if res.StatusCode == 302 || res.StatusCode == 301 {
			location := res.Header.Get("Location")
			res, err = http.Get(location)
			if err != nil {
				return nil, fmt.Errorf("Could not follow download redirect: %w", err)
			}
		}

		if res.StatusCode == 401 {
			return nil, fmt.Errorf("GITHUB_TOKEN is invalid")
		}
	}

	if res.StatusCode == 401 && len(GITHUB_TOKEN) == 0 {
		return nil, fmt.Errorf("GITHUB_TOKEN is required to access this repository")
	}

	return res, nil
}

var getConfig = config.GetConfig

func fetchGithubAPIUrl(filename string) (string, string, error) {
	cfg, err := getConfig()
	if err != nil {
		return "", "", err
	}
	releaseURL, err := url.JoinPath(cfg.CHALLENGE_SERVER, "/releases/latest")
	if err != nil {
		return "", "", err
	}
	res, err := githubGet(releaseURL, false)
	if err != nil {
		return "", "", fmt.Errorf("Could not fetch latest release metadata for %s: %w", filename, err)
	}
	if res.StatusCode != 200 {
		return "", "", fmt.Errorf("Could not fetch latest release metadata for %s: server returned %s", filename, res.Status)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return "", "", fmt.Errorf("Could not read GitHub API response: %w", err)
	}

	var release ghRelease
	if err := json.Unmarshal(body, &release); err != nil {
		return "", "", fmt.Errorf("Could not parse GitHub release metadata: %w", err)
	}

	for _, asset := range release.Assets {
		if asset.Name == filename {
			return asset.URL, release.TagName, nil
		}
	}
	return "", "", fmt.Errorf("Release asset not found: %s", filename)
}
