package workspace

import (
	"fmt"
	"juri/internals/constants"
	"juri/internals/utils"
	"os"
	"path/filepath"
)

// CreateFiles prepares a temporary job directory and writes the submission source file into it.
func CreateFiles(language constants.Language, sourceCode string, root string) (string, error) {
	// Generate a short unique ID for the job folder.
	id := utils.GenerateNanoId(6)

	// Build the workspace path for this submission.
	jobDir := filepath.Join(root, "job-"+id)
	// Create the directory structure needed for the job.
	if err := os.MkdirAll(jobDir, 0777); err != nil {
		return "", fmt.Errorf("create job directory: %w", err)
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
		_ = os.RemoveAll(jobDir)
		return "", fmt.Errorf("unsupported language %q", language)
	}
	// Write the source code into the job directory.
	filePath := filepath.Join(jobDir, filename)
	err := os.WriteFile(filePath, []byte(sourceCode), 0644)

	if err != nil {
		_ = os.RemoveAll(jobDir)
		return "", fmt.Errorf("write source file: %w", err)
	}
	// Return the source file path when the workspace is ready.
	return filePath, nil
}
