package executor

import (
	"fmt"
	"os/exec"
	"path/filepath"
)

func (e *Executor) RunDocker(filePath string) error {
	//mounting requires absolute path
	absFilePath, err := filepath.Abs(filepath.Dir(filePath))
	if err != nil {
		fmt.Printf("Error resolving path: %v\n", err)
		return err
	}
	//executing docker
	runCmd := exec.Command(
		"docker",
		"run",
		"-v",
		absFilePath+":/workspace",
		"--rm",
		"--network=none",
		"--memory=128m",
		"--cpus=1",
		"juri-cpp",
	)
	output, err := runCmd.CombinedOutput()
	fmt.Print(string(output))
	return err
}
