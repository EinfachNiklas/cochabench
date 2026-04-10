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

	printIdOnly := cmd.Bool("print-id-only")

	store, err := cochabenchdata.LoadCochabenchStore(dirPath)
	if err != nil {
		return err
	}
	defer store.Close()

	err = store.SaveEntry(&entry)
	if err != nil {
		return err
	}

	solutionPath := filepath.Join(dirPath, "solutions", entry.RunID)
	err = os.MkdirAll(solutionPath, 0777)

	if err != nil {
		return fmt.Errorf("Could not create solution directory: %w", err)
	}
	err = os.CopyFS(solutionPath, os.DirFS(filepath.Join(dirPath, "src")))
	if err != nil {
		return fmt.Errorf("Could not copy starter files into the run directory: %w", err)
	}
	if printIdOnly {
		fmt.Println(entry.RunID)
	} else {
		fmt.Printf("Initialized run %s[%s] successfully\n", entry.RunName, entry.RunID)
	}
	return nil
}

func Start(ctx context.Context, cmd *cli.Command) error {
	dirPath := cmd.Args().Get(0)
	if len(dirPath) == 0 {
		dirPath = "./"
	}
	id := cmd.String("id")

	store, err := cochabenchdata.LoadCochabenchStore(dirPath)
	if err != nil {
		return err
	}
	defer store.Close()

	entry, found, err := store.GetEntry(id)
	if err != nil {
		return err
	}
	if !found {
		return errors.New("Run not found")
	}
	switch entry.RunStatus {
	case "R":
		return errors.New("Run is already running")
	case "F":
		return errors.New("Run is already finished")
	case "I", "C":
		entry.StartTime = time.Now()
		entry.RunStatus = "R"
	}

	err = store.SaveEntry(entry)
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

	store, err := cochabenchdata.LoadCochabenchStore(dirPath)
	if err != nil {
		return err
	}
	defer store.Close()

	entry, found, err := store.GetEntry(id)
	if err != nil {
		return err
	}
	if !found {
		return errors.New("Run not found")
	}

	switch entry.RunStatus {
	case "R":
		entry.EndTime = time.Now()
		entry.RunStatus = "F"
		entry.Duration = entry.EndTime.Sub(entry.StartTime)
	case "F":
		return errors.New("Run is already finished")
	case "I", "C":
		return errors.New("Run is not running")
	}

	err = store.SaveEntry(entry)
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

	store, err := cochabenchdata.LoadCochabenchStore(dirPath)
	if err != nil {
		return err
	}
	defer store.Close()

	entry, found, err := store.GetEntry(id)
	if err != nil {
		return err
	}
	if !found {
		return errors.New("Run not found")
	}

	switch entry.RunStatus {
	case "R":
		entry.RunStatus = "C"
	case "F":
		return errors.New("Run is already finished")
	case "I", "C":
		return errors.New("Run is not running")
	}

	err = store.SaveEntry(entry)
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
	defer store.Close()
	fmt.Println(store.ToString())
	return nil
}
