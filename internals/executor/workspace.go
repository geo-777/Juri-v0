package executor

import (
	"fmt"
	"juri/internals/constants"
	"juri/internals/utils"
	"os"
	"path/filepath"
)

// responsible for creating the temporary file-folder system
func CreateWorkspace(language constants.Language, sourceCode string, root string) (string, error) {
	//generate job-id
	id := utils.GenerateNanoId(6)

	//prepare path
	jobDir := filepath.Join(root, "job-"+id)
	//creates necessary stuff
	if err := os.MkdirAll(jobDir, 0777); err != nil {
		return "", err
	}

	var filename string

	switch language {

	case constants.C:
		filename = "main.c"

	case constants.CPP:
		filename = "main.cpp"

	case constants.Java:
		filename = "Main.java"

	case constants.Python:
		filename = "main.py"

	default:
		return "", fmt.Errorf("unsupported language")
	}
	//write the code into file
	filePath := filepath.Join(jobDir, filename)
	err := os.WriteFile(filePath, []byte(sourceCode), 0644)

	if err != nil {
		return "", err
	}
	//returns file path if successful creation
	return filePath, nil
}

func RemoveWorkspace(filePath string) error {
	targetDir := filepath.Dir(filePath)

	err := os.RemoveAll(targetDir)

	return err
}
