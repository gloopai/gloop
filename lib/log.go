package lib

import (
	"github.com/sirupsen/logrus"
)

var logger = logrus.New()

// LogLevel is a custom type that wraps logrus.Level
type LogLevel logrus.Level

const (
	LogLevelPanic LogLevel = LogLevel(logrus.PanicLevel)
	LogLevelFatal LogLevel = LogLevel(logrus.FatalLevel)
	LogLevelError LogLevel = LogLevel(logrus.ErrorLevel)
	LogLevelWarn  LogLevel = LogLevel(logrus.WarnLevel)
	LogLevelInfo  LogLevel = LogLevel(logrus.InfoLevel)
	LogLevelDebug LogLevel = LogLevel(logrus.DebugLevel)
	LogLevelTrace LogLevel = LogLevel(logrus.TraceLevel)
)

// SetLogLevel sets the log level for the logger
func SetLogLevel(level LogLevel) {
	logger.SetLevel(logrus.Level(level))
}

// InitLogger initializes the logger with default settings
func InitLogger(level LogLevel, format LogFormatter) {
	logger.SetLevel(logrus.Level(level))
	logger.SetFormatter(logrus.Formatter(format))
}

// Info logs an info level message
func Info(args ...interface{}) {
	logger.Info(args...)
}

// Warn logs a warning level message
func Warn(args ...interface{}) {
	logger.Warn(args...)
}

// Error logs an error level message
func Error(args ...interface{}) {
	logger.Error(args...)
}

// Debug logs a debug level message
func Debug(args ...interface{}) {
	logger.Debug(args...)
}

// Fatal logs a fatal level message and exits the application
func Fatal(args ...interface{}) {
	logger.Fatal(args...)
}

// LogFormatter is a custom type that wraps logrus.Formatter
type LogFormatter logrus.Formatter

// SetLogFormatter sets the log formatter for the logger
func SetLogFormatter(formatter LogFormatter) {
	logger.SetFormatter(logrus.Formatter(formatter))
}
