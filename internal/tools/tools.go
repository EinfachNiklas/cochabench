package tools

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
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
		return errors.New("This directory does not exist")
	}
	if !stat.IsDir() {
		return errors.New("The provided path is not a directory")
	}
	return nil
}

func ValidateDirStructure(path string) error {
	// Check for src directory
	stat, err := os.Stat(filepath.Join(path, "src"))
	if err != nil || !stat.IsDir() {
		return errors.New("Missing Directory 'src' in provided path: " + path)
	}

	// Check for test directory
	stat, err = os.Stat(filepath.Join(path, "test"))
	if os.IsNotExist(err) || !stat.IsDir() {
		return errors.New("Missing Directory 'test' in provided path: " + path)
	}

	// Check for config.json file
	stat, err = os.Stat(filepath.Join(path, "challenge.config.json"))
	if os.IsNotExist(err) {
		return errors.New("Missing challenge.config.json file in provided path: " + path)
	}
	if err != nil {
		return errors.New("Error accessing challenge.config.json in provided path: " + path)
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
		return nil, errors.New("Malformed configuration in " + path)
	}
	return &config, nil
}

func LoadEnv() (*Env, error) {
	LLM_API_KEY := os.Getenv("LLM_API_KEY")
	if len(LLM_API_KEY) == 0 {
		return nil, fmt.Errorf("Required Environment Variable LLM_API_KEY is not set")
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
