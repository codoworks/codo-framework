// Package logger provides a logger client for the framework.
package logger

import (
	"fmt"
	"io"
	"os"

	"github.com/codoworks/codo-framework/core/clients"
	"github.com/sirupsen/logrus"
)

const (
	// ClientName is the name of the logger client.
	ClientName = "logger"
)

// Level represents log level.
type Level string

const (
	LevelDebug Level = "debug"
	LevelInfo  Level = "info"
	LevelWarn  Level = "warn"
	LevelError Level = "error"
	LevelFatal Level = "fatal"
	LevelPanic Level = "panic"
)

// Format represents log format.
type Format string

const (
	FormatJSON Format = "json"
	FormatText Format = "text"
)

// Config holds logger configuration.
type Config struct {
	Level  Level  `json:"level" yaml:"level"`
	Format Format `json:"format" yaml:"format"`
	Output io.Writer
}

// DefaultConfig returns default logger configuration.
func DefaultConfig() *Config {
	return &Config{
		Level:  LevelInfo,
		Format: FormatJSON,
		Output: os.Stdout,
	}
}

// Logger is the logger client.
type Logger struct {
	clients.BaseClient
	logger *logrus.Logger
	config *Config
}

// New creates a new logger client.
func New() *Logger {
	return &Logger{
		BaseClient: clients.NewBaseClient(ClientName),
		logger:     logrus.New(),
	}
}

// Name returns the client name.
func (l *Logger) Name() string {
	return ClientName
}

// Initialize sets up the logger with configuration.
func (l *Logger) Initialize(cfg any) error {
	config := DefaultConfig()

	if cfg != nil {
		switch c := cfg.(type) {
		case *Config:
			config = c
		case Config:
			config = &c
		case map[string]any:
			if level, ok := c["level"].(string); ok {
				config.Level = Level(level)
			}
			if format, ok := c["format"].(string); ok {
				config.Format = Format(format)
			}
		default:
			return fmt.Errorf("invalid config type: %T", cfg)
		}
	}

	l.config = config

	// Set level
	level, err := logrus.ParseLevel(string(config.Level))
	if err != nil {
		level = logrus.InfoLevel
	}
	l.logger.SetLevel(level)

	// Set format
	if config.Format == FormatJSON {
		l.logger.SetFormatter(&logrus.JSONFormatter{})
	} else {
		l.logger.SetFormatter(&logrus.TextFormatter{})
	}

	// Set output
	if config.Output != nil {
		l.logger.SetOutput(config.Output)
	}

	return l.BaseClient.Initialize(cfg)
}

// Health checks if the logger is healthy.
func (l *Logger) Health() error {
	if l.logger == nil {
		return fmt.Errorf("logger not initialized")
	}
	return nil
}

// Shutdown shuts down the logger.
func (l *Logger) Shutdown() error {
	return l.BaseClient.Shutdown()
}

// GetLogger returns the underlying logrus logger.
func (l *Logger) GetLogger() *logrus.Logger {
	return l.logger
}

// WithField creates an entry with a single field.
func (l *Logger) WithField(key string, value any) *logrus.Entry {
	return l.logger.WithField(key, value)
}

// WithFields creates an entry with multiple fields.
func (l *Logger) WithFields(fields map[string]any) *logrus.Entry {
	return l.logger.WithFields(logrus.Fields(fields))
}

// WithError creates an entry with an error.
func (l *Logger) WithError(err error) *logrus.Entry {
	return l.logger.WithError(err)
}

// Debug logs a debug message.
func (l *Logger) Debug(args ...any) {
	l.logger.Debug(args...)
}

// Debugf logs a formatted debug message.
func (l *Logger) Debugf(format string, args ...any) {
	l.logger.Debugf(format, args...)
}

// Info logs an info message.
func (l *Logger) Info(args ...any) {
	l.logger.Info(args...)
}

// Infof logs a formatted info message.
func (l *Logger) Infof(format string, args ...any) {
	l.logger.Infof(format, args...)
}

// Warn logs a warning message.
func (l *Logger) Warn(args ...any) {
	l.logger.Warn(args...)
}

// Warnf logs a formatted warning message.
func (l *Logger) Warnf(format string, args ...any) {
	l.logger.Warnf(format, args...)
}

// Error logs an error message.
func (l *Logger) Error(args ...any) {
	l.logger.Error(args...)
}

// Errorf logs a formatted error message.
func (l *Logger) Errorf(format string, args ...any) {
	l.logger.Errorf(format, args...)
}

// Fatal logs a fatal message and exits.
func (l *Logger) Fatal(args ...any) {
	l.logger.Fatal(args...)
}

// Fatalf logs a formatted fatal message and exits.
func (l *Logger) Fatalf(format string, args ...any) {
	l.logger.Fatalf(format, args...)
}

// Panic logs a panic message and panics.
func (l *Logger) Panic(args ...any) {
	l.logger.Panic(args...)
}

// Panicf logs a formatted panic message and panics.
func (l *Logger) Panicf(format string, args ...any) {
	l.logger.Panicf(format, args...)
}

// SetLevel sets the log level.
func (l *Logger) SetLevel(level Level) {
	parsedLevel, err := logrus.ParseLevel(string(level))
	if err != nil {
		parsedLevel = logrus.InfoLevel
	}
	l.logger.SetLevel(parsedLevel)
}

// GetLevel returns the current log level.
func (l *Logger) GetLevel() Level {
	return Level(l.logger.GetLevel().String())
}
