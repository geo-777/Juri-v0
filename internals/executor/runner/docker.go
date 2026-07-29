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
	"log"
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
	var exitCode int
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	var status constants.Status

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
		if _, err := io.WriteString(attachResp.Conn, stdin); err != nil {
			log.Printf("Attatching input error : %v", err)
		}
	}()
	copyDone := make(chan error, 1)

	go func() {
		_, err := stdcopy.StdCopy(&stdout, &stderr, attachResp.Reader)
		copyDone <- err
	}()

	for {
		select {
		case <-ctx.Done():
			_, _ = d.docker.ContainerKill(context.Background(), wsArtifact.ContainerID, client.ContainerKillOptions{Signal: "SIGKILL"})
			<-copyDone // let StdCopy finish
			return nil, ctx.Err()

		default:
		}

		inspect, err := d.docker.ExecInspect(ctx, execRes.ID, client.ExecInspectOptions{})
		if err != nil {
			return nil, err
		}
		if !inspect.Running {
			exitCode = inspect.ExitCode
			break
		}

		time.Sleep(50 * time.Millisecond)
	}
	//waiting for shit
	if err := <-copyDone; err != nil {
		return nil, err
	}
	//inspecting container for oom status
	inspect, err := d.docker.ContainerInspect(ctx, wsArtifact.ContainerID, client.ContainerInspectOptions{})
	if err != nil {
		return nil, err
	}

	if inspect.Container.State.OOMKilled {
		status = constants.StatusMLE
	} else if exitCode != 0 {
		status = constants.StatusRuntimeError
	} else {
		status = constants.StatusSuccess
	}

	return &executor.RunnerResponse{
		ExitCode: exitCode,
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
		Status:   status,
	}, nil
}
