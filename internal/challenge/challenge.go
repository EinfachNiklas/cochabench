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
	"strings"

	"github.com/urfave/cli/v3"
)

type Challenge struct {
	Title    string
	Location string
}

type Manifest struct {
	Challenges map[string]Challenge
}

func downloadManifest() (*Manifest, error) {
	var manifest Manifest
	fmt.Println("Fetching Manifest")
	resp, err := http.Get("https://github.com/EinfachNiklas/cochabench-challenges-test/releases/latest/download/maninfest.json")
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
	for _, challenge := range manifest.Challenges {
		fmt.Printf("%s\n", challenge.Title)
	}

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
