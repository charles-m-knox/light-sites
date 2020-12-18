package helpers

import (
	"fmt"
	"lightsites/constants"
	"os"
	"path/filepath"
	"strings"
)

// LoadFileDirectory locates all markdown files in a directory and returns
// a slice of os.FileInfo items
// Unused currently
func LoadFileDirectory(fileInfo []os.FileInfo) (result []string) {
	for _, file := range fileInfo {
		fileName := file.Name()
		if strings.Contains(fileName, ".md") {
			result = append(result, strings.TrimRight(fileName, ".md"))
		}
	}

	return result
}

// ReadDirectory returns all files and directories from a top-level
// directory
// Unused currently
func ReadDirectory(directory string) ([]os.FileInfo, error) {
	var files []os.FileInfo
	f, err := os.Open(directory)
	if err != nil {
		return files, err
	}

	fInfo, err := f.Readdir(-1)
	f.Close()

	if err != nil {
		return files, err
	}

	files = append(files, fInfo...)

	return files, nil
}

type DirectoryListing struct {
	Files []string
	Path  string
}

func (dirList *DirectoryListing) WalkStep(path string, f os.FileInfo, err error) error {
	if err != nil {
		return fmt.Errorf("err walking path %v: %v", path, err.Error())
	}
	pathTrimmed := strings.TrimPrefix(path, dirList.Path)
	if pathTrimmed != "" {
		if strings.Index(pathTrimmed, ".md") == len(pathTrimmed)-len(constants.MarkdownFileSuffix) {
			newFile := strings.TrimSuffix(pathTrimmed, constants.MarkdownFileSuffix)
			dirList.Files = append(dirList.Files, strings.TrimPrefix(newFile, "/"))
		}
	}
	return nil
}

func (dirList *DirectoryListing) WalkDirectory() error {
	err := filepath.Walk(dirList.Path, dirList.WalkStep)
	if err != nil {
		return fmt.Errorf("failed to walk dir %v: %v", dirList.Path, err.Error())
	}
	return nil
}
