package executor

import "juri/internals/constants"

// ExecutionArtifact describes the temporary execution workspace created for a submission.
type ExecutionArtifact struct {
	ContainerID    string
	ContainerName  string
	SourceFilePath string
	SourceDirPath  string

	Cleanup func() error
}

// CompileResult stores the output of a compilation step.
type CompileResult struct {
	Artifact *ExecutionArtifact
	Stderr   string
	Stdout   string
	ExitCode int
}

// RunnerResponse stores the result of a runtime execution.
type RunnerResponse struct {
	Stdout string
	Stderr string
	//metadata
	RuntimeNS int64
	MemoryKB  int64
	ExitCode  int

	Status constants.Status
}
