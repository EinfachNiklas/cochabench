package tools

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

func ValidateDirPath(path string) error {
	stat, err := os.Stat(path)
	if err != nil {
		return errors.New("This directory does not exist")
	}
	if !stat.IsDir() {
		return errors.New("The provided path is not a directory")
	}
	return nil
}

func ValidateDirStructure(path string) error {
	stat, err := os.Stat(filepath.Join(path, "challenge"))
	if err != nil || !stat.IsDir() {
		return errors.New("Missing Directory 'challenge' in provided path: " + path)
	}
	stat, err = os.Stat(filepath.Join(path, "challenge", "test"))
	if os.IsNotExist(err) || !stat.IsDir() {
		return errors.New("Missing Directory /challenge/test in provided path: " + path)
	}
	stat, err = os.Stat(filepath.Join(path, "challenge", "src"))
	if os.IsNotExist(err) || !stat.IsDir() {
		return errors.New("Missing Directory /challenge/src in provided path: " + path)
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
