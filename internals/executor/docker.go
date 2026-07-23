package executor

import (
	"bytes"
	"code-runner/internals/api/dtos"
	"context"
	"fmt"
	"path/filepath"
	"time"

	gonanoid "github.com/matoous/go-nanoid/v2"
	"github.com/moby/moby/api/pkg/stdcopy"
	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/client"
)

// map for image names
var imageNames = map[dtos.Language]string{
	dtos.C:      "juri/c",
	dtos.CPP:    "juri/cpp",
	dtos.Java:   "juri/java",
	dtos.Python: "juri/python",
}

func (e *Executor) RunDocker(filePath string, language dtos.Language) (string, string, error) {

	//creating id for identify container
	id, err := gonanoid.Generate("abcdefghijklmnopqrstuvwxyz0123456789", 10)
	if err != nil {
		id = "abhcbjh"
	}
	//max time allowed for program is 3 second
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	//mounting requires absolute path
	absFilePath, err := filepath.Abs(filepath.Dir(filePath))
	if err != nil {
		fmt.Printf("Error resolving path: %v\n", err)
		return "", "", err
	}
	//executing docker
	var exitCode int64
	fmt.Println("ExitCode :", exitCode)
	resp, err := e.dockerClient.ContainerCreate(ctx, client.ContainerCreateOptions{
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
		return "", "", err
	}
	//container start
	if _, err := e.dockerClient.ContainerStart(ctx, resp.ID, client.ContainerStartOptions{}); err != nil {
		return "", "", err
	}
	//container wait
	wait := e.dockerClient.ContainerWait(ctx, resp.ID, client.ContainerWaitOptions{})
	select {
	case err := <-wait.Error:
		if err != nil {
			fmt.Println(err)
		}
	case res := <-wait.Result:
		exitCode = res.StatusCode
	}
	//inspecting container
	_, err = e.dockerClient.ContainerInspect(ctx, resp.ID, client.ContainerInspectOptions{})
	if err != nil {
		return "", "", err
	}
	//logs
	reader, err := e.dockerClient.ContainerLogs(ctx, resp.ID,
		client.ContainerLogsOptions{ShowStdout: true, ShowStderr: true})
	if err != nil {
		return "", "", err
	}
	defer reader.Close()

	//fetching input/output from reader
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	_, err = stdcopy.StdCopy(&stdout, &stderr, reader)
	if err != nil {
		return "", "", err
	}
	//always remove container
	defer e.dockerClient.ContainerRemove(
		context.Background(),
		resp.ID,
		client.ContainerRemoveOptions{
			Force: true,
		},
	)

	return stdout.String(), stderr.String(), err

}
