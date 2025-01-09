package logger

//nolint:depguard
import "github.com/sirupsen/logrus"

type Logger interface {
	Println(args ...interface{})
	Printf(format string, args ...interface{})
	Debug(args ...interface{})
	Debugf(format string, args ...interface{})
	Info(args ...interface{})
	Infof(format string, args ...interface{})
	Warn(args ...interface{})
	Warnf(format string, args ...interface{})
	Error(args ...interface{})
	Errorf(format string, args ...interface{})
	Fatal(args ...interface{})
	Fatalf(format string, args ...interface{})
}

type LogrusLogger struct {
	*logrus.Logger
}

func New(level string) Logger {
	baseLogger := logrus.New()
	lvl, err := logrus.ParseLevel(level)
	if err != nil {
		baseLogger.Errorf("Error parsing log level: %v", err)
		return nil
	}

	baseLogger.SetLevel(lvl)
	baseLogger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})

	return &LogrusLogger{baseLogger}
}

func (l *LogrusLogger) Debug(args ...interface{}) {
	l.Logger.Debug(args...)
}

func (l *LogrusLogger) Debugf(format string, args ...interface{}) {
	l.Logger.Debugf(format, args...)
}

func (l *LogrusLogger) Info(args ...interface{}) {
	l.Logger.Info(args...)
}

func (l *LogrusLogger) Infof(format string, args ...interface{}) {
	l.Logger.Infof(format, args...)
}

func (l *LogrusLogger) Warn(args ...interface{}) {
	l.Logger.Warn(args...)
}

func (l *LogrusLogger) Warnf(format string, args ...interface{}) {
	l.Logger.Warnf(format, args...)
}

func (l *LogrusLogger) Error(args ...interface{}) {
	l.Logger.Error(args...)
}

func (l *LogrusLogger) Errorf(format string, args ...interface{}) {
	l.Logger.Errorf(format, args...)
}

func (l *LogrusLogger) Fatal(args ...interface{}) {
	l.Logger.Fatal(args...)
}

func (l *LogrusLogger) Fatalf(format string, args ...interface{}) {
	l.Logger.Fatalf(format, args...)
}

func (l *LogrusLogger) Println(args ...interface{}) {
	l.Logger.Println(args...)
}

func (l *LogrusLogger) Printf(format string, args ...interface{}) {
	l.Logger.Printf(format, args...)
}
