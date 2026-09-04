package runner

import (
	"context"
	"juri/config"
	"juri/internals/constants"
	"juri/internals/executor"
	"juri/internals/executor/persistent"

	"github.com/moby/moby/client"
)

type DockerRunner struct {
	cfg    *config.Config
	docker *client.Client
}

func NewDockerRunner(cfg *config.Config, docker *client.Client) Runner {
	return &DockerRunner{cfg: cfg, docker: docker}
}

// Run executes a compiled submission inside the prepared container.
func (d *DockerRunner) Run(
	ctx context.Context,
	runner *executor.Runner,
	stdin string,
	language constants.Language) (*executor.RunnerResponse, error) {

	response, err := persistent.ExecuteRunner(ctx, runner, persistent.Request{
		Type:     "run",
		Language: language,
		Stdin:    stdin,
	})

	if err != nil {
		return nil, err
	}

	return &executor.RunnerResponse{
		Stdout: response.Output,
		Stderr: response.Error,

		RuntimeNS: response.Metadata.RuntimeNS,
		ExitCode:  response.Metadata.ExitCode,
		MemoryKB:  response.Metadata.MemoryKB,
	}, nil
}
