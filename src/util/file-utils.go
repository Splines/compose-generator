package util

import (
	"io/ioutil"
	"os"
	"strings"
)

// ---------------------------------------------------------------- Public functions ---------------------------------------------------------------

// FileExists checks if a file exists
func FileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

// IsDirectory checks if a file is a directory
func IsDirectory(path string) bool {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return false
	}
	return fileInfo.IsDir()
}

func AddFileToGitignore(path string) {
	filename := ".gitignore"
	var f *os.File
	content := ""
	if FileExists(filename) {
		// File does exist already
		b, err1 := ioutil.ReadFile(filename)
		if err1 != nil {
			Error("Could not read "+filename+" file", err1, true)
		}
		content = string(b) + "\n"
		if strings.Contains(content, path) {
			// This path is already included
			return
		}
		f, _ = os.OpenFile(filename, os.O_APPEND|os.O_WRONLY, 0777)
	} else {
		// File does not exist yet
		f, _ = os.OpenFile(filename, os.O_CREATE|os.O_WRONLY, 0777)
	}

	defer f.Close()
	_, err2 := f.WriteString(content + "# Docker secrets\n" + path)
	if err2 != nil {
		Error("Could not write to "+filename+" file", err2, true)
	}
}