package compiler

import (
	"context"
	"juri/config"
	"juri/internals/constants"
	"juri/internals/executor"
	"juri/internals/utils"
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

var imageNames = map[constants.Language]string{
	constants.C:      "juri/c",
	constants.CPP:    "juri/cpp",
	constants.Java:   "juri/java",
	constants.Python: "juri/python",
}

var compileCommands = map[constants.Language][]string{
	constants.C: {
		"gcc",
		"main.c",
		"-O2",
		"-o",
		"program",
	},
	constants.CPP: {
		"g++",
		"main.cpp",
		"-O2",
		"-o",
		"program",
	},
	constants.Java: {
		"javac",
		"Main.java",
	},
	constants.Python: nil, // No compilation required
}

func (d *DockerCompiler) Compile(
	ctx context.Context,
	language constants.Language,
	sourceCode string,
) (*executor.WorkspaceArtifact, error) {
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
			Image:       imageNames[language],
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
		return nil, err
	}
	//starting container
	if _, err := d.docker.ContainerStart(ctx, resp.ID, client.ContainerStartOptions{}); err != nil {
		return nil, err
	}

	//compilation
	cmd := compileCommands[language] //fetching cmd
	if cmd != nil {
		execRes, err := d.docker.ExecCreate(ctx, resp.ID, client.ExecCreateOptions{Cmd: cmd, AttachStdout: true, AttachStderr: true})
		if err != nil {
			return nil, err
		}

		d.docker.ExecStart(ctx, execRes.ID, client.ExecStartOptions{})

	}

	//returning artifact with details as well as cleanup func
	return &executor.WorkspaceArtifact{
		SourceFilePath: sourceFilePath,
		SourceDirPath:  sourceDirPath,
		ContainerName:  containerName,
		ContainerID:    resp.ID,
		Cleanup: func() error {
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
		},
	}, nil
}
