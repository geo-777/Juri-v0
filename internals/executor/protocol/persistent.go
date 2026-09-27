package protocol

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"juri/internals/constants"
	"juri/internals/executor"

	"github.com/moby/moby/api/pkg/stdcopy"
	"github.com/moby/moby/client"
)

//runner shouldnt create container as pool would be responsible for it later

//writing the persistent runner here.
//responsible for creating persistent runner and executing commands

type Request struct {
	Type      string             `json:"type"`
	Language  constants.Language `json:"language"`
	Stdin     string             `json:"stdin"`
	TimeLimit int                `json:"time_limit"`
}

type Response struct {
	Success bool   `json:"success"`
	Output  string `json:"output,omitempty"`
	Error   string `json:"error,omitempty"`

	Metadata Metadata `json:"metadata,omitempty"`
}

type Metadata struct {
	RuntimeNS int64 `json:"runtime_ns"`
	MemoryKB  int64 `json:"memory_kb"`
	ExitCode  int   `json:"exit_code"`
	TimedOut  bool  `json:"timed_out"`
}

func CreateRunner(
	ctx context.Context,
	docker *client.Client,
	containerID string,
) (*executor.Runner, error) {

	attachResp, err := docker.ContainerAttach(
		ctx,
		containerID,
		client.ContainerAttachOptions{
			Stream: true,
			Stdin:  true,
			Stdout: true,
			Stderr: true,
		},
	)
	if err != nil {
		return nil, err
	}

	// Without a TTY Docker multiplexes stdout and stderr in its attach stream.
	// Keep only stdout for the JSON response decoder while draining stderr.
	stdoutReader, stdoutWriter := io.Pipe()
	go func() {
		_, copyErr := stdcopy.StdCopy(stdoutWriter, io.Discard, attachResp.Reader)
		_ = stdoutWriter.CloseWithError(copyErr)
	}()

	return &executor.Runner{
		Conn:        attachResp.Conn,
		Reader:      json.NewDecoder(stdoutReader),
		Writer:      json.NewEncoder(attachResp.Conn),
		ContainerID: containerID,
	}, nil
}

func ExecuteRunner(
	ctx context.Context,
	runner *executor.Runner,
	req Request,
) (*Response, error) {

	json.MarshalIndent(req, "", "  ")

	if err := runner.Writer.Encode(req); err != nil {
		return nil, fmt.Errorf("failed to write runner request: %w", err)
	}

	var res Response

	if err := runner.Reader.Decode(&res); err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}
	return &res, nil
}
