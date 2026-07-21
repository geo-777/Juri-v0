package executor

import (
	"code-runner/internals/api/dtos"
	"code-runner/internals/config"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Executor struct {
	cfg *config.Config
}

func New(cfg *config.Config) *Executor {
	return &Executor{
		cfg: cfg,
	}
}

func (e *Executor) Execute(language dtos.Language, sourceCode string, stdIn string) (string, error) {
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
	output, err := e.RunDocker(filePath, language)

	if err != nil {
		return "", err
	}

	//parsing output to fetch memory and time
	data, err := os.ReadFile(filepath.Join(filepath.Dir(filePath), "metadata.txt"))
	if err != nil {
		return "", err
	}

	var memory string
	var runtime string
	var exitCode string

	metaData := strings.Split(string(data), "\n")

	for _, line := range metaData {
		line = strings.TrimSpace(line)

		switch {

		case strings.HasPrefix(line, "runtime_ns="):
			runtime = strings.TrimPrefix(line, "runtime_ns=")

		case strings.HasPrefix(line, "memory_kb="):
			memory = strings.TrimPrefix(line, "memory_kb=")

		case strings.HasPrefix(line, "exit_code="):
			exitCode = strings.TrimPrefix(line, "exit_code=")
		}
	}

	fmt.Println(memory, runtime, exitCode)
	// Execute
	// Cleanup

	return output, nil
}
