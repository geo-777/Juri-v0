package runner

import (
	"context"
	"fmt"
	"juri/config"
	"juri/internals/constants"
	"juri/internals/executor"
	"juri/internals/executor/docker_helpers"
	"juri/internals/executor/languages"

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
	ctx context.Context,
	wsArtifact *executor.ExecutionArtifact,
	stdin string,
	language constants.Language) (*executor.RunnerResponse, error) {
	
	runResult, err := docker_helpers.Execute(ctx, d.docker, docker_helpers.Request{
		ContainerID: wsArtifact.ContainerID,
		Command:     languages.RunCommands[language],
		Stdin:       stdin,
	})
	if err != nil {
		if ctx.Err() != nil {
			_, _ = d.docker.ContainerKill(context.Background(), wsArtifact.ContainerID, client.ContainerKillOptions{Signal: "SIGKILL"})
			return nil, ctx.Err()
		}
		return nil, fmt.Errorf("run command: %w", err)
	}
	//inspecting container for oom status
	inspect, err := d.docker.ContainerInspect(ctx, wsArtifact.ContainerID, client.ContainerInspectOptions{})
	if err != nil {
		return nil, fmt.Errorf("docker container inspect: %w", err)
	}

	var status constants.Status
	if inspect.Container.State.OOMKilled {
		status = constants.StatusMLE
	} else if runResult.ExitCode != 0 {
		status = constants.StatusRuntimeError
	} else {
		status = constants.StatusSuccess
	}

	return &executor.RunnerResponse{
		ExitCode: runResult.ExitCode,
		Stdout:   runResult.Stdout,
		Stderr:   runResult.Stderr,
		Status:   status,
	}, nil
}
