package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Environment map[string]EnvValue

// EnvValue helps to distinguish between empty files and files with the first empty line.
type EnvValue struct {
	Value      string
	NeedRemove bool
}

// ReadDir reads a specified directory and returns map of env variables.
// Variables represented as files where filename is name of variable, file first line is a value.
func ReadDir(dir string) (Environment, error) {
	env := make(Environment)
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		name := file.Name()
		if strings.Contains(name, "=") {
			return nil, fmt.Errorf("invalid environment variable name: %s", name)
		}

		f, err := os.Open(filepath.Join(dir, name))
		if err != nil {
			return nil, err
		}

		scanner := bufio.NewScanner(f)
		var content string
		if scanner.Scan() {
			content = scanner.Text()
		}
		f.Close()

		fmt.Printf("Original content of %s: %q\n", name, content)

		// Replace all null bytes with newlines
		content = strings.ReplaceAll(content, "\x00", "\n")
		fmt.Printf("Content after replacing null bytes with newlines: %q\n", content)

		// Trim trailing spaces, tabs, but not newlines
		value := strings.TrimRightFunc(content, func(r rune) bool {
			return r == ' ' || r == '\t'
		})
		fmt.Printf("Trimmed value: %q\n", value)

		envValue := EnvValue{
			Value:      value,
			NeedRemove: len(value) == 0,
		}

		env[name] = envValue
	}

	return env, nil
}
