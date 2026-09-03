package runner

import (
	"context"
	"juri/config"
	"juri/internals/constants"
	"juri/internals/executor"

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

	return nil, nil
}
