package services

import (
	"juri/internals/api/dtos"
	"juri/internals/executor"
)

type ExecutionService struct {
	executor *executor.Executor
}

func NewExecutionService(executor *executor.Executor) *ExecutionService {
	return &ExecutionService{
		executor: executor,
	}
}

func (s *ExecutionService) Run(req *dtos.RunRequestDto) (*dtos.RunResponseDto, error) {

	responseData, err := s.executor.Execute(
		req.Language,
		req.SourceCode,
		req.Stdin,
	)

	if err != nil {
		return &dtos.RunResponseDto{}, err
	}

	return &dtos.RunResponseDto{
		Stdout:          responseData.Output,
		ExecutionTimeNs: responseData.Metadata.RuntimeNS,
		MemoryKB:        responseData.Metadata.MemoryKB,
		ExitCode:        responseData.Metadata.ExitCode,
		Stderr:          responseData.Stderr,
		Status:          responseData.Status,
	}, nil
}
