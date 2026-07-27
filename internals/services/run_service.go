package services

import (
	"context"
	"juri/internals/api/dtos"
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
	_, err := s.compiler.Compile(runCtx, dto.Language, dto.SourceCode)
	//defer compRes.Artifact.Cleanup()

	if err != nil {
		return nil, err
	}

	return &dtos.RunResponseDto{}, nil
}
