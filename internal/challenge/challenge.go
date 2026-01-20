package challenges

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/urfave/cli/v3"
)

const REPO_BASE_URL = "https://github.com/EinfachNiklas/cochabench-challenges-test/"

type Challenge struct {
	Title      string
	Location   string
	Language   string
	Difficulty string
}

type Manifest struct {
	Challenges map[string]Challenge
}

func (m Manifest) toString() string {
	if len(m.Challenges) == 0 {
		return "No challenges available."
	}

	// Header
	headers := []string{"ID", "Title", "Language", "Difficulty"}

	// Sort keys for stable output
	ids := make([]string, 0, len(m.Challenges))
	for id := range m.Challenges {
		ids = append(ids, id)
	}
	sort.Strings(ids)

	// Compute column widths
	wID, wTitle, wLang, wDiff := len(headers[0]), len(headers[1]), len(headers[2]), len(headers[3])

	for _, id := range ids {
		c := m.Challenges[id]
		if len(id) > wID {
			wID = len(id)
		}
		if len(c.Title) > wTitle {
			wTitle = len(c.Title)
		}
		if len(c.Language) > wLang {
			wLang = len(c.Language)
		}
		if len(c.Difficulty) > wDiff {
			wDiff = len(c.Difficulty)
		}
	}

	// Helper to build separator line
	sep := func() string {
		return fmt.Sprintf(
			"%s-+-%s-+-%s-+-%s",
			strings.Repeat("-", wID),
			strings.Repeat("-", wTitle),
			strings.Repeat("-", wLang),
			strings.Repeat("-", wDiff),
		)
	}

	var b strings.Builder

	// Header row
	b.WriteString(fmt.Sprintf(
		"%-*s | %-*s | %-*s | %-*s\n",
		wID, headers[0],
		wTitle, headers[1],
		wLang, headers[2],
		wDiff, headers[3],
	))
	b.WriteString(sep())
	b.WriteString("\n")

	// Data rows
	for _, id := range ids {
		c := m.Challenges[id]
		b.WriteString(fmt.Sprintf(
			"%-*s | %-*s | %-*s | %-*s\n",
			wID, id,
			wTitle, c.Title,
			wLang, c.Language,
			wDiff, c.Difficulty,
		))
	}

	return b.String()
}

func downloadManifest() (*Manifest, error) {
	var manifest Manifest
	fmt.Println("Fetching Manifest")
	resp, err := http.Get(REPO_BASE_URL + "releases/latest/download/maninfest.json")
	if err != nil {
		return nil, errors.Join(errors.New("Cannot download challenge manifest"), err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, errors.Join(errors.New("Error when downloading challenge manifest"), err)
	}

	d, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Join(errors.New("Error when downloading challenge manifest"), err)
	}
	err = json.Unmarshal(d, &manifest)
	if err != nil {
		return nil, errors.Join(errors.New("Bad manifest format"), err)
	}
	return &manifest, nil

}

func downloadChallenge(location string, path string) error {
	fmt.Printf("Fetching Challenge: %s\n", location)
	resp, err := http.Get(location)
	if err != nil {
		return errors.Join(errors.New("Cannot download challenge"), err)
	}
	defer resp.Body.Close()
	d, err := io.ReadAll(resp.Body)
	if err != nil {
		return errors.Join(errors.New("Could not read http response"), err)
	}

	zr, err := zip.NewReader(bytes.NewReader(d), resp.ContentLength)
	if err != nil {
		return errors.Join(errors.New("Cannot extract challenge"), err)
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
		out, err := os.OpenFile(targetPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
		if err != nil {
			return errors.Join(errors.New("Could not create extracting-file: "+targetPath), err)
		}
		_, err = io.Copy(out, fr)
		if err != nil {
			return errors.Join(errors.New("Error when writing file: "+targetPath), err)
		}
	}
	return nil
}

func List(ctx context.Context, cmd *cli.Command) error {
	manifest, err := downloadManifest()
	if err != nil {
		return errors.Join(errors.New("Could not download manifest"), err)
	}
	fmt.Println(manifest.toString())
	return nil
}

func Get(ctx context.Context, cmd *cli.Command) error {
	dirPath, err := os.Getwd()
	if err != nil {
		return err
	}
	id := strings.TrimSpace(cmd.Args().Get(0))
	if len(id) == 0 {
		return errors.New("No challenge id provided")
	}
	manifest, err := downloadManifest()
	if err != nil {
		return err
	}
	challenge, ok := manifest.Challenges[id]
	if !ok {
		return errors.New("The provided challenge id does not exist: " + id)
	}

	zipExtractPath := filepath.Join(dirPath)
	err = downloadChallenge(challenge.Location, zipExtractPath)
	if err != nil {
		return errors.Join(errors.New("Failed to download challenge"), err)
	}

	fmt.Printf("Successfully downloaded challenge %s\n", id)
	return nil
}
