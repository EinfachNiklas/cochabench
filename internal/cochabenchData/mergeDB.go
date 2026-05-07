package cochabenchdata

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/EinfachNiklas/cochabench/internal/tools"
	"github.com/urfave/cli/v3"
)

type MergedDB struct {
	db   *sql.DB
	path string
}

type MergedCochabenchEntry struct {
	RunName              string
	challenge            string
	RunID                string
	RunStatus            string
	StartTime            time.Time
	EndTime              time.Time
	Duration             time.Duration
	TimedOut             bool
	NumTotalTests        int
	NumPassedTests       int
	NumFailedTests       int
	QualityScore         float64
	MaintainabilityScore float64
	SecurityScore        float64
}

func validatePath(path string) error {
	_, err := os.Stat(filepath.Join(path, "challenge.config.json"))
	if !os.IsNotExist(err) {
		return fmt.Errorf("You cannot merge from inside a challenge directory.")
	}
	return nil
}

func newMergedDB(path string) (*MergedDB, error) {
	err := validatePath(path)
	if err != nil {
		return nil, err
	}

	db, err := setupMergedDB(path)
	if err != nil {
		return nil, err
	}
	return &MergedDB{db: db, path: path}, nil
}

func getAllRuns(path string) ([]*MergedCochabenchEntry, []string, error) {
	var cochabenchEntries []*MergedCochabenchEntry
	var challenges []string

	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, nil, fmt.Errorf("Failed to read directory: %s", path)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		challengeConf, err := tools.LoadChallengeConfig(filepath.Join(path, entry.Name(), "challenge.config.json"))
		if err != nil {
			return nil, nil, fmt.Errorf("Could not open challegne config for %s: %v", entry.Name(), err)
		}

		challenges = append(challenges, challengeConf.ChallengeID)

		store, err := LoadCochabenchStore(filepath.Join(path, entry.Name()))
		if err != nil {
			return nil, nil, fmt.Errorf("Could not access database for challenge %s: %v", challengeConf.ChallengeID, err)
		}

		rows, found, err := store.GetAllEntries()
		if err != nil {
			return nil, nil, fmt.Errorf("Could not get runs from database for challenge %s: %v", challengeConf.ChallengeID, err)
		}
		if !found {
			continue
		}

		for _, row := range rows {
			cochabenchEntries = append(cochabenchEntries, &MergedCochabenchEntry{
				RunName:              row.RunName,
				challenge:            challengeConf.ChallengeID,
				RunID:                row.RunID,
				RunStatus:            row.RunStatus,
				StartTime:            row.StartTime,
				EndTime:              row.EndTime,
				Duration:             row.Duration,
				TimedOut:             row.TimedOut,
				NumTotalTests:        row.NumTotalTests,
				NumPassedTests:       row.NumPassedTests,
				NumFailedTests:       row.NumFailedTests,
				QualityScore:         row.QualityScore,
				MaintainabilityScore: row.MaintainabilityScore,
				SecurityScore:        row.SecurityScore,
			})
		}
	}

	return cochabenchEntries, nil, nil
}

func (db *MergedDB) merge() error {
	runs, challenges, err := getAllRuns(db.path)
	if err != nil {
		return err
	}

	stmt, err := db.db.Prepare(`INSERT INTO challenges(challengeId) 
								VALUES(?)
								ON CONFLICT(challengeId) ABORT`)
	if err != nil {
		return fmt.Errorf("Could not insert challenge into db: %v", err)
	}

	for _, challenge := range challenges {
		_, err = stmt.Exec(challenge)
		if err != nil {
			stmt.Close()
			return fmt.Errorf("Could not insert challenge %s into table: %v", challenge, err)
		}
	}
	stmt.Close()

	stmt, err = db.db.Prepare(`
		INSERT INTO runs(runId, challengeId, runName, runStatus, startTime, endTime, duration, testTimedOut, numTotalTests, numPassedTests, numFailedTests, qualityScore, maintainabilityScore, securityScore)
		VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(runId) REPLACE
	`)
	if err != nil {
		return fmt.Errorf("Could not insert merged runs into table: %v", err)
	}
	defer stmt.Close()

	for _, run := range runs {
		_, err := stmt.Exec(
			run.RunID,
			run.challenge,
			run.RunName,
			run.RunStatus,
			run.StartTime,
			run.EndTime,
			run.Duration,
			run.TimedOut,
			run.NumTotalTests,
			run.NumPassedTests,
			run.NumFailedTests,
			run.QualityScore,
			run.MaintainabilityScore,
			run.SecurityScore,
		)
		if err != nil {
			return fmt.Errorf("Could not insert run %s: %v", run.RunID, err)
		}
	}
	return nil
}

func MergeDB(ctx context.Context, cmd *cli.Command) error {
	path := cmd.String("path")

	mergedDB, err := newMergedDB(path)
	if err != nil {
		return err
	}

	err = mergedDB.merge()
	if err != nil {
		return err
	}

	return nil
}
