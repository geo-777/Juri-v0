package services

import (
	"context"
	"fmt"
	"juri/internals/api/dtos"
	"juri/internals/constants"
	"juri/internals/executor/protocol"
	"juri/internals/executor/workspace"
	"log"

	"time"
)

// RunService coordinates compilation and execution for a single submission.
type RunService struct {
	runnerFactor *workspace.DockerRunnerFactory
}

func NewRunService(runnerFactor *workspace.DockerRunnerFactory) *RunService {
	return &RunService{runnerFactor: runnerFactor}
}

// Run compiles a submission, executes it, and returns the observed result.
func (s *RunService) Run(dto dtos.RunRequestDto) (*dtos.RunResponseDto, error) {
	// Use a short timeout to prevent runaway code from blocking the service indefinitely.
	runCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Create runner artifact.
	runner, err := s.runnerFactor.CreateRunner(runCtx, dto.Language, dto.SourceCode)
	if err != nil {
		return nil, fmt.Errorf("runner object creation: %w", err)
	}
	//handles cleanup
	defer func() {
		if err := runner.Cleanup(); err != nil {
			log.Printf("run service cleanup artifact: %v", err)
		}
	}()
	//handles compilation
	compRes, err := protocol.ExecuteRunner(runCtx, runner, protocol.Request{
		Type:     "compile",
		Language: dto.Language,
	})
	if err != nil {
		return nil, fmt.Errorf("execute compile request: %w", err)
	}
	if !compRes.Success {
		exitCode := compRes.Metadata.ExitCode
		if exitCode == 0 {
			exitCode = 1
		}
		return &dtos.RunResponseDto{
			Status:   constants.StatusCompilationError,
			Stdout:   compRes.Output,
			Stderr:   compRes.Error,
			ExitCode: exitCode,
		}, nil
	}
	// Execute the compiled artifact with the provided stdin.

	runnerResp, err := protocol.ExecuteRunner(runCtx, runner, protocol.Request{
		Type:     "run",
		Language: dto.Language,
		Stdin:    dto.Stdin,
	})

	if !runnerResp.Success {
		runnerResp.Metadata.ExitCode = 1

	}

	return &dtos.RunResponseDto{
		Stdout:          runnerResp.Output,
		Stderr:          runnerResp.Error,
		ExitCode:        runnerResp.Metadata.ExitCode,
		MemoryKB:        runnerResp.Metadata.MemoryKB,
		ExecutionTimeNs: runnerResp.Metadata.RuntimeNS,
	}, nil
}
