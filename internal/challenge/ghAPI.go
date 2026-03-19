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
		return nil, fmt.Errorf("Could not create Request to url%s: %v\n", url, err)
	}

	if len(GITHUB_TOKEN) != 0 {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", GITHUB_TOKEN))
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
		return nil, fmt.Errorf("Failed to send http request: %v\n", err)
	}

	if res.StatusCode == 302 || res.StatusCode == 301 {
		location := res.Header.Get("Location")
		res, err = http.Get(location)
		if err != nil {
			return nil, fmt.Errorf("Failed to follow redirect: %v\n", err)
		}
	}

	if res.StatusCode == 401 && len(GITHUB_TOKEN) == 0 {
		return nil, fmt.Errorf("A Github Token is required. Please provide it via the environment variable GITHUB_TOKEN")
	} else if res.StatusCode == 401 {
		return nil, fmt.Errorf("The Github Token provided in the environment variable GITHUB_TOKEN is invalid")
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
		return "", "", fmt.Errorf("Could not fetch Github-Tag of file %s: %v\n", filename, err)
	}
	if res.StatusCode != 200 {
		return "", "", fmt.Errorf("Could not fetch Github-Tag of file %s: The http response was %s\nIf the repository you are trying to access is private, you may need to provide a Github Token via the environment variable GITHUB_TOKEN\n", filename, res.Status)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return "", "", fmt.Errorf("Could not read response body: %v\n", err)
	}

	var release ghRelease
	if err := json.Unmarshal(body, &release); err != nil {
		return "", "", fmt.Errorf("Could not parse release JSON: %v\n", err)
	}

	for _, asset := range release.Assets {
		if asset.Name == filename {
			return asset.URL, release.TagName, nil
		}
	}
	return "", "", fmt.Errorf("Asset %s not found in latest release %s", filename, release.TagName)
}
