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
	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Feature: k-admin-system
// Property 36: Panic Recovery Without Crash
// For any request handler that panics, the panic recovery middleware SHALL catch the panic, log it, return 500 error, and keep the server running
// Validates: Requirements 11.7, 16.6

func TestProperty_PanicRecoveryWithoutCrash(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	// Generator for panic messages
	panicMessageGen := gen.OneConstOf(
		"runtime error: invalid memory address or nil pointer dereference",
		"runtime error: index out of range",
		"runtime error: slice bounds out of range",
		"custom panic: something went wrong",
		"database connection failed",
		"unexpected error occurred",
	)

	// Generator for HTTP methods
	methodGen := gen.OneConstOf("GET", "POST", "PUT", "DELETE", "PATCH")

	// Generator for paths
	pathGen := gen.OneConstOf(
		"/api/v1/users",
		"/api/v1/roles",
		"/api/v1/menus",
		"/api/v1/test",
		"/panic",
	)

	properties.Property("Panic recovery catches panic, logs it, returns 500, and keeps server running", prop.ForAll(
		func(panicMessage, method, path string) bool {
			// Setup logger to capture logs
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
				zapcore.ErrorLevel, // Only capture error level logs
			)
			logger := zap.New(core)
			global.Logger = logger
			defer func() { global.Logger = nil }()

			// Setup router with recovery middleware
			router := gin.New()
			router.Use(Recovery())
			router.Handle(method, path, func(c *gin.Context) {
				// Trigger panic
				panic(panicMessage)
			})

			// Execute request
			req := httptest.NewRequest(method, path, nil)
			w := httptest.NewRecorder()

			// This should not crash the server
			router.ServeHTTP(w, req)

			// Verify response status is 200 (unified response format)
			if w.Code != http.StatusOK {
				t.Logf("Expected status 200, got %d", w.Code)
				return false
			}

			// Verify response body contains error code 500
			var response common.Response
			err := json.Unmarshal(w.Body.Bytes(), &response)
			if err != nil {
				t.Logf("Failed to parse response: %v", err)
				return false
			}

			if response.Code != 500 {
				t.Logf("Expected error code 500, got %d", response.Code)
				return false
			}

			if response.Msg == "" {
				t.Logf("Expected error message, got empty string")
				return false
			}

			// Verify error message contains panic information
			if !strings.Contains(response.Msg, "Internal server error") {
				t.Logf("Error message should contain 'Internal server error', got: %s", response.Msg)
				return false
			}

			// Verify panic was logged
			logOutput := buf.String()
			if logOutput == "" {
				t.Logf("Expected panic to be logged, but log is empty")
				return false
			}

			var logEntry map[string]any
			err = json.Unmarshal(buf.Bytes(), &logEntry)
			if err != nil {
				t.Logf("Failed to parse log entry: %v", err)
				return false
			}

			// Verify log contains "Panic recovered" message
			msg, hasMsg := logEntry["msg"]
			if !hasMsg || msg != "Panic recovered" {
				t.Logf("Expected log message 'Panic recovered', got: %v", msg)
				return false
			}

			// Verify log contains error field
			if _, hasError := logEntry["error"]; !hasError {
				t.Logf("Log entry missing 'error' field")
				return false
			}

			// Verify log contains path field
			loggedPath, hasPath := logEntry["path"]
			if !hasPath || loggedPath != path {
				t.Logf("Expected path %s in log, got: %v", path, loggedPath)
				return false
			}

			// Verify log contains method field
			loggedMethod, hasMethod := logEntry["method"]
			if !hasMethod || loggedMethod != method {
				t.Logf("Expected method %s in log, got: %v", method, loggedMethod)
				return false
			}

			// Server should still be running (we can make another request)
			req2 := httptest.NewRequest("GET", "/health", nil)
			w2 := httptest.NewRecorder()
			router.GET("/health", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"status": "ok"})
			})
			router.ServeHTTP(w2, req2)

			if w2.Code != http.StatusOK {
				t.Logf("Server should still be running after panic, but health check failed")
				return false
			}

			return true
		},
		panicMessageGen,
		methodGen,
		pathGen,
	))

	properties.TestingRun(t)
}

