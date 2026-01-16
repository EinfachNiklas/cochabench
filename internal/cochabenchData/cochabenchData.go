package cochabenchdata

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
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
