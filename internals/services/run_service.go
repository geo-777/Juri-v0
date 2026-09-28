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
	runnerFactory *workspace.DockerRunnerFactory
}

func NewRunService(runnerFactory *workspace.DockerRunnerFactory) *RunService {
	return &RunService{runnerFactory: runnerFactory}
}

// Run compiles a submission, executes it, and returns the observed result.
func (s *RunService) Run(dto dtos.RunRequestDto) (*dtos.RunResponseDto, error) {
	timeLimit := dto.TimeLimitMs
	if timeLimit == 0 {
		timeLimit = 2000
	}
	memoryLimit := dto.MemoryLimit
	if memoryLimit == 0 {
		memoryLimit = 128 * 1024
	}
	setupCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Create runner artifact.
	runner, err := s.runnerFactory.CreateRunner(setupCtx, dto.Language, dto.SourceCode, memoryLimit)
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
	compRes, err := protocol.ExecuteRunner(setupCtx, runner, protocol.Request{
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

	runCtx, runCancel := context.WithTimeout(context.Background(), time.Duration(timeLimit+250)*time.Millisecond)
	defer runCancel()
	runnerResp, err := protocol.ExecuteRunner(runCtx, runner, protocol.Request{
		Type:          "run",
		Language:      dto.Language,
		Stdin:         dto.Stdin,
		TimeLimit:     timeLimit,
		MemoryLimitKB: memoryLimit,
	})
	if err != nil {
		if runCtx.Err() != nil {
			return &dtos.RunResponseDto{Status: constants.StatusTLE, Stderr: "time limit exceeded", ExitCode: 1}, nil
		}
		return nil, fmt.Errorf("execute run request: %w", err)
	}
	status := constants.StatusSuccess
	switch {
	case runnerResp.Metadata.TimedOut:
		status = constants.StatusTLE
	case runnerResp.Metadata.MemoryExceeded:
		status = constants.StatusMLE
	case !runnerResp.Success:
		status = constants.StatusRuntimeError
	}
	if !runnerResp.Success && runnerResp.Metadata.ExitCode == 0 {
		runnerResp.Metadata.ExitCode = 1
	}

	return &dtos.RunResponseDto{
		Status:          status,
		Stdout:          runnerResp.Output,
		Stderr:          runnerResp.Error,
		ExitCode:        runnerResp.Metadata.ExitCode,
		MemoryKB:        runnerResp.Metadata.MemoryKB,
		ExecutionTimeNs: runnerResp.Metadata.RuntimeNS,
	}, nil
}
