package cochabenchdata

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/EinfachNiklas/cochabench/internal/tools"
)

type CochabenchEntry struct {
	RunName              string
	RunID                string
	RunStatus            string
	StartTime            time.Time
	EndTime              time.Time
	TestDuration         time.Duration
	TimedOut             bool
	NumTotalTests        int
	NumPassedTests       int
	NumFailedTests       int
	QualityScore         float64
	MaintainabilityScore float64
	SecurityScore        float64
}

type Store struct {
	db *sql.DB
}

func LoadCochabenchStore(path string) (*Store, error) {
	db, err := setupDB(path)
	if err != nil {
		return nil, err
	}
	return &Store{db: db}, nil
}

func (store *Store) GetEntry(runID string) (*CochabenchEntry, bool, error) {
	stmt, err := store.db.Prepare("SELECT runId, runName, runStatus, startTime, endTime, duration, testTimedOut, numTotalTests, numPassedTests, numFailedTests, qualityScore, maintainabilityScore, securityScore FROM runs WHERE runId = ?")
	if err != nil {
		return nil, false, fmt.Errorf("Failed to prepare SELECT statement: %v\n", err)
	}
	defer stmt.Close()
	var entry CochabenchEntry
	err = stmt.QueryRow(runID).Scan(
		&entry.RunID,
		&entry.RunName,
		&entry.RunStatus,
		&entry.StartTime,
		&entry.EndTime,
		&entry.TestDuration,
		&entry.TimedOut,
		&entry.NumTotalTests,
		&entry.NumPassedTests,
		&entry.NumFailedTests,
		&entry.QualityScore,
		&entry.MaintainabilityScore,
		&entry.SecurityScore,
	)
	if err == sql.ErrNoRows {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, fmt.Errorf("Failed to scan row: %v\n", err)
	}
	return &entry, true, nil
}

func (store *Store) SaveEntry(entry *CochabenchEntry) error {
	_, err := store.db.Exec(`
		INSERT INTO runs(runId, runName, runStatus, startTime, endTime, duration, testTimedOut, numTotalTests, numPassedTests, numFailedTests, qualityScore, maintainabilityScore, securityScore)
		VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(runId) DO UPDATE SET
			runName = excluded.runName,
			runStatus = excluded.runStatus,
			startTime = excluded.startTime,
			endTime = excluded.endTime,
			duration = excluded.duration,
			testTimedOut = excluded.testTimedOut,
			numTotalTests = excluded.numTotalTests,
			numPassedTests = excluded.numPassedTests,
			numFailedTests = excluded.numFailedTests,
			qualityScore = excluded.qualityScore,
			maintainabilityScore = excluded.maintainabilityScore,
			securityScore = excluded.securityScore`,
		entry.RunID, entry.RunName, entry.RunStatus, entry.StartTime, entry.EndTime, entry.TestDuration, entry.TimedOut, entry.NumTotalTests, entry.NumPassedTests, entry.NumFailedTests, entry.QualityScore, entry.MaintainabilityScore, entry.SecurityScore,
	)
	if err != nil {
		return fmt.Errorf("Failed to save entry %v: %v\n", entry, err)
	}
	return nil
}

func (store *Store) ToString() string {
	rows, err := store.db.Query("SELECT runId, runName, runStatus, startTime, endTime, duration, testTimedOut, numTotalTests, numPassedTests, numFailedTests, qualityScore, maintainabilityScore, securityScore FROM runs ORDER BY runId")
	if err != nil {
		return fmt.Sprintf("(error: %v)\n", err)
	}
	defer rows.Close()

	tb := tools.NewTableBuilder([]string{"RunID", "RunName", "Status", "StartTime", "EndTime", "Duration", "TimedOut", "Total", "Passed", "Failed", "Quality", "Maintainability", "Security"})
	hasRows := false

	for rows.Next() {
		var entry CochabenchEntry
		if err := rows.Scan(&entry.RunID, &entry.RunName, &entry.RunStatus, &entry.StartTime, &entry.EndTime, &entry.TestDuration, &entry.TimedOut, &entry.NumTotalTests, &entry.NumPassedTests, &entry.NumFailedTests, &entry.QualityScore, &entry.MaintainabilityScore, &entry.SecurityScore); err != nil {
			log.Printf("Warning: failed to scan row: %v", err)
			continue
		}
		hasRows = true
		tb.AddRow([]string{
			entry.RunID,
			entry.RunName,
			entry.RunStatus,
			tools.FmtTime(entry.StartTime),
			tools.FmtTime(entry.EndTime),
			entry.TestDuration.String(),
			fmt.Sprintf("%v", entry.TimedOut),
			fmt.Sprintf("%d", entry.NumTotalTests),
			fmt.Sprintf("%d", entry.NumPassedTests),
			fmt.Sprintf("%d", entry.NumFailedTests),
			fmt.Sprintf("%.2f", entry.QualityScore),
			fmt.Sprintf("%.2f", entry.MaintainabilityScore),
			fmt.Sprintf("%.2f", entry.SecurityScore),
		})
	}

	if !hasRows {
		return "(no entries)\n"
	}
	return tb.Build()
}

func (store *Store) Close() {
	store.db.Close()
}
