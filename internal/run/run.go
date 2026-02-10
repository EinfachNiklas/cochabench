package run

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/urfave/cli/v3"

	cochabenchdata "github.com/EinfachNiklas/cochabench/internal/cochabenchData"
	"github.com/EinfachNiklas/cochabench/internal/tools"
)

func loadEntry(dirPath string, id string) (*cochabenchdata.CochabenchEntry, error) {

	if len(id) == 0 {
		return nil, errors.New("No ID provided")
	}

	err := tools.ValidateDirPath(dirPath)
	if err != nil {
		return nil, err
	}
	err = tools.ValidateDirStructure(dirPath)
	if err != nil {
		return nil, err
	}

	store, err := cochabenchdata.LoadCochabenchStore(dirPath)
	if err != nil {
		return nil, err
	}
	entry, ok := (*store)[id]
	if !ok {
		return nil, errors.New("RunID not found")
	}
	return &entry, nil
}

func Init(ctx context.Context, cmd *cli.Command) error {
	dirPath := cmd.Args().Get(0)
	if len(dirPath) == 0 {
		dirPath = "./"
	}
	err := tools.ValidateDirPath(dirPath)
	if err != nil {
		return err
	}
	err = tools.ValidateDirStructure(dirPath)
	if err != nil {
		return err
	}
	name := cmd.String("name")
	id := uuid.NewString()
	entry := cochabenchdata.CochabenchEntry{
		RunName:   name,
		RunID:     id,
		RunStatus: "I",
	}

	err = entry.Write(dirPath)
	if err != nil {
		return err
	}

	solutionPath := filepath.Join(dirPath, "solutions", entry.RunID)
	err = os.MkdirAll(solutionPath, 0777)

	if err != nil {
		return errors.New("Could not create solution directory " + solutionPath)
	}
	err = os.CopyFS(solutionPath, os.DirFS(filepath.Join(dirPath, "src")))
	if err != nil {
		return fmt.Errorf("Could not copy source files to solutions directory: %w", err)
	}
	fmt.Printf("Initialized run %s[%s] successfully\n", entry.RunName, entry.RunID)
	return nil
}

func Start(ctx context.Context, cmd *cli.Command) error {
	dirPath := cmd.Args().Get(0)
	if len(dirPath) == 0 {
		dirPath = "./"
	}
	id := cmd.String("id")
	entry, err := loadEntry(dirPath, id)
	if err != nil {
		return err
	}
	switch entry.RunStatus {
	case "R":
		return errors.New("Run " + entry.RunName + "[" + id + "] is already running\n")
	case "F":
		return errors.New("Run " + entry.RunName + "[" + id + "] is already finnished\n")
	case "I", "C":
		entry.StartTime = time.Now()
		entry.RunStatus = "R"
	}

	err = entry.Write(dirPath)
	if err != nil {
		return err
	}
	fmt.Printf("Run %s[%s] started successfully\n", entry.RunName, entry.RunID)
	return nil
}

func Stop(ctx context.Context, cmd *cli.Command) error {
	dirPath := cmd.Args().Get(0)
	if len(dirPath) == 0 {
		dirPath = "./"
	}
	id := cmd.String("id")
	entry, err := loadEntry(dirPath, id)
	if err != nil {
		return err
	}
	switch entry.RunStatus {
	case "R":
		entry.EndTime = time.Now()
		entry.RunStatus = "F"
	case "F":
		return errors.New("Run " + entry.RunName + "[" + id + "] is already finnished\n")
	case "I", "C":
		return errors.New("Run " + entry.RunName + "[" + id + "] is not running\n")
	}

	err = entry.Write(dirPath)
	if err != nil {
		return err
	}
	fmt.Printf("Run %s[%s] stopped successfully\n", entry.RunName, entry.RunID)
	return nil
}

func Cancel(ctx context.Context, cmd *cli.Command) error {
	dirPath := cmd.Args().Get(0)
	if len(dirPath) == 0 {
		dirPath = "./"
	}
	id := cmd.String("id")
	entry, err := loadEntry(dirPath, id)
	if err != nil {
		return err
	}
	switch entry.RunStatus {
	case "R":
		entry.RunStatus = "C"
	case "F":
		return errors.New("Run " + entry.RunName + "[" + id + "] is already finnished\n")
	case "I", "C":
		return errors.New("Run " + entry.RunName + "[" + id + "] is not running\n")
	}

	err = entry.Write(dirPath)
	if err != nil {
		return err
	}
	fmt.Printf("Run %s[%s] was canceled successfully\n", entry.RunName, entry.RunID)
	return nil
}

func List(ctx context.Context, cmd *cli.Command) error {
	dirPath := cmd.Args().Get(0)
	if len(dirPath) == 0 {
		dirPath = "./"
	}
	err := tools.ValidateDirPath(dirPath)
	if err != nil {
		return err
	}
	err = tools.ValidateDirStructure(dirPath)
	if err != nil {
		return err
	}
	store, err := cochabenchdata.LoadCochabenchStore(dirPath)
	if err != nil {
		return err
	}
	fmt.Println(store.ToString())
	return nil
}
