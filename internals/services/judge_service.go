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

// JudgeService is responsible for evaluating submissions against judge criteria.
type JudgeService struct {
	runnerFactory *workspace.DockerRunnerFactory
}

func NewJudgeService(runnerFactory *workspace.DockerRunnerFactory) *JudgeService {
	return &JudgeService{runnerFactory: runnerFactory}
}

// Judge is the entry point for judge-related evaluation flow.
func (s *JudgeService) Judge(dto dtos.JudgeRequestDto) (*dtos.JudgeResponseDto, error) {
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
	// Always tear down the per-submission container and workspace.
	defer func() {
		if err := runner.Cleanup(); err != nil {
			log.Printf("judge service cleanup artifact: %v", err)
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
		return &dtos.JudgeResponseDto{
			Status: constants.StatusCompilationError,
			Stderr: compRes.Error,
		}, nil
	}

	// Each testcase gets its own execution deadline. Compilation is excluded
	// from the user's execution time limit.
	var judgeResult dtos.JudgeResultDto
	var peakMemoryKB, peakTimeNS int64

	for _, testCase := range dto.TestCases {
		runCtx, runCancel := context.WithTimeout(context.Background(), time.Duration(timeLimit+250)*time.Millisecond)
		runnerResp, err := protocol.ExecuteRunner(runCtx, runner, protocol.Request{
			Type:          "run",
			Language:      dto.Language,
			Stdin:         testCase.Input,
			TimeLimit:     timeLimit,
			MemoryLimitKB: memoryLimit,
		})
		runCancel()
		if err != nil {
			if runCtx.Err() != nil {
				judgeResult.Total = len(dto.TestCases)
				return &dtos.JudgeResponseDto{Status: constants.StatusTLE, Stderr: "time limit exceeded", Result: judgeResult, MemoryKB: peakMemoryKB, ExecutionTimeNs: peakTimeNS}, nil
			}
			return nil, fmt.Errorf("execute run request: %w", err)
		}
		peakMemoryKB = max(peakMemoryKB, runnerResp.Metadata.MemoryKB)
		peakTimeNS = max(peakTimeNS, runnerResp.Metadata.RuntimeNS)

		passed := runnerResp.Output == testCase.ExpectedOutput

		judgeResult.TestCases = append(judgeResult.TestCases, dtos.JudgeTestCaseResultDto{
			Input:          testCase.Input,
			ExpectedOutput: testCase.ExpectedOutput,
			ActualOutput:   runnerResp.Output,
			Passed:         passed,
		})
		if !runnerResp.Success {
			status := constants.StatusRuntimeError
			if runnerResp.Metadata.TimedOut {
				status = constants.StatusTLE
			} else if runnerResp.Metadata.MemoryExceeded {
				status = constants.StatusMLE
			}
			judgeResult.Total = len(dto.TestCases)
			return &dtos.JudgeResponseDto{Status: status, Stderr: runnerResp.Error, Result: judgeResult, MemoryKB: peakMemoryKB, ExecutionTimeNs: peakTimeNS}, nil
		}

		if !passed {
			break
		}
		judgeResult.Passed += 1
	}
	judgeResult.Total = len(dto.TestCases)

	return &dtos.JudgeResponseDto{Result: judgeResult, Status: constants.StatusSuccess, MemoryKB: peakMemoryKB,
		ExecutionTimeNs: peakTimeNS}, nil
}
