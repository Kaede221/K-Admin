package middleware

import (
	"bytes"
	"encoding/json"
	"fmt"
	"k-admin-system/global"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Feature: k-admin-system
// Property 41: HTTP Request Logging Completeness
// For any HTTP request, the log entry SHALL contain timestamp, method, path, status code, latency, and client IP
// Validates: Requirements 13.4

func TestProperty_HTTPRequestLoggingCompleteness(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	// Generator for HTTP methods
	methodGen := gen.OneConstOf("GET", "POST", "PUT", "DELETE", "PATCH", "HEAD", "OPTIONS")

	// Generator for paths
	pathGen := gen.OneConstOf(
		"/api/v1/users",
		"/api/v1/roles",
		"/api/v1/menus",
		"/api/v1/login",
		"/api/v1/users/123",
		"/health",
		"/",
	)

	// Generator for status codes
	statusGen := gen.OneConstOf(200, 201, 204, 400, 401, 403, 404, 500, 502, 503)

	// Generator for client IPs
	ipGen := gen.OneConstOf(
		"192.168.1.1",
		"10.0.0.1",
		"172.16.0.1",
		"127.0.0.1",
		"203.0.113.1",
	)

	properties.Property("HTTP request logging contains all required fields", prop.ForAll(
		func(method, path string, statusCode int, clientIP string) bool {
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
				zapcore.InfoLevel,
			)
			logger := zap.New(core)
			global.Logger = logger
			defer func() { global.Logger = nil }()

			router := gin.New()
			router.Use(Logger())
			router.Handle(method, path, func(c *gin.Context) {
				c.JSON(statusCode, gin.H{"message": "test"})
			})

			// Execute
			req := httptest.NewRequest(method, path, nil)
			req.RemoteAddr = fmt.Sprintf("%s:12345", clientIP)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Verify log entry contains all required fields
			var logEntry map[string]interface{}
			err := json.Unmarshal(buf.Bytes(), &logEntry)
			if err != nil {
				t.Logf("Failed to parse log entry: %v", err)
				return false
			}

			// Check for timestamp field
			timestamp, hasTimestamp := logEntry["timestamp"]
			if !hasTimestamp || timestamp == "" {
				t.Logf("Missing or empty timestamp field")
				return false
			}

			// Check for method field
			loggedMethod, hasMethod := logEntry["method"]
			if !hasMethod || loggedMethod != method {
				t.Logf("Missing or incorrect method field: expected %s, got %v", method, loggedMethod)
				return false
			}

			// Check for path field
			loggedPath, hasPath := logEntry["path"]
			if !hasPath || loggedPath != path {
				t.Logf("Missing or incorrect path field: expected %s, got %v", path, loggedPath)
				return false
			}

			// Check for status field
			loggedStatus, hasStatus := logEntry["status"]
			if !hasStatus {
				t.Logf("Missing status field")
				return false
			}
			// JSON unmarshals numbers as float64
			statusFloat, ok := loggedStatus.(float64)
			if !ok || int(statusFloat) != statusCode {
				t.Logf("Incorrect status field: expected %d, got %v", statusCode, loggedStatus)
				return false
			}

			// Check for latency field
			latency, hasLatency := logEntry["latency"]
			if !hasLatency || latency == "" {
				t.Logf("Missing or empty latency field")
				return false
			}

			// Check for client_ip field
			loggedIP, hasIP := logEntry["client_ip"]
			if !hasIP || loggedIP == "" {
				t.Logf("Missing or empty client_ip field")
				return false
			}
			// Client IP should be a string
			if _, ok := loggedIP.(string); !ok {
				t.Logf("client_ip is not a string: %v", loggedIP)
				return false
			}

			return true
		},
		methodGen,
		pathGen,
		statusGen,
		ipGen,
	))

	properties.TestingRun(t)
}

// Property: Log entry structure consistency
// For any HTTP request, the log entry SHALL have consistent structure with expected field types
func TestProperty_LogEntryStructureConsistency(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	// Generator for various request scenarios
	methodGen := gen.OneConstOf("GET", "POST", "PUT", "DELETE")
	pathGen := gen.RegexMatch(`/api/v[0-9]/[a-z]+(/[0-9]+)?`)
	statusGen := gen.IntRange(200, 599)

	properties.Property("Log entry has consistent structure and types", prop.ForAll(
		func(method, path string, statusCode int) bool {
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
				zapcore.InfoLevel,
			)
			logger := zap.New(core)
			global.Logger = logger
			defer func() { global.Logger = nil }()

			router := gin.New()
			router.Use(Logger())
			router.Handle(method, path, func(c *gin.Context) {
				c.JSON(statusCode, gin.H{"message": "test"})
			})

			// Execute
			req := httptest.NewRequest(method, path, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Verify structure
			var logEntry map[string]interface{}
			err := json.Unmarshal(buf.Bytes(), &logEntry)
			if err != nil {
				return false
			}

			// Verify field types
			if _, ok := logEntry["timestamp"].(string); !ok {
				return false
			}
			if _, ok := logEntry["level"].(string); !ok {
				return false
			}
			if _, ok := logEntry["msg"].(string); !ok {
				return false
			}
			if _, ok := logEntry["method"].(string); !ok {
				return false
			}
			if _, ok := logEntry["path"].(string); !ok {
				return false
			}
			if _, ok := logEntry["status"].(float64); !ok {
				return false
			}
			if _, ok := logEntry["latency"].(string); !ok {
				return false
			}
			if _, ok := logEntry["client_ip"].(string); !ok {
				return false
			}

			return true
		},
		methodGen,
		pathGen,
		statusGen,
	))

	properties.TestingRun(t)
}
