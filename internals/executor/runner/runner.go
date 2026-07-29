package runner

//interface that determines contract.
//can be reused if i later switch to nsjails

import (
	"context"
	"juri/internals/constants"
	"juri/internals/executor"
)

type Runner interface {
	Run(ctx context.Context, wsArtifact *executor.ExecutionArtifact, stdin string, languages constants.Language) (*executor.RunnerResponse, error)
}
