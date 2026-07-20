package executor

import (
	"code-runner/internals/api/dtos"
	"code-runner/internals/config"
)

type Executor struct {
	cfg *config.Config
}

func New(cfg *config.Config) *Executor {
	return &Executor{
		cfg: cfg,
	}
}

func (e *Executor) Execute(language dtos.Language, sourceCode string) (string, error) {
	//handles creation of workspace
	filePath, err := CreateWorkspace(
		language,
		sourceCode,
		e.cfg.WorkspaceRoot,
	)

	if err != nil {
		return "", err
	}

	// Docker
	err = e.RunDocker(filePath)

	// Execute
	// Cleanup

	return filePath, nil
}
