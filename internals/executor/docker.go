package executor

import (
	"code-runner/internals/api/dtos"
	"fmt"
	"os/exec"
	"path/filepath"
)

// map for image names
var imageNames = map[dtos.Language]string{
	dtos.C:      "juri/c",
	dtos.CPP:    "juri/cpp",
	dtos.Java:   "juri/java",
	dtos.Python: "juri/python",
}

func (e *Executor) RunDocker(filePath string, language dtos.Language) error {
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
		imageNames[language],
	)
	fmt.Println(imageNames[language])
	output, err := runCmd.CombinedOutput()
	fmt.Print(string(output))
	return err
}
