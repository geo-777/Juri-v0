package dtos

import "juri/internals/constants"

// RunRequestDto captures the input payload for a compile-and-run request.
type RunRequestDto struct {
	Language    constants.Language `json:"language" binding:"required,oneof=cpp c java python"`
	SourceCode  string             `json:"source_code" binding:"required"`
	Stdin       string             `json:"stdin"`
	TimeLimitMs int                `json:"time_limit_ms" binding:"omitempty,gte=1,lte=15000"`
	MemoryLimit int                `json:"memory_limit_kb" binding:"omitempty,gte=1,lte=1048576"`
}

// RunResponseDto carries the outcome of a submission execution.
type RunResponseDto struct {
	Status          constants.Status `json:"status"`
	Stdout          string           `json:"stdout"`
	Stderr          string           `json:"stderr"`
	ExitCode        int              `json:"exit_code"`
	ExecutionTimeNs int64            `json:"execution_time_ns"`
	MemoryKB        int64            `json:"memory_kb"`
}
