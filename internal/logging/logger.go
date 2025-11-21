package logging

import (
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// NewLogger creates a new zap logger with the specified log level
func NewLogger(level string) (*zap.Logger, error) {
	var zapLevel zapcore.Level
	switch level {
	case "debug":
		zapLevel = zapcore.DebugLevel
	case "info":
		zapLevel = zapcore.InfoLevel
	case "warn":
		zapLevel = zapcore.WarnLevel
	case "error":
		zapLevel = zapcore.ErrorLevel
	default:
		zapLevel = zapcore.InfoLevel
	}

	// Create encoder config
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "timestamp",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "message",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// Use lumberjack for log rotation if configured
	var writeSyncer zapcore.WriteSyncer
	if shouldUseFileLogging() {
		// Create a file syncer with rotation
		lumberjackLogger := &lumberjack.Logger{
			Filename:   "./logs/algonius-supervisor.log",
			MaxSize:    100, // megabytes
			MaxBackups: 3,
			MaxAge:     28, // days
			Compress:   true,
		}
		writeSyncer = zapcore.AddSync(lumberjackLogger)
	} else {
		// Write to stdout
		writeSyncer = zapcore.AddSync(os.Stdout)
	}

	// Create the logger
	encoder := zapcore.NewJSONEncoder(encoderConfig)
	logger := zapcore.NewCore(
		encoder,
		writeSyncer,
		zapLevel,
	)

	return zap.New(logger), nil
}

// shouldUseFileLogging determines whether to log to file based on environment
func shouldUseFileLogging() bool {
	logTo := os.Getenv("LOG_TO")
	if logTo == "file" {
		return true
	}
	return false
}

// Middleware creates a Gin middleware for logging HTTP requests
func Middleware(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// Process request
		c.Next()

		// Log after request completes
		end := time.Now()
		latency := end.Sub(start)
		clientIP := c.ClientIP()
		method := c.Request.Method
		statusCode := c.Writer.Status()

		if raw != "" {
			path = path + "?" + raw
		}

		logger.Info("HTTP Request",
			zap.String("client_ip", clientIP),
			zap.String("method", method),
			zap.String("path", path),
			zap.Int("status_code", statusCode),
			zap.Duration("latency", latency),
			zap.String("user_agent", c.Request.UserAgent()),
			zap.Int("bytes", c.Writer.Size()),
		)
	}
}

// WithExecutionID adds execution ID to the logger context
func WithExecutionID(logger *zap.Logger, executionID string) *zap.Logger {
	return logger.With(zap.String("execution_id", executionID))
}

// WithAgentID adds agent ID to the logger context
func WithAgentID(logger *zap.Logger, agentID string) *zap.Logger {
	return logger.With(zap.String("agent_id", agentID))
}

// WithTaskID adds task ID to the logger context
func WithTaskID(logger *zap.Logger, taskID string) *zap.Logger {
	return logger.With(zap.String("task_id", taskID))
}

// LogSensitiveData is a helper to safely log data that might contain sensitive information
func LogSensitiveData(logger *zap.Logger, message string, data string, shouldLog bool) {
	if shouldLog {
		logger.Info(message, zap.String("data", data))
	} else {
		logger.Info(message, zap.String("data", "[REDACTED]"))
	}
}