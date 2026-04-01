package tools

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"
	"time"
)

type ChallengeConfig struct {
	Name          string
	ChallengeID   string
	ChallengeType string
}

// TableBuilder helps build formatted ASCII tables with automatic column width calculation
type TableBuilder struct {
	headers []string
	rows    [][]string
}

type Env struct {
	LLM_API_KEY string
}

// NewTableBuilder creates a new table builder with the specified headers
func NewTableBuilder(headers []string) *TableBuilder {
	return &TableBuilder{
		headers: headers,
		rows:    make([][]string, 0),
	}
}

// AddRow adds a data row to the table
func (tb *TableBuilder) AddRow(row []string) {
	tb.rows = append(tb.rows, row)
}

// Build generates the formatted table string
func (tb *TableBuilder) Build() string {
	if len(tb.headers) == 0 {
		return ""
	}

	// Calculate column widths based on headers and data
	widths := make([]int, len(tb.headers))
	for i, header := range tb.headers {
		widths[i] = len(header)
	}

	for _, row := range tb.rows {
		for i, cell := range row {
			if i < len(widths) && len(cell) > widths[i] {
				widths[i] = len(cell)
			}
		}
	}

	var b strings.Builder

	// Build header row
	for i, header := range tb.headers {
		if i > 0 {
			b.WriteString(" | ")
		}
		fmt.Fprintf(&b, "%-*s", widths[i], header)
	}
	b.WriteString("\n")

	// Build separator line
	for i, width := range widths {
		if i > 0 {
			b.WriteString("-+-")
		}
		b.WriteString(strings.Repeat("-", width))
	}
	b.WriteString("\n")

	// Build data rows
	for _, row := range tb.rows {
		for i, cell := range row {
			if i > 0 {
				b.WriteString(" | ")
			}
			if i < len(widths) {
				fmt.Fprintf(&b, "%-*s", widths[i], cell)
			} else {
				b.WriteString(cell)
			}
		}
		b.WriteString("\n")
	}

	return b.String()
}

func ValidateDirPath(path string) error {
	stat, err := os.Stat(path)
	if err != nil {
		return errors.New("Directory does not exist")
	}
	if !stat.IsDir() {
		return errors.New("Path is not a directory")
	}
	return nil
}

func ValidateDirStructure(path string) error {
	// Check for src directory
	stat, err := os.Stat(filepath.Join(path, "src"))
	if err != nil || !stat.IsDir() {
		return errors.New("Challenge directory is missing required folder: src")
	}

	// Check for test directory
	stat, err = os.Stat(filepath.Join(path, "test"))
	if err != nil || !stat.IsDir() {
		return errors.New("Challenge directory is missing required folder: test")
	}

	// Check for config.json file
	stat, err = os.Stat(filepath.Join(path, "challenge.config.json"))
	if os.IsNotExist(err) {
		return errors.New("Challenge directory is missing required file: challenge.config.json")
	}
	if err != nil {
		return errors.New("Could not access challenge.config.json")
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
		return nil, errors.New("challenge.config.json is invalid")
	}
	return &config, nil
}

func LoadEnv() (*Env, error) {
	LLM_API_KEY := os.Getenv("LLM_API_KEY")
	if len(LLM_API_KEY) == 0 {
		return nil, errors.New("LLM_API_KEY is not set")
	}

	return &Env{
		LLM_API_KEY: LLM_API_KEY,
	}, nil
}
func FmtTime(t time.Time) string {
	if t.IsZero() {
		return "-"
	}
	return t.Local().Format("2006-01-02 15:04:05")
}

func GetBuildVersion(externalVersion string) string {
	info, ok := debug.ReadBuildInfo()
	return resolveBuildVersion(externalVersion, info, ok)
}

func resolveBuildVersion(externalVersion string, info *debug.BuildInfo, ok bool) string {
	if externalVersion != "" && externalVersion != "dev" {
		return externalVersion
	}

	if !ok || info == nil {
		return externalVersion
	}

	if info.Main.Version != "" && info.Main.Version != "(devel)" {
		return info.Main.Version
	}

	for _, setting := range info.Settings {
		if setting.Key != "vcs.revision" || setting.Value == "" {
			continue
		}
		if len(setting.Value) > 7 {
			return setting.Value[:7]
		}
		return setting.Value
	}

	return externalVersion
}
