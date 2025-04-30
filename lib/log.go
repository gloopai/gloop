package lib

import (
	"github.com/sirupsen/logrus"
)

type log struct {
	logger *logrus.Logger
}

var (
	Log = &log{
		logger: logrus.New(),
	}
)

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
func (l *log) SetLogLevel(level LogLevel) {
	l.logger.SetLevel(logrus.Level(level))
}

// InitLogger initializes the logger with default settings
func (l *log) InitLogger(level LogLevel, format LogFormatter) {
	l.logger.SetLevel(logrus.Level(level))
	l.logger.SetFormatter(logrus.Formatter(format))
}

// Info logs an info level message
func (l *log) Info(args ...interface{}) {
	l.logger.Info(args...)
}

// Warn logs a warning level message
func (l *log) Warn(args ...interface{}) {
	l.logger.Warn(args...)
}

// Error logs an error level message
func (l *log) Error(args ...interface{}) {
	l.logger.Error(args...)
}

// Debug logs a debug level message
func (l *log) Debug(args ...interface{}) {
	l.logger.Debug(args...)
}

// Fatal logs a fatal level message and exits the application
func (l *log) Fatal(args ...interface{}) {
	l.logger.Fatal(args...)
}

// Infof logs a formatted info level message
func (l *log) Infof(format string, args ...interface{}) {
	l.logger.Infof(format, args...)
}

// Warnf logs a formatted warning level message
func (l *log) Warnf(format string, args ...interface{}) {
	l.logger.Warnf(format, args...)
}

// Errorf logs a formatted error level message
func (l *log) Errorf(format string, args ...interface{}) {
	l.logger.Errorf(format, args...)
}

// Debugf logs a formatted debug level message
func (l *log) Debugf(format string, args ...interface{}) {
	l.logger.Debugf(format, args...)
}

// Fatalf logs a formatted fatal level message and exits the application
func (l *log) Fatalf(format string, args ...interface{}) {
	l.logger.Fatalf(format, args...)
}

// LogFormatter is a custom type that wraps logrus.Formatter
type LogFormatter logrus.Formatter

// SetLogFormatter sets the log formatter for the logger
func (l *log) SetLogFormatter(formatter LogFormatter) {
	l.logger.SetFormatter(logrus.Formatter(formatter))
}
