package services

import (
	"code-runner/internals/api/dtos"
	"code-runner/internals/config"
	"code-runner/pkg/utils"
	"fmt"
)

func RunService(req *dtos.RunRequestDto, cfg *config.Config) error {

	//creates file
	filePath, err := utils.TempFileCreator(req.Language, req.SourceCode, cfg.WorkspaceRoot)
	fmt.Println(filePath)
	if err != nil {
		return err
	}

	return nil
}
