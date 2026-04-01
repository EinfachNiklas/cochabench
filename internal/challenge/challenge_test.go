package challenge

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/EinfachNiklas/cochabench/internal/config"
)

// createTestZip builds an in-memory ZIP archive from a map of filename→content.
// It automatically adds directory entries for nested paths.
func createTestZip(t *testing.T, files map[string]string) []byte {
	t.Helper()
	var buf bytes.Buffer
	w := zip.NewWriter(&buf)

	// Collect and create directory entries
	dirs := map[string]bool{}
	for name := range files {
		dir := filepath.Dir(name)
		if dir != "." {
			dirs[dir+"/"] = true
		}
	}
	for dir := range dirs {
		hdr := &zip.FileHeader{
			Name: dir,
		}
		hdr.SetMode(os.ModeDir | 0755)
		_, err := w.CreateHeader(hdr)
		if err != nil {
			t.Fatalf("zip.CreateHeader dir(%q): %v", dir, err)
		}
	}

	for name, content := range files {
		f, err := w.Create(name)
		if err != nil {
			t.Fatalf("zip.Create(%q): %v", name, err)
		}
		f.Write([]byte(content))
	}
	w.Close()
	return buf.Bytes()
}

func TestManifest_toString(t *testing.T) {
	tests := []struct {
		name     string
		manifest Manifest
		tag      string
		contains []string
		exact    string
	}{
		{
			name:     "EmptyChallenges",
			manifest: Manifest{Challenges: map[string]Challenge{}},
			tag:      "v1.0",
			exact:    "No challenges available.",
		},
		{
			name:     "NilChallenges",
			manifest: Manifest{Challenges: nil},
			tag:      "v1.0",
			exact:    "No challenges available.",
		},
		{
			name: "SingleChallenge",
			manifest: Manifest{Challenges: map[string]Challenge{
				"001": {Title: "Foo", Language: "go", Difficulty: "easy"},
			}},
			tag:      "v1.0",
			contains: []string{"Release: v1.0", "001", "Foo", "go", "easy"},
		},
		{
			name: "ReleaseTag",
			manifest: Manifest{Challenges: map[string]Challenge{
				"001": {Title: "X"},
			}},
			tag:      "v2.3.1",
			contains: []string{"Release: v2.3.1"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.manifest.toString(tt.tag)

			if tt.exact != "" {
				if got != tt.exact {
					t.Errorf("toString() = %q, want %q", got, tt.exact)
				}
				return
			}

			for _, s := range tt.contains {
				if !strings.Contains(got, s) {
					t.Errorf("toString() missing %q in output:\n%s", s, got)
				}
			}

			// Verify sort order for SortedByID case
			if tt.name == "SortedByID" {
				idx1 := strings.Index(got, "001")
				idx2 := strings.Index(got, "002")
				idx3 := strings.Index(got, "003")
				if idx1 >= idx2 || idx2 >= idx3 {
					t.Errorf("IDs not in sorted order: 001@%d, 002@%d, 003@%d\n%s", idx1, idx2, idx3, got)
				}
			}
		})
	}
}

func TestDownloadManifest(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		manifestJSON, _ := json.Marshal(Manifest{
			Challenges: map[string]Challenge{
				"001": {Title: "Hello Go", Filename: "hello-go.zip", Language: "go", Difficulty: "easy"},
			},
		})

		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch {
			case strings.HasSuffix(r.URL.Path, "/releases/latest"):
				release := ghRelease{
					TagName: "v1.0.0",
					Assets:  []ghAsset{{ID: 1, Name: "manifest.json", URL: "http://" + r.Host + "/assets/manifest"}},
				}
				b, _ := json.Marshal(release)
				w.Write(b)
			case r.URL.Path == "/assets/manifest":
				w.Write(manifestJSON)
			default:
				w.WriteHeader(404)
			}
		}))
		t.Cleanup(srv.Close)
		orig := getConfig
		getConfig = func() (*config.Config, error) {
			return &config.Config{CHALLENGE_SERVER: srv.URL}, nil
		}
		t.Cleanup(func() { getConfig = orig })
		t.Setenv("GITHUB_TOKEN", "test-token")

		manifest, tag, err := downloadManifest()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if tag != "v1.0.0" {
			t.Errorf("tag = %q, want %q", tag, "v1.0.0")
		}
		if len(manifest.Challenges) != 1 {
			t.Errorf("challenges count = %d, want 1", len(manifest.Challenges))
		}
		c, ok := manifest.Challenges["001"]
		if !ok {
			t.Fatal("challenge 001 not found")
		}
		if c.Title != "Hello Go" {
			t.Errorf("title = %q, want %q", c.Title, "Hello Go")
		}
	})

	t.Run("InvalidManifestJSON", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch {
			case strings.HasSuffix(r.URL.Path, "/releases/latest"):
				release := ghRelease{
					TagName: "v1.0.0",
					Assets:  []ghAsset{{ID: 1, Name: "manifest.json", URL: "PLACEHOLDER"}},
				}
				// Replace placeholder with actual server URL (we don't know it yet in the struct)
				b, _ := json.Marshal(release)
				body := strings.Replace(string(b), "PLACEHOLDER", "http://"+r.Host+"/assets/manifest", 1)
				w.Write([]byte(body))
			case r.URL.Path == "/assets/manifest":
				w.Write([]byte("{broken json"))
			default:
				w.WriteHeader(404)
			}
		}))
		t.Cleanup(srv.Close)
		orig := getConfig
		getConfig = func() (*config.Config, error) {
			return &config.Config{CHALLENGE_SERVER: srv.URL}, nil
		}
		t.Cleanup(func() { getConfig = orig })
		t.Setenv("GITHUB_TOKEN", "test-token")

		_, _, err := downloadManifest()
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !strings.Contains(err.Error(), "Challenge manifest is invalid") {
			t.Errorf("error = %q, want substring %q", err.Error(), "Challenge manifest is invalid")
		}
	})

	t.Run("ReleaseAPIError", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(500)
		}))
		t.Cleanup(srv.Close)
		orig := getConfig
		getConfig = func() (*config.Config, error) {
			return &config.Config{CHALLENGE_SERVER: srv.URL}, nil
		}
		t.Cleanup(func() { getConfig = orig })
		t.Setenv("GITHUB_TOKEN", "test-token")

		_, _, err := downloadManifest()
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("ConfigError", func(t *testing.T) {
		orig := getConfig
		getConfig = func() (*config.Config, error) {
			return nil, fmt.Errorf("config unavailable")
		}
		t.Cleanup(func() { getConfig = orig })

		_, _, err := downloadManifest()
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !strings.Contains(err.Error(), "config unavailable") {
			t.Errorf("error = %q, want substring %q", err.Error(), "config unavailable")
		}
	})
}

