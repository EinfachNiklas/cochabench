package config

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/urfave/cli/v3"
)

type Config struct {
	LLM_PROVIDER     string
	LLM_BASE_PATH    string
	LLM_MODEL        string
	CHALLENGE_SERVER string
}

func GetConfig() (*Config, error) {
	userConfigDirPath, err := os.UserConfigDir()
	if err != nil {
		return nil, fmt.Errorf("Could not get UserConfigDir: %v\n", err)
	}
	configPath := filepath.Join(userConfigDirPath, "cochabench", "config.json")

	var config Config

	_, err = os.Stat(configPath)
	if os.IsNotExist(err) {
		config = Config{LLM_PROVIDER: "anthropic", LLM_BASE_PATH: "https://api.anthropic.com/v1", LLM_MODEL: "claude-sonnet-4-6", CHALLENGE_SERVER: "https://github.com/EinfachNiklas/cochabench-challenges-test/"}
		err = os.MkdirAll(filepath.Dir(configPath), 0755)
		if err != nil {
			return nil, fmt.Errorf("Could not create config directory: %v\n", err)
		}
		data, err := json.MarshalIndent(config, "", "  ")
		if err != nil {
			return nil, fmt.Errorf("Could not marshal config: %v\n", err)
		}
		err = os.WriteFile(configPath, data, 0644)
		if err != nil {
			return nil, fmt.Errorf("Could not write config file: %v\n", err)
		}
	} else if err != nil {
		return nil, fmt.Errorf("Can not access config file: %v\n", err)
	} else {
		d, err := os.ReadFile(configPath)
		if err != nil {
			return nil, fmt.Errorf("Could not read config file %v\n", err)
		}
		err = json.Unmarshal(d, &config)
		if err != nil {
			return nil, fmt.Errorf("Could not parse config file to json: %v\n", err)
		}
	}
	return &config, nil
}

func Initialize(ctx context.Context, cmd *cli.Command) error {
	userConfigDirPath, err := os.UserConfigDir()
	if err != nil {
		return fmt.Errorf("Could not get UserConfigDir: %v\n", err)
	}
	configPath := filepath.Join(userConfigDirPath, "cochabench", "config.json")

	_, err = os.Stat(configPath)
	if os.IsNotExist(err) {
		_, err = GetConfig()
		if err != nil {
			return fmt.Errorf("Could not initialize config: %v\n", err)
		}
		fmt.Printf("Initialized Config at %s\n", configPath)
		return nil
	}

	fmt.Printf("Config already exists at %s\n", configPath)
	return nil
}
