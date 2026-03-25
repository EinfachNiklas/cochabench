package challenge

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/EinfachNiklas/cochabench/internal/config"
)

func TestGithubGet(t *testing.T) {
	tests := []struct {
		name            string
		token           string
		downloadingFile bool
		serverStatus    int
		serverHeaders   map[string]string
		wantErr         bool
		errSubstr       string
		checkRequest    func(t *testing.T, r *http.Request)
	}{
		{
			name:         "Success_WithToken",
			token:        "ghp_testtoken123",
			serverStatus: 200,
			checkRequest: func(t *testing.T, r *http.Request) {
				auth := r.Header.Get("Authorization")
				if auth != "" {
					t.Errorf("Authorization = %q, want empty (token only used on retry)", auth)
				}
			},
		},
		{
			name:         "Success_WithToken_After401",
			token:        "ghp_testtoken123",
			serverStatus: 401,
			wantErr:      true,
			errSubstr:    "Token provided",
		},
		{
			name:         "Success_WithoutToken",
			token:        "",
			serverStatus: 200,
			checkRequest: func(t *testing.T, r *http.Request) {
				auth := r.Header.Get("Authorization")
				if auth != "" {
					t.Errorf("Authorization should be empty, got %q", auth)
				}
			},
		},
		{
			name:            "AcceptHeader_Download",
			token:           "",
			downloadingFile: true,
			serverStatus:    200,
			checkRequest: func(t *testing.T, r *http.Request) {
				accept := r.Header.Get("Accept")
				if accept != "application/octet-stream" {
					t.Errorf("Accept = %q, want %q", accept, "application/octet-stream")
				}
			},
		},
		{
			name:            "AcceptHeader_API",
			token:           "",
			downloadingFile: false,
			serverStatus:    200,
			checkRequest: func(t *testing.T, r *http.Request) {
				accept := r.Header.Get("Accept")
				if accept == "application/octet-stream" {
					t.Errorf("Accept should not be application/octet-stream for API calls")
				}
			},
		},
		{
			name:         "Auth_401_NoToken",
			token:        "",
			serverStatus: 401,
			wantErr:      true,
			errSubstr:    "Github Token is required",
		},
		{
			name:         "Auth_401_BadToken",
			token:        "ghp_invalid",
			serverStatus: 401,
			wantErr:      true,
			errSubstr:    "Token provided",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var lastRequest *http.Request
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				lastRequest = r.Clone(r.Context())
				for k, v := range tt.serverHeaders {
					w.Header().Set(k, v)
				}
				w.WriteHeader(tt.serverStatus)
				fmt.Fprint(w, "OK")
			}))
			defer srv.Close()

			t.Setenv("GITHUB_TOKEN", tt.token)

			res, err := githubGet(srv.URL, tt.downloadingFile)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if !strings.Contains(err.Error(), tt.errSubstr) {
					t.Errorf("error = %q, want substring %q", err.Error(), tt.errSubstr)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			defer res.Body.Close()

			if tt.checkRequest != nil && lastRequest != nil {
				tt.checkRequest(t, lastRequest)
			}
		})
	}
}

func TestGithubGet_TokenRetry(t *testing.T) {
	// Test that token is used on retry after 401
	requestCount := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		auth := r.Header.Get("Authorization")

		if requestCount == 1 {
			if auth != "" {
				t.Errorf("First request: Authorization = %q, want empty", auth)
			}
			w.WriteHeader(401)
			return
		}

		if auth != "Bearer ghp_testtoken123" {
			t.Errorf("Second request: Authorization = %q, want %q", auth, "Bearer ghp_testtoken123")
		}
		w.WriteHeader(200)
		fmt.Fprint(w, "OK")
	}))
	defer srv.Close()

	t.Setenv("GITHUB_TOKEN", "ghp_testtoken123")

	res, err := githubGet(srv.URL, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer res.Body.Close()

	if requestCount != 2 {
		t.Errorf("requestCount = %d, want 2 (initial + retry)", requestCount)
	}
}

func TestGithubGet_Redirect(t *testing.T) {
	// Target server that the redirect points to
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		fmt.Fprint(w, "redirected-content")
	}))
	defer target.Close()

	tests := []struct {
		name       string
		statusCode int
	}{
		{"Redirect_301", 301},
		{"Redirect_302", 302},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Location", target.URL)
				w.WriteHeader(tt.statusCode)
			}))
			defer srv.Close()

			t.Setenv("GITHUB_TOKEN", "")

			res, err := githubGet(srv.URL, false)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			defer res.Body.Close()

			if res.StatusCode != 200 {
				t.Errorf("status = %d, want 200 after redirect", res.StatusCode)
			}
		})
	}
}

func TestFetchReleaseAssetURL(t *testing.T) {
	tests := []struct {
		name      string
		filename  string
		status    int
		body      string
		wantURL   string
		wantTag   string
		wantErr   bool
		errSubstr string
	}{
		{
			name:     "Success_AssetFound",
			filename: "challenge.zip",
			status:   200,
			body: toJSON(ghRelease{
				TagName: "v1.2.0",
				Assets: []ghAsset{
					{ID: 1, Name: "manifest.json", URL: "https://api.github.com/assets/1"},
					{ID: 2, Name: "challenge.zip", URL: "https://api.github.com/assets/2"},
				},
			}),
			wantURL: "https://api.github.com/assets/2",
			wantTag: "v1.2.0",
		},
		{
			name:     "AssetNotFound",
			filename: "nonexistent.zip",
			status:   200,
			body: toJSON(ghRelease{
				TagName: "v1.0.0",
				Assets:  []ghAsset{{ID: 1, Name: "other.zip", URL: "https://example.com"}},
			}),
			wantErr:   true,
			errSubstr: "Asset nonexistent.zip not found",
		},
		{
			name:     "EmptyAssets",
			filename: "challenge.zip",
			status:   200,
			body: toJSON(ghRelease{
				TagName: "v1.0.0",
				Assets:  []ghAsset{},
			}),
			wantErr:   true,
			errSubstr: "Asset challenge.zip not found",
		},
		{
			name:      "Non200Response",
			filename:  "challenge.zip",
			status:    404,
			body:      "Not Found",
			wantErr:   true,
			errSubstr: "http response was",
		},
		{
			name:      "InvalidJSON",
			filename:  "challenge.zip",
			status:    200,
			body:      "{broken json",
			wantErr:   true,
			errSubstr: "Could not parse release JSON",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.status)
				fmt.Fprint(w, tt.body)
			}))
			defer srv.Close()

			t.Setenv("GITHUB_TOKEN", "test-token")

			// Mock config so CHALLENGE_SERVER points to test server
			orig := getConfig
			getConfig = func() (*config.Config, error) {
				return &config.Config{CHALLENGE_SERVER: srv.URL}, nil
			}
			t.Cleanup(func() { getConfig = orig })

			gotURL, gotTag, err := fetchGithubAPIUrl(tt.filename)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if !strings.Contains(err.Error(), tt.errSubstr) {
					t.Errorf("error = %q, want substring %q", err.Error(), tt.errSubstr)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if gotURL != tt.wantURL {
				t.Errorf("URL = %q, want %q", gotURL, tt.wantURL)
			}
			if gotTag != tt.wantTag {
				t.Errorf("Tag = %q, want %q", gotTag, tt.wantTag)
			}
		})
	}
}

func toJSON(v any) string {
	b, _ := json.Marshal(v)
	return string(b)
}
