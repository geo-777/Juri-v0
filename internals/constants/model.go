package constants

type SubmissionEntity struct {
	ID              int64                      `db:"id"`
	Language        Language                   `db:"language"`
	TestCases       []SubmissionTestCaseEntity `db:"test_cases"`
	CallbackURL     string                     `db:"callback_url"`
	TimeLimitMs     int                        `db:"time_limit_ms"`
	MemoryLimit     int                        `db:"memory_limit_kb"`
	SourceCode      string                     `db:"source_code"`
	Status          Status                     `db:"status"`
	Stderr          string                     `db:"stderr"`
	ExecutionTimeNs int64                      `db:"execution_time_ns"`
	MemoryKB        int64                      `db:"memory_kb"`
	Result          JudgeResultEntity          `db:"result"`
}

type SubmissionTestCaseEntity struct {
	Input          string `json:"input"`
	ExpectedOutput string `json:"expected_output"`
}

type JudgeTestCaseResultEntity struct {
	Input          string `json:"input"`
	ExpectedOutput string `json:"expected_output"`
	ActualOutput   string `json:"actual_output"`
	Passed         bool   `json:"passed"`
}

type JudgeResultEntity struct {
	Passed    int                         `json:"passed"`
	Total     int                         `json:"total"`
	TestCases []JudgeTestCaseResultEntity `json:"test_cases,omitempty"`
}
