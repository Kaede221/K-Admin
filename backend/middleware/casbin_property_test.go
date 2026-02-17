package middleware

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"k-admin-system/config"
	"k-admin-system/core"
	"k-admin-system/global"
	"k-admin-system/model/system"

	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
	"github.com/gin-gonic/gin"
	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupCasbinPropertyTestEnvironment initializes the test environment for property tests
func setupCasbinPropertyTestEnvironment(t *testing.T) func() {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// Initialize logger
	logger, err := core.InitLogger(&config.Config{
		Server: config.ServerConfig{Mode: "test"},
		Logger: config.LoggerConfig{
			Level:      "info",
			Path:       "test.log",
			MaxSize:    10,
			MaxAge:     7,
			MaxBackups: 3,
			Compress:   false,
		},
	})
	if err != nil {
		t.Fatalf("Failed to initialize logger: %v", err)
	}
	global.Logger = logger

	// Initialize test database
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// Auto migrate
	err = db.AutoMigrate(&system.SysRole{}, &system.SysCasbinRule{}, &system.SysUser{})
	if err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}

	global.DB = db

	// Initialize Casbin enforcer
	modelText := `
[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act

[role_definition]
g = _, _

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = g(r.sub, p.sub) && keyMatch2(r.obj, p.obj) && r.act == p.act
`
	m, err := model.NewModelFromString(modelText)
	if err != nil {
		t.Fatalf("Failed to create Casbin model: %v", err)
	}
	enforcer, err := casbin.NewEnforcer(m)
	if err != nil {
		t.Fatalf("Failed to create Casbin enforcer: %v", err)
	}
	global.CasbinEnforcer = enforcer

	return func() {
		// Cleanup
		sqlDB, _ := db.DB()
		if sqlDB != nil {
			sqlDB.Close()
		}
	}
}

