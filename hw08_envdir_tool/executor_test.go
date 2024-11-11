package main

import (
	"bytes"
	"os"
	"os/exec"
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRunCmd(t *testing.T) {
	env := Environment{
		"FOO":   {Value: "123", NeedRemove: false},
		"BAR":   {Value: "value", NeedRemove: false},
		"EMPTY": {Value: "", NeedRemove: true},
	}

	var cmd []string
	if runtime.GOOS == "windows" {
		cmd = []string{"cmd", "/C", "set"}
	} else {
		cmd = []string{"env"}
	}

	// Перенаправим вывод команды в буфер
	var stdout bytes.Buffer
	execCmd := exec.Command(cmd[0], cmd[1:]...)
	execCmd.Env = os.Environ()
	for key, value := range env {
		if value.NeedRemove {
			execCmd.Env = removeEnv(execCmd.Env, key)
		} else {
			execCmd.Env = append(execCmd.Env, key+"="+value.Value)
		}
	}
	execCmd.Stdout = &stdout
	execCmd.Stderr = os.Stderr

	err := execCmd.Run()
	require.NoError(t, err)

	output := stdout.String()

	// Проверка, что переменные окружения установлены правильно
	require.Contains(t, output, "FOO=123")
	require.Contains(t, output, "BAR=value")

	// Проверка, что переменная окружения удалена правильно
	require.NotContains(t, output, "EMPTY")
}

func TestCmdDirWithReadDirWithTestData(t *testing.T) {
	// Используем директорию с тестовыми данными
	testDataDir := "testdata/env"

	env, err := ReadDir(testDataDir)
	require.NoError(t, err)

	var cmd []string
	if runtime.GOOS == "windows" {
		cmd = []string{"cmd", "/C", "set"}
	} else {
		cmd = []string{"env"}
	}

	// Перенаправим вывод команды в буфер
	var stdout bytes.Buffer
	execCmd := exec.Command(cmd[0], cmd[1:]...)
	execCmd.Env = os.Environ()
	for key, value := range env {
		if value.NeedRemove {
			execCmd.Env = removeEnv(execCmd.Env, key)
		} else {
			execCmd.Env = append(execCmd.Env, key+"="+value.Value)
		}
	}
	execCmd.Stdout = &stdout
	execCmd.Stderr = os.Stderr

	errCmd := execCmd.Run()
	require.NoError(t, errCmd)

	output := stdout.String()

	// Проверка, что переменные окружения установлены правильно
	require.Contains(t, output, "FOO=   foo\nwith new line")
	require.Contains(t, output, "BAR=bar")
	require.Contains(t, output, "HELLO=\"hello\"")

	// Проверка, что переменная окружения удалена правильно
	require.NotContains(t, output, "EMPTY", "UNSET")
}
