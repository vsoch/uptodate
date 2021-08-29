package utils

import (
	"io/ioutil"
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

// ReadFile reads a file from the system
func ReadFile(path string) string {
	filey, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err = filey.Close(); err != nil {
			log.Fatal(err)
		}
	}()
	bytes, err := ioutil.ReadAll(filey)
	return string(bytes)
}

// WriteFile and some content to the filesystem
func WriteFile(path string, content string) error {
	filey, err := os.Create(path)
	if err != nil {
		return err
	}
	defer filey.Close()

	_, err = filey.WriteString(content)
	if err != nil {
		return err
	}
	err = filey.Sync()
	return err
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
