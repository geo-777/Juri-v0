package constants

// Language represents the supported submission languages.
type Language string

const (
	CPP    Language = "cpp"
	C      Language = "c"
	Java   Language = "java"
	Python Language = "python"
)

// Status describes the outcome of a submission run.
type Status string

const (
	StatusPending Status = "pending"
	StatusQueued  Status = "queued"
	StatusRunning Status = "running"

	StatusSuccess          Status = "success"
	StatusWrongAnswer      Status = "wrong_answer"
	StatusCompilationError Status = "compilation_error"
	StatusRuntimeError     Status = "runtime_error"
	StatusTLE              Status = "time_limit_exceeded"
	StatusMLE              Status = "memory_limit_exceeded"

	StatusSystemError Status = "system_error"
)
