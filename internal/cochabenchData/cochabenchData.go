package cochabenchdata

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/EinfachNiklas/cochabench/internal/tools"
)

type CochabenchEntry struct {
	RunName         string
	RunID           string
	RunStatus       string
	StartTime       time.Time
	EndTime         time.Time
	TestDuration    time.Duration
	PassedTests     bool
	TimedOut        bool
	NumTotalTests   int
	NumPassedTests  int
	NumFailedTests  int
	NumSkippedTests int
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

	// Sort for stable order
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
		return t.Local().Format("2006-01-02 15:04:05")
	}

	// Build table using shared utility
	tb := tools.NewTableBuilder([]string{"ID", "RunName", "RunID", "Status", "StartTime", "EndTime"})

	for _, id := range ids {
		e := (*store)[id]
		tb.AddRow([]string{
			id,
			e.RunName,
			e.RunID,
			e.RunStatus,
			fmtTime(e.StartTime),
			fmtTime(e.EndTime),
		})
	}

	return tb.Build()
}