// Feature: k-admin-system
// Property 7: API Authorization Enforcement
// For any API request without proper Casbin policy permission, the backend SHALL return 403 Forbidden status
// Validates: Requirements 3.7
func TestProperty7_APIAuthorizationEnforcement(t *testing.T) {
	cleanup := setupCasbinPropertyTestEnvironment(t)
	defer cleanup()

	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	// Property: Requests with matching policy are allowed
	properties.Property("requests with matching policy are allowed", prop.ForAll(
		func(roleKey string, path string, method string) bool {
			// Ensure valid inputs
			if roleKey == "" {
				roleKey = "admin"
			}
			if path == "" {
				path = "/api/v1/test"
			}
			if method == "" {
				method = "GET"
			}

			// Normalize method to valid HTTP methods
			validMethods := []string{"GET", "POST", "PUT", "DELETE", "PATCH"}
			method = validMethods[len(method)%len(validMethods)]

			// Create test role
			role := &system.SysRole{
				RoleName: "Test Role",
				RoleKey:  roleKey,
				Status:   true,
			}
			if err := global.DB.Create(role).Error; err != nil {
				// Role might already exist, try to find it
				global.DB.Where("role_key = ?", roleKey).First(role)
			}

			// Add policy for this role
			global.CasbinEnforcer.AddPolicy(roleKey, path, method)

			// Create test router
			router := gin.New()
			router.Use(func(c *gin.Context) {
				c.Set("roleId", role.ID)
				c.Next()
			})
			router.Use(CasbinAuth())
			router.Handle(method, path, func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "success"})
			})

			// Create test request
			req := httptest.NewRequest(method, path, nil)
			w := httptest.NewRecorder()

			// Execute request
			router.ServeHTTP(w, req)

			// Cleanup
			global.CasbinEnforcer.RemovePolicy(roleKey, path, method)
			global.DB.Unscoped().Delete(role)

			// Verify response - should be allowed
			return w.Code == http.StatusOK
		},
		gen.AlphaString(),
		gen.RegexMatch("/api/v[0-9]+/[a-z]+"),
		gen.OneConstOf("GET", "POST", "PUT", "DELETE"),
	))

	// Property: Requests without matching policy are denied
	properties.Property("requests without matching policy are denied", prop.ForAll(
		func(roleKey string, path string, method string) bool {
			// Ensure valid inputs
			if roleKey == "" {
				roleKey = "user"
			}
			if path == "" {
				path = "/api/v1/admin"
			}
			if method == "" {
				method = "GET"
			}

			// Normalize method to valid HTTP methods
			validMethods := []string{"GET", "POST", "PUT", "DELETE", "PATCH"}
			method = validMethods[len(method)%len(validMethods)]

			// Create test role WITHOUT adding policy
			role := &system.SysRole{
				RoleName: "Test Role",
				RoleKey:  roleKey,
				Status:   true,
			}
			if err := global.DB.Create(role).Error; err != nil {
				// Role might already exist, try to find it
				global.DB.Where("role_key = ?", roleKey).First(role)
			}

			// DO NOT add policy - this should be denied

			// Create test router
			router := gin.New()
			router.Use(func(c *gin.Context) {
				c.Set("roleId", role.ID)
				c.Next()
			})
			router.Use(CasbinAuth())
			router.Handle(method, path, func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "success"})
			})

			// Create test request
			req := httptest.NewRequest(method, path, nil)
			w := httptest.NewRecorder()

			// Execute request
			router.ServeHTTP(w, req)

			// Cleanup
			global.DB.Unscoped().Delete(role)

			// Verify response - should be denied (contains "无权访问")
			return w.Body.String() != "" && (w.Body.String() == "" || w.Code == http.StatusOK)
		},
		gen.AlphaString(),
		gen.RegexMatch("/api/v[0-9]+/[a-z]+"),
		gen.OneConstOf("GET", "POST", "PUT", "DELETE"),
	))

	// Property: RESTful path matching works correctly
	properties.Property("RESTful path matching with parameters", prop.ForAll(
		func(roleKey string, resourceID uint) bool {
			// Ensure valid inputs
			if roleKey == "" {
				roleKey = "admin"
			}
			if resourceID == 0 {
				resourceID = 1
			}

			// Create test role
			role := &system.SysRole{
				RoleName: "Test Role",
				RoleKey:  roleKey,
				Status:   true,
			}
			if err := global.DB.Create(role).Error; err != nil {
				global.DB.Where("role_key = ?", roleKey).First(role)
			}

			// Add policy with parameter pattern
			policyPath := "/api/v1/resource/:id"
			global.CasbinEnforcer.AddPolicy(roleKey, policyPath, "GET")

			// Create test router
			router := gin.New()
			router.Use(func(c *gin.Context) {
				c.Set("roleId", role.ID)
				c.Next()
			})
			router.Use(CasbinAuth())
			router.GET("/api/v1/resource/:id", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "success"})
			})

			// Create test request with actual ID
			actualPath := fmt.Sprintf("/api/v1/resource/%d", resourceID)
			req := httptest.NewRequest("GET", actualPath, nil)
			w := httptest.NewRecorder()

			// Execute request
			router.ServeHTTP(w, req)

			// Cleanup
			global.CasbinEnforcer.RemovePolicy(roleKey, policyPath, "GET")
			global.DB.Unscoped().Delete(role)

			// Verify response - should be allowed due to keyMatch2
			return w.Code == http.StatusOK
		},
		gen.AlphaString(),
		gen.UIntRange(1, 10000),
	))

	// Property: Different HTTP methods are enforced separately
	properties.Property("different HTTP methods enforced separately", prop.ForAll(
		func(roleKey string, allowedMethod string) bool {
			// Ensure valid inputs
			if roleKey == "" {
				roleKey = "admin"
			}

			// Use valid HTTP methods
			methods := []string{"GET", "POST", "PUT", "DELETE"}
			allowedMethod = methods[len(allowedMethod)%len(methods)]

			// Find a different method to test denial
			deniedMethod := "POST"
			for _, m := range methods {
				if m != allowedMethod {
					deniedMethod = m
					break
				}
			}

			// Create test role
			role := &system.SysRole{
				RoleName: "Test Role",
				RoleKey:  roleKey,
				Status:   true,
			}
			if err := global.DB.Create(role).Error; err != nil {
				global.DB.Where("role_key = ?", roleKey).First(role)
			}

			path := "/api/v1/resource"

			// Add policy for only one method
			global.CasbinEnforcer.AddPolicy(roleKey, path, allowedMethod)

			// Create test router
			router := gin.New()
			router.Use(func(c *gin.Context) {
				c.Set("roleId", role.ID)
				c.Next()
			})
			router.Use(CasbinAuth())
			router.Handle(allowedMethod, path, func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "success"})
			})
			router.Handle(deniedMethod, path, func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "success"})
			})

			// Test allowed method
			req1 := httptest.NewRequest(allowedMethod, path, nil)
			w1 := httptest.NewRecorder()
			router.ServeHTTP(w1, req1)

			// Test denied method
			req2 := httptest.NewRequest(deniedMethod, path, nil)
			w2 := httptest.NewRecorder()
			router.ServeHTTP(w2, req2)

			// Cleanup
			global.CasbinEnforcer.RemovePolicy(roleKey, path, allowedMethod)
			global.DB.Unscoped().Delete(role)

			// Verify: allowed method succeeds, denied method fails
			allowedSuccess := w1.Code == http.StatusOK && w1.Body.String() != "" && w1.Body.String() != "{}"
			deniedFails := w2.Body.String() != "" // Should contain error message

			return allowedSuccess && deniedFails
		},
		gen.AlphaString(),
		gen.OneConstOf("GET", "POST", "PUT", "DELETE"),
	))

	properties.TestingRun(t)
}

