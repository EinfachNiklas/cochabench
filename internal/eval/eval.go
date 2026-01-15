package eval

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

type ChallengeConfig struct {
	Name          string
	ChallengeID   string
	ChallengeType string
}

func Evaluate() bool {
	return true
}

func ValidateDirStructure(path string) error {
	stat, err := os.Stat(filepath.Join(path, "solution"))
	if err != nil || !stat.IsDir() {
		return errors.New("Missing Directory 'solution' in provided path: " + path)
	}
	stat, err = os.Stat(filepath.Join(path, "config.json"))
	if err != nil {
		return errors.New("Missing Config File 'config.json' in provided path: " + path)
	}
	return nil
}

func LoadChallengeConfig(path string) (*ChallengeConfig, error) {
	var config ChallengeConfig
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(data, &config)
	if err != nil {
		return nil, errors.New("Malfomed configuration in " + path)
	}
	return &config, nil
}
