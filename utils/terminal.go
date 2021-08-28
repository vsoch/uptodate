package utils

import (
	"log"
	"os"
	"os/exec"
	"strings"
)

// Get the environment into a map
func GetEnvironment() map[string]string {
	vars := make(map[string]string)
	for _, env := range os.Environ() {
		pair := strings.SplitN(env, "=", 2)
		vars[pair[0]] = pair[1]
	}
	return vars
}

// Run one command!
func RunCommand(cmd []string, env []string) string {

	// Define the command!
	Cmd := exec.Command(cmd[0], cmd[1:]...)
	Cmd.Env = os.Environ()

	// If we have environment strings, add them
	if len(env) > 0 {
		Cmd.Env = append(Cmd.Env, env...)
	}

	// The output will go to the buffer
	Cmd.Start()
	Cmd.Wait()

	output, err := Cmd.Output()
	if err != nil {
		log.Fatal(err)
	}
	return string(output)
}
