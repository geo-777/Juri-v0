package dtos

// Language enums and type
type Language string

const (
	CPP    Language = "cpp"
	C      Language = "c"
	Java   Language = "java"
	Python Language = "python"
)

// Status enums and type
type Status string

const (
	StatusSuccess          Status = "success"
	StatusCompilationError Status = "compilation_error"
	StatusRuntimeError     Status = "runtime_error"
	StatusTLE              Status = "time_limit_exceeded"
	StatusMLE              Status = "memory_limit_exceeded"
)

// request dto
type RunRequestDto struct {
	Language   Language `json:"language" binding:"required,oneof=cpp c java python"`
	SourceCode string   `json:"source_code" binding:"required"`
	Stdin      string   `json:"stdin"`
}

// response dto
type RunResponseDto struct {
	Status          Status `json:"status"`
	Stdout          string `json:"stdout"`
	Stderr          string `json:"stderr"`
	ExitCode        int    `json:"exit_code"`
	ExecutionTimeMs int64  `json:"execution_time_ms"`
}
