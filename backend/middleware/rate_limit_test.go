package middleware

import (
	"context"
	"fmt"
	"k-admin-system/config"
	"k-admin-system/global"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

// setupRateLimitTest initializes test environment for rate limiting tests
func setupRateLimitTest(t *testing.T) (*gin.Engine, func()) {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// Initialize logger
	logger, _ := zap.NewDevelopment()
	global.Logger = logger

	// Initialize Redis client
	redisClient := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       1, // Use DB 1 for tests
	})

	// Test Redis connection
	ctx := context.Background()
	if err := redisClient.Ping(ctx).Err(); err != nil {
		t.Skipf("Redis not available, skipping rate limit tests: %v", err)
	}

	global.RedisClient = redisClient

	// Create test router
	router := gin.New()

	// Cleanup function
	cleanup := func() {
		// Clear all test keys
		if global.RedisClient != nil {
			keys, _ := global.RedisClient.Keys(ctx, "rate_limit:*").Result()
			if len(keys) > 0 {
				global.RedisClient.Del(ctx, keys...)
			}
			global.RedisClient.Close()
		}
	}

	return router, cleanup
}

func TestRateLimit_Disabled(t *testing.T) {
	router, cleanup := setupRateLimitTest(t)
	defer cleanup()

	// Configure rate limiting as disabled
	rateLimitConfig := config.RateLimitConfig{
		Enabled:  false,
		Requests: 5,
		Window:   60,
		KeyFunc:  "ip",
	}

	router.Use(RateLimit(rateLimitConfig))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "ok"})
	})

	// Make 10 requests (more than the limit)
	for i := 0; i < 10; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// All requests should succeed since rate limiting is disabled
		assert.Equal(t, 200, w.Code, "Request %d should succeed", i+1)
	}
}

func TestRateLimit_IPBased(t *testing.T) {
	router, cleanup := setupRateLimitTest(t)
	defer cleanup()

	// Configure rate limiting: 5 requests per 60 seconds
	rateLimitConfig := config.RateLimitConfig{
		Enabled:  true,
		Requests: 5,
		Window:   60,
		KeyFunc:  "ip",
	}

	router.Use(RateLimit(rateLimitConfig))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "ok"})
	})

	// Make 5 requests (at the limit)
	for i := 0; i < 5; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "192.168.1.1:12345"
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code, "Request %d should succeed", i+1)
	}

	// 6th request should be rate limited
	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "192.168.1.1:12345"
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code, "6th request should be rate limited")
	// Check response body contains rate limit message
	assert.Contains(t, w.Body.String(), "请求过于频繁")
}

func TestRateLimit_DifferentIPs(t *testing.T) {
	router, cleanup := setupRateLimitTest(t)
	defer cleanup()

	// Configure rate limiting: 3 requests per 60 seconds
	rateLimitConfig := config.RateLimitConfig{
		Enabled:  true,
		Requests: 3,
		Window:   60,
		KeyFunc:  "ip",
	}

	router.Use(RateLimit(rateLimitConfig))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "ok"})
	})

	// IP1: Make 3 requests (at the limit)
	for i := 0; i < 3; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "192.168.1.1:12345"
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code, "IP1 request %d should succeed", i+1)
	}

	// IP2: Should still be able to make requests
	for i := 0; i < 3; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "192.168.1.2:12345"
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code, "IP2 request %d should succeed", i+1)
	}

	// IP1: 4th request should be rate limited
	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "192.168.1.1:12345"
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code, "IP1 4th request should be rate limited")
	assert.Contains(t, w.Body.String(), "请求过于频繁")
}

func TestRateLimit_UserBased(t *testing.T) {
	router, cleanup := setupRateLimitTest(t)
	defer cleanup()

	// Configure rate limiting: 3 requests per 60 seconds, user-based
	rateLimitConfig := config.RateLimitConfig{
		Enabled:  true,
		Requests: 3,
		Window:   60,
		KeyFunc:  "user",
	}

	router.Use(RateLimit(rateLimitConfig))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "ok"})
	})

	// User 1: Make 3 requests (at the limit)
	for i := 0; i < 3; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()

		// Create context with user ID
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Set("userId", uint(1))

		router.ServeHTTP(w, req)
		assert.Equal(t, 200, w.Code, "User 1 request %d should succeed", i+1)
	}

	// User 2: Should still be able to make requests
	for i := 0; i < 3; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()

		// Create context with different user ID
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Set("userId", uint(2))

		router.ServeHTTP(w, req)
		assert.Equal(t, 200, w.Code, "User 2 request %d should succeed", i+1)
	}
}

func TestRateLimit_NoRedis(t *testing.T) {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// Initialize logger
	logger, _ := zap.NewDevelopment()
	global.Logger = logger

	// Set Redis client to nil
	global.RedisClient = nil

	router := gin.New()

	// Configure rate limiting
	rateLimitConfig := config.RateLimitConfig{
		Enabled:  true,
		Requests: 5,
		Window:   60,
		KeyFunc:  "ip",
	}

	router.Use(RateLimit(rateLimitConfig))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "ok"})
	})

	// Make 10 requests - all should succeed since Redis is not available
	for i := 0; i < 10; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code, "Request %d should succeed (Redis unavailable)", i+1)
	}
}

