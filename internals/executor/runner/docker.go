package runner

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"juri/config"
	"juri/internals/constants"
	"juri/internals/executor"
	"juri/internals/executor/languages"
	"time"

	"github.com/moby/moby/api/pkg/stdcopy"
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
	var exitCode int = 0 //bydefault
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	//fetching run cmd from map
	runCmd := languages.RunCommands[language]

	//exec Creation
	execRes, err := d.docker.ExecCreate(ctx, wsArtifact.ContainerID, client.ExecCreateOptions{
		Cmd: runCmd, AttachStdin: true, AttachStderr: true, AttachStdout: true, TTY: false,
	})
	if err != nil {
		return nil, fmt.Errorf("runner (exec create): %w", err)
	}

	//attatching to exec
	attachResp, err := d.docker.ExecAttach(ctx, execRes.ID, client.ExecAttachOptions{})
	if err != nil {
		return nil, fmt.Errorf("runner (attatch) : %w", err)
	}
	defer attachResp.Close()

	//start exec
	if _, err := d.docker.ExecStart(ctx, execRes.ID, client.ExecStartOptions{}); err != nil {
		return nil, fmt.Errorf("runner (exec start) : %w", err)
	}

	//passing input via goroutine
	go func() {
		defer attachResp.CloseWrite()
		io.WriteString(attachResp.Conn, stdin)
	}()

	for {
		inspect, err := d.docker.ExecInspect(ctx, execRes.ID, client.ExecInspectOptions{})
		if err != nil {
			return nil, err
		}

		if !inspect.Running {
			exitCode = inspect.ExitCode
			fmt.Println("Exit Code:", inspect.ExitCode)
			_, err = stdcopy.StdCopy(&stdout, &stderr, attachResp.Reader)
			if err != nil {
				return nil, err
			}
			break
		}

		time.Sleep(50 * time.Millisecond)
	}

	return &executor.RunnerResponse{
		ExitCode: exitCode,
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
	}, nil
}
