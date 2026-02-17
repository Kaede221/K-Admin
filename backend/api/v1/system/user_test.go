package system

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"k-admin-system/config"
	"k-admin-system/global"
	"k-admin-system/model/common"
	"k-admin-system/model/system"
	"k-admin-system/utils"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupTestRouter initializes a test Gin router with the User API
func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	return router
}

// setupTestDB initializes an in-memory SQLite database for testing
func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// Run migrations
	err = db.AutoMigrate(&system.SysUser{})
	if err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}

	return db
}

// cleanupTestDB cleans all data from the test database
func cleanupTestDB(db *gorm.DB) {
	db.Exec("DELETE FROM sys_users")
}

// createTestUser creates a test user in the database
func createTestUser(db *gorm.DB, username, password string, roleID uint, active bool) (*system.SysUser, error) {
	user := &system.SysUser{
		Username: username,
		Password: password,
		RoleID:   roleID,
		Active:   active,
	}
	err := db.Create(user).Error
	return user, err
}

// TestLogin_ValidCredentials tests successful login with valid credentials
func TestLogin_ValidCredentials(t *testing.T) {
	// Setup
	db := setupTestDB(t)
	defer cleanupTestDB(db)

	originalDB := global.DB
	global.DB = db
	defer func() { global.DB = originalDB }()

	// Initialize minimal config for JWT
	originalConfig := global.Config
	global.Config = &config.Config{
		JWT: config.JWTConfig{
			Secret:            "test-secret-key",
			AccessExpiration:  15,
			RefreshExpiration: 10080,
		},
	}
	defer func() { global.Config = originalConfig }()

	// Hash the password properly using bcrypt
	hashedPassword, err := utils.HashPassword("password123")
	assert.NoError(t, err)

	// Create a test user with hashed password
	testUser := &system.SysUser{
		Username: "testuser",
		Password: hashedPassword,
		RoleID:   1,
		Active:   true,
	}
	db.Create(testUser)

	// Setup router
	router := setupTestRouter()
	userApi := &UserApi{}
	router.POST("/api/v1/user/login", userApi.Login)

	// Create request
	loginReq := LoginRequest{
		Username: "testuser",
		Password: "password123",
	}
	body, _ := json.Marshal(loginReq)
	req, _ := http.NewRequest("POST", "/api/v1/user/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	// Execute request
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)

	var response common.Response
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, 0, response.Code)

	// Verify response contains tokens and user
	responseData, ok := response.Data.(map[string]interface{})
	assert.True(t, ok)
	assert.NotEmpty(t, responseData["accessToken"])
	assert.NotEmpty(t, responseData["refreshToken"])
	assert.NotNil(t, responseData["user"])
}

// TestLogin_InvalidCredentials tests login failure with invalid credentials
func TestLogin_InvalidCredentials(t *testing.T) {
	// Setup
	db := setupTestDB(t)
	defer cleanupTestDB(db)

	originalDB := global.DB
	global.DB = db
	defer func() { global.DB = originalDB }()

	// Hash the password properly
	hashedPassword, err := utils.HashPassword("password123")
	assert.NoError(t, err)

	// Create a test user
	testUser := &system.SysUser{
		Username: "testuser",
		Password: hashedPassword,
		RoleID:   1,
		Active:   true,
	}
	db.Create(testUser)

	// Setup router
	router := setupTestRouter()
	userApi := &UserApi{}
	router.POST("/api/v1/user/login", userApi.Login)

	// Test cases for invalid credentials
	testCases := []struct {
		name     string
		username string
		password string
	}{
		{"Wrong password", "testuser", "wrongpassword"},
		{"Non-existent user", "nonexistent", "password123"},
		{"Empty password", "testuser", ""},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			loginReq := LoginRequest{
				Username: tc.username,
				Password: tc.password,
			}
			body, _ := json.Marshal(loginReq)
			req, _ := http.NewRequest("POST", "/api/v1/user/login", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)

			var response common.Response
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.NotEqual(t, 0, response.Code, "Expected non-zero error code for invalid credentials")
			assert.NotEmpty(t, response.Msg, "Expected error message")
		})
	}
}

