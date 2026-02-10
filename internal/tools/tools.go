package tools

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
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
	stat, err := os.Stat(filepath.Join(path, "src"))
	if err != nil || !stat.IsDir() {
		return errors.New("Missing Directory 'src' in provided path: " + path)
	}
	stat, err = os.Stat(filepath.Join(path, "test"))
	if os.IsNotExist(err) || !stat.IsDir() {
		return errors.New("Missing Directory 'test' in provided path: " + path)
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
