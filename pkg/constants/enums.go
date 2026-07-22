package constants

// Status enums and type
type Status string

const (
	StatusSuccess          Status = "success"
	StatusCompilationError Status = "compilation_error"
	StatusRuntimeError     Status = "runtime_error"
	StatusTLE              Status = "time_limit_exceeded"
	StatusMLE              Status = "memory_limit_exceeded"
)
