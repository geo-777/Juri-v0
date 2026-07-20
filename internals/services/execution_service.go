package services

import (
	"code-runner/internals/api/dtos"
	"code-runner/internals/executor"
)

type ExecutionService struct {
	executor *executor.Executor
}

func NewExecutionService(executor *executor.Executor) *ExecutionService {
	return &ExecutionService{
		executor: executor,
	}
}

func (s *ExecutionService) Run(req *dtos.RunRequestDto) error {

	_, err := s.executor.Execute(
		req.Language,
		req.SourceCode,
	)

	if err != nil {
		return err
	}

	return nil
}
