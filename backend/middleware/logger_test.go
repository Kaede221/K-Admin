package middleware

import (
	"bytes"
	"encoding/json"
	"k-admin-system/global"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// setupLoggerWithBuffer creates a test logger that writes to a buffer
func setupLoggerWithBuffer() (*zap.Logger, *bytes.Buffer) {
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
		zapcore.InfoLevel,
	)
	logger := zap.New(core)
	return logger, &buf
}

func TestLogger_LogsRequestDetails(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	logger, buf := setupLoggerWithBuffer()
	global.Logger = logger
	defer func() { global.Logger = nil }()

	router := gin.New()
	router.Use(Logger())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "success"})
	})

	// Execute
	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "192.168.1.1:12345"
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Verify
	assert.Equal(t, 200, w.Code)

	// Parse log output
	var logEntry map[string]interface{}
	err := json.Unmarshal(buf.Bytes(), &logEntry)
	assert.NoError(t, err)

	// Verify log fields
	assert.Equal(t, "info", logEntry["level"])
	assert.Equal(t, "HTTP Request", logEntry["msg"])
	assert.Equal(t, "GET", logEntry["method"])
	assert.Equal(t, "/test", logEntry["path"])
	assert.Equal(t, float64(200), logEntry["status"])
	assert.NotEmpty(t, logEntry["latency"])
	assert.NotEmpty(t, logEntry["client_ip"])
}

func TestLogger_LogsDifferentMethods(t *testing.T) {
	methods := []string{"GET", "POST", "PUT", "DELETE", "PATCH"}

	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			// Setup
			gin.SetMode(gin.TestMode)
			logger, buf := setupLoggerWithBuffer()
			global.Logger = logger
			defer func() { global.Logger = nil }()

			router := gin.New()
			router.Use(Logger())
			router.Handle(method, "/test", func(c *gin.Context) {
				c.JSON(200, gin.H{"message": "success"})
			})

			// Execute
			req := httptest.NewRequest(method, "/test", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Verify
			var logEntry map[string]interface{}
			err := json.Unmarshal(buf.Bytes(), &logEntry)
			assert.NoError(t, err)
			assert.Equal(t, method, logEntry["method"])
		})
	}
}

func TestLogger_LogsDifferentStatusCodes(t *testing.T) {
	testCases := []struct {
		name       string
		statusCode int
	}{
		{"Success", 200},
		{"Created", 201},
		{"BadRequest", 400},
		{"Unauthorized", 401},
		{"Forbidden", 403},
		{"NotFound", 404},
		{"InternalServerError", 500},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup
			gin.SetMode(gin.TestMode)
			logger, buf := setupLoggerWithBuffer()
			global.Logger = logger
			defer func() { global.Logger = nil }()

			router := gin.New()
			router.Use(Logger())
			router.GET("/test", func(c *gin.Context) {
				c.JSON(tc.statusCode, gin.H{"message": "test"})
			})

			// Execute
			req := httptest.NewRequest("GET", "/test", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Verify
			var logEntry map[string]interface{}
			err := json.Unmarshal(buf.Bytes(), &logEntry)
			assert.NoError(t, err)
			assert.Equal(t, float64(tc.statusCode), logEntry["status"])
		})
	}
}

func TestLogger_LogsClientIP(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	logger, buf := setupLoggerWithBuffer()
	global.Logger = logger
	defer func() { global.Logger = nil }()

	router := gin.New()
	router.Use(Logger())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "success"})
	})

	// Execute with specific IP
	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "10.0.0.1:54321"
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Verify
	var logEntry map[string]interface{}
	err := json.Unmarshal(buf.Bytes(), &logEntry)
	assert.NoError(t, err)
	assert.Contains(t, logEntry["client_ip"], "10.0.0.1")
}

func TestLogger_MeasuresLatency(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	logger, buf := setupLoggerWithBuffer()
	global.Logger = logger
	defer func() { global.Logger = nil }()

	router := gin.New()
	router.Use(Logger())
	router.GET("/test", func(c *gin.Context) {
		// Simulate some processing time
		// Note: In tests this will be very fast, but we can verify the field exists
		c.JSON(200, gin.H{"message": "success"})
	})

	// Execute
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Verify
	var logEntry map[string]interface{}
	err := json.Unmarshal(buf.Bytes(), &logEntry)
	assert.NoError(t, err)
	assert.NotEmpty(t, logEntry["latency"])
	// Latency should be a string with time unit (e.g., "1.234ms")
	latency, ok := logEntry["latency"].(string)
	assert.True(t, ok)
	assert.NotEmpty(t, latency)
}

func TestLogger_HandlesNilLogger(t *testing.T) {
	// Setup - no logger initialized
	gin.SetMode(gin.TestMode)
	global.Logger = nil

	router := gin.New()
	router.Use(Logger())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "success"})
	})

	// Execute - should not panic
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	// This should not panic even with nil logger
	assert.NotPanics(t, func() {
		router.ServeHTTP(w, req)
	})

	assert.Equal(t, 200, w.Code)
}

func TestLogger_LogsPathWithQueryParams(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	logger, buf := setupLoggerWithBuffer()
	global.Logger = logger
	defer func() { global.Logger = nil }()

	router := gin.New()
	router.Use(Logger())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "success"})
	})

	// Execute with query parameters
	req := httptest.NewRequest("GET", "/test?page=1&size=10", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Verify - path should not include query params
	var logEntry map[string]interface{}
	err := json.Unmarshal(buf.Bytes(), &logEntry)
	assert.NoError(t, err)
	assert.Equal(t, "/test", logEntry["path"])
}