func TestDownloadChallenge(t *testing.T) {
	// Helper: create a mock server that serves a release + ZIP asset
	setupChallengeServer := func(t *testing.T, filename string, zipData []byte) {
		t.Helper()
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch {
			case strings.HasSuffix(r.URL.Path, "/releases/latest"):
				release := ghRelease{
					TagName: "v2.0.0",
					Assets:  []ghAsset{{ID: 1, Name: filename, URL: "PLACEHOLDER"}},
				}
				b, _ := json.Marshal(release)
				body := strings.Replace(string(b), "PLACEHOLDER", "http://"+r.Host+"/assets/challenge", 1)
				w.Write([]byte(body))
			case r.URL.Path == "/assets/challenge":
				w.Header().Set("Content-Length", fmt.Sprintf("%d", len(zipData)))
				w.Write(zipData)
			default:
				w.WriteHeader(404)
			}
		}))
		t.Cleanup(srv.Close)
		orig := getConfig
		getConfig = func() (*config.Config, error) {
			return &config.Config{CHALLENGE_SERVER: srv.URL}, nil
		}
		t.Cleanup(func() { getConfig = orig })
		t.Setenv("GITHUB_TOKEN", "test-token")
	}

	t.Run("Success_ExtractsFiles", func(t *testing.T) {
		zipData := createTestZip(t, map[string]string{
			"main.go": "package main\n",
		})
		setupChallengeServer(t, "challenge.zip", zipData)

		destDir := t.TempDir()
		err := downloadChallenge("challenge.zip", destDir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		content, err := os.ReadFile(filepath.Join(destDir, "main.go"))
		if err != nil {
			t.Fatalf("file not extracted: %v", err)
		}
		if string(content) != "package main\n" {
			t.Errorf("content = %q, want %q", string(content), "package main\n")
		}
	})

	t.Run("Success_NestedDirs", func(t *testing.T) {
		zipData := createTestZip(t, map[string]string{
			"src/main.go":       "package main\n",
			"test/main_test.go": "package main\n",
		})
		setupChallengeServer(t, "challenge.zip", zipData)

		destDir := t.TempDir()
		err := downloadChallenge("challenge.zip", destDir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		for _, path := range []string{"src/main.go", "test/main_test.go"} {
			if _, err := os.Stat(filepath.Join(destDir, path)); os.IsNotExist(err) {
				t.Errorf("expected file %s not found", path)
			}
		}
	})

	t.Run("InvalidZipContent", func(t *testing.T) {
		setupChallengeServer(t, "bad.zip", []byte("not a zip"))

		destDir := t.TempDir()
		err := downloadChallenge("bad.zip", destDir)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !strings.Contains(err.Error(), "Downloaded challenge archive is invalid") {
			t.Errorf("error = %q, want substring %q", err.Error(), "Downloaded challenge archive is invalid")
		}
	})

	t.Run("DownloadError_AssetNotFound", func(t *testing.T) {
		// Server has release but asset name doesn't match
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch {
			case strings.HasSuffix(r.URL.Path, "/releases/latest"):
				release := ghRelease{
					TagName: "v1.0.0",
					Assets:  []ghAsset{{ID: 1, Name: "other.zip", URL: "http://example.com"}},
				}
				b, _ := json.Marshal(release)
				w.Write(b)
			default:
				w.WriteHeader(404)
			}
		}))
		t.Cleanup(srv.Close)
		orig := getConfig
		getConfig = func() (*config.Config, error) {
			return &config.Config{CHALLENGE_SERVER: srv.URL}, nil
		}
		t.Cleanup(func() { getConfig = orig })
		t.Setenv("GITHUB_TOKEN", "test-token")

		err := downloadChallenge("missing.zip", t.TempDir())
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !strings.Contains(err.Error(), "Release asset not found: missing.zip") {
			t.Errorf("error = %q, want substring %q", err.Error(), "Release asset not found: missing.zip")
		}
	})
}
