package runner

//interface that determines contract.
//can be reused if i later switch to nsjails

import "juri/internals/executor"

type Runner interface {
	Run(wsArtifact *executor.ExecutionArtifact, stdin string) (*executor.RunnerResponse, error)
}
