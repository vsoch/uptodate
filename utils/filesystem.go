package utils

import (
	"log"
	"os"
	"path/filepath"
	"strings"
)

// Wrapper to get the PWD
func GetPwd() string {
	path, err := os.Getwd()
	if err != nil {
		log.Fatal("Cannot derive the present working directory.")
	}
	return path
}

// Find files in a directory based on a pattern
func RecursiveFind(root string, pattern string, allowPrefix bool) ([]string, error) {

	// Create a list of results to return
	results := []string{}

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		match, _ := filepath.Match(pattern, filepath.Base(path))

		// If we don't have a match and the parser allows a prefix
		if !match && allowPrefix {
			fileBasename := filepath.Base(path)
			match = strings.HasPrefix(fileBasename, pattern)
		}

		if match {
			results = append(results, path)
		}
		return nil

	})

	if err != nil {
		log.Fatal("Error running RecursiveFind to find files %s", err)
	}
	return results, err
}
