package logger

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"strings"
	"testing"
)

func TestLogger_Levels(t *testing.T) {
	// Создаем буфер для перехвата вывода логгера
	var buf bytes.Buffer

	// Инициализируем логгер с наивысшим уровнем логирования
	log := New("debug") // Предполагается, что New возвращает *LogrusLogger
	log.(*LogrusLogger).Logger.SetOutput(&buf)

	// Словарь уровней логирования и сообщений для тестирования
	tests := []struct {
		logFunc     func(args ...interface{})
		expectedMsg string
	}{
		{log.Debug, "debug message"},
		{log.Info, "info message"},
		{log.Warn, "warn message"},
		{log.Error, "error message"},
	}

	for _, tc := range tests {
		buf.Reset() // Очищаем буфер перед каждым тестом
		tc.logFunc(tc.expectedMsg)

		if !strings.Contains(buf.String(), tc.expectedMsg) {
			t.Errorf("Expected '%s' to be in log output", tc.expectedMsg)
		}
	}
}

// Так как Fatal вызывает os.Exit(), мы проводим тест отдельно, через запуск отдельного процесса.
func TestLogFatal(t *testing.T) {
	if os.Getenv("BE_CRASHER") == "1" {
		log := New("debug") // Создайте ваш логгер
		log.Fatal("fatal message")
		return
	}
	cmd := exec.Command(os.Args[0], "-test.run=TestLogFatal")
	cmd.Env = append(os.Environ(), "BE_CRASHER=1")
	err := cmd.Run()
	var e *exec.ExitError
	if errors.As(err, &e) && !e.Success() {
		return // Тест успешен, если процесс завершился с ошибкой
	}
	t.Fatalf("process ran with err %v, want exit status 1", err)
}

func TestLogger_FormattedLevels(t *testing.T) {
	var buf bytes.Buffer
	log := New("debug")
	log.(*LogrusLogger).Logger.SetOutput(&buf)

	tests := []struct {
		logFunc     func(format string, args ...interface{})
		format      string
		args        []interface{}
		expectedMsg string
	}{
		{log.Debugf, "debug %s", []interface{}{"formatted message"}, "debug formatted message"},
		{log.Infof, "info %s", []interface{}{"formatted message"}, "info formatted message"},
		{log.Warnf, "warn %s", []interface{}{"formatted message"}, "warn formatted message"},
		{log.Errorf, "error %s", []interface{}{"formatted message"}, "error formatted message"},
	}

	for _, tc := range tests {
		buf.Reset()
		tc.logFunc(tc.format, tc.args...)

		if !strings.Contains(buf.String(), tc.expectedMsg) {
			t.Errorf("Expected '%s' to be in log output", tc.expectedMsg)
		}
	}
}

// Тестирование Fatalf следует проводить отдельно, так как он вызывает os.Exit(), подобно тесту для Fatal.
func TestLogFatalf(t *testing.T) {
	if os.Getenv("BE_CRASHER_F") == "1" {
		log := New("debug")
		log.Fatalf("fatal formatted message %d", 1)
		return
	}
	cmd := exec.Command(os.Args[0], "-test.run=TestLogFatalf")
	cmd.Env = append(os.Environ(), "BE_CRASHER_F=1")
	err := cmd.Run()
	var e *exec.ExitError
	if errors.As(err, &e) && !e.Success() {
		return // Тест успешен, если процесс завершился с ошибкой
	}
	t.Fatalf("process ran with err %v, want exit status 1", err)
}
