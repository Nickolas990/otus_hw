package main

import (
	"fmt"
	"os"
)

func main() {
	// Place your code here.
	if len(os.Args) < 3 {
		fmt.Printf("Usage: %s /path/to/env/dir command arg1 arg2 ...\n", os.Args[0])
		os.Exit(1)
	}

	envDir := os.Args[1]
	command := os.Args[2]
	args := os.Args[3:]

	env, err := ReadDir(envDir)
	if err != nil {
		fmt.Printf("Error reading env dir: %v\n", err)
		os.Exit(1)
	}

	exitCode := RunCmd(append([]string{command}, args...), env)
	os.Exit(exitCode)
}
