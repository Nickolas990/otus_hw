package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	//nolint:depguard
	"al.essio.dev/pkg/shellescape"
)

// RunCmd runs a command + arguments (cmd) with environment variables from env.
func RunCmd(cmd []string, env Environment) int {
	// Create a new command
	safeCmd := []string{cmd[0]}
	for _, arg := range cmd[1:] {
		safeCmd = append(safeCmd, shellescape.Quote(arg))
	}

	// Use the safeExecCommand function
	execCmd, err := safeExecCommand(safeCmd[0], safeCmd[1:]...)
	if err != nil {
		fmt.Printf("Error validating command: %v\n", err)
		return 1
	}

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

	// Run the command
	if err := execCmd.Run(); err != nil {
		fmt.Printf("Error running command: %v\n", err)
		var exitError *exec.ExitError
		if errors.As(err, &exitError) {
			return exitError.ExitCode()
		}
		return 1
	}

	return 0
}

// Helper function to remove an environment variable from a slice of environment variables.
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

func safeExecCommand(name string, arg ...string) (*exec.Cmd, error) {
	// Validate command arguments
	if err := validateCmd(append([]string{name}, arg...)); err != nil {
		return nil, err
	}

	return exec.Command(name, arg...), nil // #nosec G204
}

func validateCmd(cmd []string) error {
	if len(cmd) == 0 {
		return errors.New("command is empty")
	}

	// Check for forbidden characters
	for _, arg := range cmd {
		if strings.Contains(arg, ";") || strings.Contains(arg, "&") || strings.Contains(arg, "|") {
			return errors.New("invalid character in command argument")
		}
	}

	// Check if the command exists in the system
	if _, err := exec.LookPath(cmd[0]); err != nil {
		return errors.New("command not found")
	}

	return nil
}
