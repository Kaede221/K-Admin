package core

import (
	"os"
	"path/filepath"
	"testing"

	"k-admin-system/config"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func TestInitLogger(t *testing.T) {
	tests := []struct {
		name    string
		cfg     *config.Config
		wantErr bool
	}{
		{
			name: "valid configuration with debug level",
			cfg: &config.Config{
				Server: config.ServerConfig{
					Mode: "debug",
					Port: ":8080",
				},
				Logger: config.LoggerConfig{
					Level:      "debug",
					Path:       "./test_logs/test.log",
					MaxSize:    10,
					MaxAge:     7,
					MaxBackups: 3,
					Compress:   false,
				},
			},
			wantErr: false,
		},
		{
			name: "valid configuration with info level",
			cfg: &config.Config{
				Server: config.ServerConfig{
					Mode: "release",
					Port: ":8080",
				},
				Logger: config.LoggerConfig{
					Level:      "info",
					Path:       "./test_logs/test_info.log",
					MaxSize:    100,
					MaxAge:     7,
					MaxBackups: 3,
					Compress:   true,
				},
			},
			wantErr: false,
		},
		{
			name: "invalid log level",
			cfg: &config.Config{
				Server: config.ServerConfig{
					Mode: "debug",
					Port: ":8080",
				},
				Logger: config.LoggerConfig{
					Level:      "invalid",
					Path:       "./test_logs/test_invalid.log",
					MaxSize:    10,
					MaxAge:     7,
					MaxBackups: 3,
					Compress:   false,
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger, err := InitLogger(tt.cfg)
			if (err != nil) != tt.wantErr {
				t.Errorf("InitLogger() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && logger == nil {
				t.Error("InitLogger() returned nil logger without error")
			}
			if logger != nil {
				_ = logger.Sync()
			}
		})
	}

	// Cleanup test logs
	_ = os.RemoveAll("./test_logs")
}

func TestParseLogLevel(t *testing.T) {
	tests := []struct {
		name      string
		level     string
		wantLevel zapcore.Level
		wantErr   bool
	}{
		{"debug level", "debug", zapcore.DebugLevel, false},
		{"info level", "info", zapcore.InfoLevel, false},
		{"warn level", "warn", zapcore.WarnLevel, false},
		{"error level", "error", zapcore.ErrorLevel, false},
		{"fatal level", "fatal", zapcore.FatalLevel, false},
		{"invalid level", "invalid", zapcore.InfoLevel, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotLevel, err := parseLogLevel(tt.level)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseLogLevel() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotLevel != tt.wantLevel {
				t.Errorf("parseLogLevel() = %v, want %v", gotLevel, tt.wantLevel)
			}
		})
	}
}

func TestLogHelpers(t *testing.T) {
	// Create a temporary logger for testing
	cfg := &config.Config{
		Server: config.ServerConfig{
			Mode: "test",
			Port: ":8080",
		},
		Logger: config.LoggerConfig{
			Level:      "debug",
			Path:       "./test_logs/helpers.log",
			MaxSize:    10,
			MaxAge:     7,
			MaxBackups: 3,
			Compress:   false,
		},
	}

	logger, err := InitLogger(cfg)
	if err != nil {
		t.Fatalf("Failed to initialize logger: %v", err)
	}
	defer func() {
		_ = logger.Sync()
		_ = os.RemoveAll("./test_logs")
	}()

	// Test log helper functions
	t.Run("LogInfo", func(t *testing.T) {
		LogInfo(logger, "test info message", zap.String("key", "value"))
	})

	t.Run("LogDebug", func(t *testing.T) {
		LogDebug(logger, "test debug message", zap.Int("count", 42))
	})

	t.Run("LogWarn", func(t *testing.T) {
		LogWarn(logger, "test warning message", zap.Bool("flag", true))
	})

	t.Run("LogError", func(t *testing.T) {
		LogError(logger, "test error message", zap.Error(os.ErrNotExist))
	})

	// Verify log file was created and contains data
	logPath := filepath.Join(".", "test_logs", "helpers.log")
	info, err := os.Stat(logPath)
	if err != nil {
		t.Errorf("Log file not created: %v", err)
	}
	if info.Size() == 0 {
		t.Error("Log file is empty")
	}
}

func TestLogFileRotation(t *testing.T) {
	cfg := &config.Config{
		Server: config.ServerConfig{
			Mode: "release",
			Port: ":8080",
		},
		Logger: config.LoggerConfig{
			Level:      "info",
			Path:       "./test_logs/rotation.log",
			MaxSize:    1, // 1MB for testing
			MaxAge:     7,
			MaxBackups: 2,
			Compress:   false,
		},
	}

	logger, err := InitLogger(cfg)
	if err != nil {
		t.Fatalf("Failed to initialize logger: %v", err)
	}
	defer func() {
		_ = logger.Sync()
		_ = os.RemoveAll("./test_logs")
	}()

	// Write some logs
	for i := 0; i < 100; i++ {
		logger.Info("test message", zap.Int("iteration", i))
	}

	// Verify log file exists
	logPath := filepath.Join(".", "test_logs", "rotation.log")
	if _, err := os.Stat(logPath); err != nil {
		t.Errorf("Log file not created: %v", err)
	}
}

func TestEnvironmentSpecificOutput(t *testing.T) {
	tests := []struct {
		name string
		mode string
	}{
		{"debug mode", "debug"},
		{"test mode", "test"},
		{"release mode", "release"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{
				Server: config.ServerConfig{
					Mode: tt.mode,
					Port: ":8080",
				},
				Logger: config.LoggerConfig{
					Level:      "info",
					Path:       "./test_logs/" + tt.mode + ".log",
					MaxSize:    10,
					MaxAge:     7,
					MaxBackups: 3,
					Compress:   false,
				},
			}

			logger, err := InitLogger(cfg)
			if err != nil {
				t.Fatalf("Failed to initialize logger: %v", err)
			}
			defer func() {
				_ = logger.Sync()
			}()

			// Log a test message
			logger.Info("test message for " + tt.mode)

			// Verify log file was created
			logPath := filepath.Join(".", "test_logs", tt.mode+".log")
			if _, err := os.Stat(logPath); err != nil {
				t.Errorf("Log file not created for mode %s: %v", tt.mode, err)
			}
		})
	}

	// Cleanup
	_ = os.RemoveAll("./test_logs")
}