// TestCreateUser_DuplicateUsername tests user creation with duplicate username
func TestCreateUser_DuplicateUsername(t *testing.T) {
	// Setup
	db := setupTestDB(t)
	defer cleanupTestDB(db)

	originalDB := global.DB
	global.DB = db
	defer func() { global.DB = originalDB }()

	// Create first user
	firstUser := &system.SysUser{
		Username: "duplicateuser",
		Password: "password123",
		RoleID:   1,
		Active:   true,
	}
	db.Create(firstUser)

	// Setup router
	router := setupTestRouter()
	userApi := &UserApi{}
	router.POST("/api/v1/user", userApi.CreateUser)

	// Attempt to create second user with same username
	createReq := CreateUserRequest{
		Username: "duplicateuser",
		Password: "password456",
		RoleID:   1,
		Active:   true,
	}
	body, _ := json.Marshal(createReq)
	req, _ := http.NewRequest("POST", "/api/v1/user", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)

	var response common.Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.NotEqual(t, 0, response.Code, "Expected non-zero error code for duplicate username")
	assert.Contains(t, response.Msg, "username already exists", "Expected error message about duplicate username")
}

// TestCreateUser_Success tests successful user creation
func TestCreateUser_Success(t *testing.T) {
	// Setup
	db := setupTestDB(t)
	defer cleanupTestDB(db)

	originalDB := global.DB
	global.DB = db
	defer func() { global.DB = originalDB }()

	// Setup router
	router := setupTestRouter()
	userApi := &UserApi{}
	router.POST("/api/v1/user", userApi.CreateUser)

	// Create user request
	createReq := CreateUserRequest{
		Username: "newuser",
		Password: "password123",
		Nickname: "New User",
		Email:    "newuser@test.com",
		RoleID:   1,
		Active:   true,
	}
	body, _ := json.Marshal(createReq)
	req, _ := http.NewRequest("POST", "/api/v1/user", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)

	var response common.Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, 0, response.Code)

	// Verify user was created in database
	var createdUser system.SysUser
	err = db.Where("username = ?", "newuser").First(&createdUser).Error
	assert.NoError(t, err)
	assert.Equal(t, "newuser", createdUser.Username)
	assert.Equal(t, "New User", createdUser.Nickname)
}

// TestPasswordMasking_InResponses tests that password field is masked in API responses
func TestPasswordMasking_InResponses(t *testing.T) {
	// Setup
	db := setupTestDB(t)
	defer cleanupTestDB(db)

	originalDB := global.DB
	global.DB = db
	defer func() { global.DB = originalDB }()

	// Create a test user
	testUser := &system.SysUser{
		Username: "testuser",
		Password: "$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy",
		Nickname: "Test User",
		RoleID:   1,
		Active:   true,
	}
	db.Create(testUser)

	// Setup router
	router := setupTestRouter()
	userApi := &UserApi{}
	router.GET("/api/v1/user/:id", userApi.GetUser)
	router.GET("/api/v1/user/list", userApi.GetUserList)

	// Test GetUser endpoint
	t.Run("GetUser masks password", func(t *testing.T) {
		req, _ := http.NewRequest("GET", fmt.Sprintf("/api/v1/user/%d", testUser.ID), nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response common.Response
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		// Check that password field is not in JSON response
		responseJSON := w.Body.String()
		assert.NotContains(t, responseJSON, "password", "Password field should not be in response")
		assert.NotContains(t, responseJSON, "$2a$10$", "Password hash should not be in response")
	})

	// Test GetUserList endpoint
	t.Run("GetUserList masks password", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/v1/user/list?page=1&pageSize=10", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response common.Response
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		// Check that password field is not in JSON response
		responseJSON := w.Body.String()
		assert.NotContains(t, responseJSON, "password", "Password field should not be in response")
		assert.NotContains(t, responseJSON, "$2a$10$", "Password hash should not be in response")
	})
}

