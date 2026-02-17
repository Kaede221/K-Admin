package middleware

import (
	"bytes"
	"encoding/json"
	"k-admin-system/global"
	"k-admin-system/model/common"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func TestRecovery_CatchesPanic(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	var buf bytes.Buffer
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "timestamp",
		LevelKey:       "level",
		MessageKey:     "msg",
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
	}
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),
		zapcore.AddSync(&buf),
		zapcore.ErrorLevel,
	)
	logger := zap.New(core)
	global.Logger = logger
	defer func() { global.Logger = nil }()

	router := gin.New()
	router.Use(Recovery())
	router.GET("/panic", func(c *gin.Context) {
		panic("test panic")
	})

	// Execute
	req := httptest.NewRequest("GET", "/panic", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code, "Should return 200 status")

	var response common.Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err, "Should parse response JSON")
	assert.Equal(t, 500, response.Code, "Response code should be 500")
	assert.Contains(t, response.Msg, "Internal server error", "Error message should contain 'Internal server error'")
	assert.Contains(t, response.Msg, "test panic", "Error message should contain panic message")
}

func TestRecovery_LogsStackTrace(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	var buf bytes.Buffer
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "timestamp",
		LevelKey:       "level",
		MessageKey:     "msg",
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
	}
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),
		zapcore.AddSync(&buf),
		zap.ErrorLevel,
	)
	logger := zap.New(core)
	global.Logger = logger
	defer func() { global.Logger = nil }()

	router := gin.New()
	router.Use(Recovery())
	router.GET("/panic", func(c *gin.Context) {
		panic("test panic with stack trace")
	})

	// Execute
	req := httptest.NewRequest("GET", "/panic", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	logOutput := buf.String()
	assert.NotEmpty(t, logOutput, "Should log panic")

	var logEntry map[string]any
	err := json.Unmarshal(buf.Bytes(), &logEntry)
	assert.NoError(t, err, "Should parse log entry")

	assert.Equal(t, "Panic recovered", logEntry["msg"], "Log message should be 'Panic recovered'")
	assert.NotNil(t, logEntry["error"], "Log should contain error field")
	assert.NotNil(t, logEntry["stack"], "Log should contain stack trace")

	stackStr, ok := logEntry["stack"].(string)
	assert.True(t, ok, "Stack should be a string")
	assert.Contains(t, stackStr, "goroutine", "Stack trace should contain goroutine info")
	assert.Contains(t, stackStr, ".go:", "Stack trace should contain file and line info")
}

func TestRecovery_LogsRequestInfo(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	var buf bytes.Buffer
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "timestamp",
		LevelKey:       "level",
		MessageKey:     "msg",
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
	}
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),
		zapcore.AddSync(&buf),
		zapcore.ErrorLevel,
	)
	logger := zap.New(core)
	global.Logger = logger
	defer func() { global.Logger = nil }()

	router := gin.New()
	router.Use(Recovery())
	router.POST("/api/v1/users", func(c *gin.Context) {
		panic("database error")
	})

	// Execute
	req := httptest.NewRequest("POST", "/api/v1/users", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	var logEntry map[string]any
	err := json.Unmarshal(buf.Bytes(), &logEntry)
	assert.NoError(t, err, "Should parse log entry")

	assert.Equal(t, "/api/v1/users", logEntry["path"], "Log should contain request path")
	assert.Equal(t, "POST", logEntry["method"], "Log should contain request method")
}

func TestRecovery_ServerContinuesRunning(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	var buf bytes.Buffer
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "timestamp",
		LevelKey:       "level",
		MessageKey:     "msg",
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
	}
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),
		zapcore.AddSync(&buf),
		zapcore.ErrorLevel,
	)
	logger := zap.New(core)
	global.Logger = logger
	defer func() { global.Logger = nil }()

	router := gin.New()
	router.Use(Recovery())
	router.GET("/panic", func(c *gin.Context) {
		panic("crash test")
	})
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Execute panic request
	req1 := httptest.NewRequest("GET", "/panic", nil)
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)

	// Execute normal request to verify server is still running
	req2 := httptest.NewRequest("GET", "/health", nil)
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)

	// Assert
	assert.Equal(t, http.StatusOK, w2.Code, "Server should still respond after panic")

	var healthResponse map[string]any
	err := json.Unmarshal(w2.Body.Bytes(), &healthResponse)
	assert.NoError(t, err, "Should parse health response")
	assert.Equal(t, "ok", healthResponse["status"], "Health check should return ok")
}

