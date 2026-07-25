package services

import (
	"juri/internals/executor/compiler"
	"juri/internals/executor/runner"
)

type JudgeService struct {
	compiler compiler.Compiler
	runner   runner.Runner
}

func NewJudgeService(c compiler.Compiler, r runner.Runner) *JudgeService {
	return &JudgeService{compiler: c, runner: r}
}

func (s *JudgeService) Judge() {

}
