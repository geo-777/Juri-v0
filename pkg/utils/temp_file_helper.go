package utils

import (
	"code-runner/internals/api/dtos"
	"fmt"
	"os"
	"path/filepath"

	gonanoid "github.com/matoous/go-nanoid/v2"
)

//helps to create temp jobs files

func TempFileCreator(lang dtos.Language, sourceCode string, location string) (string, error) {
	//creates job id
	id, err := gonanoid.Generate("abcdefghijklmnopqrstuvwxyz0123456789", 4)
	if err != nil {
		return "", err
	}
	location = location + "/job-" + id
	//creates stuff required
	err = os.MkdirAll(location, 0755)
	if err != nil {
		return "", err
	}
	//decides the file name based on lang
	var fileName string

	switch lang {
	case dtos.C:
		fileName = "main.c"
	case dtos.CPP:
		fileName = "main.cpp"
	case dtos.Java:
		fileName = "main.java"
	case dtos.Python:
		fileName = "main.py"
	default:
		return "", fmt.Errorf("unsupported language")
	}
	//preparing file location
	filePath := filepath.Join(location, fileName)
	// writing code to file
	err = os.WriteFile(filePath, []byte(sourceCode), 0644)
	if err != nil {
		return "", err
	}

	return filePath, nil
}
