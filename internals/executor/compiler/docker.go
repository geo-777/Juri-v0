package compiler

import (
	"bytes"
	"context"
	"juri/config"
	"juri/internals/constants"
	"juri/internals/executor"
	"juri/internals/executor/languages"
	"juri/internals/utils"
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
	//generating container name
	containerName := utils.GenerateNanoId(10)

	//creating workspace
	sourceFilePath, err := executor.CreateWorkspace(language, sourceCode, d.cfg.WorkspaceRoot)
	if err != nil {
		return nil, err
	}
	sourceDirPath, err := filepath.Abs(filepath.Dir(sourceFilePath))
	if err != nil {
		return nil, err
	}

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
				Memory:   128 * 1024 * 1024, //128MB
				NanoCPUs: 1_000_000_000,     //1 CPU
			},
			Binds: []string{sourceDirPath + ":/workspace"},
		},
		Name: containerName,
	})
	if err != nil {
		_ = os.RemoveAll(sourceDirPath) //cleaning ws
		return nil, err
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
			cleanup()
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
	if cmd != nil {
		execRes, err := d.docker.ExecCreate(ctx, resp.ID, client.ExecCreateOptions{
			Cmd: cmd, AttachStdout: true, AttachStderr: true, TTY: false,
		})
		if err != nil {
			return nil, err
		}
		//attatching to exec
		attachResp, _ := d.docker.ExecAttach(ctx, execRes.ID, client.ExecAttachOptions{})
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
			time.Sleep(10 * time.Millisecond)
		}

	}
	success = true
	//returning artifact with details as well as cleanup func
	return &executor.CompileResult{
		Artifact: &executor.WorkspaceArtifact{
			SourceFilePath: sourceFilePath,
			SourceDirPath:  sourceDirPath,
			ContainerName:  containerName,
			ContainerID:    resp.ID,
			Cleanup:        cleanup,
		},
		ExitCode: exitCode,
		Stderr:   stderr.String(),
	}, nil
}
