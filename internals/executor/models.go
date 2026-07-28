package executor

import "juri/internals/constants"

type ExecutionArtifact struct {
	ContainerID    string
	ContainerName  string
	SourceFilePath string
	SourceDirPath  string

	Cleanup func() error
}

type CompileResult struct {
	Artifact *ExecutionArtifact
	Stderr   string
	Stdout   string
	ExitCode int
}

type RunnerResponse struct {
	Stdout string
	Stderr string
	//metadata
	RuntimeNS int64
	MemoryKB  int64
	ExitCode  int

	Status constants.Status
}
