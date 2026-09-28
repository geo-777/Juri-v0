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
	fmt.Println("I WAS SUMMONED")
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
		return &dtos.JudgeResponseDto{
			Status: constants.StatusCompilationError,
			Stderr: compRes.Error,
		}, nil
	}

	//judgement starts here
	runCtx, runCancel := context.WithTimeout(context.Background(), time.Duration(timeLimit+250)*time.Millisecond)
	defer runCancel()
	var judgeResult dtos.JudgeResultDto

	var memoryTotal int64
	var timeTotal int64

	for _, testCase := range dto.TestCases {
		runnerResp, err := protocol.ExecuteRunner(runCtx, runner, protocol.Request{
			Type:          "run",
			Language:      dto.Language,
			Stdin:         testCase.Input,
			TimeLimit:     timeLimit,
			MemoryLimitKB: memoryLimit,
		})
		passed := runnerResp.Output == testCase.ExpectedOutput

		if err != nil {
			if runCtx.Err() != nil {
				return &dtos.JudgeResponseDto{Status: constants.StatusTLE, Stderr: "time limit exceeded"}, nil
			}
			return nil, fmt.Errorf("execute run request: %w", err)
		}

		judgeResult.TestCases = append(judgeResult.TestCases, dtos.JudgeTestCaseResultDto{
			Input:          testCase.Input,
			ExpectedOutput: testCase.ExpectedOutput,
			ActualOutput:   runnerResp.Output,
			Passed:         passed,
		})

		if !passed {
			break
		}
		judgeResult.Passed += 1
		passed = true
		memoryTotal += max(runnerResp.Metadata.MemoryKB, memoryTotal)
		timeTotal += runnerResp.Metadata.RuntimeNS
	}
	judgeResult.Total = len(dto.TestCases)

	return &dtos.JudgeResponseDto{Result: judgeResult, Status: constants.StatusSuccess, MemoryKB: memoryTotal,
		ExecutionTimeNs: timeTotal}, nil
}