// Feature: k-admin-system
// Property 37: Error Logging with Stack Traces
// For any error that occurs in the backend, the error SHALL be logged with timestamp, error message, and full stack trace
// Validates: Requirements 11.8

func TestProperty_ErrorLoggingWithStackTraces(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	// Generator for different types of panics
	panicTypeGen := gen.OneConstOf(
		"nil pointer dereference",
		"index out of range",
		"division by zero",
		"type assertion failed",
		"custom error",
	)

	// Generator for paths
	pathGen := gen.OneConstOf(
		"/api/v1/users",
		"/api/v1/roles",
		"/api/v1/menus",
		"/api/v1/error",
	)

	properties.Property("Panic logs contain stack trace information", prop.ForAll(
		func(panicType, path string) bool {
			// Setup logger to capture logs
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

			// Setup router with recovery middleware
			router := gin.New()
			router.Use(Recovery())
			router.GET(path, func(c *gin.Context) {
				// Trigger panic with specific type
				panic(panicType)
			})

			// Execute request
			req := httptest.NewRequest("GET", path, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Verify panic was logged
			logOutput := buf.String()
			if logOutput == "" {
				t.Logf("Expected panic to be logged, but log is empty")
				return false
			}

			var logEntry map[string]any
			err := json.Unmarshal(buf.Bytes(), &logEntry)
			if err != nil {
				t.Logf("Failed to parse log entry: %v", err)
				return false
			}

			// Verify log contains timestamp
			timestamp, hasTimestamp := logEntry["timestamp"]
			if !hasTimestamp || timestamp == "" {
				t.Logf("Log entry missing timestamp")
				return false
			}

			// Verify log contains error message
			errorField, hasError := logEntry["error"]
			if !hasError || errorField == "" {
				t.Logf("Log entry missing error field")
				return false
			}

			// Verify log contains stack trace
			stack, hasStack := logEntry["stack"]
			if !hasStack {
				t.Logf("Log entry missing stack trace")
				return false
			}

			stackStr, ok := stack.(string)
			if !ok || stackStr == "" {
				t.Logf("Stack trace is empty or not a string")
				return false
			}

			// Verify stack trace contains goroutine information
			if !strings.Contains(stackStr, "goroutine") {
				t.Logf("Stack trace should contain 'goroutine' keyword")
				return false
			}

			// Verify stack trace contains file and line information
			if !strings.Contains(stackStr, ".go:") {
				t.Logf("Stack trace should contain file and line information (.go:)")
				return false
			}

			// Verify log contains path and method
			if _, hasPath := logEntry["path"]; !hasPath {
				t.Logf("Log entry missing path field")
				return false
			}

			if _, hasMethod := logEntry["method"]; !hasMethod {
				t.Logf("Log entry missing method field")
				return false
			}

			return true
		},
		panicTypeGen,
		pathGen,
	))

	properties.TestingRun(t)
}

// Property: Multiple panics don't crash server
// For any sequence of requests that panic, the server SHALL handle all of them and remain operational
func TestProperty_MultiplePanicsServerStability(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 50 // Fewer iterations since we're testing multiple requests

	properties := gopter.NewProperties(parameters)

	// Generator for number of panic requests
	numRequestsGen := gen.IntRange(1, 10)

	properties.Property("Server handles multiple panics without crashing", prop.ForAll(
		func(numRequests int) bool {
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
			router.GET("/health", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"status": "ok"})
			})

			// Execute multiple panic requests
			for i := 0; i < numRequests; i++ {
				req := httptest.NewRequest("GET", "/panic", nil)
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				// Each request should return 500 error
				var response common.Response
				err := json.Unmarshal(w.Body.Bytes(), &response)
				if err != nil || response.Code != 500 {
					t.Logf("Request %d failed to return proper error response", i)
					return false
				}
			}

			// Verify server is still operational
			req := httptest.NewRequest("GET", "/health", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				t.Logf("Server not operational after %d panics", numRequests)
				return false
			}

			return true
		},
		numRequestsGen,
	))

	properties.TestingRun(t)
}
