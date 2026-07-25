package executor

import (
	"juri/internals/config"
	"juri/pkg/constants"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/moby/moby/client"
)

type Executor struct {
	cfg          *config.Config
	dockerClient *client.Client
}

func New(cfg *config.Config, client *client.Client) *Executor {
	return &Executor{
		cfg:          cfg,
		dockerClient: client,
	}
}

type ExecutionMetadata struct {
	RuntimeNS int64
	MemoryKB  int64
	ExitCode  int
	Phase     string
}
type ExecutionData struct {
	Stderr   string
	Metadata ExecutionMetadata
	Output   string
	Status   constants.Status
}

func (e *Executor) Execute(language constants.Language, sourceCode string, stdIn string) (ExecutionData, error) {

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
	dockerResponse, err := e.RunDocker(filePath, language, stdIn)

	if err != nil {
		erroredOutput := ExecutionData{
			Metadata: ExecutionMetadata{ExitCode: 1},
			Output:   dockerResponse.Stdout, Stderr: dockerResponse.Stderr,
			Status: dockerResponse.Status,
		}

		return erroredOutput, nil
	}

	//parsing output to fetch memory,time and exit code
	metaData := ExecutionMetadata{}
	if dockerResponse.Status != constants.StatusTLE {
		metaData, err = getMetaData(filePath)
		if err != nil {
			return ExecutionData{}, err
		}

		if metaData.Phase == "COMPILATION" {
			dockerResponse.Status = constants.StatusCompilationError
		}
	} else {
		metaData.ExitCode = 1 //for TLEs
	}

	return ExecutionData{
		Metadata: metaData,
		Output:   dockerResponse.Stdout,
		Stderr:   dockerResponse.Stderr,
		Status:   dockerResponse.Status,
	}, nil
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
		case strings.HasPrefix(line, "phase="):
			meta.Phase = strings.TrimPrefix(line, "phase=")
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
