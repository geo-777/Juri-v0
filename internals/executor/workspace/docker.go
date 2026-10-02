package workspace

import (
	"context"
	"fmt"
	"juri/config"
	"juri/internals/constants"
	"juri/internals/executor"
	"juri/internals/executor/protocol"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/client"
)

type DockerRunnerFactory struct {
	cfg    *config.Config
	docker *client.Client
}

func NewDockerRunnerFactory(
	cfg *config.Config,
	docker *client.Client,
) *DockerRunnerFactory {
	return &DockerRunnerFactory{
		cfg:    cfg,
		docker: docker,
	}
}

// DONT SPLIT THIS FUNCITON NOW.
// This function size would decrease greatly once container pools are implemented.

func (d *DockerRunnerFactory) CreateRunner(ctx context.Context, language constants.Language, sourceCode string, memoryLimitKB int) (*executor.Runner, error) {
	// Create a temporary workspace for the submitted source file.
	sourceFilePath, err := CreateFiles(language, sourceCode, d.cfg.WorkspaceRoot)
	if err != nil {
		return nil, fmt.Errorf("create workspace: %w", err)
	}

	sourceDirPath, err := filepath.Abs(filepath.Dir(sourceFilePath))
	if err != nil {
		_ = os.RemoveAll(filepath.Dir(sourceFilePath))
		return nil, fmt.Errorf("resolve workspace path: %w", err)
	}

	pidLimit := int64(128)
	if memoryLimitKB <= 0 {
		memoryLimitKB = constants.DefaultMemoryLimitKB
	}
	// Create a container with resource limits and the mounted workspace.
	resp, err := d.docker.ContainerCreate(ctx, client.ContainerCreateOptions{
		Config: &container.Config{
			Image:        constants.ImageNames[language],
			WorkingDir:   "/workspace",
			AttachStdin:  true,
			AttachStdout: true,
			AttachStderr: true,
			OpenStdin:    true,
			// The persistent runner speaks a JSON protocol over stdin/stdout.
			// A TTY echoes stdin back into stdout and corrupts that protocol.
			Tty: false,
		},
		HostConfig: &container.HostConfig{
			NetworkMode: "none",
			Resources: container.Resources{
				Memory:    int64(memoryLimitKB) * 1024,
				NanoCPUs:  1_000_000_000, //1 CPU
				PidsLimit: &pidLimit,     // takes only pointer to int64
			},

			Binds:          []string{sourceDirPath + ":/workspace"},
			ReadonlyRootfs: true,
			SecurityOpt: []string{
				"no-new-privileges:true",
			},
			CapDrop: []string{"ALL"},
			Tmpfs: map[string]string{
				"/tmp": "rw,noexec,nosuid,size=64m",
			},
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
	runner, err := protocol.CreateRunner(ctx, d.docker, resp.ID)
	if err != nil {
		return nil, fmt.Errorf("Failed to start docker attach : %w", err)
	}
	runner.SourceFilePath = sourceDirPath
	runner.Cleanup = cleanup

	// Attach before starting so the runner cannot exit before its protocol
	// stream is connected.
	if _, err := d.docker.ContainerStart(ctx, resp.ID, client.ContainerStartOptions{}); err != nil {
		return nil, fmt.Errorf("docker container start: %w", err)
	}

	success = true
	// Return the execution artifact and compile output.
	return runner, nil

}
