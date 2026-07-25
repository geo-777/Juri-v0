package executor

import "juri/internals/constants"

type WorkspaceArtifact struct {
	DockerContainer string
	FilePath        string
}

type RunnerResponse struct {
	Stdout string
	Stderr string

	RuntimeNS int64
	MemoryKB  int64
	ExitCode  int

	Status constants.Status
}
