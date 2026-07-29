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

type RunService struct {
	compiler compiler.Compiler
	runner   runner.Runner
}

func NewRunService(c compiler.Compiler, r runner.Runner) *RunService {
	return &RunService{compiler: c, runner: r}
}

func (s *RunService) Run(dto dtos.RunRequestDto) (*dtos.RunResponseDto, error) {
	runCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	//calling compiler service
	//this returns an artifact with container ID and workspace path
	compRes, err := s.compiler.Compile(runCtx, dto.Language, dto.SourceCode)
	if err != nil {
		return nil, fmt.Errorf("Compile submission : %w", err)
	}
	if compRes.ExitCode != 0 {
		return &dtos.RunResponseDto{
			ExitCode: compRes.ExitCode,
			Status:   constants.StatusCompilationError,
			Stdout:   compRes.Stdout,
			Stderr:   compRes.Stderr,
		}, nil
	}
	defer compRes.Artifact.Cleanup()

	//calling runner service
	runResp, err := s.runner.Run(runCtx, compRes.Artifact, dto.Stdin, dto.Language)
	if err != nil {
		return nil, fmt.Errorf("Run submission : %w", err)
	}
	return &dtos.RunResponseDto{
		Stdout:   runResp.Stdout,
		Stderr:   runResp.Stderr,
		ExitCode: runResp.ExitCode,
	}, nil
}
