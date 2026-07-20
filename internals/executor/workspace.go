package executor

import (
	"code-runner/internals/api/dtos"
	"fmt"
	"os"
	"path/filepath"

	gonanoid "github.com/matoous/go-nanoid/v2"
)

// responsible for creating the temporary file-folder system
func CreateWorkspace(language dtos.Language, sourceCode string, root string) (string, error) {
	//generate job-id
	id, err := gonanoid.Generate(
		"abcdefghijklmnopqrstuvwxyz0123456789",
		6,
	)
	if err != nil {
		return "", err
	}
	//prepare path
	jobDir := filepath.Join(root, "job-"+id)
	//creates necessary stuff
	if err := os.MkdirAll(jobDir, 0777); err != nil {
		return "", err
	}

	var filename string

	switch language {

	case dtos.C:
		filename = "main.c"

	case dtos.CPP:
		filename = "main.cpp"

	case dtos.Java:
		filename = "Main.java"

	case dtos.Python:
		filename = "main.py"

	default:
		return "", fmt.Errorf("unsupported language")
	}
	//write the code into file
	filePath := filepath.Join(jobDir, filename)
	err = os.WriteFile(filePath, []byte(sourceCode), 0644)

	if err != nil {
		return "", err
	}
	//returns file path if successful creation
	return filePath, nil
}
