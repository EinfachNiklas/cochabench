package config

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/EinfachNiklas/cochabench/internal/tools"
	"github.com/urfave/cli/v3"
)

type Config struct {
	LLM_PROVIDER     string
	LLM_BASE_PATH    string
	LLM_MODEL        string
	CHALLENGE_SERVER string
}

func getConfigPath() (string, error) {
	userConfigDirPath, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("Could not get UserConfigDir: %v\n", err)
	}
	return filepath.Join(userConfigDirPath, "cochabench", "config.json"), nil
}

func GetConfig() (*Config, error) {
	configPath, err := getConfigPath()
	if err != nil {
		return nil, err
	}

	var config Config

	_, err = os.Stat(configPath)
	if os.IsNotExist(err) {
		config = Config{LLM_PROVIDER: "anthropic", LLM_BASE_PATH: "https://api.anthropic.com/v1", LLM_MODEL: "claude-sonnet-4-6", CHALLENGE_SERVER: "https://github.com/EinfachNiklas/cochabench-challenges-test/"}
		if err := saveConfig(&config); err != nil {
			return nil, err
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
	configPath, err := getConfigPath()
	if err != nil {
		return err
	}

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

func configFields(c *Config) map[string]*string {
	return map[string]*string{
		"LLM_PROVIDER":     &c.LLM_PROVIDER,
		"LLM_BASE_PATH":    &c.LLM_BASE_PATH,
		"LLM_MODEL":        &c.LLM_MODEL,
		"CHALLENGE_SERVER": &c.CHALLENGE_SERVER,
	}
}

func saveConfig(c *Config) error {
	configPath, err := getConfigPath()
	if err != nil {
		return err
	}
	err = os.MkdirAll(filepath.Dir(configPath), 0755)
	if err != nil {
		return fmt.Errorf("Could not create config directory: %v\n", err)
	}
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("Could not marshal config: %v\n", err)
	}
	return os.WriteFile(configPath, data, 0644)
}

func Show(ctx context.Context, cmd *cli.Command) error {
	config, err := GetConfig()
	if err != nil {
		return err
	}
	tb := tools.NewTableBuilder([]string{"Key", "Value"})
	for key, val := range configFields(config) {
		tb.AddRow([]string{key, *val})
	}
	fmt.Print(tb.Build())
	return nil
}

func Get(ctx context.Context, cmd *cli.Command) error {
	key := cmd.Args().Get(0)
	if len(key) == 0 {
		return fmt.Errorf("No key provided. Usage: cochabench config get <key>")
	}
	config, err := GetConfig()
	if err != nil {
		return err
	}
	fields := configFields(config)
	val, ok := fields[key]
	if !ok {
		return fmt.Errorf("Unknown config key: %s", key)
	}
	fmt.Println(*val)
	return nil
}

func Set(ctx context.Context, cmd *cli.Command) error {
	key := cmd.Args().Get(0)
	value := cmd.Args().Get(1)
	if len(key) == 0 || len(value) == 0 {
		return fmt.Errorf("Usage: cochabench config set <key> <value>")
	}
	config, err := GetConfig()
	if err != nil {
		return err
	}
	fields := configFields(config)
	ptr, ok := fields[key]
	if !ok {
		return fmt.Errorf("Unknown config key: %s", key)
	}
	*ptr = value
	err = saveConfig(config)
	if err != nil {
		return err
	}
	fmt.Printf("Set %s = %s\n", key, value)
	return nil
}
