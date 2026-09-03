package executor

import (
	"encoding/json"
	"juri/internals/constants"
	"net"
)

// Runner represents a reusable persistent runner.
type Runner struct {
	Conn        net.Conn
	Reader      *json.Decoder
	Writer      *json.Encoder
	ContainerID string

	SourceFilePath string
	Cleanup        func() error
}

// CompileResult stores the output of a compilation step.
type CompileResult struct {
	Runner   *Runner
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
