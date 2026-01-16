package timer

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/urfave/cli/v3"

	"github.com/EinfachNiklas/cochabench/internal/tools"
)

func Start(ctx context.Context, cmd *cli.Command) error {
	dirPath := cmd.Args().Get(0)
	err := tools.ValidateDirPath(dirPath)
	if err != nil {
		return err
	}
	err = tools.ValidateDirStructure(dirPath)
	if err != nil {
		return err
	}

	var cochabenchData *tools.CochabenchData
	dataPath := filepath.Join(dirPath, "cochabench.json")
	_, err = os.Stat(dataPath)
	if os.IsNotExist(err) {
		cochabenchData = &tools.CochabenchData{
			RunName:   uuid.NewString(),
			StartTime: time.Now(),
		}
	} else {
		cochabenchData, err = tools.LoadCochabenchData(dataPath)
		if err != nil {
			return err
		}
		if len(cochabenchData.RunName) == 0 {
			cochabenchData = &tools.CochabenchData{
				RunName: uuid.NewString(),
			}
		}
		if cochabenchData.StartTime.IsZero() {
			cochabenchData.StartTime = time.Now()
		} else {
			return errors.New("Run " + cochabenchData.RunName + " is already running")
		}
	}

	err = tools.WriteCochabenchData(cochabenchData, dataPath)
	if err != nil {
		return err
	}
	fmt.Printf("Successfully started run %s\n", cochabenchData.RunName)
	return nil
}

func Stop(ctx context.Context, cmd *cli.Command) error {
	dirPath := cmd.Args().Get(0)
	err := tools.ValidateDirPath(dirPath)
	if err != nil {
		return err
	}
	err = tools.ValidateDirStructure(dirPath)
	if err != nil {
		return err
	}

	var cochabenchData *tools.CochabenchData
	dataPath := filepath.Join(dirPath, "cochabench.json")
	_, err = os.Stat(dataPath)
	if os.IsNotExist(err) {
		return errors.New("There is no run to stop")
	} else {
		cochabenchData, err = tools.LoadCochabenchData(dataPath)
		if err != nil {
			return err
		}
		if cochabenchData.StartTime.IsZero() {
			return errors.New("There is no run to stop")
		} else if !cochabenchData.EndTime.IsZero() {
			return errors.New("Run " + cochabenchData.RunName + " is already finished")
		} else {
			cochabenchData.EndTime = time.Now()
		}
	}

	err = tools.WriteCochabenchData(cochabenchData, dataPath)
	if err != nil {
		return err
	}
	fmt.Printf("Successfully stopped run %s\n", cochabenchData.RunName)
	return nil
}

func Cancel(ctx context.Context, cmd *cli.Command) error {
	dirPath := cmd.Args().Get(0)

	err := tools.ValidateDirPath(dirPath)
	if err != nil {
		return err
	}
	err = tools.ValidateDirStructure(dirPath)
	if err != nil {
		return err
	}

	var cochabenchData *tools.CochabenchData
	dataPath := filepath.Join(dirPath, "cochabench.json")
	_, err = os.Stat(dataPath)
	if os.IsNotExist(err) {
		return errors.New("There is no run to cancel")
	}

	cochabenchData, err = tools.LoadCochabenchData(dataPath)
	if err != nil {
		return err
	}
	if cochabenchData.StartTime.IsZero() {
		return errors.New("There is no run to cancel")
	} else if !cochabenchData.EndTime.IsZero() {
		return errors.New("Run " + cochabenchData.RunName + " is already finished")
	} else {
		err = tools.WriteCochabenchData(&tools.CochabenchData{}, dataPath)
	}

	if err != nil {
		return err
	}
	fmt.Printf("Successfully canceled run %s \n", cochabenchData.RunName)
	return nil
}
