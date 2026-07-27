package executor

import "juri/internals/constants"

type WorkspaceArtifact struct {
	ContainerID    string
	ContainerName  string
	SourceFilePath string
	SourceDirPath  string

	Cleanup func() error
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
