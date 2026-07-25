package services

import (
	"juri/internals/api/dtos"
	"juri/internals/executor/compiler"
	"juri/internals/executor/runner"
)

type RunService struct {
	compiler compiler.Compiler
	runner   runner.Runner
}

func NewRunService(c compiler.Compiler, r runner.Runner) *RunService {
	return &RunService{compiler: c, runner: r}
}

func (s *RunService) Run() (dtos.RunResponseDto, error) {
	return dtos.RunResponseDto{}, nil
}