// TestGetUserList_PaginationAndFiltering tests pagination and filtering functionality
func TestGetUserList_PaginationAndFiltering(t *testing.T) {
	// Setup
	db := setupTestDB(t)
	defer cleanupTestDB(db)

	originalDB := global.DB
	global.DB = db
	defer func() { global.DB = originalDB }()

	// Create multiple test users
	for i := 1; i <= 15; i++ {
		user := &system.SysUser{
			Username: fmt.Sprintf("user%d", i),
			Password: "password123",
			Nickname: fmt.Sprintf("User %d", i),
			Email:    fmt.Sprintf("user%d@test.com", i),
			RoleID:   uint(i%3 + 1), // Role IDs: 1, 2, 3
			Active:   i%2 == 0,      // Alternate active/inactive
		}
		db.Create(user)
	}

	// Setup router
	router := setupTestRouter()
	userApi := &UserApi{}
	router.GET("/api/v1/user/list", userApi.GetUserList)

	// Test pagination
	t.Run("Pagination - Page 1", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/v1/user/list?page=1&pageSize=5", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response common.Response
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, 0, response.Code)

		responseData, ok := response.Data.(map[string]interface{})
		assert.True(t, ok)

		list, ok := responseData["list"].([]interface{})
		assert.True(t, ok)
		assert.Equal(t, 5, len(list), "Expected 5 users on page 1")

		total, ok := responseData["total"].(float64)
		assert.True(t, ok)
		assert.Equal(t, float64(15), total, "Expected total of 15 users")
	})

	t.Run("Pagination - Page 2", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/v1/user/list?page=2&pageSize=5", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response common.Response
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		responseData, ok := response.Data.(map[string]interface{})
		assert.True(t, ok)

		list, ok := responseData["list"].([]interface{})
		assert.True(t, ok)
		assert.Equal(t, 5, len(list), "Expected 5 users on page 2")
	})

	t.Run("Pagination - Last page", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/v1/user/list?page=3&pageSize=5", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response common.Response
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		responseData, ok := response.Data.(map[string]interface{})
		assert.True(t, ok)

		list, ok := responseData["list"].([]interface{})
		assert.True(t, ok)
		assert.Equal(t, 5, len(list), "Expected 5 users on page 3")
	})

	// Test filtering by username
	t.Run("Filter by username", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/v1/user/list?page=1&pageSize=10&username=user1", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response common.Response
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		responseData, ok := response.Data.(map[string]interface{})
		assert.True(t, ok)

		list, ok := responseData["list"].([]interface{})
		assert.True(t, ok)
		// Should match: user1, user10, user11, user12, user13, user14, user15
		assert.GreaterOrEqual(t, len(list), 1, "Expected at least 1 user matching 'user1'")
	})

	// Test filtering by role
	t.Run("Filter by roleId", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/v1/user/list?page=1&pageSize=20&roleId=1", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response common.Response
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		responseData, ok := response.Data.(map[string]interface{})
		assert.True(t, ok)

		list, ok := responseData["list"].([]interface{})
		assert.True(t, ok)
		// Users with roleId=1: user3, user6, user9, user12, user15 (5 users)
		assert.Equal(t, 5, len(list), "Expected 5 users with roleId=1")
	})

	// Test filtering by active status
	t.Run("Filter by active status", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/v1/user/list?page=1&pageSize=20&active=true", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response common.Response
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		responseData, ok := response.Data.(map[string]interface{})
		assert.True(t, ok)

		list, ok := responseData["list"].([]interface{})
		assert.True(t, ok)

		total, ok := responseData["total"].(float64)
		assert.True(t, ok)
		// Active users: user2, user4, user6, user8, user10, user12, user14 (7 users with i%2==0)
		// But we need to verify the actual count returned
		assert.Equal(t, float64(len(list)), total, "Total should match list length")

		// Verify all returned users are active
		for _, u := range list {
			userMap, ok := u.(map[string]interface{})
			assert.True(t, ok)
			active, ok := userMap["active"].(bool)
			assert.True(t, ok)
			assert.True(t, active, "All returned users should be active")
		}
	})
}

