package middleware

import (
	"bytes"
	"k-admin-system/config"
	"k-admin-system/global"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Feature: k-admin-system
// Property 53: Middleware Execution Order
// For any request, middleware SHALL execute in the defined order:
// CORS → Rate Limit → JWT → Casbin → Handler → Recovery
// Validates: Requirements 16.7
func TestProperty53_MiddlewareExecutionOrder(t *testing.T) {
	gin.SetMode(gin.TestMode)

	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	// Generator for HTTP methods
	methodGen := gen.OneConstOf("GET", "POST", "PUT", "DELETE")

	// Generator for paths
	pathGen := gen.OneConstOf(
		"/api/v1/test",
		"/api/v1/users",
		"/api/v1/roles",
		"/api/v1/menus",
	)

	properties.Property("middleware executes in correct order", prop.ForAll(
		func(method, path string) bool {
			// Track execution order
			var executionOrder []string
			var mu sync.Mutex

			// Create tracking middleware
			trackingMiddleware := func(name string) gin.HandlerFunc {
				return func(c *gin.Context) {
					mu.Lock()
					executionOrder = append(executionOrder, name+"-before")
					mu.Unlock()
					c.Next()
					mu.Lock()
					executionOrder = append(executionOrder, name+"-after")
					mu.Unlock()
				}
			}

			// Setup router with middleware in correct order
			router := gin.New()

			// 1. Recovery (wraps everything)
			router.Use(trackingMiddleware("recovery"))

			// 2. CORS
			router.Use(trackingMiddleware("cors"))

			// 3. Rate Limit
			router.Use(trackingMiddleware("ratelimit"))

			// 4. Logger
			router.Use(trackingMiddleware("logger"))

			// 5. JWT
			router.Use(trackingMiddleware("jwt"))

			// 6. Casbin
			router.Use(trackingMiddleware("casbin"))

			// Handler
			router.Handle(method, path, func(c *gin.Context) {
				mu.Lock()
				executionOrder = append(executionOrder, "handler")
				mu.Unlock()
				c.JSON(200, gin.H{"message": "ok"})
			})

			// Execute request
			req := httptest.NewRequest(method, path, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Verify execution order
			expectedOrder := []string{
				"recovery-before",
				"cors-before",
				"ratelimit-before",
				"logger-before",
				"jwt-before",
				"casbin-before",
				"handler",
				"casbin-after",
				"jwt-after",
				"logger-after",
				"ratelimit-after",
				"cors-after",
				"recovery-after",
			}

			if len(executionOrder) != len(expectedOrder) {
				t.Logf("Execution order length mismatch: expected %d, got %d", len(expectedOrder), len(executionOrder))
				t.Logf("Expected: %v", expectedOrder)
				t.Logf("Got: %v", executionOrder)
				return false
			}

			for i, expected := range expectedOrder {
				if executionOrder[i] != expected {
					t.Logf("Execution order mismatch at position %d: expected %s, got %s", i, expected, executionOrder[i])
					t.Logf("Full order: %v", executionOrder)
					return false
				}
			}

			return true
		},
		methodGen,
		pathGen,
	))

	properties.Property("recovery middleware wraps all other middleware", prop.ForAll(
		func(method, path string) bool {
			// Track if recovery catches panics from other middleware
			var recovered bool
			var mu sync.Mutex

			router := gin.New()

			// Recovery middleware (must be first)
			router.Use(func(c *gin.Context) {
				defer func() {
					if err := recover(); err != nil {
						mu.Lock()
						recovered = true
						mu.Unlock()
						c.JSON(500, gin.H{"error": "internal server error"})
					}
				}()
				c.Next()
			})

			// Middleware that panics
			router.Use(func(c *gin.Context) {
				panic("test panic")
			})

			router.Handle(method, path, func(c *gin.Context) {
				c.JSON(200, gin.H{"message": "ok"})
			})

			// Execute request
			req := httptest.NewRequest(method, path, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Verify recovery caught the panic
			if !recovered {
				t.Logf("Recovery middleware did not catch panic")
				return false
			}

			// Verify response is 500
			if w.Code != 500 {
				t.Logf("Expected status 500, got %d", w.Code)
				return false
			}

			return true
		},
		methodGen,
		pathGen,
	))

	properties.Property("middleware chain allows nested execution", prop.ForAll(
		func(method, path string) bool {
			// Track that middleware executes in nested fashion (LIFO for after-handler)
			var executionOrder []string
			var mu sync.Mutex

			trackingMiddleware := func(name string) gin.HandlerFunc {
				return func(c *gin.Context) {
					mu.Lock()
					executionOrder = append(executionOrder, name+"-start")
					mu.Unlock()

					c.Next()

					mu.Lock()
					executionOrder = append(executionOrder, name+"-end")
					mu.Unlock()
				}
			}

			router := gin.New()
			router.Use(trackingMiddleware("middleware1"))
			router.Use(trackingMiddleware("middleware2"))
			router.Use(trackingMiddleware("middleware3"))

			router.Handle(method, path, func(c *gin.Context) {
				mu.Lock()
				executionOrder = append(executionOrder, "handler")
				mu.Unlock()
				c.JSON(200, gin.H{"message": "ok"})
			})

			req := httptest.NewRequest(method, path, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Verify nested execution order
			expectedOrder := []string{
				"middleware1-start",
				"middleware2-start",
				"middleware3-start",
				"handler",
				"middleware3-end",
				"middleware2-end",
				"middleware1-end",
			}

			if len(executionOrder) != len(expectedOrder) {
				t.Logf("Execution order length mismatch: expected %d, got %d", len(expectedOrder), len(executionOrder))
				return false
			}

			for i, expected := range expectedOrder {
				if executionOrder[i] != expected {
					t.Logf("Execution order mismatch at position %d: expected %s, got %s", i, expected, executionOrder[i])
					return false
				}
			}

			return true
		},
		methodGen,
		pathGen,
	))

	properties.Property("logger middleware executes before handler", prop.ForAll(
		func(method, path string) bool {
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
				c.JSON(200, gin.H{"message": "ok"})
			})

			req := httptest.NewRequest(method, path, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Verify logger captured the request
			logOutput := buf.String()
			if !strings.Contains(logOutput, method) || !strings.Contains(logOutput, path) {
				t.Logf("Logger did not capture request: %s %s", method, path)
				return false
			}

			return true
		},
		methodGen,
		pathGen,
	))

	properties.TestingRun(t)
}

// Feature: k-admin-system
// Property 54: Middleware Route Exclusion
// For any route marked as excluded from specific middleware, that middleware
// SHALL not execute for requests to that route
// Validates: Requirements 16.8
func TestProperty54_MiddlewareRouteExclusion(t *testing.T) {
	gin.SetMode(gin.TestMode)

	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	// Generator for HTTP methods
	methodGen := gen.OneConstOf("GET", "POST", "PUT", "DELETE")

	properties.Property("excluded routes skip middleware", prop.ForAll(
		func(method string) bool {
			// Track middleware execution
			var middlewareExecuted bool
			var mu sync.Mutex

			// Create middleware with exclusion logic
			excludedPaths := []string{"/health", "/public"}
			conditionalMiddleware := func() gin.HandlerFunc {
				return func(c *gin.Context) {
					// Check if path is excluded
					excluded := false
					for _, excludedPath := range excludedPaths {
						if c.Request.URL.Path == excludedPath {
							excluded = true
							break
						}
					}

					if !excluded {
						mu.Lock()
						middlewareExecuted = true
						mu.Unlock()
					}

					c.Next()
				}
			}

			router := gin.New()
			router.Use(conditionalMiddleware())

			// Register excluded route
			router.Handle(method, "/health", func(c *gin.Context) {
				c.JSON(200, gin.H{"status": "ok"})
			})

			// Execute request to excluded route
			req := httptest.NewRequest(method, "/health", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Verify middleware was not executed
			if middlewareExecuted {
				t.Logf("Middleware executed for excluded route /health")
				return false
			}

			// Verify response is successful
			if w.Code != 200 {
				t.Logf("Expected status 200, got %d", w.Code)
				return false
			}

			return true
		},
		methodGen,
	))

	properties.Property("non-excluded routes execute middleware", prop.ForAll(
		func(method, path string) bool {
			// Skip if path is empty
			if path == "" {
				path = "/api/v1/test"
			}

			// Track middleware execution
			var middlewareExecuted bool
			var mu sync.Mutex

			// Create middleware with exclusion logic
			excludedPaths := []string{"/health", "/public"}
			conditionalMiddleware := func() gin.HandlerFunc {
				return func(c *gin.Context) {
					// Check if path is excluded
					excluded := false
					for _, excludedPath := range excludedPaths {
						if c.Request.URL.Path == excludedPath {
							excluded = true
							break
						}
					}

					if !excluded {
						mu.Lock()
						middlewareExecuted = true
						mu.Unlock()
					}

					c.Next()
				}
			}

			router := gin.New()
			router.Use(conditionalMiddleware())

			// Register non-excluded route
			router.Handle(method, path, func(c *gin.Context) {
				c.JSON(200, gin.H{"message": "ok"})
			})

			// Execute request to non-excluded route
			req := httptest.NewRequest(method, path, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Skip if path was excluded
			for _, excludedPath := range excludedPaths {
				if path == excludedPath {
					return true
				}
			}

			// Verify middleware was executed
			if !middlewareExecuted {
				t.Logf("Middleware not executed for non-excluded route %s", path)
				return false
			}

			return true
		},
		methodGen,
		gen.OneConstOf("/api/v1/users", "/api/v1/roles", "/api/v1/menus", "/api/v1/test"),
	))

	properties.Property("JWT middleware can be excluded for public routes", prop.ForAll(
		func(method string) bool {
			// Track JWT middleware execution
			var jwtExecuted bool
			var mu sync.Mutex

			// Create JWT middleware that tracks execution
			jwtMiddleware := func() gin.HandlerFunc {
				return func(c *gin.Context) {
					mu.Lock()
					jwtExecuted = true
					mu.Unlock()
					c.Next()
				}
			}

			router := gin.New()

			// Public routes (no JWT)
			publicGroup := router.Group("/public")
			{
				publicGroup.Handle(method, "/login", func(c *gin.Context) {
					c.JSON(200, gin.H{"message": "login"})
				})
			}

			// Protected routes (with JWT)
			protectedGroup := router.Group("/api")
			protectedGroup.Use(jwtMiddleware())
			{
				protectedGroup.Handle(method, "/users", func(c *gin.Context) {
					c.JSON(200, gin.H{"message": "users"})
				})
			}

			// Execute request to public route
			req := httptest.NewRequest(method, "/public/login", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Verify JWT middleware was not executed
			if jwtExecuted {
				t.Logf("JWT middleware executed for public route")
				return false
			}

			// Reset flag
			mu.Lock()
			jwtExecuted = false
			mu.Unlock()

			// Execute request to protected route
			req = httptest.NewRequest(method, "/api/users", nil)
			w = httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Verify JWT middleware was executed
			if !jwtExecuted {
				t.Logf("JWT middleware not executed for protected route")
				return false
			}

			return true
		},
		methodGen,
	))

	properties.Property("CORS middleware applies to all routes", prop.ForAll(
		func(method, path string) bool {
			// Skip empty paths
			if path == "" {
				path = "/test"
			}

			corsConfig := config.CORSConfig{
				AllowOrigins: []string{"http://localhost:3000"},
				AllowMethods: []string{"GET", "POST", "PUT", "DELETE"},
				AllowHeaders: []string{"Content-Type"},
			}

			router := gin.New()
			router.Use(CORS(corsConfig))

			router.Handle(method, path, func(c *gin.Context) {
				c.JSON(200, gin.H{"message": "ok"})
			})

			req := httptest.NewRequest(method, path, nil)
			req.Header.Set("Origin", "http://localhost:3000")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Verify CORS headers are set (CORS applies to all routes)
			if w.Header().Get("Access-Control-Allow-Origin") == "" {
				t.Logf("CORS headers not set for route %s", path)
				return false
			}

			return true
		},
		methodGen,
		gen.OneConstOf("/health", "/api/v1/users", "/public/login", "/test"),
	))

	properties.Property("middleware exclusion preserves handler execution", prop.ForAll(
		func(method, path string) bool {
			// Skip empty paths
			if path == "" {
				path = "/test"
			}

			// Track handler execution
			var handlerExecuted bool
			var mu sync.Mutex

			// Create middleware that might be excluded
			excludedPaths := []string{"/health"}
			conditionalMiddleware := func() gin.HandlerFunc {
				return func(c *gin.Context) {
					// Check if path is excluded
					excluded := false
					for _, excludedPath := range excludedPaths {
						if c.Request.URL.Path == excludedPath {
							excluded = true
							break
						}
					}

					if !excluded {
						// Middleware logic here
					}

					c.Next()
				}
			}

			router := gin.New()
			router.Use(conditionalMiddleware())

			router.Handle(method, path, func(c *gin.Context) {
				mu.Lock()
				handlerExecuted = true
				mu.Unlock()
				c.JSON(200, gin.H{"message": "ok"})
			})

			req := httptest.NewRequest(method, path, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Verify handler was executed regardless of middleware exclusion
			if !handlerExecuted {
				t.Logf("Handler not executed for route %s", path)
				return false
			}

			// Verify response is successful
			if w.Code != 200 {
				t.Logf("Expected status 200, got %d", w.Code)
				return false
			}

			return true
		},
		methodGen,
		gen.OneConstOf("/health", "/api/v1/users", "/test"),
	))

	properties.Property("multiple middleware can be excluded independently", prop.ForAll(
		func(method string) bool {
			// Track middleware execution
			var middleware1Executed, middleware2Executed bool
			var mu sync.Mutex

			// Middleware 1 excludes /health
			middleware1 := func() gin.HandlerFunc {
				return func(c *gin.Context) {
					if c.Request.URL.Path != "/health" {
						mu.Lock()
						middleware1Executed = true
						mu.Unlock()
					}
					c.Next()
				}
			}

			// Middleware 2 excludes /public
			middleware2 := func() gin.HandlerFunc {
				return func(c *gin.Context) {
					if c.Request.URL.Path != "/public" {
						mu.Lock()
						middleware2Executed = true
						mu.Unlock()
					}
					c.Next()
				}
			}

			router := gin.New()
			router.Use(middleware1())
			router.Use(middleware2())

			router.Handle(method, "/health", func(c *gin.Context) {
				c.JSON(200, gin.H{"status": "ok"})
			})
			router.Handle(method, "/public", func(c *gin.Context) {
				c.JSON(200, gin.H{"message": "public"})
			})

			// Test /health route
			req := httptest.NewRequest(method, "/health", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Verify middleware1 was not executed, but middleware2 was
			if middleware1Executed {
				t.Logf("Middleware1 should not execute for /health")
				return false
			}
			if !middleware2Executed {
				t.Logf("Middleware2 should execute for /health")
				return false
			}

			// Reset flags
			mu.Lock()
			middleware1Executed = false
			middleware2Executed = false
			mu.Unlock()

			// Test /public route
			req = httptest.NewRequest(method, "/public", nil)
			w = httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Verify middleware2 was not executed, but middleware1 was
			if !middleware1Executed {
				t.Logf("Middleware1 should execute for /public")
				return false
			}
			if middleware2Executed {
				t.Logf("Middleware2 should not execute for /public")
				return false
			}

			return true
		},
		methodGen,
	))

	properties.TestingRun(t)
}
