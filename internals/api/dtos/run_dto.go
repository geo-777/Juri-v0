package dtos

import "juri/internals/constants"

// request dto
type RunRequestDto struct {
	Language   constants.Language `json:"language" binding:"required,oneof=cpp c java python"`
	SourceCode string             `json:"source_code" binding:"required"`
	Stdin      string             `json:"stdin"`
}

// response dto
type RunResponseDto struct {
	Status          constants.Status `json:"status"`
	Stdout          string           `json:"stdout"`
	Stderr          string           `json:"stderr"`
	ExitCode        int              `json:"exit_code"`
	ExecutionTimeNs int64            `json:"execution_time_ns"`
	MemoryKB        int64            `json:"memory_kb"`
}
