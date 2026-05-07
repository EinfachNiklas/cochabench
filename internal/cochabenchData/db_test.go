package cochabenchdata

import (
	"database/sql"
	"path/filepath"
	"testing"

	_ "modernc.org/sqlite"
)

// TestSetupMergedDB tests the setupMergedDB function
func TestSetupMergedDB(t *testing.T) {
	tempDir := t.TempDir()

	// Test successful DB creation
	db, err := setupMergedDB(tempDir)
	if err != nil {
		t.Fatalf("Failed to setup merged DB: %v", err)
	}
	defer db.Close()

	// Verify the database file was created
	if _, err := sql.Open("sqlite", filepath.Join(tempDir, "cochabenchMerged.db")); err != nil {
		t.Errorf("Expected cochabenchMerged.db to exist, got error: %v", err)
	}

	// Verify challenges table exists
	var tableName string
	err = db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='challenges'").Scan(&tableName)
	if err != nil || tableName != "challenges" {
		t.Errorf("Expected 'challenges' table to exist, got error: %v", err)
	}

	// Verify runs table exists
	err = db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='runs'").Scan(&tableName)
	if err != nil || tableName != "runs" {
		t.Errorf("Expected 'runs' table to exist, got error: %v", err)
	}
}

// TestSetupMergedDBTableSchema tests the schema of the merged DB tables
func TestSetupMergedDBTableSchema(t *testing.T) {
	tempDir := t.TempDir()
	db, err := setupMergedDB(tempDir)
	if err != nil {
		t.Fatalf("Failed to setup merged DB: %v", err)
	}
	defer db.Close()

	// Expected columns for challenges table
	challengesColumns := map[string]string{
		"challengeId": "CHAR(256)",
	}

	// Verify challenges table schema
	rows, err := db.Query("PRAGMA table_info(challenges)")
	if err != nil {
		t.Fatalf("Failed to get challenges table info: %v", err)
	}
	defer rows.Close()

	actualChallengesColumns := make(map[string]string)
	for rows.Next() {
		var cid, notnull, pk, dfltValue interface{}
		var name, ctype string
		if err := rows.Scan(&cid, &name, &ctype, &notnull, &dfltValue, &pk); err != nil {
			t.Fatalf("Failed to scan challenges table info: %v", err)
		}
		actualChallengesColumns[name] = ctype
	}

	for col, expectedType := range challengesColumns {
		if actualType, ok := actualChallengesColumns[col]; !ok {
			t.Errorf("Column %s not found in challenges table", col)
		} else if actualType != expectedType {
			t.Errorf("Column %s: expected type %s, got %s", col, expectedType, actualType)
		}
	}

	// Expected columns for runs table
	runsColumns := map[string]string{
		"runId":                "CHAR(36)",
		"challengeId":          "CHAR(256)",
		"runName":              "VARCHAR(256)",
		"runStatus":            "CHAR(1)",
		"startTime":            "TIMESTAMP",
		"endTime":              "TIMESTAMP",
		"duration":             "INTEGER",
		"testTimedOut":         "BOOLEAN",
		"numTotalTests":        "INTEGER",
		"numPassedTests":       "INTEGER",
		"numFailedTests":       "INTEGER",
		"qualityScore":         "DECIMAL(15, 2)",
		"maintainabilityScore": "DECIMAL(15, 2)",
		"securityScore":        "DECIMAL(15, 2)",
	}

	// Verify runs table schema
	rows, err = db.Query("PRAGMA table_info(runs)")
	if err != nil {
		t.Fatalf("Failed to get runs table info: %v", err)
	}
	defer rows.Close()

	actualRunsColumns := make(map[string]string)
	for rows.Next() {
		var cid, notnull, pk, dfltValue interface{}
		var name, ctype string
		if err := rows.Scan(&cid, &name, &ctype, &notnull, &dfltValue, &pk); err != nil {
			t.Fatalf("Failed to scan runs table info: %v", err)
		}
		actualRunsColumns[name] = ctype
	}

	for col, expectedType := range runsColumns {
		if actualType, ok := actualRunsColumns[col]; !ok {
			t.Errorf("Column %s not found in runs table", col)
		} else if actualType != expectedType {
			t.Errorf("Column %s: expected type %s, got %s", col, expectedType, actualType)
		}
	}
}

// TestSetupMergedDBForeignKey tests the foreign key constraint in the merged DB
func TestSetupMergedDBForeignKey(t *testing.T) {
	tempDir := t.TempDir()
	db, err := setupMergedDB(tempDir)
	if err != nil {
		t.Fatalf("Failed to setup merged DB: %v", err)
	}
	defer db.Close()

	// Verify foreign key constraint exists using PRAGMA foreign_key_list
	rows, err := db.Query("PRAGMA foreign_key_list('runs')")
	if err != nil {
		t.Fatalf("Failed to get foreign key list: %v", err)
	}
	defer rows.Close()

	found := false
	for rows.Next() {
		var id, seq, table, from, to, onUpdate, onDelete, match interface{}
		if err := rows.Scan(&id, &seq, &table, &from, &to, &onUpdate, &onDelete, &match); err != nil {
			t.Fatalf("Failed to scan foreign key list: %v", err)
		}
		if table == "challenges" && from == "challengeId" && to == "challengeId" {
			found = true
			break
		}
	}

	if !found {
		// Fallback: Check schema directly
		var schema string
		err = db.QueryRow("SELECT sql FROM sqlite_master WHERE name='runs'").Scan(&schema)
		if err != nil {
			t.Fatalf("Failed to get runs table schema: %v", err)
		}
		if !containsSubstring(schema, "FOREIGN KEY (challengeId) REFERENCES challenges(challengeId)") {
			t.Errorf("Expected foreign key constraint in runs table schema: %s", schema)
		}
	}
}

// containsString is a helper function to check if a string contains a substring
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsSubstring(s, substr))
}

// containsSubstring checks if substr is a substring of s
func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
