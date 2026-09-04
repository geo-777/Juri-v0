package services

import (
	"context"
	"fmt"
	"juri/internals/api/dtos"
	"juri/internals/constants"
	"juri/internals/executor/compiler"
	"juri/internals/executor/runner"
	"time"
)

// RunService coordinates compilation and execution for a single submission.
type RunService struct {
	compiler compiler.Compiler
	runner   runner.Runner
}

func NewRunService(c compiler.Compiler, r runner.Runner) *RunService {
	return &RunService{compiler: c, runner: r}
}

// Run compiles a submission, executes it, and returns the observed result.
func (s *RunService) Run(dto dtos.RunRequestDto) (*dtos.RunResponseDto, error) {
	// Use a short timeout to prevent runaway code from blocking the service indefinitely.
	runCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	// Compile the submission into an executable artifact.
	compRes, err := s.compiler.Compile(runCtx, dto.Language, dto.SourceCode)
	if err != nil {
		return nil, fmt.Errorf("run service compile submission: %w", err)
	}
	// defer func() {
	// 	if err := compRes.Runner.Cleanup(); err != nil {
	// 		log.Printf("run service cleanup artifact: %v", err)
	// 	}
	// }()
	if compRes.ExitCode != 0 {
		return &dtos.RunResponseDto{
			ExitCode: compRes.ExitCode,
			Status:   constants.StatusCompilationError,
			Stdout:   compRes.Stdout,
			Stderr:   compRes.Stderr,
		}, nil
	}
	// Execute the compiled artifact with the provided stdin.

	resp, err := s.runner.Run(runCtx, compRes.Runner, dto.Stdin, dto.Language)
	if err != nil {
		return nil, err
	}

	fmt.Println("Output :", resp)

	return &dtos.RunResponseDto{
		Stdout:          resp.Stdout,
		Stderr:          resp.Stderr,
		ExitCode:        resp.ExitCode,
		MemoryKB:        resp.MemoryKB,
		ExecutionTimeNs: resp.RuntimeNS,
	}, nil
}
