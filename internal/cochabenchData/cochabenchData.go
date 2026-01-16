package cochabenchdata

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type CochabenchEntry struct {
	RunName   string
	RunID     string
	RunStatus string
	StartTime time.Time
	EndTime   time.Time
}
type Store map[string]CochabenchEntry

const STORE_FILE_NAME = "cochabenchStore.json"

func initStore(path string) (*Store, error) {
	store := Store{}
	jsonData, err := json.Marshal(store)
	if err != nil {
		return nil, err
	}
	err = os.WriteFile(filepath.Join(path, STORE_FILE_NAME), jsonData, 0666)
	if err != nil {
		return nil, err
	}
	return &store, nil
}

func LoadCochabenchStore(path string) (*Store, error) {
	var store Store
	data, err := os.ReadFile(filepath.Join(path, STORE_FILE_NAME))
	if os.IsNotExist(err) {
		return initStore(path)
	} else if err != nil {
		return nil, err
	}
	err = json.Unmarshal(data, &store)
	if err != nil {
		return nil, errors.New("Malformed cochabench.json file in " + path)
	}
	return &store, nil
}

func (data *CochabenchEntry) Write(path string) error {
	store, err := LoadCochabenchStore(path)
	if err != nil {
		return err
	}
	(*store)[data.RunID] = *data
	jsonData, err := json.Marshal(store)
	if err != nil {
		return err
	}
	err = os.WriteFile(filepath.Join(path, STORE_FILE_NAME), jsonData, 0666)
	if err != nil {
		return err
	}
	return nil
}
func (store *Store) ToString() string {
	if store == nil || *store == nil || len(*store) == 0 {
		return "(no entries)\n"
	}

	// Stable order
	ids := make([]string, 0, len(*store))
	for id := range *store {
		ids = append(ids, id)
	}
	sort.Strings(ids)

	// Helper for time formatting
	fmtTime := func(t time.Time) string {
		if t.IsZero() {
			return "-"
		}
		// Choose a readable, sortable format
		return t.Local().Format("2006-01-02 15:04:05")
	}

	// Compute column widths (based on headers + values)
	hID, hName, hRunID, hStatus, hStart, hEnd := "ID", "RunName", "RunID", "Status", "StartTime", "EndTime"
	wID, wName, wRunID, wStatus, wStart, wEnd := len(hID), len(hName), len(hRunID), len(hStatus), len(hStart), len(hEnd)

	for _, id := range ids {
		e := (*store)[id]
		if len(id) > wID {
			wID = len(id)
		}
		if len(e.RunName) > wName {
			wName = len(e.RunName)
		}
		if len(e.RunID) > wRunID {
			wRunID = len(e.RunID)
		}
		if len(e.RunStatus) > wStatus {
			wStatus = len(e.RunStatus)
		}
		if l := len(fmtTime(e.StartTime)); l > wStart {
			wStart = l
		}
		if l := len(fmtTime(e.EndTime)); l > wEnd {
			wEnd = l
		}
	}

	// Build table
	var b strings.Builder

	// Header
	fmt.Fprintf(&b, "%-*s  %-*s  %-*s  %-*s  %-*s  %-*s\n",
		wID, hID, wName, hName, wRunID, hRunID, wStatus, hStatus, wStart, hStart, wEnd, hEnd,
	)

	// Separator
	sep := func(n int) string { return strings.Repeat("-", n) }
	fmt.Fprintf(&b, "%s  %s  %s  %s  %s  %s\n",
		sep(wID), sep(wName), sep(wRunID), sep(wStatus), sep(wStart), sep(wEnd),
	)

	// Rows
	for _, id := range ids {
		e := (*store)[id]
		fmt.Fprintf(&b, "%-*s  %-*s  %-*s  %-*s  %-*s  %-*s\n",
			wID, id,
			wName, e.RunName,
			wRunID, e.RunID,
			wStatus, e.RunStatus,
			wStart, fmtTime(e.StartTime),
			wEnd, fmtTime(e.EndTime),
		)
	}

	return b.String()
}
