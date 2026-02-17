package middleware

import (
	"k-admin-system/config"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// Feature: k-admin-system
// Property 51: CORS Header Configuration
// For any cross-origin request from an allowed origin, the CORS middleware SHALL
// set appropriate Access-Control headers
// Validates: Requirements 16.3
func TestProperty51_CORSHeaderConfiguration(t *testing.T) {
	gin.SetMode(gin.TestMode)

	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	properties.Property("allowed origin receives CORS headers", prop.ForAll(
		func(port uint16) bool {
			// Generate origin from port
			origin := "http://localhost:" + string(rune(port%10000+3000))

			corsConfig := config.CORSConfig{
				AllowOrigins:     []string{origin, "http://localhost:3000"},
				AllowMethods:     []string{"GET", "POST", "PUT", "DELETE"},
				AllowHeaders:     []string{"Content-Type", "Authorization"},
				ExposeHeaders:    []string{"X-Total-Count"},
				AllowCredentials: true,
				MaxAge:           3600,
			}

			router := gin.New()
			router.Use(CORS(corsConfig))
			router.GET("/test", func(c *gin.Context) {
				c.JSON(200, gin.H{"message": "ok"})
			})

			req := httptest.NewRequest("GET", "/test", nil)
			req.Header.Set("Origin", origin)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Verify CORS headers are set
			if w.Header().Get("Access-Control-Allow-Origin") != origin {
				t.Logf("Expected Access-Control-Allow-Origin: %s, got: %s", origin, w.Header().Get("Access-Control-Allow-Origin"))
				return false
			}

			if w.Header().Get("Access-Control-Allow-Methods") == "" {
				t.Logf("Access-Control-Allow-Methods header not set")
				return false
			}

			if w.Header().Get("Access-Control-Allow-Headers") == "" {
				t.Logf("Access-Control-Allow-Headers header not set")
				return false
			}

			if w.Header().Get("Access-Control-Allow-Credentials") != "true" {
				t.Logf("Access-Control-Allow-Credentials should be 'true'")
				return false
			}

			return true
		},
		gen.UInt16(),
	))

	properties.Property("wildcard origin allows any origin", prop.ForAll(
		func(domain string) bool {
			// Skip empty domains
			if domain == "" {
				domain = "example.com"
			}

			origin := "http://" + domain

			corsConfig := config.CORSConfig{
				AllowOrigins: []string{"*"},
				AllowMethods: []string{"GET", "POST"},
				AllowHeaders: []string{"Content-Type"},
			}

			router := gin.New()
			router.Use(CORS(corsConfig))
			router.GET("/test", func(c *gin.Context) {
				c.JSON(200, gin.H{"message": "ok"})
			})

			req := httptest.NewRequest("GET", "/test", nil)
			req.Header.Set("Origin", origin)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Verify origin is allowed
			if w.Header().Get("Access-Control-Allow-Origin") != origin {
				t.Logf("Wildcard should allow origin %s", origin)
				return false
			}

			return true
		},
		gen.AlphaString(),
	))

	properties.Property("disallowed origin receives no CORS headers", prop.ForAll(
		func(domain string) bool {
			// Skip empty domains
			if domain == "" {
				return true
			}

			origin := "http://" + domain

			corsConfig := config.CORSConfig{
				AllowOrigins: []string{"http://localhost:3000", "https://example.com"},
				AllowMethods: []string{"GET", "POST"},
				AllowHeaders: []string{"Content-Type"},
			}

			// Skip if origin happens to match allowed origins
			if origin == "http://localhost:3000" || origin == "https://example.com" {
				return true
			}

			router := gin.New()
			router.Use(CORS(corsConfig))
			router.GET("/test", func(c *gin.Context) {
				c.JSON(200, gin.H{"message": "ok"})
			})

			req := httptest.NewRequest("GET", "/test", nil)
			req.Header.Set("Origin", origin)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Verify no CORS headers are set for disallowed origin
			if w.Header().Get("Access-Control-Allow-Origin") != "" {
				t.Logf("Disallowed origin %s should not receive CORS headers", origin)
				return false
			}

			return true
		},
		gen.AlphaString(),
	))

	properties.Property("OPTIONS preflight returns 204", prop.ForAll(
		func() bool {
			corsConfig := config.CORSConfig{
				AllowOrigins: []string{"http://localhost:3000"},
				AllowMethods: []string{"GET", "POST", "PUT", "DELETE"},
				AllowHeaders: []string{"Content-Type", "Authorization"},
				MaxAge:       86400,
			}

			router := gin.New()
			router.Use(CORS(corsConfig))
			router.POST("/test", func(c *gin.Context) {
				c.JSON(200, gin.H{"message": "ok"})
			})

			req := httptest.NewRequest("OPTIONS", "/test", nil)
			req.Header.Set("Origin", "http://localhost:3000")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Verify preflight returns 204
			if w.Code != 204 {
				t.Logf("Expected status 204 for OPTIONS, got %d", w.Code)
				return false
			}

			// Verify CORS headers are set
			if w.Header().Get("Access-Control-Allow-Origin") == "" {
				t.Logf("CORS headers should be set for OPTIONS request")
				return false
			}

			return true
		},
	))

	properties.Property("wildcard subdomain matches correctly", prop.ForAll(
		func(subdomain string) bool {
			// Skip empty subdomains
			if subdomain == "" {
				subdomain = "api"
			}

			// Clean subdomain to be valid
			subdomain = strings.Map(func(r rune) rune {
				if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' {
					return r
				}
				return -1
			}, subdomain)

			if subdomain == "" {
				subdomain = "api"
			}

			origin := "http://" + subdomain + ".example.com"

			corsConfig := config.CORSConfig{
				AllowOrigins: []string{"*.example.com"},
				AllowMethods: []string{"GET"},
				AllowHeaders: []string{"Content-Type"},
			}

			router := gin.New()
			router.Use(CORS(corsConfig))
			router.GET("/test", func(c *gin.Context) {
				c.JSON(200, gin.H{"message": "ok"})
			})

			req := httptest.NewRequest("GET", "/test", nil)
			req.Header.Set("Origin", origin)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Verify subdomain is allowed
			if w.Header().Get("Access-Control-Allow-Origin") != origin {
				t.Logf("Wildcard subdomain should allow %s", origin)
				return false
			}

			return true
		},
		gen.AlphaString(),
	))

	properties.Property("no origin header means no CORS headers", prop.ForAll(
		func() bool {
			corsConfig := config.CORSConfig{
				AllowOrigins: []string{"http://localhost:3000"},
				AllowMethods: []string{"GET"},
				AllowHeaders: []string{"Content-Type"},
			}

			router := gin.New()
			router.Use(CORS(corsConfig))
			router.GET("/test", func(c *gin.Context) {
				c.JSON(200, gin.H{"message": "ok"})
			})

			req := httptest.NewRequest("GET", "/test", nil)
			// No Origin header set
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Verify no CORS headers are set
			if w.Header().Get("Access-Control-Allow-Origin") != "" {
				t.Logf("No CORS headers should be set without Origin header")
				return false
			}

			return true
		},
	))

	properties.Property("credentials flag controls header", prop.ForAll(
		func(allowCredentials bool) bool {
			corsConfig := config.CORSConfig{
				AllowOrigins:     []string{"http://localhost:3000"},
				AllowMethods:     []string{"GET"},
				AllowHeaders:     []string{"Content-Type"},
				AllowCredentials: allowCredentials,
			}

			router := gin.New()
			router.Use(CORS(corsConfig))
			router.GET("/test", func(c *gin.Context) {
				c.JSON(200, gin.H{"message": "ok"})
			})

			req := httptest.NewRequest("GET", "/test", nil)
			req.Header.Set("Origin", "http://localhost:3000")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			credHeader := w.Header().Get("Access-Control-Allow-Credentials")

			if allowCredentials {
				if credHeader != "true" {
					t.Logf("Expected Access-Control-Allow-Credentials: true, got: %s", credHeader)
					return false
				}
			} else {
				if credHeader != "" {
					t.Logf("Expected no Access-Control-Allow-Credentials header, got: %s", credHeader)
					return false
				}
			}

			return true
		},
		gen.Bool(),
	))

	properties.Property("expose headers are set correctly", prop.ForAll(
		func(headerCount uint8) bool {
			// Limit header count to reasonable range
			count := int(headerCount%5 + 1)

			exposeHeaders := make([]string, count)
			for i := 0; i < count; i++ {
				exposeHeaders[i] = "X-Custom-Header-" + string(rune('A'+i))
			}

			corsConfig := config.CORSConfig{
				AllowOrigins:  []string{"http://localhost:3000"},
				AllowMethods:  []string{"GET"},
				AllowHeaders:  []string{"Content-Type"},
				ExposeHeaders: exposeHeaders,
			}

			router := gin.New()
			router.Use(CORS(corsConfig))
			router.GET("/test", func(c *gin.Context) {
				c.JSON(200, gin.H{"message": "ok"})
			})

			req := httptest.NewRequest("GET", "/test", nil)
			req.Header.Set("Origin", "http://localhost:3000")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			exposeHeader := w.Header().Get("Access-Control-Expose-Headers")

			// Verify all expose headers are present
			for _, header := range exposeHeaders {
				if !strings.Contains(exposeHeader, header) {
					t.Logf("Expected expose header %s not found in %s", header, exposeHeader)
					return false
				}
			}

			return true
		},
		gen.UInt8(),
	))

	properties.TestingRun(t)
}
