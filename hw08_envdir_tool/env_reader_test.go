package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestReadDir(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "envdirtest")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create test files
	files := map[string]string{
		"VALID":   "value",
		"EMPTY":   "",
		"TRIMMED": "value \t",
		"NULL":    "value\x00",
	}

	for name, content := range files {
		err := os.WriteFile(filepath.Join(tempDir, name), []byte(content), 0644)
		require.NoError(t, err)
	}

	env, err := ReadDir(tempDir)
	require.NoError(t, err)

	expected := Environment{
		"VALID":   {Value: "value", NeedRemove: false},
		"EMPTY":   {Value: "", NeedRemove: true},
		"TRIMMED": {Value: "value", NeedRemove: false},
		"NULL":    {Value: "value\n", NeedRemove: false},
	}

	require.Equal(t, expected, env)
}

func TestReadDirWithTestData(t *testing.T) {
	// Use test data directory
	testDataDir := "testdata/env"

	env, err := ReadDir(testDataDir)
	require.NoError(t, err)

	expected := Environment{
		"BAR":   {Value: "bar", NeedRemove: false},
		"EMPTY": {Value: "", NeedRemove: true},
		"FOO":   {Value: "   foo\nwith new line", NeedRemove: false},
		"HELLO": {Value: "\"hello\"", NeedRemove: false},
		"UNSET": {Value: "", NeedRemove: true},
	}

	require.Equal(t, expected, env)
}
