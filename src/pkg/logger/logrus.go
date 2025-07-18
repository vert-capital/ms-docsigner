package logger

import (
	"os"

	"github.com/sirupsen/logrus"
)

func NewLogrusLogger(level string) *logrus.Logger {
	logger := logrus.New()

	logger.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: "2006-01-02T15:04:05Z07:00",
	})

	logger.SetOutput(os.Stdout)

	switch LogLevel(level) {
	case DEBUG:
		logger.SetLevel(logrus.DebugLevel)
	case INFO:
		logger.SetLevel(logrus.InfoLevel)
	case WARN:
		logger.SetLevel(logrus.WarnLevel)
	case ERROR:
		logger.SetLevel(logrus.ErrorLevel)
	case SILENT:
		logger.SetLevel(logrus.PanicLevel)
	default:
		logger.SetLevel(logrus.InfoLevel)
	}

	return logger
}
