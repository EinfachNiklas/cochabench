package cochabenchdata

import (
	"strings"
	"testing"
	"time"
)

func TestSetupDB_CreatesTableAndReturnsDB(t *testing.T) {
	db, err := setupDB(t.TempDir())
	if err != nil {
		t.Fatalf("setupDB failed: %v", err)
	}
	defer db.Close()

	rows, err := db.Query("SELECT COUNT(*) FROM runs")
	if err != nil {
		t.Fatalf("runs table not accessible: %v", err)
	}
	defer rows.Close()

	if !rows.Next() {
		t.Fatal("expected a row from COUNT(*)")
	}
	var count int
	if err := rows.Scan(&count); err != nil {
		t.Fatalf("failed to scan count: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected 0 rows in fresh table, got %d", count)
	}
}

func TestSetupDB_IsIdempotent(t *testing.T) {
	dir := t.TempDir()

	db1, err := setupDB(dir)
	if err != nil {
		t.Fatalf("first setupDB failed: %v", err)
	}
	db1.Close()

	// Calling setupDB again on the same dir should not fail
	db2, err := setupDB(dir)
	if err != nil {
		t.Fatalf("second setupDB failed: %v", err)
	}
	db2.Close()
}

func TestLoadCochabenchStore(t *testing.T) {
	store, err := LoadCochabenchStore(t.TempDir())
	if err != nil {
		t.Fatalf("LoadCochabenchStore failed: %v", err)
	}
	defer store.Close()
}

func TestSaveAndGetEntry(t *testing.T) {
	store, err := LoadCochabenchStore(t.TempDir())
	if err != nil {
		t.Fatalf("LoadCochabenchStore failed: %v", err)
	}
	defer store.Close()

	now := time.Now().Truncate(time.Second)
	entry := &CochabenchEntry{
		RunID:                "test-run-123",
		RunName:              "my-test-run",
		RunStatus:            "F",
		StartTime:            now,
		EndTime:              now.Add(30 * time.Second),
		TestDuration:         10 * time.Second,
		TimedOut:             false,
		NumTotalTests:        5,
		NumPassedTests:       4,
		NumFailedTests:       1,
		QualityScore:         8.5,
		MaintainabilityScore: 7.0,
		SecurityScore:        9.0,
	}

	if err := store.SaveEntry(entry); err != nil {
		t.Fatalf("SaveEntry failed: %v", err)
	}

	got, found, err := store.GetEntry("test-run-123")
	if err != nil {
		t.Fatalf("GetEntry failed: %v", err)
	}
	if !found {
		t.Fatal("expected entry to be found")
	}

	if got.RunID != entry.RunID {
		t.Errorf("RunID = %q, want %q", got.RunID, entry.RunID)
	}
	if got.RunName != entry.RunName {
		t.Errorf("RunName = %q, want %q", got.RunName, entry.RunName)
	}
	if got.RunStatus != entry.RunStatus {
		t.Errorf("RunStatus = %q, want %q", got.RunStatus, entry.RunStatus)
	}
	if got.NumTotalTests != entry.NumTotalTests {
		t.Errorf("NumTotalTests = %d, want %d", got.NumTotalTests, entry.NumTotalTests)
	}
	if got.NumPassedTests != entry.NumPassedTests {
		t.Errorf("NumPassedTests = %d, want %d", got.NumPassedTests, entry.NumPassedTests)
	}
	if got.NumFailedTests != entry.NumFailedTests {
		t.Errorf("NumFailedTests = %d, want %d", got.NumFailedTests, entry.NumFailedTests)
	}
	if got.QualityScore != entry.QualityScore {
		t.Errorf("QualityScore = %f, want %f", got.QualityScore, entry.QualityScore)
	}
	if got.MaintainabilityScore != entry.MaintainabilityScore {
		t.Errorf("MaintainabilityScore = %f, want %f", got.MaintainabilityScore, entry.MaintainabilityScore)
	}
	if got.SecurityScore != entry.SecurityScore {
		t.Errorf("SecurityScore = %f, want %f", got.SecurityScore, entry.SecurityScore)
	}
}

func TestGetEntry_NotFound(t *testing.T) {
	store, err := LoadCochabenchStore(t.TempDir())
	if err != nil {
		t.Fatalf("LoadCochabenchStore failed: %v", err)
	}
	defer store.Close()

	_, found, err := store.GetEntry("nonexistent-id")
	if err != nil {
		t.Fatalf("GetEntry failed: %v", err)
	}
	if found {
		t.Fatal("expected entry to not be found")
	}
}

func TestSaveEntry_UpsertOnConflict(t *testing.T) {
	store, err := LoadCochabenchStore(t.TempDir())
	if err != nil {
		t.Fatalf("LoadCochabenchStore failed: %v", err)
	}
	defer store.Close()

	entry := &CochabenchEntry{
		RunID:     "upsert-id",
		RunName:   "original-name",
		RunStatus: "R",
	}
	if err := store.SaveEntry(entry); err != nil {
		t.Fatalf("first SaveEntry failed: %v", err)
	}

	// Update with same ID
	entry.RunName = "updated-name"
	entry.RunStatus = "F"
	entry.QualityScore = 9.0
	if err := store.SaveEntry(entry); err != nil {
		t.Fatalf("second SaveEntry failed: %v", err)
	}

	got, found, err := store.GetEntry("upsert-id")
	if err != nil {
		t.Fatalf("GetEntry failed: %v", err)
	}
	if !found {
		t.Fatal("expected entry to be found")
	}
	if got.RunName != "updated-name" {
		t.Errorf("RunName = %q, want %q", got.RunName, "updated-name")
	}
	if got.RunStatus != "F" {
		t.Errorf("RunStatus = %q, want %q", got.RunStatus, "F")
	}
	if got.QualityScore != 9.0 {
		t.Errorf("QualityScore = %f, want 9.0", got.QualityScore)
	}
}

func TestToString_Empty(t *testing.T) {
	store, err := LoadCochabenchStore(t.TempDir())
	if err != nil {
		t.Fatalf("LoadCochabenchStore failed: %v", err)
	}
	defer store.Close()

	result := store.ToString()
	if result != "(no entries)\n" {
		t.Errorf("ToString() = %q, want %q", result, "(no entries)\n")
	}
}

func TestToString_WithEntries(t *testing.T) {
	store, err := LoadCochabenchStore(t.TempDir())
	if err != nil {
		t.Fatalf("LoadCochabenchStore failed: %v", err)
	}
	defer store.Close()

	entry := &CochabenchEntry{
		RunID:                "run-1",
		RunName:              "test",
		RunStatus:            "F",
		NumTotalTests:        3,
		NumPassedTests:       3,
		QualityScore:         8.0,
		MaintainabilityScore: 7.0,
		SecurityScore:        9.0,
	}
	if err := store.SaveEntry(entry); err != nil {
		t.Fatalf("SaveEntry failed: %v", err)
	}

	result := store.ToString()
	if result == "(no entries)\n" {
		t.Error("expected table output, got empty marker")
	}
	// Verify key data appears in the table
	if !strings.Contains(result, "run-1") {
		t.Error("expected RunID 'run-1' in output")
	}
	if !strings.Contains(result, "test") {
		t.Error("expected RunName 'test' in output")
	}
}

func TestClose(t *testing.T) {
	store, err := LoadCochabenchStore(t.TempDir())
	if err != nil {
		t.Fatalf("LoadCochabenchStore failed: %v", err)
	}

	store.Close()

	// After close, operations should fail
	_, _, err = store.GetEntry("any")
	if err == nil {
		t.Error("expected error after Close, got nil")
	}
}
