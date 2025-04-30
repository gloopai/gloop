package lib

import (
	"sync"
	"sync/atomic"

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

// 优化了性能和线程安全性，使用了sync.Once来确保InitLogger只被初始化一次
var initOnce sync.Once

// InitLogger initializes the logger with default settings
func (l *log) InitLogger(level LogLevel, format LogFormatter) {
	initOnce.Do(func() {
		l.logger.SetLevel(logrus.Level(level))

		if format == nil {
			format = &logrus.TextFormatter{}
		}
		l.logger.SetFormatter(logrus.Formatter(format))
	})
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

var debugEnabled int32 = 1 // 1 for true, 0 for false

// Debug logs a debug level message
func (l *log) Debug(args ...interface{}) {
	if atomic.LoadInt32(&debugEnabled) == 0 {
		return
	}
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
	if atomic.LoadInt32(&debugEnabled) == 0 {
		return
	}
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

// SetDebugEnabled sets the debugEnabled flag
func (l *log) SetDebugEnabled(enabled bool) {
	if enabled {
		atomic.StoreInt32(&debugEnabled, 1)
	} else {
		atomic.StoreInt32(&debugEnabled, 0)
	}
}
