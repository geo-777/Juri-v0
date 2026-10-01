package judge

import (
	"context"
	"fmt"
	"log"
	"time"

	"juri/internals/api"
	"juri/internals/constants"
	"juri/internals/executor"
	"juri/internals/executor/protocol"
)

const runnerSetupTimeout = 30 * time.Second
const testCaseTimeoutBuffer = 250 * time.Millisecond

// RunnerFactory creates the isolated execution environment for one submission.
// TODO (later) : implement container pools
type RunnerFactory interface {
	CreateRunner(context.Context, constants.Language, string, int) (*executor.Runner, error)
}

// JudgeService compiles and evaluates submissions using an execution backend.
type JudgeService struct {
	runnerFactory RunnerFactory
}

func NewJudgeService(runnerFactory RunnerFactory) *JudgeService {
	return &JudgeService{runnerFactory: runnerFactory}
}

// Judge creates an isolated runner, compiles once, and evaluates test cases in order.
func (s *JudgeService) Judge(ctx context.Context, submission api.SubmissionRequestDto) (*api.JudgeResponseDto, error) {
	// Set defaults and stuff if dto doesnt contain limits
	limits := resolveLimits(submission)
	setupCtx, cancel := context.WithTimeout(ctx, runnerSetupTimeout)
	defer cancel()

	// Creates workspace and runner
	runner, err := s.runnerFactory.CreateRunner(setupCtx, submission.Language, submission.SourceCode, limits.memoryKB)
	if err != nil {
		return nil, fmt.Errorf("create runner: %w", err)
	}
	if runner.Cleanup != nil {
		defer cleanupRunner(runner)
	}

	compileStderr, compilationFailed, err := compileSubmission(setupCtx, runner, submission.Language)
	if err != nil {
		return nil, err
	}
	if compilationFailed {
		return &api.JudgeResponseDto{Status: constants.StatusCompilationError, Stderr: compileStderr}, nil
	}

	return evaluateTestCases(setupCtx, runner, submission, limits)
}

type executionLimits struct {
	timeLimitMs int
	memoryKB    int
}

func resolveLimits(submission api.SubmissionRequestDto) executionLimits {
	timeLimit := submission.TimeLimitMs
	if timeLimit == 0 {
		timeLimit = constants.DefaultTimeLimitMs
	}
	memoryLimit := submission.MemoryLimit
	if memoryLimit == 0 {
		memoryLimit = constants.DefaultMemoryLimitKB
	}
	return executionLimits{timeLimitMs: timeLimit, memoryKB: memoryLimit}
}

func cleanupRunner(runner *executor.Runner) {
	if err := runner.Cleanup(); err != nil {
		log.Printf("judge runner cleanup failed: %v", err)
	}
}

func compileSubmission(ctx context.Context, runner *executor.Runner, language constants.Language) (string, bool, error) {
	response, err := protocol.ExecuteRunner(ctx, runner, protocol.Request{
		Type:     "compile",
		Language: language,
	})
	if err != nil {
		return "", false, fmt.Errorf("execute compile request: %w", err)
	}

	if !response.Success {
		return response.Error, true, nil
	}
	return "", false, nil
}

func evaluateTestCases(ctx context.Context, runner *executor.Runner, submission api.SubmissionRequestDto, limits executionLimits) (*api.JudgeResponseDto, error) {
	result := api.JudgeResultDto{Total: len(submission.TestCases)}
	var peakMemoryKB, peakTimeNS int64

	for _, testCase := range submission.TestCases {
		response, timedOut, err := runTestCase(ctx, runner, submission.Language, testCase.Input, limits)
		if timedOut {
			return &api.JudgeResponseDto{
				Status: constants.StatusTLE, Stderr: "time limit exceeded", Result: result,
				MemoryKB: peakMemoryKB, ExecutionTimeNs: peakTimeNS,
			}, nil
		}
		if err != nil {
			return nil, err
		}

		peakMemoryKB = max(peakMemoryKB, response.Metadata.MemoryKB)
		peakTimeNS = max(peakTimeNS, response.Metadata.RuntimeNS)

		// Only returns the failed test case
		if !response.Success {
			// This handles TLE and MLE for testcase
			appendFailedCase(&result, testCase, response.Output)
			status := runtimeStatus(response.Metadata)
			return &api.JudgeResponseDto{
				Status: status, Stderr: response.Error, Result: result,
				MemoryKB: peakMemoryKB, ExecutionTimeNs: peakTimeNS,
			}, nil
		}
		if response.Output != testCase.ExpectedOutput {
			// This handles incorrect answer for testcase
			appendFailedCase(&result, testCase, response.Output)
			break
		}

		result.Passed++
	}

	status := constants.StatusSuccess
	if result.Passed != result.Total {
		status = constants.StatusWrongAnswer
	}
	return &api.JudgeResponseDto{
		Status: status, Result: result, MemoryKB: peakMemoryKB, ExecutionTimeNs: peakTimeNS,
	}, nil
}

func runTestCase(ctx context.Context, runner *executor.Runner, language constants.Language, input string, limits executionLimits) (*protocol.Response, bool, error) {
	// Timeout = provided limit + buffer time
	timeout := time.Duration(limits.timeLimitMs)*time.Millisecond + testCaseTimeoutBuffer

	runCtx, cancel := context.WithTimeout(ctx, timeout)
	response, err := protocol.ExecuteRunner(runCtx, runner, protocol.Request{
		Type: "run", Language: language, Stdin: input,
		TimeLimit: limits.timeLimitMs, MemoryLimitKB: limits.memoryKB,
	})

	timedOut := runCtx.Err() != nil && ctx.Err() == nil
	cancel()

	if err != nil {
		if timedOut {
			return nil, true, err
		}
		return nil, false, fmt.Errorf("execute run request: %w", err)
	}
	return response, false, nil
}

func appendFailedCase(result *api.JudgeResultDto, testCase api.SubmissionTestCaseDto, actualOutput string) {
	result.TestCases = append(result.TestCases, api.JudgeTestCaseResultDto{
		Input: testCase.Input, ExpectedOutput: testCase.ExpectedOutput,
		ActualOutput: actualOutput, Passed: false,
	})
}

func runtimeStatus(metadata protocol.Metadata) constants.Status {
	if metadata.TimedOut {
		return constants.StatusTLE
	}
	if metadata.MemoryExceeded {
		return constants.StatusMLE
	}
	return constants.StatusRuntimeError
}
