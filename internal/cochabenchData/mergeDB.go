package cochabenchdata

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	"github.com/urfave/cli/v3"
)

type MergedDB struct {
	db   *sql.DB
	path string
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

func getAllRuns() ([]*CochabenchEntry, error) {
	var cochabenchEntries []*CochabenchEntry

	entries, err := os.ReadDir(db.path)
	if err != nil {
		return nil, fmt.Errorf("Failed to read directory: %s", db.path)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		store, err := LoadCochabenchStore(filepath.Join(db.path, entry.Name()))
		if err != nil {
			return nil, fmt.Errorf("Could not access database for challenge %s: %v", entry.Name(), err)
		}

		rows, found, err := store.GetAllEntries()
		if err != nil {
			return nil, fmt.Errorf("Could not get runs from database for challenge %s: %v", entry.Name(), err)
		}
		if !found {
			continue
		}

		cochabenchEntries = append(cochabenchEntries, rows...)
	}

	return cochabenchEntries, nil
}

func (db *MergedDB) merge() error {
	runs, err := getAllRuns()
	if err != nil {
		return err
	}

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
