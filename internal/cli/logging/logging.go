package logging

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// CLILogger provides logging for the CLI application
type CLILogger struct {
	logger *zap.Logger
	sugar  *zap.SugaredLogger
}

// NewCLILogger creates a new CLI logger
func NewCLILogger(verbose bool) (*CLILogger, error) {
	config := zap.NewDevelopmentConfig()
	config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	config.Level = zap.NewAtomicLevelAt(zapcore.InfoLevel)

	if verbose {
		config.Level = zap.NewAtomicLevelAt(zapcore.DebugLevel)
	}

	logger, err := config.Build()
	if err != nil {
		return nil, err
	}

	return &CLILogger{
		logger: logger,
		sugar:  logger.Sugar(),
	}, nil
}

// Debug logs a debug message
func (l *CLILogger) Debug(msg string, fields ...zapcore.Field) {
	l.logger.Debug(msg, fields...)
}

// Info logs an info message
func (l *CLILogger) Info(msg string, fields ...zapcore.Field) {
	l.logger.Info(msg, fields...)
}

// Warn logs a warning message
func (l *CLILogger) Warn(msg string, fields ...zapcore.Field) {
	l.logger.Warn(msg, fields...)
}

// Error logs an error message
func (l *CLILogger) Error(msg string, fields ...zapcore.Field) {
	l.logger.Error(msg, fields...)
}

// Sugar returns the sugared logger for structured logging
func (l *CLILogger) Sugar() *zap.SugaredLogger {
	return l.sugar
}

// Sync syncs the logger
func (l *CLILogger) Sync() error {
	return l.logger.Sync()
}

// Global logger instance
var globalLogger *CLILogger

// InitGlobalLogger initializes the global logger
func InitGlobalLogger(verbose bool) error {
	logger, err := NewCLILogger(verbose)
	if err != nil {
		return err
	}
	globalLogger = logger
	return nil
}

// Debug logs a debug message to the global logger
func Debug(msg string, fields ...zapcore.Field) {
	if globalLogger != nil {
		globalLogger.Debug(msg, fields...)
	}
}

// Info logs an info message to the global logger
func Info(msg string, fields ...zapcore.Field) {
	if globalLogger != nil {
		globalLogger.Info(msg, fields...)
	}
}

// Warn logs a warning message to the global logger
func Warn(msg string, fields ...zapcore.Field) {
	if globalLogger != nil {
		globalLogger.Warn(msg, fields...)
	}
}

// Error logs an error message to the global logger
func Error(msg string, fields ...zapcore.Field) {
	if globalLogger != nil {
		globalLogger.Error(msg, fields...)
	}
}

// Sugar returns the global sugared logger
func Sugar() *zap.SugaredLogger {
	if globalLogger != nil {
		return globalLogger.sugar
	}
	// Fallback to stderr if logger not initialized
	return zap.New(zapcore.NewCore(
		zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
		zapcore.AddSync(os.Stderr),
		zapcore.InfoLevel,
	)).Sugar()
}