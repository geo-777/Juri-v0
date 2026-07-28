package compiler

import (
	"bytes"
	"context"
	"fmt"
	"juri/config"
	"juri/internals/constants"
	"juri/internals/executor"
	"juri/internals/executor/languages"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/moby/moby/api/pkg/stdcopy"
	"github.com/moby/moby/api/types/container"
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
	ctx context.Context,
	language constants.Language,
	sourceCode string,
) (*executor.CompileResult, error) {
	//creating workspace
	sourceFilePath, err := executor.CreateWorkspace(language, sourceCode, d.cfg.WorkspaceRoot)
	if err != nil {
		return nil, err
	}
	sourceDirPath, err := filepath.Abs(filepath.Dir(sourceFilePath))
	if err != nil {
		return nil, err
	}
	pidLimit := int64(128)

	//creating container
	resp, err := d.docker.ContainerCreate(ctx, client.ContainerCreateOptions{
		Config: &container.Config{
			Image:       languages.ImageNames[language],
			WorkingDir:  "/workspace",
			AttachStdin: true,
			OpenStdin:   true,
			Tty:         false,
		},
		HostConfig: &container.HostConfig{
			NetworkMode: "none",
			Resources: container.Resources{
				Memory:    128 * 1024 * 1024, //128MB
				NanoCPUs:  1_000_000_000,     //1 CPU
				PidsLimit: &pidLimit,         // takes only pointer to int64
			},

			Binds: []string{sourceDirPath + ":/workspace"},
		},
	})
	if err != nil {
		_ = os.RemoveAll(sourceDirPath) //cleaning ws
		return nil, fmt.Errorf("create container: %w", err)
	}
	//cleanup function
	cleanup := func() error {
		cleanupCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		//removing container
		_, err := d.docker.ContainerRemove(cleanupCtx, resp.ID, client.ContainerRemoveOptions{Force: true})
		if err != nil {
			return err
		}
		//removing workspace
		if err := os.RemoveAll(sourceDirPath); err != nil {
			log.Printf("Failed to remove workspace : %v", err)
			return err
		}
		return nil
	}

	success := false
	defer func() {
		if !success {
			if err := cleanup(); err != nil {
				log.Printf("cleanup failed: %v", err)
			}
		}
	}()
	//starting container
	if _, err := d.docker.ContainerStart(ctx, resp.ID, client.ContainerStartOptions{}); err != nil {
		return nil, err
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	var exitCode int = 0
	//compilation
	cmd := languages.CompileCommands[language] //fetching cmd

	if cmd != nil { // Interpreted languages (e.g. Python) have no compilation step.
		execRes, err := d.docker.ExecCreate(ctx, resp.ID, client.ExecCreateOptions{
			Cmd: cmd, AttachStdout: true, AttachStderr: true, TTY: false,
		})
		if err != nil {
			return nil, err
		}
		//attatching to exec
		attachResp, err := d.docker.ExecAttach(ctx, execRes.ID, client.ExecAttachOptions{})
		if err != nil {
			return nil, err
		}

		defer attachResp.Close()
		//runnning exec
		if _, err := d.docker.ExecStart(ctx, execRes.ID, client.ExecStartOptions{}); err != nil {
			return nil, err
		}

		//polling to get compilation status
		for {
			inspect, err := d.docker.ExecInspect(ctx, execRes.ID, client.ExecInspectOptions{})
			if err != nil {
				return nil, err
			}
			//if not running
			if !inspect.Running {
				exitCode = inspect.ExitCode
				_, err = stdcopy.StdCopy(&stdout, &stderr, attachResp.Reader)
				if err != nil {
					return nil, err
				}
				break
			}
			time.Sleep(50 * time.Millisecond)
		}

	}
	success = true
	//returning artifact with details as well as cleanup func
	return &executor.CompileResult{
		Artifact: &executor.ExecutionArtifact{
			SourceFilePath: sourceFilePath,
			SourceDirPath:  sourceDirPath,
			ContainerID:    resp.ID,
			Cleanup:        cleanup,
		},
		ExitCode: exitCode,
		Stderr:   stderr.String(),
		Stdout:   stdout.String(),
	}, nil
}
