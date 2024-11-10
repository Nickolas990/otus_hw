package main

import (
	"log"
	"os"
	"runtime"
	"testing"

	//nolint:depguard
	"github.com/stretchr/testify/require"
)

func TestCopy(t *testing.T) {
	srcContent := []byte("Hello, World!")
	srcFile, err := os.CreateTemp("", "srcFile")
	require.NoError(t, err)

	_, err = srcFile.Write(srcContent)
	require.NoError(t, err)
	srcFile.Close()

	tests := []struct {
		name        string
		fromPath    string
		toPath      string
		offset      int64
		limit       int64
		expected    string
		expectError error
	}{
		{
			name:        "Copy full file",
			fromPath:    srcFile.Name(),
			toPath:      "destFile",
			offset:      0,
			limit:       0,
			expected:    "Hello, World!",
			expectError: nil,
		},
		{
			name:        "Copy with offset",
			fromPath:    srcFile.Name(),
			toPath:      "destFile",
			offset:      7,
			limit:       0,
			expected:    "World!",
			expectError: nil,
		},
		{
			name:        "Copy with limit",
			fromPath:    srcFile.Name(),
			toPath:      "destFile",
			offset:      0,
			limit:       5,
			expected:    "Hello",
			expectError: nil,
		},
		{
			name:        "Copy with offset and limit",
			fromPath:    srcFile.Name(),
			toPath:      "destFile",
			offset:      7,
			limit:       5,
			expected:    "World",
			expectError: nil,
		},
		{
			name:        "Offset exceeds file size",
			fromPath:    srcFile.Name(),
			toPath:      "destFile",
			offset:      int64(len(srcContent)) + 1,
			limit:       0,
			expected:    "",
			expectError: ErrOffsetExceedsFileSize,
		},
		{
			name:        "Source and destination are the same",
			fromPath:    srcFile.Name(),
			toPath:      srcFile.Name(),
			offset:      0,
			limit:       0,
			expected:    "Hello, World!",
			expectError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			destFile, err := os.CreateTemp("", "destFile")
			require.NoError(t, err)
			defer func() {
				err := destFile.Close()
				if err != nil {
					log.Fatalf("failed to close destination file: %v", err)
					return
				}
				err = os.Remove(destFile.Name())
				if err != nil {
					log.Fatalf("failed to remove destination file: %v", err)
					return
				}
			}()

			err = Copy(srcFile.Name(), destFile.Name(), tt.offset, tt.limit)
			if tt.expectError != nil {
				require.Equal(t, tt.expectError, err)
				return
			}
			require.NoError(t, err)

			destContent, err := os.ReadFile(destFile.Name())
			require.NoError(t, err)
			require.Equal(t, tt.expected, string(destContent))
		})
	}

	// Удаление временного исходного файла после всех тестов
	err = os.Remove(srcFile.Name())
	if err != nil {
		log.Fatalf("failed to remove source file: %v", err)
		return
	}
}

func TestUnsupportedFile(t *testing.T) {
	var unsupportedFile string

	if runtime.GOOS == "windows" {
		unsupportedFile = "CON"
	} else {
		unsupportedFile = "/dev/null"
	}

	destFile, err := os.CreateTemp("", "destFile")
	require.NoError(t, err)
	defer func() {
		destFile.Close()
		os.Remove(destFile.Name())
	}()

	err = Copy(unsupportedFile, destFile.Name(), 0, 0)
	require.Equal(t, ErrUnsupportedFile, err)
}
