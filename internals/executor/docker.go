package executor

import (
	"bytes"
	"code-runner/internals/api/dtos"
	"code-runner/pkg/constants"
	"context"
	"fmt"
	"path/filepath"
	"time"

	gonanoid "github.com/matoous/go-nanoid/v2"
	"github.com/moby/moby/api/pkg/stdcopy"
	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/client"
)

type DockerRunResponse struct {
	Stdout string
	Stderr string
	Status constants.Status
}

// map for image names
var imageNames = map[dtos.Language]string{
	dtos.C:      "juri/c",
	dtos.CPP:    "juri/cpp",
	dtos.Java:   "juri/java",
	dtos.Python: "juri/python",
}

func (e *Executor) RunDocker(filePath string, language dtos.Language) (*DockerRunResponse, error) {

	//creating id for identify container
	id, err := gonanoid.Generate("abcdefghijklmnopqrstuvwxyz0123456789", 10)
	if err != nil {
		id = "abhcbjh"
	}
	//max time allowed for program is 3 second
	dockerCtx := context.Background()

	runCtx, cancel := context.WithTimeout(dockerCtx, 2*time.Second)
	defer cancel()
	//mounting requires absolute path
	absFilePath, err := filepath.Abs(filepath.Dir(filePath))
	if err != nil {
		fmt.Printf("Error resolving path: %v\n", err)
		return &DockerRunResponse{}, err
	}
	//executing docker
	var exitCode int64
	fmt.Println("ExitCode :", exitCode)
	//creates docker container
	resp, err := e.dockerClient.ContainerCreate(dockerCtx, client.ContainerCreateOptions{
		Config: &container.Config{
			Image:      imageNames[language],
			WorkingDir: "/workspace",
		},
		HostConfig: &container.HostConfig{
			NetworkMode: "none",
			Resources: container.Resources{
				Memory:   128 * 1024 * 1024, //128MB
				NanoCPUs: 1_000_000_000,     //1 CPU
			},
			Binds: []string{absFilePath + ":/workspace"},
		},
		Name: id,
	})

	if err != nil {
		return &DockerRunResponse{}, err
	}
	//container start
	if _, err := e.dockerClient.ContainerStart(dockerCtx, resp.ID, client.ContainerStartOptions{}); err != nil {
		return &DockerRunResponse{}, err
	}
	//container wait
	wait := e.dockerClient.ContainerWait(runCtx, resp.ID, client.ContainerWaitOptions{})
	select {
	case err := <-wait.Error:
		//kill the container
		if err != nil {
			if runCtx.Err() == context.DeadlineExceeded {
				_, _ = e.dockerClient.ContainerKill(dockerCtx, resp.ID, client.ContainerKillOptions{Signal: "SIGKILL"})
			} else {
				return nil, err
			}
		}
	case res := <-wait.Result:
		exitCode = res.StatusCode
	}
	//inspecting container
	inspect, err := e.dockerClient.ContainerInspect(dockerCtx, resp.ID, client.ContainerInspectOptions{})
	if err != nil {
		return &DockerRunResponse{}, err
	}
	//logs
	reader, err := e.dockerClient.ContainerLogs(dockerCtx, resp.ID,
		client.ContainerLogsOptions{ShowStdout: true, ShowStderr: true})
	if err != nil {
		return &DockerRunResponse{}, err
	}
	defer reader.Close()

	//fetching input/output from reader
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	_, err = stdcopy.StdCopy(&stdout, &stderr, reader)
	if err != nil {
		return &DockerRunResponse{}, err
	}
	//always remove container
	defer e.dockerClient.ContainerRemove(
		context.Background(),
		resp.ID,
		client.ContainerRemoveOptions{
			Force: true,
		},
	)
	var status constants.Status
	//determines appropriate status
	if runCtx.Err() == context.DeadlineExceeded {
		status = constants.StatusTLE
	} else if inspect.Container.State.OOMKilled {
		status = constants.StatusMLE
	} else if inspect.Container.State.ExitCode != 0 {
		status = constants.StatusRuntimeError
	} else {
		status = constants.StatusSuccess
	}

	return &DockerRunResponse{
		Stdout: stdout.String(),
		Stderr: stderr.String(),
		Status: status,
	}, err

}
