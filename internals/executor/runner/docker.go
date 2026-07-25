package runner

import (
	"juri/config"
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

func (d *DockerRunner) Run(
	wsArtifact *executor.WorkspaceArtifact,
	stdin string) (*executor.RunnerResponse, error) {

	return nil, nil
}