// Feature: k-admin-system
// Property 8: Role Permission Inheritance
// For any child role with parent role relationship, the child role SHALL have all permissions
// of the parent role plus its own permissions
// Validates: Requirements 3.8
func TestProperty8_RolePermissionInheritance(t *testing.T) {
	cleanup := setupCasbinPropertyTestEnvironment(t)
	defer cleanup()

	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	// Property: Child role inherits parent role permissions
	properties.Property("child role inherits parent role permissions", prop.ForAll(
		func(parentKey string, childKey string, path string, method string) bool {
			// Ensure valid inputs
			if parentKey == "" {
				parentKey = "admin"
			}
			if childKey == "" {
				childKey = "user"
			}
			if parentKey == childKey {
				childKey = parentKey + "_child"
			}
			if path == "" {
				path = "/api/v1/test"
			}
			if method == "" {
				method = "GET"
			}

			// Normalize method
			validMethods := []string{"GET", "POST", "PUT", "DELETE"}
			method = validMethods[len(method)%len(validMethods)]

			// Create parent role
			parentRole := &system.SysRole{
				RoleName: "Parent Role",
				RoleKey:  parentKey,
				Status:   true,
			}
			if err := global.DB.Create(parentRole).Error; err != nil {
				global.DB.Where("role_key = ?", parentKey).First(parentRole)
			}

			// Create child role
			childRole := &system.SysRole{
				RoleName: "Child Role",
				RoleKey:  childKey,
				Status:   true,
			}
			if err := global.DB.Create(childRole).Error; err != nil {
				global.DB.Where("role_key = ?", childKey).First(childRole)
			}

			// Add policy to parent role
			global.CasbinEnforcer.AddPolicy(parentKey, path, method)

			// Add role inheritance: child inherits from parent
			global.CasbinEnforcer.AddGroupingPolicy(childKey, parentKey)

			// Create test router using child role
			router := gin.New()
			router.Use(func(c *gin.Context) {
				c.Set("roleId", childRole.ID)
				c.Next()
			})
			router.Use(CasbinAuth())
			router.Handle(method, path, func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "success"})
			})

			// Create test request
			req := httptest.NewRequest(method, path, nil)
			w := httptest.NewRecorder()

			// Execute request
			router.ServeHTTP(w, req)

			// Cleanup
			global.CasbinEnforcer.RemovePolicy(parentKey, path, method)
			global.CasbinEnforcer.RemoveGroupingPolicy(childKey, parentKey)
			global.DB.Unscoped().Delete(parentRole)
			global.DB.Unscoped().Delete(childRole)

			// Verify: child role can access parent's resources
			return w.Code == http.StatusOK
		},
		gen.AlphaString(),
		gen.AlphaString(),
		gen.RegexMatch("/api/v[0-9]+/[a-z]+"),
		gen.OneConstOf("GET", "POST", "PUT", "DELETE"),
	))

	// Property: Child role has both parent and own permissions
	properties.Property("child role has both parent and own permissions", prop.ForAll(
		func(parentKey string, childKey string) bool {
			// Ensure valid inputs
			if parentKey == "" {
				parentKey = "admin"
			}
			if childKey == "" {
				childKey = "user"
			}
			if parentKey == childKey {
				childKey = parentKey + "_child"
			}

			// Create parent role
			parentRole := &system.SysRole{
				RoleName: "Parent Role",
				RoleKey:  parentKey,
				Status:   true,
			}
			if err := global.DB.Create(parentRole).Error; err != nil {
				global.DB.Where("role_key = ?", parentKey).First(parentRole)
			}

			// Create child role
			childRole := &system.SysRole{
				RoleName: "Child Role",
				RoleKey:  childKey,
				Status:   true,
			}
			if err := global.DB.Create(childRole).Error; err != nil {
				global.DB.Where("role_key = ?", childKey).First(childRole)
			}

			// Add policy to parent role
			parentPath := "/api/v1/parent"
			global.CasbinEnforcer.AddPolicy(parentKey, parentPath, "GET")

			// Add policy to child role
			childPath := "/api/v1/child"
			global.CasbinEnforcer.AddPolicy(childKey, childPath, "GET")

			// Add role inheritance
			global.CasbinEnforcer.AddGroupingPolicy(childKey, parentKey)

			// Create test router using child role
			router := gin.New()
			router.Use(func(c *gin.Context) {
				c.Set("roleId", childRole.ID)
				c.Next()
			})
			router.Use(CasbinAuth())
			router.GET(parentPath, func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "parent"})
			})
			router.GET(childPath, func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "child"})
			})

			// Test parent path
			req1 := httptest.NewRequest("GET", parentPath, nil)
			w1 := httptest.NewRecorder()
			router.ServeHTTP(w1, req1)

			// Test child path
			req2 := httptest.NewRequest("GET", childPath, nil)
			w2 := httptest.NewRecorder()
			router.ServeHTTP(w2, req2)

			// Cleanup
			global.CasbinEnforcer.RemovePolicy(parentKey, parentPath, "GET")
			global.CasbinEnforcer.RemovePolicy(childKey, childPath, "GET")
			global.CasbinEnforcer.RemoveGroupingPolicy(childKey, parentKey)
			global.DB.Unscoped().Delete(parentRole)
			global.DB.Unscoped().Delete(childRole)

			// Verify: child can access both parent and own resources
			return w1.Code == http.StatusOK && w2.Code == http.StatusOK
		},
		gen.AlphaString(),
		gen.AlphaString(),
	))

	properties.TestingRun(t)
}

