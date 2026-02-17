package middleware

import (
	"context"
	"fmt"
	"k-admin-system/config"
	"k-admin-system/global"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
	"github.com/redis/go-redis/v9"
)

// setupRateLimitTestEnvironment initializes Redis for rate limit testing
func setupRateLimitTestEnvironment(t *testing.T) func() {
	gin.SetMode(gin.TestMode)

	// Initialize Redis client for testing
	global.RedisClient = redis.NewClient(&redis.Options{
		Addr:     "127.0.0.1:6379", // Use IPv4 explicitly
		Password: "",
		DB:       15, // Use DB 15 for testing to avoid conflicts
	})

	// Test Redis connection
	ctx := context.Background()
	if err := global.RedisClient.Ping(ctx).Err(); err != nil {
		t.Skipf("Redis not available for testing: %v", err)
	}

	// Cleanup function
	return func() {
		// Clean up all test keys
		global.RedisClient.FlushDB(ctx)
		global.RedisClient.Close()
		global.RedisClient = nil
	}
}

// Feature: k-admin-system
// Property 52: Rate Limiting Enforcement
// For any client exceeding the configured request rate limit, subsequent requests
// SHALL be rejected with 429 Too Many Requests status
// Validates: Requirements 16.4
func TestProperty52_RateLimitingEnforcement(t *testing.T) {
	cleanup := setupRateLimitTestEnvironment(t)
	defer cleanup()

	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	properties.Property("requests within limit are allowed", prop.ForAll(
		func(requestCount uint8) bool {
			// Limit request count to 1-10 to keep test fast
			count := int(requestCount%10 + 1)
			maxRequests := count + 5 // Set limit higher than request count

			rateLimitConfig := config.RateLimitConfig{
				Enabled:  true,
				Requests: maxRequests,
				Window:   60,
				KeyFunc:  "ip",
			}

			router := gin.New()
			router.Use(RateLimit(rateLimitConfig))
			router.GET("/test", func(c *gin.Context) {
				c.JSON(200, gin.H{"message": "ok"})
			})

			// Make requests within the limit
			for i := 0; i < count; i++ {
				req := httptest.NewRequest("GET", "/test", nil)
				req.RemoteAddr = fmt.Sprintf("192.168.1.%d:12345", requestCount%250+1)
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				if w.Code != 200 {
					t.Logf("Request %d/%d should succeed, got status %d", i+1, count, w.Code)
					return false
				}
			}

			return true
		},
		gen.UInt8(),
	))

	properties.Property("requests exceeding limit are rejected with 429", prop.ForAll(
		func(maxRequests uint8) bool {
			// Limit to 1-20 requests to keep test fast
			limit := int(maxRequests%20 + 1)

			rateLimitConfig := config.RateLimitConfig{
				Enabled:  true,
				Requests: limit,
				Window:   60,
				KeyFunc:  "ip",
			}

			router := gin.New()
			router.Use(RateLimit(rateLimitConfig))
			router.GET("/test", func(c *gin.Context) {
				c.JSON(200, gin.H{"message": "ok"})
			})

			// Use unique IP for each test to avoid interference
			testIP := fmt.Sprintf("10.0.%d.%d:12345", int(maxRequests)/256, int(maxRequests)%256)

			// Make requests up to the limit - all should succeed
			for i := 0; i < limit; i++ {
				req := httptest.NewRequest("GET", "/test", nil)
				req.RemoteAddr = testIP
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				if w.Code != 200 {
					t.Logf("Request %d/%d within limit should succeed, got status %d", i+1, limit, w.Code)
					return false
				}
			}

			// Next request should be rate limited
			req := httptest.NewRequest("GET", "/test", nil)
			req.RemoteAddr = testIP
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != 429 {
				t.Logf("Request exceeding limit should return 429, got %d", w.Code)
				return false
			}

			return true
		},
		gen.UInt8(),
	))

	properties.Property("different IPs have independent rate limits", prop.ForAll(
		func(ip1Octet, ip2Octet uint8) bool {
			// Ensure different IPs
			if ip1Octet == ip2Octet {
				ip2Octet = ip1Octet + 1
			}

			rateLimitConfig := config.RateLimitConfig{
				Enabled:  true,
				Requests: 5,
				Window:   60,
				KeyFunc:  "ip",
			}

			router := gin.New()
			router.Use(RateLimit(rateLimitConfig))
			router.GET("/test", func(c *gin.Context) {
				c.JSON(200, gin.H{"message": "ok"})
			})

			ip1 := fmt.Sprintf("192.168.1.%d:12345", ip1Octet)
			ip2 := fmt.Sprintf("192.168.1.%d:12345", ip2Octet)

			// Exhaust limit for IP1
			for i := 0; i < 5; i++ {
				req := httptest.NewRequest("GET", "/test", nil)
				req.RemoteAddr = ip1
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				if w.Code != 200 {
					t.Logf("IP1 request %d should succeed", i+1)
					return false
				}
			}

			// IP1 should now be rate limited
			req1 := httptest.NewRequest("GET", "/test", nil)
			req1.RemoteAddr = ip1
			w1 := httptest.NewRecorder()
			router.ServeHTTP(w1, req1)

			if w1.Code != 429 {
				t.Logf("IP1 should be rate limited, got status %d", w1.Code)
				return false
			}

			// IP2 should still be allowed
			req2 := httptest.NewRequest("GET", "/test", nil)
			req2.RemoteAddr = ip2
			w2 := httptest.NewRecorder()
			router.ServeHTTP(w2, req2)

			if w2.Code != 200 {
				t.Logf("IP2 should not be rate limited, got status %d", w2.Code)
				return false
			}

			return true
		},
		gen.UInt8(),
		gen.UInt8(),
	))

	properties.Property("disabled rate limiting allows all requests", prop.ForAll(
		func(requestCount uint8) bool {
			// Make 10-30 requests
			count := int(requestCount%20 + 10)

			rateLimitConfig := config.RateLimitConfig{
				Enabled:  false, // Disabled
				Requests: 5,     // Low limit, but should be ignored
				Window:   60,
				KeyFunc:  "ip",
			}

			router := gin.New()
			router.Use(RateLimit(rateLimitConfig))
			router.GET("/test", func(c *gin.Context) {
				c.JSON(200, gin.H{"message": "ok"})
			})

			testIP := fmt.Sprintf("172.16.%d.%d:12345", int(requestCount)/256, int(requestCount)%256)

			// All requests should succeed even though we exceed the limit
			for i := 0; i < count; i++ {
				req := httptest.NewRequest("GET", "/test", nil)
				req.RemoteAddr = testIP
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				if w.Code != 200 {
					t.Logf("Request %d should succeed when rate limiting disabled, got %d", i+1, w.Code)
					return false
				}
			}

			return true
		},
		gen.UInt8(),
	))

	properties.Property("sliding window allows requests after time passes", prop.ForAll(
		func(windowSeconds uint8) bool {
			// Use small window (1-5 seconds) for faster testing
			window := int(windowSeconds%5 + 1)

			rateLimitConfig := config.RateLimitConfig{
				Enabled:  true,
				Requests: 2,
				Window:   window,
				KeyFunc:  "ip",
			}

			router := gin.New()
			router.Use(RateLimit(rateLimitConfig))
			router.GET("/test", func(c *gin.Context) {
				c.JSON(200, gin.H{"message": "ok"})
			})

			testIP := fmt.Sprintf("10.10.%d.%d:12345", int(windowSeconds)/256, int(windowSeconds)%256)

			// Make 2 requests (at limit)
			for i := 0; i < 2; i++ {
				req := httptest.NewRequest("GET", "/test", nil)
				req.RemoteAddr = testIP
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				if w.Code != 200 {
					t.Logf("Request %d should succeed", i+1)
					return false
				}
			}

			// Third request should be rate limited
			req := httptest.NewRequest("GET", "/test", nil)
			req.RemoteAddr = testIP
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != 429 {
				t.Logf("Third request should be rate limited, got %d", w.Code)
				return false
			}

			// Wait for window to pass
			time.Sleep(time.Duration(window+1) * time.Second)

			// Now request should succeed again
			req = httptest.NewRequest("GET", "/test", nil)
			req.RemoteAddr = testIP
			w = httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != 200 {
				t.Logf("Request after window should succeed, got %d", w.Code)
				return false
			}

			return true
		},
		gen.UInt8(),
	))

	properties.Property("user-based rate limiting uses user ID", prop.ForAll(
		func(userID uint16) bool {
			// Skip user ID 0
			if userID == 0 {
				userID = 1
			}

			rateLimitConfig := config.RateLimitConfig{
				Enabled:  true,
				Requests: 3,
				Window:   60,
				KeyFunc:  "user",
			}

			router := gin.New()
			router.Use(func(c *gin.Context) {
				// Simulate JWT middleware setting userId
				c.Set("userId", uint(userID))
				c.Next()
			})
			router.Use(RateLimit(rateLimitConfig))
			router.GET("/test", func(c *gin.Context) {
				c.JSON(200, gin.H{"message": "ok"})
			})

			// Make 3 requests (at limit) from same user but different IPs
			for i := 0; i < 3; i++ {
				req := httptest.NewRequest("GET", "/test", nil)
				req.RemoteAddr = fmt.Sprintf("192.168.%d.%d:12345", i, userID%256)
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				if w.Code != 200 {
					t.Logf("Request %d should succeed", i+1)
					return false
				}
			}

			// Fourth request should be rate limited regardless of IP
			req := httptest.NewRequest("GET", "/test", nil)
			req.RemoteAddr = "10.0.0.1:12345" // Different IP
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != 429 {
				t.Logf("User should be rate limited across IPs, got %d", w.Code)
				return false
			}

			return true
		},
		gen.UInt16(),
	))

	properties.Property("user rate limiting falls back to IP when no user", prop.ForAll(
		func(ipOctet uint8) bool {
			rateLimitConfig := config.RateLimitConfig{
				Enabled:  true,
				Requests: 3,
				Window:   60,
				KeyFunc:  "user", // User-based, but no user in context
			}

			router := gin.New()
			// No JWT middleware - userId not set
			router.Use(RateLimit(rateLimitConfig))
			router.GET("/test", func(c *gin.Context) {
				c.JSON(200, gin.H{"message": "ok"})
			})

			testIP := fmt.Sprintf("192.168.100.%d:12345", ipOctet)

			// Make 3 requests (at limit)
			for i := 0; i < 3; i++ {
				req := httptest.NewRequest("GET", "/test", nil)
				req.RemoteAddr = testIP
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				if w.Code != 200 {
					t.Logf("Request %d should succeed", i+1)
					return false
				}
			}

			// Fourth request should be rate limited by IP
			req := httptest.NewRequest("GET", "/test", nil)
			req.RemoteAddr = testIP
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != 429 {
				t.Logf("Should fall back to IP-based rate limiting, got %d", w.Code)
				return false
			}

			return true
		},
		gen.UInt8(),
	))

	properties.TestingRun(t)
}
