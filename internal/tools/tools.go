package tools

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"time"
)

type ChallengeConfig struct {
	Name          string
	ChallengeID   string
	ChallengeType string
}

type CochabenchData struct {
	RunName   string
	StartTime time.Time
	EndTime   time.Time
}

func ValidateDirPath(path string) error {
	stat, err := os.Stat(path)
	if len(path) == 0 {
		return errors.New("No path provided")
	}
	if err != nil {
		return errors.New("This directory does not exist")
	}
	if !stat.IsDir() {
		return errors.New("The provided path is not a directory")
	}
	return nil
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

func LoadCochabenchData(path string) (*CochabenchData, error) {
	var cochabencData CochabenchData
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(data, &cochabencData)
	if err != nil {
		return nil, errors.New("Malformed cochabench.json file in " + path)
	}
	return &cochabencData, nil
}

func WriteCochabenchData(data *CochabenchData, path string) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}
	err = os.WriteFile("cochabench.json", jsonData, 0666)
	if err != nil {
		return err
	}
	return nil
}
