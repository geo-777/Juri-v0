package executor

import (
	"code-runner/internals/api/dtos"
	"fmt"
	"os"
	"path/filepath"

	gonanoid "github.com/matoous/go-nanoid/v2"
)

// creates the temporary file
func CreateWorkspace(language dtos.Language, sourceCode string, root string) (string, error) {

	id, err := gonanoid.Generate(
		"abcdefghijklmnopqrstuvwxyz0123456789",
		6,
	)

	if err != nil {
		return "", err
	}

	jobDir := filepath.Join(root, "job-"+id)

	if err := os.MkdirAll(jobDir, 0755); err != nil {
		return "", err
	}

	var filename string

	switch language {

	case "c":
		filename = "main.c"

	case "cpp":
		filename = "main.cpp"

	case "java":
		filename = "Main.java"

	case "python":
		filename = "main.py"

	default:
		return "", fmt.Errorf("unsupported language")
	}

	filePath := filepath.Join(jobDir, filename)

	err = os.WriteFile(
		filePath,
		[]byte(sourceCode),
		0644,
	)

	if err != nil {
		return "", err
	}

	return filePath, nil
}