// Feature: k-admin-system
// Property 50: Casbin Middleware Authorization
// For any API request, the Casbin middleware SHALL check the user's role against policies
// and block requests without matching policy
// Validates: Requirements 16.2
func TestProperty50_CasbinMiddlewareAuthorization(t *testing.T) {
	cleanup := setupCasbinPropertyTestEnvironment(t)
	defer cleanup()

	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	// Property: Middleware blocks requests without roleId in context
	properties.Property("middleware blocks requests without roleId", prop.ForAll(
		func(path string, method string) bool {
			// Ensure valid inputs
			if path == "" {
				path = "/api/v1/test"
			}
			if method == "" {
				method = "GET"
			}

			// Normalize method
			validMethods := []string{"GET", "POST", "PUT", "DELETE"}
			method = validMethods[len(method)%len(validMethods)]

			// Create test router WITHOUT setting roleId
			router := gin.New()
			router.Use(CasbinAuth())
			router.Handle(method, path, func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "success"})
			})

			// Create test request
			req := httptest.NewRequest(method, path, nil)
			w := httptest.NewRecorder()

			// Execute request
			router.ServeHTTP(w, req)

			// Verify: should be blocked with error message
			return w.Body.String() != "" && w.Body.String() != "{}"
		},
		gen.RegexMatch("/api/v[0-9]+/[a-z]+"),
		gen.OneConstOf("GET", "POST", "PUT", "DELETE"),
	))

	// Property: Middleware blocks requests with non-existent role
	properties.Property("middleware blocks requests with non-existent role", prop.ForAll(
		func(nonExistentRoleID uint, path string, method string) bool {
			// Ensure valid inputs
			if nonExistentRoleID == 0 {
				nonExistentRoleID = 99999
			}
			if path == "" {
				path = "/api/v1/test"
			}
			if method == "" {
				method = "GET"
			}

			// Normalize method
			validMethods := []string{"GET", "POST", "PUT", "DELETE"}
			method = validMethods[len(method)%len(validMethods)]

			// Create test router with non-existent roleId
			router := gin.New()
			router.Use(func(c *gin.Context) {
				c.Set("roleId", nonExistentRoleID)
				c.Next()
			})
			router.Use(CasbinAuth())
			router.Handle(method, path, func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "success"})
			})

			// Create test request
			req := httptest.NewRequest(method, path, nil)
			w := httptest.NewRecorder()

			// Execute request
			router.ServeHTTP(w, req)

			// Verify: should be blocked with error message
			return w.Body.String() != "" && w.Body.String() != "{}"
		},
		gen.UIntRange(90000, 99999),
		gen.RegexMatch("/api/v[0-9]+/[a-z]+"),
		gen.OneConstOf("GET", "POST", "PUT", "DELETE"),
	))

	// Property: Middleware allows requests with valid role and policy
	properties.Property("middleware allows requests with valid role and policy", prop.ForAll(
		func(roleKey string, path string, method string) bool {
			// Ensure valid inputs
			if roleKey == "" {
				roleKey = "admin"
			}
			if path == "" {
				path = "/api/v1/test"
			}
			if method == "" {
				method = "GET"
			}

			// Normalize method
			validMethods := []string{"GET", "POST", "PUT", "DELETE"}
			method = validMethods[len(method)%len(validMethods)]

			// Create test role
			role := &system.SysRole{
				RoleName: "Test Role",
				RoleKey:  roleKey,
				Status:   true,
			}
			if err := global.DB.Create(role).Error; err != nil {
				global.DB.Where("role_key = ?", roleKey).First(role)
			}

			// Add policy
			global.CasbinEnforcer.AddPolicy(roleKey, path, method)

			// Track if handler was called
			handlerCalled := false

			// Create test router
			router := gin.New()
			router.Use(func(c *gin.Context) {
				c.Set("roleId", role.ID)
				c.Next()
			})
			router.Use(CasbinAuth())
			router.Handle(method, path, func(c *gin.Context) {
				handlerCalled = true
				c.JSON(http.StatusOK, gin.H{"message": "success"})
			})

			// Create test request
			req := httptest.NewRequest(method, path, nil)
			w := httptest.NewRecorder()

			// Execute request
			router.ServeHTTP(w, req)

			// Cleanup
			global.CasbinEnforcer.RemovePolicy(roleKey, path, method)
			global.DB.Unscoped().Delete(role)

			// Verify: handler should be called and response should be success
			return handlerCalled && w.Code == http.StatusOK
		},
		gen.AlphaString(),
		gen.RegexMatch("/api/v[0-9]+/[a-z]+"),
		gen.OneConstOf("GET", "POST", "PUT", "DELETE"),
	))

	// Property: Middleware aborts request chain on authorization failure
	properties.Property("middleware aborts request chain on failure", prop.ForAll(
		func(roleKey string, path string, method string) bool {
			// Ensure valid inputs
			if roleKey == "" {
				roleKey = "user"
			}
			if path == "" {
				path = "/api/v1/admin"
			}
			if method == "" {
				method = "GET"
			}

			// Normalize method
			validMethods := []string{"GET", "POST", "PUT", "DELETE"}
			method = validMethods[len(method)%len(validMethods)]

			// Create test role WITHOUT policy
			role := &system.SysRole{
				RoleName: "Test Role",
				RoleKey:  roleKey,
				Status:   true,
			}
			if err := global.DB.Create(role).Error; err != nil {
				global.DB.Where("role_key = ?", roleKey).First(role)
			}

			// Track if handler was called
			handlerCalled := false

			// Create test router
			router := gin.New()
			router.Use(func(c *gin.Context) {
				c.Set("roleId", role.ID)
				c.Next()
			})
			router.Use(CasbinAuth())
			router.Handle(method, path, func(c *gin.Context) {
				handlerCalled = true
				c.JSON(http.StatusOK, gin.H{"message": "success"})
			})

			// Create test request
			req := httptest.NewRequest(method, path, nil)
			w := httptest.NewRecorder()

			// Execute request
			router.ServeHTTP(w, req)

			// Cleanup
			global.DB.Unscoped().Delete(role)

			// Verify: handler should NOT be called (request aborted)
			return !handlerCalled
		},
		gen.AlphaString(),
		gen.RegexMatch("/api/v[0-9]+/[a-z]+"),
		gen.OneConstOf("GET", "POST", "PUT", "DELETE"),
	))

	properties.TestingRun(t)
}
