package docker_helpers

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"time"

	"github.com/moby/moby/api/pkg/stdcopy"
	"github.com/moby/moby/client"
)

const pollInterval = 50 * time.Millisecond

type Request struct {
	ContainerID string
	Command     []string
	Stdin       string
}

type Result struct {
	Stdout   string
	Stderr   string
	ExitCode int
}

// Execute runs a command in an existing container and collects its output.
func Execute(ctx context.Context, docker *client.Client, request Request) (*Result, error) {
	execResult, err := docker.ExecCreate(ctx, request.ContainerID, client.ExecCreateOptions{
		Cmd: request.Command, AttachStdin: request.Stdin != "", AttachStdout: true, AttachStderr: true,
	})
	if err != nil {
		return nil, fmt.Errorf("docker exec create: %w", err)
	}

	attachResponse, err := docker.ExecAttach(ctx, execResult.ID, client.ExecAttachOptions{})
	if err != nil {
		return nil, fmt.Errorf("docker exec attach: %w", err)
	}
	defer attachResponse.Close()

	if _, err := docker.ExecStart(ctx, execResult.ID, client.ExecStartOptions{}); err != nil {
		return nil, fmt.Errorf("docker exec start: %w", err)
	}

	if request.Stdin != "" {
		go func() {
			defer attachResponse.CloseWrite()
			_, _ = io.WriteString(attachResponse.Conn, request.Stdin)
		}()
	}

	var stdout, stderr bytes.Buffer
	copyDone := make(chan error, 1)
	go func() {
		_, err := stdcopy.StdCopy(&stdout, &stderr, attachResponse.Reader)
		copyDone <- err
	}()

	exitCode, err := waitForCompletion(ctx, docker, execResult.ID)
	if err != nil {
		return nil, err
	}
	if err := <-copyDone; err != nil {
		return nil, fmt.Errorf("docker exec read output: %w", err)
	}

	return &Result{ExitCode: exitCode, Stdout: stdout.String(), Stderr: stderr.String()}, nil
}

func waitForCompletion(ctx context.Context, docker *client.Client, execID string) (int, error) {
	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	for {
		inspect, err := docker.ExecInspect(ctx, execID, client.ExecInspectOptions{})
		if err != nil {
			return 0, fmt.Errorf("docker exec inspect: %w", err)
		}
		if !inspect.Running {
			return inspect.ExitCode, nil
		}
		select {
		case <-ctx.Done():
			return 0, ctx.Err()
		case <-ticker.C:
		}
	}
}
