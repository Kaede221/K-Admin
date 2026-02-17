package core

import (
	"fmt"
	"os"
	"path/filepath"

	"k-admin-system/config"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// InitLogger initializes the Zap logger with Lumberjack for log rotation
// Returns a configured logger instance based on the application configuration
func InitLogger(cfg *config.Config) (*zap.Logger, error) {
	// Parse log level from configuration
	level, err := parseLogLevel(cfg.Logger.Level)
	if err != nil {
		return nil, fmt.Errorf("invalid log level: %w", err)
	}

	// Create encoder configuration
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "timestamp",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// Create encoder (JSON format for structured logging)
	encoder := zapcore.NewJSONEncoder(encoderConfig)

	// Create log file writer with rotation using Lumberjack
	logDir := filepath.Dir(cfg.Logger.Path)
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}

	fileWriter := &lumberjack.Logger{
		Filename:   cfg.Logger.Path,
		MaxSize:    cfg.Logger.MaxSize,    // megabytes
		MaxAge:     cfg.Logger.MaxAge,     // days
		MaxBackups: cfg.Logger.MaxBackups, // number of backups
		Compress:   cfg.Logger.Compress,   // compress rotated files
		LocalTime:  true,                  // use local time for filenames
	}

	// Determine output destinations based on server mode
	var core zapcore.Core
	if cfg.Server.Mode == "debug" || cfg.Server.Mode == "test" {
		// Development mode: output to both console and file
		consoleEncoder := zapcore.NewConsoleEncoder(encoderConfig)
		consoleCore := zapcore.NewCore(
			consoleEncoder,
			zapcore.AddSync(os.Stdout),
			level,
		)
		fileCore := zapcore.NewCore(
			encoder,
			zapcore.AddSync(fileWriter),
			level,
		)
		core = zapcore.NewTee(consoleCore, fileCore)
	} else {
		// Production mode: output to file only
		core = zapcore.NewCore(
			encoder,
			zapcore.AddSync(fileWriter),
			level,
		)
	}

	// Create logger with caller information and stack traces for errors
	logger := zap.New(core,
		zap.AddCaller(),
		zap.AddCallerSkip(0),
		zap.AddStacktrace(zapcore.ErrorLevel),
	)

	return logger, nil
}

// parseLogLevel converts string log level to zapcore.Level
func parseLogLevel(level string) (zapcore.Level, error) {
	switch level {
	case "debug":
		return zapcore.DebugLevel, nil
	case "info":
		return zapcore.InfoLevel, nil
	case "warn":
		return zapcore.WarnLevel, nil
	case "error":
		return zapcore.ErrorLevel, nil
	case "fatal":
		return zapcore.FatalLevel, nil
	default:
		return zapcore.InfoLevel, fmt.Errorf("unknown log level: %s", level)
	}
}

// LogInfo logs an informational message
func LogInfo(logger *zap.Logger, msg string, fields ...zap.Field) {
	logger.Info(msg, fields...)
}

// LogDebug logs a debug message
func LogDebug(logger *zap.Logger, msg string, fields ...zap.Field) {
	logger.Debug(msg, fields...)
}

// LogWarn logs a warning message
func LogWarn(logger *zap.Logger, msg string, fields ...zap.Field) {
	logger.Warn(msg, fields...)
}

// LogError logs an error message
func LogError(logger *zap.Logger, msg string, fields ...zap.Field) {
	logger.Error(msg, fields...)
}

// LogFatal logs a fatal message and exits the application
func LogFatal(logger *zap.Logger, msg string, fields ...zap.Field) {
	logger.Fatal(msg, fields...)
}

// SyncLogger flushes any buffered log entries
// Should be called before application shutdown
func SyncLogger(logger *zap.Logger) error {
	return logger.Sync()
}