func TestRecovery_NilPointerDereference(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	var buf bytes.Buffer
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "timestamp",
		LevelKey:       "level",
		MessageKey:     "msg",
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
	}
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),
		zapcore.AddSync(&buf),
		zapcore.ErrorLevel,
	)
	logger := zap.New(core)
	global.Logger = logger
	defer func() { global.Logger = nil }()

	router := gin.New()
	router.Use(Recovery())
	router.GET("/nil", func(c *gin.Context) {
		var ptr *string
		_ = *ptr // This will cause nil pointer dereference
	})

	// Execute
	req := httptest.NewRequest("GET", "/nil", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code, "Should return 200 status")

	var response common.Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err, "Should parse response JSON")
	assert.Equal(t, 500, response.Code, "Response code should be 500")
	assert.Contains(t, response.Msg, "Internal server error", "Error message should indicate internal error")
}

func TestRecovery_WithoutLogger(t *testing.T) {
	// Setup - no logger configured
	gin.SetMode(gin.TestMode)
	global.Logger = nil

	router := gin.New()
	router.Use(Recovery())
	router.GET("/panic", func(c *gin.Context) {
		panic("test panic without logger")
	})

	// Execute
	req := httptest.NewRequest("GET", "/panic", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert - should still handle panic gracefully even without logger
	assert.Equal(t, http.StatusOK, w.Code, "Should return 200 status even without logger")

	var response common.Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err, "Should parse response JSON")
	assert.Equal(t, 500, response.Code, "Response code should be 500")
	assert.Contains(t, response.Msg, "Internal server error", "Error message should be present")
}

func TestRecovery_DifferentPanicTypes(t *testing.T) {
	testCases := []struct {
		name        string
		panicValue  any
		shouldCatch bool
	}{
		{
			name:        "String panic",
			panicValue:  "string error",
			shouldCatch: true,
		},
		{
			name:        "Integer panic",
			panicValue:  42,
			shouldCatch: true,
		},
		{
			name:        "Struct panic",
			panicValue:  struct{ msg string }{msg: "struct error"},
			shouldCatch: true,
		},
		{
			name:        "Nil panic",
			panicValue:  nil,
			shouldCatch: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup
			gin.SetMode(gin.TestMode)
			var buf bytes.Buffer
			encoderConfig := zapcore.EncoderConfig{
				TimeKey:        "timestamp",
				LevelKey:       "level",
				MessageKey:     "msg",
				EncodeLevel:    zapcore.LowercaseLevelEncoder,
				EncodeTime:     zapcore.ISO8601TimeEncoder,
				EncodeDuration: zapcore.StringDurationEncoder,
			}
			core := zapcore.NewCore(
				zapcore.NewJSONEncoder(encoderConfig),
				zapcore.AddSync(&buf),
				zapcore.ErrorLevel,
			)
			logger := zap.New(core)
			global.Logger = logger
			defer func() { global.Logger = nil }()

			router := gin.New()
			router.Use(Recovery())
			router.GET("/panic", func(c *gin.Context) {
				panic(tc.panicValue)
			})

			// Execute
			req := httptest.NewRequest("GET", "/panic", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Assert
			if tc.shouldCatch {
				assert.Equal(t, http.StatusOK, w.Code, "Should catch panic and return 200")

				var response common.Response
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err, "Should parse response JSON")
				assert.Equal(t, 500, response.Code, "Response code should be 500")

				// Verify logging occurred
				logOutput := buf.String()
				assert.NotEmpty(t, logOutput, "Should log the panic")
				assert.Contains(t, strings.ToLower(logOutput), "panic", "Log should mention panic")
			}
		})
	}
}
