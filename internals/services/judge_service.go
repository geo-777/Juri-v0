package services

import (
	"juri/internals/executor/compiler"
	"juri/internals/executor/runner"
)

// JudgeService is responsible for evaluating submissions against judge criteria.
type JudgeService struct {
	compiler compiler.Compiler
	runner   runner.Runner
}

func NewJudgeService(c compiler.Compiler, r runner.Runner) *JudgeService {
	return &JudgeService{compiler: c, runner: r}
}

// Judge is the entry point for judge-related evaluation flow.
func (s *JudgeService) Judge() {
	// TODO: implement the judge workflow.
}