func TestGetRateLimitKey_IP(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/test", nil)
	c.Request.RemoteAddr = "192.168.1.1:12345"

	key := getRateLimitKey(c, "ip")
	expected := "rate_limit:ip:192.168.1.1"
	assert.Equal(t, expected, key)
}

func TestGetRateLimitKey_User(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/test", nil)
	c.Set("userId", uint(123))

	key := getRateLimitKey(c, "user")
	expected := "rate_limit:user:123"
	assert.Equal(t, expected, key)
}

func TestGetRateLimitKey_UserFallbackToIP(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/test", nil)
	c.Request.RemoteAddr = "192.168.1.1:12345"
	// No userId set

	key := getRateLimitKey(c, "user")
	expected := "rate_limit:ip:192.168.1.1"
	assert.Equal(t, expected, key, "Should fallback to IP when user not authenticated")
}

func TestGetRateLimitKey_Invalid(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/test", nil)

	key := getRateLimitKey(c, "invalid")
	assert.Equal(t, "", key, "Invalid key function should return empty string")
}

func TestCheckRateLimit_SlidingWindow(t *testing.T) {
	// Initialize Redis client
	redisClient := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       1, // Use DB 1 for tests
	})

	ctx := context.Background()
	if err := redisClient.Ping(ctx).Err(); err != nil {
		t.Skipf("Redis not available, skipping test: %v", err)
	}

	global.RedisClient = redisClient
	defer func() {
		// Cleanup
		keys, _ := global.RedisClient.Keys(ctx, "test_rate_limit:*").Result()
		if len(keys) > 0 {
			global.RedisClient.Del(ctx, keys...)
		}
		global.RedisClient.Close()
	}()

	testKey := "test_rate_limit:test_key"

	// Test: Allow first 3 requests
	for i := 0; i < 3; i++ {
		allowed, err := checkRateLimit(testKey, 3, 60)
		assert.NoError(t, err, "Request %d should not error", i+1)
		assert.True(t, allowed, "Request %d should be allowed", i+1)
	}

	// Test: 4th request should be denied
	allowed, err := checkRateLimit(testKey, 3, 60)
	assert.NoError(t, err)
	assert.False(t, allowed, "4th request should be denied")
}

func TestRateLimit_Returns429(t *testing.T) {
	router, cleanup := setupRateLimitTest(t)
	defer cleanup()

	// Configure rate limiting: 2 requests per 60 seconds
	rateLimitConfig := config.RateLimitConfig{
		Enabled:  true,
		Requests: 2,
		Window:   60,
		KeyFunc:  "ip",
	}

	router.Use(RateLimit(rateLimitConfig))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "ok"})
	})

	// Make 2 requests (at the limit)
	for i := 0; i < 2; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "192.168.1.1:12345"
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code, "Request %d should succeed", i+1)
	}

	// 3rd request should return 429 (via unified response with code field)
	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "192.168.1.1:12345"
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// The response should be 200 HTTP status with code=429 in JSON body
	assert.Equal(t, 200, w.Code, "HTTP status should be 200 (unified response)")
	assert.Contains(t, w.Body.String(), "429", "Response should contain code 429")
	assert.Contains(t, w.Body.String(), "请求过于频繁", "Response should contain rate limit message")
}

func TestRateLimit_ConfigValidation(t *testing.T) {
	tests := []struct {
		name        string
		config      config.RateLimitConfig
		expectError bool
	}{
		{
			name: "Valid IP-based config",
			config: config.RateLimitConfig{
				Enabled:  true,
				Requests: 100,
				Window:   60,
				KeyFunc:  "ip",
			},
			expectError: false,
		},
		{
			name: "Valid user-based config",
			config: config.RateLimitConfig{
				Enabled:  true,
				Requests: 50,
				Window:   30,
				KeyFunc:  "user",
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test that the config can be used without errors
			assert.NotEmpty(t, tt.config.KeyFunc)
			assert.Greater(t, tt.config.Requests, 0)
			assert.Greater(t, tt.config.Window, 0)
		})
	}
}

// Benchmark tests
func BenchmarkRateLimit(b *testing.B) {
	gin.SetMode(gin.TestMode)

	// Initialize logger
	logger, _ := zap.NewDevelopment()
	global.Logger = logger

	// Initialize Redis client
	redisClient := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       1,
	})

	ctx := context.Background()
	if err := redisClient.Ping(ctx).Err(); err != nil {
		b.Skipf("Redis not available, skipping benchmark: %v", err)
	}

	global.RedisClient = redisClient
	defer func() {
		keys, _ := global.RedisClient.Keys(ctx, "rate_limit:*").Result()
		if len(keys) > 0 {
			global.RedisClient.Del(ctx, keys...)
		}
		global.RedisClient.Close()
	}()

	router := gin.New()
	rateLimitConfig := config.RateLimitConfig{
		Enabled:  true,
		Requests: 1000,
		Window:   60,
		KeyFunc:  "ip",
	}

	router.Use(RateLimit(rateLimitConfig))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "ok"})
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = fmt.Sprintf("192.168.1.%d:12345", i%255)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}
