package logging

import (
"context"
"fmt"
"io"
"os"
"strings"
"time"

"github.com/sirupsen/logrus"
)

// LogLevel represents the logging level
type LogLevel string

const (
LogLevelTrace LogLevel = "trace"
LogLevelDebug LogLevel = "debug"
LogLevelInfo  LogLevel = "info"
LogLevelWarn  LogLevel = "warn"
LogLevelError LogLevel = "error"
LogLevelFatal LogLevel = "fatal"
)

// Logger interface defines the logging contract
type Logger interface {
Trace(args ...interface{})
Debug(args ...interface{})
Info(args ...interface{})
Warn(args ...interface{})
Error(args ...interface{})
Fatal(args ...interface{})

Tracef(format string, args ...interface{})
Debugf(format string, args ...interface{})
Infof(format string, args ...interface{})
Warnf(format string, args ...interface{})
Errorf(format string, args ...interface{})
Fatalf(format string, args ...interface{})

WithField(key string, value interface{}) Logger
WithFields(fields map[string]interface{}) Logger
WithContext(ctx context.Context) Logger
}

// Config holds the logger configuration
type Config struct {
Level      LogLevel `yaml:"level" json:"level"`
Format     string   `yaml:"format" json:"format"`
Output     string   `yaml:"output" json:"output"`
TimeFormat string   `yaml:"time_format" json:"time_format"`
}

// DefaultConfig returns the default logger configuration
func DefaultConfig() *Config {
return &Config{
Level:      LogLevelInfo,
Format:     "json",
Output:     "stdout",
TimeFormat: time.RFC3339,
}
}

type logrusLogger struct {
logger *logrus.Logger
entry  *logrus.Entry
}

// NewLogger creates a new logger instance
func NewLogger(config *Config) (Logger, error) {
if config == nil {
config = DefaultConfig()
}

logger := logrus.New()

level, err := logrus.ParseLevel(string(config.Level))
if err != nil {
return nil, fmt.Errorf("invalid log level %s: %w", config.Level, err)
}
logger.SetLevel(level)

switch strings.ToLower(config.Format) {
case "json":
logger.SetFormatter(&logrus.JSONFormatter{
TimestampFormat: config.TimeFormat,
})
case "text":
logger.SetFormatter(&logrus.TextFormatter{
FullTimestamp:   true,
TimestampFormat: config.TimeFormat,
})
default:
return nil, fmt.Errorf("unsupported log format: %s", config.Format)
}

var output io.Writer
switch strings.ToLower(config.Output) {
case "stdout":
output = os.Stdout
case "stderr":
output = os.Stderr
default:
file, err := os.OpenFile(config.Output, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
if err != nil {
return nil, fmt.Errorf("failed to open log file %s: %w", config.Output, err)
}
output = file
}
logger.SetOutput(output)

return &logrusLogger{
logger: logger,
entry:  logrus.NewEntry(logger),
}, nil
}

func (l *logrusLogger) Trace(args ...interface{}) {
l.entry.Trace(args...)
}

func (l *logrusLogger) Debug(args ...interface{}) {
l.entry.Debug(args...)
}

func (l *logrusLogger) Info(args ...interface{}) {
l.entry.Info(args...)
}

func (l *logrusLogger) Warn(args ...interface{}) {
l.entry.Warn(args...)
}

func (l *logrusLogger) Error(args ...interface{}) {
l.entry.Error(args...)
}

func (l *logrusLogger) Fatal(args ...interface{}) {
l.entry.Fatal(args...)
}

func (l *logrusLogger) Tracef(format string, args ...interface{}) {
l.entry.Tracef(format, args...)
}

func (l *logrusLogger) Debugf(format string, args ...interface{}) {
l.entry.Debugf(format, args...)
}

func (l *logrusLogger) Infof(format string, args ...interface{}) {
l.entry.Infof(format, args...)
}

func (l *logrusLogger) Warnf(format string, args ...interface{}) {
l.entry.Warnf(format, args...)
}

func (l *logrusLogger) Errorf(format string, args ...interface{}) {
l.entry.Errorf(format, args...)
}

func (l *logrusLogger) Fatalf(format string, args ...interface{}) {
l.entry.Fatalf(format, args...)
}

func (l *logrusLogger) WithField(key string, value interface{}) Logger {
return &logrusLogger{
logger: l.logger,
entry:  l.entry.WithField(key, value),
}
}

func (l *logrusLogger) WithFields(fields map[string]interface{}) Logger {
return &logrusLogger{
logger: l.logger,
entry:  l.entry.WithFields(logrus.Fields(fields)),
}
}

func (l *logrusLogger) WithContext(ctx context.Context) Logger {
return &logrusLogger{
logger: l.logger,
entry:  l.entry.WithContext(ctx),
}
}

var globalLogger Logger

func InitGlobalLogger(config *Config) error {
logger, err := NewLogger(config)
if err != nil {
return err
}
globalLogger = logger
return nil
}

func GetLogger() Logger {
if globalLogger == nil {
logger, _ := NewLogger(DefaultConfig())
globalLogger = logger
}
return globalLogger
}

func Trace(args ...interface{}) {
GetLogger().Trace(args...)
}

func Debug(args ...interface{}) {
GetLogger().Debug(args...)
}

func Info(args ...interface{}) {
GetLogger().Info(args...)
}

func Warn(args ...interface{}) {
GetLogger().Warn(args...)
}

func Error(args ...interface{}) {
GetLogger().Error(args...)
}

func Fatal(args ...interface{}) {
GetLogger().Fatal(args...)
}

func Tracef(format string, args ...interface{}) {
GetLogger().Tracef(format, args...)
}

func Debugf(format string, args ...interface{}) {
GetLogger().Debugf(format, args...)
}

func Infof(format string, args ...interface{}) {
GetLogger().Infof(format, args...)
}

func Warnf(format string, args ...interface{}) {
GetLogger().Warnf(format, args...)
}

func Errorf(format string, args ...interface{}) {
GetLogger().Errorf(format, args...)
}

func Fatalf(format string, args ...interface{}) {
GetLogger().Fatalf(format, args...)
}

func WithField(key string, value interface{}) Logger {
return GetLogger().WithField(key, value)
}

func WithFields(fields map[string]interface{}) Logger {
return GetLogger().WithFields(fields)
}

func WithContext(ctx context.Context) Logger {
return GetLogger().WithContext(ctx)
}
