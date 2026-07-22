package executor

import (
	"code-runner/internals/api/dtos"
	"code-runner/internals/config"
	"code-runner/pkg/constants"
	"log"
	"os"
	"path/filepath"
	"strconv"
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

type ExecutionMetadata struct {
	RuntimeNS int64
	MemoryKB  int64
	ExitCode  int
}
type ExecutionData struct {
	Stderr   string
	Metadata ExecutionMetadata
	Output   string
	Status   constants.Status
}

func (e *Executor) Execute(language dtos.Language, sourceCode string, stdIn string) (ExecutionData, error) {
	var executionData ExecutionData

	//handles creation of workspace
	filePath, err := CreateWorkspace(
		language,
		sourceCode,
		e.cfg.WorkspaceRoot,
	)

	if err != nil {
		return ExecutionData{}, err
	}
	//iniates cleanup on return
	defer cleanupDir(filePath)
	// Docker
	stdout, stderr, err := e.RunDocker(filePath, language)

	if err != nil {
		erroredOutput := ExecutionData{
			Metadata: ExecutionMetadata{ExitCode: 1},
			Output:   stdout, Stderr: stderr,
		}

		if err.Error() == "time limit exceeded" {
			erroredOutput.Status = constants.StatusTLE
		}

		return erroredOutput, nil
	}
	executionData.Output = stdout

	//parsing output to fetch memory and time
	metaData, err := getMetaData(filePath)

	if err != nil {
		return ExecutionData{}, err
	}
	executionData.Metadata = metaData
	executionData.Status = constants.StatusSuccess
	return executionData, nil
}

// fetchees metaData.txt data from job dir
func getMetaData(filePath string) (ExecutionMetadata, error) {
	data, err := os.ReadFile(filepath.Join(filepath.Dir(filePath), "metadata.txt"))
	if err != nil {
		return ExecutionMetadata{}, err
	}

	var meta ExecutionMetadata

	metaData := strings.Split(string(data), "\n")
	//parsing by line
	for _, line := range metaData {
		line = strings.TrimSpace(line)

		switch {

		case strings.HasPrefix(line, "runtime_ns="):
			meta.RuntimeNS, err = strconv.ParseInt(strings.TrimPrefix(line, "runtime_ns="), 10, 64)

		case strings.HasPrefix(line, "memory_kb="):
			meta.MemoryKB, err = strconv.ParseInt(strings.TrimPrefix(line, "memory_kb="), 10, 64)

		case strings.HasPrefix(line, "exit_code="):
			meta.ExitCode, err = strconv.Atoi(strings.TrimPrefix(line, "exit_code="))
		}
		if err != nil {
			return ExecutionMetadata{}, err
		}
	}

	return meta, nil
}

// cleansup workspace-dir
func cleanupDir(filePath string) {
	if err := RemoveWorkspace(filePath); err != nil {
		log.Printf("Failed to remove workspace : %v", err)
	}
}
