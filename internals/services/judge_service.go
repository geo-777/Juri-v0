package services

import (
	"context"
	"errors"
	"fmt"
	"juri/internals/api/dtos"
	"juri/internals/constants"
	"juri/internals/executor/compiler"
	"juri/internals/executor/runner"
	"log"
	"time"
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
func (s *JudgeService) Judge(dto dtos.JudgeRequestDto) (*dtos.JudgeResponseDto, error) {
	var judgeResult dtos.JudgeResultDto
	//context setup
	runCtx, cancel := context.WithTimeout(context.Background(), time.Second*100)
	defer cancel()

	// Compile the submission into an executable artifact.
	compRes, err := s.compiler.Compile(runCtx, dto.Language, dto.SourceCode)
	if err != nil {
		return nil, fmt.Errorf("run service compile submission: %w", err)
	}
	defer func() {
		if err := compRes.Artifact.Cleanup(); err != nil {
			log.Printf("run service cleanup artifact: %v", err)
		}
	}()
	if compRes.ExitCode != 0 {
		return &dtos.JudgeResponseDto{
			Status: constants.StatusCompilationError,
			Stderr: compRes.Stderr,
		}, nil
	}

	for _, testCase := range dto.TestCases {
		passed := false
		//running code with input
		runResp, err := s.runner.Run(runCtx, compRes.Artifact, testCase.Input, dto.Language)
		if err != nil {
			if errors.Is(err, context.DeadlineExceeded) {
				return &dtos.JudgeResponseDto{Status: constants.StatusTLE}, nil
			}
			return nil, fmt.Errorf("run service execute submission: %w", err)
		}

		judgeResult.TestCases = append(judgeResult.TestCases, dtos.JudgeTestCaseResultDto{
			Input:          testCase.Input,
			ExpectedOutput: testCase.ExpectedOutput,
			ActualOutput:   runResp.Stdout,
			Passed:         passed,
		})

		if runResp.Stdout != testCase.ExpectedOutput {

			break
		}
		judgeResult.Passed += 1
		passed = true

	}
	judgeResult.Total = len(dto.TestCases)

	return &dtos.JudgeResponseDto{Result: judgeResult}, nil
}
