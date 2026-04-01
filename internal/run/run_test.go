package run

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	cochabenchdata "github.com/EinfachNiklas/cochabench/internal/cochabenchData"
	"github.com/urfave/cli/v3"
)

// setupChallengeDir creates a temporary directory with the required challenge structure.
func setupChallengeDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "src"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "src", "main.go"), []byte("package main\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(dir, "test"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "challenge.config.json"), []byte(`{"Name":"test","ChallengeID":"c1","ChallengeType":"go"}`), 0644); err != nil {
		t.Fatal(err)
	}
	return dir
}

// runWithArgs wraps a run function in a cli.Command and executes it with the given flags and args.
func runWithArgs(t *testing.T, fn func(context.Context, *cli.Command) error, flags []cli.Flag, osArgs []string) error {
	t.Helper()
	var actionErr error
	cmd := &cli.Command{
		Name:  "test",
		Flags: flags,
		Action: func(ctx context.Context, cmd *cli.Command) error {
			actionErr = fn(ctx, cmd)
			return nil
		},
	}
	if err := cmd.Run(context.Background(), append([]string{"test"}, osArgs...)); err != nil {
		t.Fatalf("cli.Run failed unexpectedly: %v", err)
	}
	return actionErr
}

func nameFlag() []cli.Flag {
	return []cli.Flag{&cli.StringFlag{Name: "name", Aliases: []string{"n"}}}
}

func idFlag() []cli.Flag {
	return []cli.Flag{&cli.StringFlag{Name: "id", Aliases: []string{"i"}}}
}

// seedEntry inserts a CochabenchEntry directly into the store.
func seedEntry(t *testing.T, dirPath string, entry *cochabenchdata.CochabenchEntry) {
	t.Helper()
	store, err := cochabenchdata.LoadCochabenchStore(dirPath)
	if err != nil {
		t.Fatalf("seedEntry: LoadCochabenchStore failed: %v", err)
	}
	defer store.Close()
	if err := store.SaveEntry(entry); err != nil {
		t.Fatalf("seedEntry: SaveEntry failed: %v", err)
	}
}

// getEntry reads back an entry from the store for assertions.
func getEntry(t *testing.T, dirPath, id string) *cochabenchdata.CochabenchEntry {
	t.Helper()
	store, err := cochabenchdata.LoadCochabenchStore(dirPath)
	if err != nil {
		t.Fatalf("getEntry: LoadCochabenchStore failed: %v", err)
	}
	defer store.Close()
	entry, found, err := store.GetEntry(id)
	if err != nil {
		t.Fatalf("getEntry: GetEntry failed: %v", err)
	}
	if !found {
		t.Fatalf("getEntry: entry %q not found", id)
	}
	return entry
}

