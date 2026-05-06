package cochabenchdata

import (
	"database/sql"
	"fmt"
	"path/filepath"

	_ "modernc.org/sqlite"
)

func setupDB(dirPath string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", filepath.Join(dirPath, "cochabench.db"))
	if err != nil {
		return nil, fmt.Errorf("Could not open run database: %w", err)
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS runs(
			runId CHAR(36) PRIMARY KEY,
			runName VARCHAR(256) NOT NULL,
			runStatus CHAR(1) NOT NULL,
			startTime TIMESTAMP,
			endTime TIMESTAMP,
			duration INTEGER,
			testTimedOut BOOLEAN,
			numTotalTests INTEGER,
			numPassedTests INTEGER,
			numFailedTests INTEGER,
			qualityScore DECIMAL(15, 2),
			maintainabilityScore DECIMAL(15, 2),
			securityScore DECIMAL(15, 2)
	);`)
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("Could not initialize run database: %w", err)
	}
	return db, nil
}

func setupMergedDB(dirPath string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", filepath.Join(dirPath, "cochabenchMerged.db"))
	if err != nil {
		return nil, fmt.Errorf("Could not open run database: %w", err)
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS challenges(
			challengeId CHAR(36) PRIMARY KEY
		)	
	`)
	if err != nil {
		return nil, fmt.Errorf("Could not initialize challenge db table: %w", err)
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS runs(
			runId CHAR(36) PRIMARY KEY,
			challengeId CHAR(36) NOT NULL,
			runName VARCHAR(256) NOT NULL,
			runStatus CHAR(1) NOT NULL,
			startTime TIMESTAMP,
			endTime TIMESTAMP,
			duration INTEGER,
			testTimedOut BOOLEAN,
			numTotalTests INTEGER,
			numPassedTests INTEGER,
			numFailedTests INTEGER,
			qualityScore DECIMAL(15, 2),
			maintainabilityScore DECIMAL(15, 2),
			securityScore DECIMAL(15, 2),
			FOREIGN KEY (challengeId) REFERENCES challenges(challengeId)
	);`)
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("Could not initialize run db table: %w", err)
	}
	return db, nil
}
