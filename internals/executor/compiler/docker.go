package compiler

import (
	"juri/config"
	"juri/internals/constants"
	"juri/internals/executor"

	"github.com/moby/moby/client"
)

type DockerCompiler struct {
	cfg    *config.Config
	docker *client.Client
}

func NewDockerCompiler(cfg *config.Config, docker *client.Client) Compiler {
	return &DockerCompiler{cfg: cfg, docker: docker}
}

func (d *DockerCompiler) Compile(
	language constants.Language,
	sourceCode string,
) (*executor.WorkspaceArtifact, error) {

	return nil, nil
}