// discoverRunID finds the single RunID by listing the solutions/ directory.
func discoverRunID(t *testing.T, dirPath string) string {
	t.Helper()
	entries, err := os.ReadDir(filepath.Join(dirPath, "solutions"))
	if err != nil {
		t.Fatalf("discoverRunID: ReadDir failed: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("discoverRunID: expected 1 entry in solutions/, got %d", len(entries))
	}
	return entries[0].Name()
}

func TestInit(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(t *testing.T) string
		osArgs    func(dir string) []string
		wantErr   bool
		errSubstr string
	}{
		{
			name:  "Success",
			setup: setupChallengeDir,
			osArgs: func(dir string) []string {
				return []string{"--name=my-run", dir}
			},
		},
		{
			name:  "DefaultName",
			setup: setupChallengeDir,
			osArgs: func(dir string) []string {
				return []string{dir}
			},
		},
		{
			name: "InvalidDir",
			setup: func(t *testing.T) string {
				return filepath.Join(t.TempDir(), "nonexistent")
			},
			osArgs: func(dir string) []string {
				return []string{"--name=x", dir}
			},
			wantErr:   true,
			errSubstr: "Directory does not exist",
		},
		{
			name: "MissingSrc",
			setup: func(t *testing.T) string {
				dir := t.TempDir()
				os.Mkdir(filepath.Join(dir, "test"), 0755)
				os.WriteFile(filepath.Join(dir, "challenge.config.json"), []byte("{}"), 0644)
				return dir
			},
			osArgs: func(dir string) []string {
				return []string{"--name=x", dir}
			},
			wantErr:   true,
			errSubstr: "missing required folder: src",
		},
		{
			name: "MissingTest",
			setup: func(t *testing.T) string {
				dir := t.TempDir()
				os.MkdirAll(filepath.Join(dir, "src"), 0755)
				os.WriteFile(filepath.Join(dir, "challenge.config.json"), []byte("{}"), 0644)
				return dir
			},
			osArgs: func(dir string) []string {
				return []string{"--name=x", dir}
			},
			wantErr:   true,
			errSubstr: "missing required folder: test",
		},
		{
			name: "MissingConfig",
			setup: func(t *testing.T) string {
				dir := t.TempDir()
				os.MkdirAll(filepath.Join(dir, "src"), 0755)
				os.Mkdir(filepath.Join(dir, "test"), 0755)
				return dir
			},
			osArgs: func(dir string) []string {
				return []string{"--name=x", dir}
			},
			wantErr:   true,
			errSubstr: "missing required file: challenge.config.json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := tt.setup(t)
			err := runWithArgs(t, Init, nameFlag(), tt.osArgs(dir))

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if !strings.Contains(err.Error(), tt.errSubstr) {
					t.Errorf("error %q missing substring %q", err, tt.errSubstr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			// Verify solutions directory was created with copied files
			runID := discoverRunID(t, dir)
			copiedFile := filepath.Join(dir, "solutions", runID, "main.go")
			if _, err := os.Stat(copiedFile); os.IsNotExist(err) {
				t.Errorf("expected copied file %s to exist", copiedFile)
			}

			// Verify entry in store
			entry := getEntry(t, dir, runID)
			if entry.RunStatus != "I" {
				t.Errorf("RunStatus = %q, want %q", entry.RunStatus, "I")
			}
			if tt.name == "Success" && entry.RunName != "my-run" {
				t.Errorf("RunName = %q, want %q", entry.RunName, "my-run")
			}
		})
	}
}

func TestStart(t *testing.T) {
	tests := []struct {
		name       string
		seedStatus string
		wantErr    bool
		errSubstr  string
		wantStatus string
	}{
		{"FromInitialized", "I", false, "", "R"},
		{"FromCanceled", "C", false, "", "R"},
		{"AlreadyRunning", "R", true, "already running", ""},
		{"AlreadyFinished", "F", true, "already finished", ""},
		{"NotFound", "", true, "Run not found", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			id := "test-id-001"

			if tt.seedStatus != "" {
				seedEntry(t, dir, &cochabenchdata.CochabenchEntry{
					RunID:     id,
					RunName:   "test-run",
					RunStatus: tt.seedStatus,
				})
			}

			err := runWithArgs(t, Start, idFlag(), []string{"--id=" + id, dir})

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if !strings.Contains(err.Error(), tt.errSubstr) {
					t.Errorf("error %q missing substring %q", err, tt.errSubstr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			entry := getEntry(t, dir, id)
			if entry.RunStatus != tt.wantStatus {
				t.Errorf("RunStatus = %q, want %q", entry.RunStatus, tt.wantStatus)
			}
			if entry.StartTime.IsZero() {
				t.Error("expected non-zero StartTime")
			}
		})
	}
}

func TestStop(t *testing.T) {
	tests := []struct {
		name       string
		seedStatus string
		wantErr    bool
		errSubstr  string
		wantStatus string
	}{
		{"FromRunning", "R", false, "", "F"},
		{"AlreadyFinished", "F", true, "finished", ""},
		{"NotRunningInit", "I", true, "not running", ""},
		{"NotRunningCanceled", "C", true, "not running", ""},
		{"NotFound", "", true, "Run not found", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			id := "test-id-002"

			if tt.seedStatus != "" {
				seedEntry(t, dir, &cochabenchdata.CochabenchEntry{
					RunID:     id,
					RunName:   "test-run",
					RunStatus: tt.seedStatus,
				})
			}

			err := runWithArgs(t, Stop, idFlag(), []string{"--id=" + id, dir})

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if !strings.Contains(err.Error(), tt.errSubstr) {
					t.Errorf("error %q missing substring %q", err, tt.errSubstr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			entry := getEntry(t, dir, id)
			if entry.RunStatus != tt.wantStatus {
				t.Errorf("RunStatus = %q, want %q", entry.RunStatus, tt.wantStatus)
			}
			if entry.EndTime.IsZero() {
				t.Error("expected non-zero EndTime")
			}
		})
	}
}

func TestCancel(t *testing.T) {
	tests := []struct {
		name       string
		seedStatus string
		wantErr    bool
		errSubstr  string
		wantStatus string
	}{
		{"FromRunning", "R", false, "", "C"},
		{"AlreadyFinished", "F", true, "finished", ""},
		{"NotRunningInit", "I", true, "not running", ""},
		{"NotRunningCanceled", "C", true, "not running", ""},
		{"NotFound", "", true, "Run not found", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			id := "test-id-003"

			if tt.seedStatus != "" {
				seedEntry(t, dir, &cochabenchdata.CochabenchEntry{
					RunID:     id,
					RunName:   "test-run",
					RunStatus: tt.seedStatus,
				})
			}

			err := runWithArgs(t, Cancel, idFlag(), []string{"--id=" + id, dir})

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if !strings.Contains(err.Error(), tt.errSubstr) {
					t.Errorf("error %q missing substring %q", err, tt.errSubstr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			entry := getEntry(t, dir, id)
			if entry.RunStatus != tt.wantStatus {
				t.Errorf("RunStatus = %q, want %q", entry.RunStatus, tt.wantStatus)
			}
		})
	}
}

func TestList(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(t *testing.T) string
		wantErr   bool
		errSubstr string
	}{
		{
			name:  "EmptyStore",
			setup: setupChallengeDir,
		},
		{
			name: "WithEntries",
			setup: func(t *testing.T) string {
				dir := setupChallengeDir(t)
				seedEntry(t, dir, &cochabenchdata.CochabenchEntry{
					RunID:     "list-run-1",
					RunName:   "run-one",
					RunStatus: "I",
				})
				seedEntry(t, dir, &cochabenchdata.CochabenchEntry{
					RunID:     "list-run-2",
					RunName:   "run-two",
					RunStatus: "F",
				})
				return dir
			},
		},
		{
			name: "InvalidDir",
			setup: func(t *testing.T) string {
				return filepath.Join(t.TempDir(), "nonexistent")
			},
			wantErr:   true,
			errSubstr: "Directory does not exist",
		},
		{
			name: "MissingStructure",
			setup: func(t *testing.T) string {
				return t.TempDir()
			},
			wantErr:   true,
			errSubstr: "missing required folder: src",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := tt.setup(t)
			err := runWithArgs(t, List, nil, []string{dir})

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if !strings.Contains(err.Error(), tt.errSubstr) {
					t.Errorf("error %q missing substring %q", err, tt.errSubstr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestInitStartStopLifecycle(t *testing.T) {
	dir := setupChallengeDir(t)

	// Init
	err := runWithArgs(t, Init, nameFlag(), []string{"--name=lifecycle", dir})
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	runID := discoverRunID(t, dir)
	entry := getEntry(t, dir, runID)
	if entry.RunStatus != "I" {
		t.Fatalf("after Init: RunStatus = %q, want %q", entry.RunStatus, "I")
	}

	// Start
	err = runWithArgs(t, Start, idFlag(), []string{"--id=" + runID, dir})
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	entry = getEntry(t, dir, runID)
	if entry.RunStatus != "R" {
		t.Fatalf("after Start: RunStatus = %q, want %q", entry.RunStatus, "R")
	}
	if entry.StartTime.IsZero() {
		t.Fatal("after Start: expected non-zero StartTime")
	}

	// Stop
	err = runWithArgs(t, Stop, idFlag(), []string{"--id=" + runID, dir})
	if err != nil {
		t.Fatalf("Stop failed: %v", err)
	}

	entry = getEntry(t, dir, runID)
	if entry.RunStatus != "F" {
		t.Fatalf("after Stop: RunStatus = %q, want %q", entry.RunStatus, "F")
	}
	if entry.EndTime.IsZero() {
		t.Fatal("after Stop: expected non-zero EndTime")
	}
	if !entry.EndTime.After(entry.StartTime) {
		t.Errorf("EndTime %v should be after StartTime %v", entry.EndTime, entry.StartTime)
	}
}

func TestInitStartCancelRestartLifecycle(t *testing.T) {
	dir := setupChallengeDir(t)

	// Init
	err := runWithArgs(t, Init, nameFlag(), []string{"--name=cancel-test", dir})
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	runID := discoverRunID(t, dir)

	// Start
	err = runWithArgs(t, Start, idFlag(), []string{"--id=" + runID, dir})
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	entry := getEntry(t, dir, runID)
	if entry.RunStatus != "R" {
		t.Fatalf("after Start: RunStatus = %q, want %q", entry.RunStatus, "R")
	}

	// Cancel
	err = runWithArgs(t, Cancel, idFlag(), []string{"--id=" + runID, dir})
	if err != nil {
		t.Fatalf("Cancel failed: %v", err)
	}

	entry = getEntry(t, dir, runID)
	if entry.RunStatus != "C" {
		t.Fatalf("after Cancel: RunStatus = %q, want %q", entry.RunStatus, "C")
	}

	// Restart from canceled
	err = runWithArgs(t, Start, idFlag(), []string{"--id=" + runID, dir})
	if err != nil {
		t.Fatalf("Restart failed: %v", err)
	}

	entry = getEntry(t, dir, runID)
	if entry.RunStatus != "R" {
		t.Fatalf("after Restart: RunStatus = %q, want %q", entry.RunStatus, "R")
	}
}
