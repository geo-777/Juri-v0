package runner

//interface that determines contract.
//can be reused if i later switch to nsjails

import "juri/internals/executor"

type Runner interface {
	Run(wsArtifact *executor.WorkspaceArtifact, stdin string) (*executor.RunnerResponse, error)
}