// TestGetUser_NotFound tests getting a non-existent user
func TestGetUser_NotFound(t *testing.T) {
	// Setup
	db := setupTestDB(t)
	defer cleanupTestDB(db)

	originalDB := global.DB
	global.DB = db
	defer func() { global.DB = originalDB }()

	// Setup router
	router := setupTestRouter()
	userApi := &UserApi{}
	router.GET("/api/v1/user/:id", userApi.GetUser)

	// Request non-existent user
	req, _ := http.NewRequest("GET", "/api/v1/user/999", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response common.Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.NotEqual(t, 0, response.Code, "Expected non-zero error code for non-existent user")
}

// TestDeleteUser_Success tests successful user deletion
func TestDeleteUser_Success(t *testing.T) {
	// Setup
	db := setupTestDB(t)
	defer cleanupTestDB(db)

	originalDB := global.DB
	global.DB = db
	defer func() { global.DB = originalDB }()

	// Create a test user
	testUser := &system.SysUser{
		Username: "deleteuser",
		Password: "password123",
		RoleID:   1,
		Active:   true,
	}
	db.Create(testUser)

	// Setup router
	router := setupTestRouter()
	userApi := &UserApi{}
	router.DELETE("/api/v1/user/:id", userApi.DeleteUser)

	// Delete user
	req, _ := http.NewRequest("DELETE", fmt.Sprintf("/api/v1/user/%d", testUser.ID), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response common.Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, 0, response.Code)

	// Verify user is soft deleted
	var deletedUser system.SysUser
	err = db.First(&deletedUser, testUser.ID).Error
	assert.Error(t, err, "User should not be found after soft delete")
}

// TestUpdateUser_Success tests successful user update
func TestUpdateUser_Success(t *testing.T) {
	// Setup
	db := setupTestDB(t)
	defer cleanupTestDB(db)

	originalDB := global.DB
	global.DB = db
	defer func() { global.DB = originalDB }()

	// Create a test user
	testUser := &system.SysUser{
		Username: "updateuser",
		Password: "password123",
		Nickname: "Old Nickname",
		RoleID:   1,
		Active:   true,
	}
	db.Create(testUser)

	// Setup router
	router := setupTestRouter()
	userApi := &UserApi{}
	router.PUT("/api/v1/user", userApi.UpdateUser)

	// Update user
	updateReq := UpdateUserRequest{
		ID:       testUser.ID,
		Username: "updateuser",
		Nickname: "New Nickname",
		Email:    "newemail@test.com",
		RoleID:   1,
		Active:   true,
	}
	body, _ := json.Marshal(updateReq)
	req, _ := http.NewRequest("PUT", "/api/v1/user", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response common.Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, 0, response.Code)

	// Verify user was updated
	var updatedUser system.SysUser
	err = db.First(&updatedUser, testUser.ID).Error
	assert.NoError(t, err)
	assert.Equal(t, "New Nickname", updatedUser.Nickname)
	assert.Equal(t, "newemail@test.com", updatedUser.Email)
}

// TestInvalidRequestParameters tests validation of request parameters
func TestInvalidRequestParameters(t *testing.T) {
	// Setup
	db := setupTestDB(t)
	defer cleanupTestDB(db)

	originalDB := global.DB
	global.DB = db
	defer func() { global.DB = originalDB }()

	userApi := &UserApi{}

	testCases := []struct {
		name   string
		method string
		path   string
		body   interface{}
		setup  func(*gin.Engine)
	}{
		{
			name:   "Login - Missing username",
			method: "POST",
			path:   "/api/v1/user/login",
			body:   map[string]string{"password": "test"},
			setup: func(r *gin.Engine) {
				r.POST("/api/v1/user/login", userApi.Login)
			},
		},
		{
			name:   "CreateUser - Missing required fields",
			method: "POST",
			path:   "/api/v1/user",
			body:   map[string]string{"username": "test"},
			setup: func(r *gin.Engine) {
				r.POST("/api/v1/user", userApi.CreateUser)
			},
		},
		{
			name:   "GetUserList - Invalid page parameter",
			method: "GET",
			path:   "/api/v1/user/list?page=0&pageSize=10",
			body:   nil,
			setup: func(r *gin.Engine) {
				r.GET("/api/v1/user/list", userApi.GetUserList)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			testRouter := setupTestRouter()
			tc.setup(testRouter)

			var req *http.Request
			if tc.body != nil {
				body, _ := json.Marshal(tc.body)
				req, _ = http.NewRequest(tc.method, tc.path, bytes.NewBuffer(body))
				req.Header.Set("Content-Type", "application/json")
			} else {
				req, _ = http.NewRequest(tc.method, tc.path, nil)
			}

			w := httptest.NewRecorder()
			testRouter.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)

			var response common.Response
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.NotEqual(t, 0, response.Code, "Expected non-zero error code for invalid parameters")
			assert.NotEmpty(t, response.Msg, "Expected error message")
		})
	}
}
