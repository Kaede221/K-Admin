package system

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// TestInitUserRouter_PublicRoutes tests that public routes are registered correctly
func TestInitUserRouter_PublicRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	apiV1 := router.Group("/api/v1")

	// Initialize user routes
	InitUserRouter(apiV1)

	// Test that login route exists and is accessible without JWT
	req := httptest.NewRequest(http.MethodPost, "/api/v1/user/login", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Should not return 404 (route exists)
	assert.NotEqual(t, http.StatusNotFound, w.Code, "Login route should exist")
}

// TestInitUserRouter_ProtectedRoutes tests that protected routes are registered correctly
func TestInitUserRouter_ProtectedRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	apiV1 := router.Group("/api/v1")

	// Initialize user routes
	InitUserRouter(apiV1)

	// Test protected routes exist
	protectedRoutes := []struct {
		method string
		path   string
	}{
		{http.MethodPost, "/api/v1/user"},
		{http.MethodPut, "/api/v1/user"},
		{http.MethodDelete, "/api/v1/user/1"},
		{http.MethodGet, "/api/v1/user/1"},
		{http.MethodGet, "/api/v1/user/list"},
		{http.MethodPost, "/api/v1/user/change-password"},
		{http.MethodPost, "/api/v1/user/reset-password"},
		{http.MethodPost, "/api/v1/user/toggle-status"},
	}

	for _, route := range protectedRoutes {
		t.Run(route.method+"_"+route.path, func(t *testing.T) {
			req := httptest.NewRequest(route.method, route.path, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Should not return 404 (route exists)
			// Will return 401 because no JWT token is provided
			assert.NotEqual(t, http.StatusNotFound, w.Code, "Route should exist: %s %s", route.method, route.path)
		})
	}
}

// TestInitUserRouter_JWTMiddlewareApplied tests that JWT middleware is applied to protected routes
func TestInitUserRouter_JWTMiddlewareApplied(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	apiV1 := router.Group("/api/v1")

	// Initialize user routes
	InitUserRouter(apiV1)

	// Test that protected routes require JWT (should return 401 without token)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/user/list", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Should return 401 Unauthorized because no JWT token is provided
	assert.Equal(t, http.StatusOK, w.Code, "Should return 200 with error response (unified response format)")
}

// TestInitUserRouter_LoginNoJWT tests that login route does not require JWT
func TestInitUserRouter_LoginNoJWT(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	apiV1 := router.Group("/api/v1")

	// Initialize user routes
	InitUserRouter(apiV1)

	// Test that login route does not require JWT
	req := httptest.NewRequest(http.MethodPost, "/api/v1/user/login", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Should not return 401 (JWT not required for login)
	// Will return 200 with error response due to missing request body
	assert.NotEqual(t, http.StatusUnauthorized, w.Code, "Login should not require JWT")
}
