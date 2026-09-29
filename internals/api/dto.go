package api

import (
	"juri/internals/constants"
)

// Request DTOs
type SubmissionTestCaseDto struct {
	Input          string `json:"input" binding:"required"`
	ExpectedOutput string `json:"expected_output" binding:"required"`
}

// SubmissionRequestDto captures the input payload for a judge request.
type SubmissionRequestDto struct {
	Language constants.Language `json:"language" binding:"required,oneof=cpp c java python"`

	SourceCode string `json:"source_code" binding:"required"`

	TestCases []SubmissionTestCaseDto `json:"test_cases" binding:"required,min=1,dive"`

	// Optional overrides
	TimeLimitMs int `json:"time_limit_ms" binding:"omitempty,gte=1,lte=15000"`
	MemoryLimit int `json:"memory_limit_kb" binding:"omitempty,gte=1,lte=1048576"`
}

// Response DTOs
type JudgeTestCaseResultDto struct {
	Input          string `json:"input"`
	ExpectedOutput string `json:"expected_output"`
	ActualOutput   string `json:"actual_output"`

	Passed bool `json:"passed"`
}

type JudgeResultDto struct {
	Passed int `json:"passed"`
	Total  int `json:"total"`

	TestCases []JudgeTestCaseResultDto `json:"test_cases,omitempty"`
}

type JudgeResponseDto struct {
	Status constants.Status `json:"status"`
	// Compilation/runtime errors.
	Stderr string `json:"stderr"`
	// Maximum resource usage across all test cases.
	ExecutionTimeNs int64 `json:"execution_time_ns"`
	MemoryKB        int64 `json:"memory_kb"`

	Result JudgeResultDto `json:"result"`
}
