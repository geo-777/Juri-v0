package compiler

import (
	"context"
	"fmt"
	"juri/config"
	"juri/internals/constants"
	"juri/internals/executor"

	"juri/internals/executor/languages"
	"juri/internals/executor/persistent"
	"log"
	"os"
	"path/filepath"
	"time"

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

// Compile creates a temporary workspace, starts a container, iniates persistent runner and finally compiles program
func (d *DockerCompiler) Compile(
	ctx context.Context,
	language constants.Language,
	sourceCode string,
) (*executor.CompileResult, error) {
	// Create a temporary workspace for the submitted source file.
	sourceFilePath, err := executor.CreateWorkspace(language, sourceCode, d.cfg.WorkspaceRoot)
	if err != nil {
		return nil, fmt.Errorf("create workspace: %w", err)
	}
	sourceDirPath, err := filepath.Abs(filepath.Dir(sourceFilePath))
	if err != nil {
		_ = os.RemoveAll(filepath.Dir(sourceFilePath))
		return nil, fmt.Errorf("resolve workspace path: %w", err)
	}
	pidLimit := int64(128)
	// Create a container with resource limits and the mounted workspace.
	resp, err := d.docker.ContainerCreate(ctx, client.ContainerCreateOptions{
		Config: &container.Config{
			Image:        languages.ImageNames[language],
			WorkingDir:   "/workspace",
			AttachStdin:  true,
			AttachStdout: true,
			AttachStderr: true,
			OpenStdin:    true,
			Tty:          true,
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
		return nil, fmt.Errorf("docker container create: %w", err)
	}
	// Define a cleanup routine for the container and workspace on failure or teardown.
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
			log.Printf("cleanup workspace: %v", err)
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
	// Start the container so the compile command can run.
	if _, err := d.docker.ContainerStart(ctx, resp.ID, client.ContainerStartOptions{}); err != nil {
		return nil, fmt.Errorf("docker container start: %w", err)
	}
	// Run the language-specific compilation command inside the container.
	// compResp, err := d.executeCompileCommand(ctx, language, resp.ID)
	// if err != nil {
	// 	return nil, fmt.Errorf("compile command: %w", err)
	// }

	runner, err := persistent.CreateRunner(ctx, d.docker, resp.ID)
	if err != nil {
		return nil, fmt.Errorf("Failed to start docker attach : %w", err)
	}
	runner.SourceFilePath = sourceDirPath
	runner.Cleanup = cleanup

	response, err := persistent.ExecuteRunner(ctx, runner, persistent.Request{
		Type:     "compile",
		Language: language,
	})

	fmt.Println(response)

	success = true
	// Return the execution artifact and compile output.
	return &executor.CompileResult{
		Runner:   runner,
		Stdout:   response.Output,
		ExitCode: response.Metadata.ExitCode,
	}, nil
}
