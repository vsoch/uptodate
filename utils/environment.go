package utils

import (
	"fmt"
	"log"
	"os"
)

// Are we running in a GitHub Action?
func IsGitHubAction() bool {
	_, inGitHub := os.LookupEnv("GITHUB_EVENT_NAME")

	return inGitHub

}

// WriteGitHubOutput writes to GITHUB_OUTPUT if the envar is set
func WriteGitHubOutput(key string, value string) {
	file, _ := os.LookupEnv("GITHUB_OUTPUT")

	// Cut out early if the envar isnt' set
	if file == "" {
		return
	}
	// If we have a filename (it might not exist) append to it
	mode := os.O_APPEND | os.O_CREATE | os.O_WRONLY
	f, err := os.OpenFile(file, mode, 0644)

	// We shouldn't exit, but alert the user
	if err != nil {
		log.Println(err)
		return
	}
	defer f.Close()

	// Write the key=value to file!
	if _, err := f.WriteString(fmt.Sprintf("%s=%s\n", key, value)); err != nil {
		log.Println(err)
	}
}
