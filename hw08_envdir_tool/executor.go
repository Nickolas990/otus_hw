package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// RunCmd runs a command + arguments (cmd) with environment variables from env.
func RunCmd(cmd []string, env Environment) int {
	// Create a new command
	execCmd := exec.Command(cmd[0], cmd[1:]...)

	// Set the environment for the command
	execCmd.Env = os.Environ()
	for key, value := range env {
		if value.NeedRemove {
			execCmd.Env = removeEnv(execCmd.Env, key)
		} else {
			execCmd.Env = append(execCmd.Env, key+"="+value.Value)
		}
	}

	// Set the standard input/output/error
	execCmd.Stdin = os.Stdin
	execCmd.Stdout = os.Stdout
	execCmd.Stderr = os.Stderr

	fmt.Printf("Running command: %s with environment: %v\n", cmd, execCmd.Env)

	// Run the command
	if err := execCmd.Run(); err != nil {
		fmt.Printf("Error running command: %v\n", err)
		if exitError, ok := err.(*exec.ExitError); ok {
			return exitError.ExitCode()
		}
		return 1
	}

	return 0
}

// Helper function to remove an environment variable from a slice of environment variables
func removeEnv(env []string, key string) []string {
	prefix := key + "="
	result := env[:0]
	for _, v := range env {
		if !strings.HasPrefix(v, prefix) {
			result = append(result, v)
		}
	}
	return result
}
