package executor

import (
	"bytes"
	"code-runner/internals/api/dtos"
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"time"

	gonanoid "github.com/matoous/go-nanoid/v2"
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
	runCmd := exec.CommandContext(
		ctx,
		"docker",
		"run",
		"--name", id,
		"--rm",
		"-v", absFilePath+":/workspace",
		"--network=none",
		"--memory=128m",
		"--cpus=1",
		imageNames[language],
	)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	//sets destination for output and error
	runCmd.Stdout = &stdout
	runCmd.Stderr = &stderr

	err = runCmd.Run()
	//handles infinte loops
	if ctx.Err() == context.DeadlineExceeded {
		exec.Command("docker", "rm", "-f", id).Run()
		return "", "", fmt.Errorf("time limit exceeded")
	}

	return stdout.String(), stderr.String(), err
}
