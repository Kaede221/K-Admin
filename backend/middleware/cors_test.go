package middleware

import (
	"k-admin-system/config"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestCORS_AllowedOrigin(t *testing.T) {
	gin.SetMode(gin.TestMode)

	corsConfig := config.CORSConfig{
		AllowOrigins:     []string{"http://localhost:3000", "https://example.com"},
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

	// Test with allowed origin
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.Equal(t, "http://localhost:3000", w.Header().Get("Access-Control-Allow-Origin"))
	assert.Equal(t, "GET, POST, PUT, DELETE", w.Header().Get("Access-Control-Allow-Methods"))
	assert.Equal(t, "Content-Type, Authorization", w.Header().Get("Access-Control-Allow-Headers"))
	assert.Equal(t, "X-Total-Count", w.Header().Get("Access-Control-Expose-Headers"))
	assert.Equal(t, "true", w.Header().Get("Access-Control-Allow-Credentials"))
	assert.Equal(t, "3600", w.Header().Get("Access-Control-Max-Age"))
}

func TestCORS_DisallowedOrigin(t *testing.T) {
	gin.SetMode(gin.TestMode)

	corsConfig := config.CORSConfig{
		AllowOrigins: []string{"http://localhost:3000"},
		AllowMethods: []string{"GET", "POST"},
		AllowHeaders: []string{"Content-Type"},
	}

	router := gin.New()
	router.Use(CORS(corsConfig))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "ok"})
	})

	// Test with disallowed origin
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Origin", "http://evil.com")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.Empty(t, w.Header().Get("Access-Control-Allow-Origin"))
}

func TestCORS_WildcardOrigin(t *testing.T) {
	gin.SetMode(gin.TestMode)

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

	// Test with any origin
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Origin", "http://any-domain.com")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.Equal(t, "http://any-domain.com", w.Header().Get("Access-Control-Allow-Origin"))
}

func TestCORS_WildcardSubdomain(t *testing.T) {
	gin.SetMode(gin.TestMode)

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

	// Test with matching subdomain
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Origin", "http://api.example.com")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.Equal(t, "http://api.example.com", w.Header().Get("Access-Control-Allow-Origin"))

	// Test with non-matching domain
	req2 := httptest.NewRequest("GET", "/test", nil)
	req2.Header.Set("Origin", "http://example.org")
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)

	assert.Equal(t, 200, w2.Code)
	assert.Empty(t, w2.Header().Get("Access-Control-Allow-Origin"))
}

func TestCORS_PreflightRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)

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

	// Test OPTIONS preflight request
	req := httptest.NewRequest("OPTIONS", "/test", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	req.Header.Set("Access-Control-Request-Method", "POST")
	req.Header.Set("Access-Control-Request-Headers", "Content-Type")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, 204, w.Code)
	assert.Equal(t, "http://localhost:3000", w.Header().Get("Access-Control-Allow-Origin"))
	assert.Equal(t, "GET, POST, PUT, DELETE", w.Header().Get("Access-Control-Allow-Methods"))
	assert.Equal(t, "Content-Type, Authorization", w.Header().Get("Access-Control-Allow-Headers"))
	assert.Equal(t, "86400", w.Header().Get("Access-Control-Max-Age"))
}

func TestCORS_NoOriginHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)

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

	// Test without Origin header (same-origin request)
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.Empty(t, w.Header().Get("Access-Control-Allow-Origin"))
}

func TestCORS_WithoutCredentials(t *testing.T) {
	gin.SetMode(gin.TestMode)

	corsConfig := config.CORSConfig{
		AllowOrigins:     []string{"http://localhost:3000"},
		AllowMethods:     []string{"GET"},
		AllowHeaders:     []string{"Content-Type"},
		AllowCredentials: false,
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

	assert.Equal(t, 200, w.Code)
	assert.Empty(t, w.Header().Get("Access-Control-Allow-Credentials"))
}

func TestIsOriginAllowed(t *testing.T) {
	tests := []struct {
		name           string
		origin         string
		allowedOrigins []string
		expected       bool
	}{
		{
			name:           "exact match",
			origin:         "http://localhost:3000",
			allowedOrigins: []string{"http://localhost:3000", "https://example.com"},
			expected:       true,
		},
		{
			name:           "wildcard all",
			origin:         "http://any-domain.com",
			allowedOrigins: []string{"*"},
			expected:       true,
		},
		{
			name:           "wildcard subdomain match",
			origin:         "http://api.example.com",
			allowedOrigins: []string{"*.example.com"},
			expected:       true,
		},
		{
			name:           "wildcard subdomain no match",
			origin:         "http://example.org",
			allowedOrigins: []string{"*.example.com"},
			expected:       false,
		},
		{
			name:           "not in list",
			origin:         "http://evil.com",
			allowedOrigins: []string{"http://localhost:3000", "https://example.com"},
			expected:       false,
		},
		{
			name:           "empty allowed list",
			origin:         "http://localhost:3000",
			allowedOrigins: []string{},
			expected:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isOriginAllowed(tt.origin, tt.allowedOrigins)
			assert.Equal(t, tt.expected, result)
		})
	}
}
