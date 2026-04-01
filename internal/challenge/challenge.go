package challenge

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/EinfachNiklas/cochabench/internal/tools"
	"github.com/urfave/cli/v3"
)

type Challenge struct {
	Title      string
	Filename   string
	Language   string
	Difficulty string
}

type Manifest struct {
	Challenges map[string]Challenge
}

func (m Manifest) toString(releaseTag string) string {
	if len(m.Challenges) == 0 {
		return "No challenges available."
	}

	// Sort keys for stable output
	ids := make([]string, 0, len(m.Challenges))
	for id := range m.Challenges {
		ids = append(ids, id)
	}
	sort.Strings(ids)

	var b strings.Builder
	fmt.Fprintf(&b, "Release: %s\n\n", releaseTag)

	// Build table using shared utility
	tb := tools.NewTableBuilder([]string{"ID", "Title", "Language", "Difficulty"})

	for _, id := range ids {
		c := m.Challenges[id]
		tb.AddRow([]string{id, c.Title, c.Language, c.Difficulty})
	}

	b.WriteString(tb.Build())
	return b.String()
}

func downloadManifest() (*Manifest, string, error) {
	var manifest Manifest
	fmt.Println("Fetching Manifest")
	manifestUrl, tag, err := fetchGithubAPIUrl("manifest.json")
	if err != nil {
		return nil, "", err
	}

	resp, err := githubGet(manifestUrl, true)
	if err != nil {
		return nil, "", fmt.Errorf("Could not download challenge manifest: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("Could not download challenge manifest: server returned %s", resp.Status)
	}

	d, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", fmt.Errorf("Could not read challenge manifest: %w", err)
	}
	err = json.Unmarshal(d, &manifest)
	if err != nil {
		return nil, "", fmt.Errorf("Challenge manifest is invalid: %w", err)
	}
	return &manifest, tag, nil

}

func downloadChallenge(filename string, path string) error {
	fmt.Printf("Fetching Challenge: %s\n", filename)

	challengeUrl, _, err := fetchGithubAPIUrl(filename)
	if err != nil {
		return err
	}

	resp, err := githubGet(challengeUrl, true)
	if err != nil {
		return fmt.Errorf("Could not download challenge: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Could not download challenge: server returned %s", resp.Status)
	}

	d, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("Could not read download response: %w", err)
	}

	zr, err := zip.NewReader(bytes.NewReader(d), resp.ContentLength)
	if err != nil {
		return fmt.Errorf("Downloaded challenge archive is invalid: %w", err)
	}
	for _, f := range zr.File {
		targetPath := filepath.Join(path, f.Name)
		if f.FileInfo().IsDir() {
			os.MkdirAll(targetPath, f.Mode())
			continue
		}
		fr, err := f.Open()
		if err != nil {
			return err
		}
		defer fr.Close()

		out, err := os.OpenFile(targetPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
		if err != nil {
			return fmt.Errorf("Could not create extracted file: %w", err)
		}
		defer out.Close()

		_, err = io.Copy(out, fr)
		if err != nil {
			return fmt.Errorf("Could not write extracted file: %w", err)
		}
	}
	return nil
}

func List(ctx context.Context, cmd *cli.Command) error {
	manifest, tag, err := downloadManifest()
	if err != nil {
		return fmt.Errorf("Could not fetch challenge list: %w", err)
	}
	fmt.Println(manifest.toString(tag))
	return nil
}

func Get(ctx context.Context, cmd *cli.Command) error {
	dirPath, err := os.Getwd()
	if err != nil {
		return err
	}
	id := strings.TrimSpace(cmd.Args().Get(0))
	if len(id) == 0 {
		return fmt.Errorf("Missing required argument: challenge ID")
	}
	manifest, _, err := downloadManifest()
	if err != nil {
		return fmt.Errorf("Could not fetch challenge list: %w", err)
	}
	challenge, ok := manifest.Challenges[id]
	if !ok {
		return fmt.Errorf("Unknown challenge ID: %s", id)
	}

	zipExtractPath := filepath.Join(dirPath)
	err = downloadChallenge(challenge.Filename, zipExtractPath)
	if err != nil {
		return err
	}

	fmt.Printf("Successfully downloaded challenge %s\n", id)
	return nil
}

func GetAll(ctx context.Context, cmd *cli.Command) error {
	dirPath, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("Could not determine current working directory: %w", err)
	}

	manifest, _, err := downloadManifest()
	if err != nil {
		return fmt.Errorf("Could not fetch challenge list: %w", err)
	}
	for _, challenge := range manifest.Challenges {
		err = downloadChallenge(challenge.Filename, dirPath)
		if err != nil {
			return fmt.Errorf("Could not download all challenges; failed on %s: %w", challenge.Filename, err)
		}
	}
	fmt.Println("Successfully downloaded all challenges")
	return nil
}
