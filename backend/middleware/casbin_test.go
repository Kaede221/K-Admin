package middleware

import (
	"k-admin-system/config"
	"k-admin-system/core"
	"k-admin-system/global"
	"k-admin-system/model/system"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupCasbinTestDB 设置测试数据库
func setupCasbinTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// 自动迁移
	err = db.AutoMigrate(&system.SysRole{}, &system.SysCasbinRule{})
	if err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}

	return db
}

// setupTestLogger 设置测试日志
func setupTestLogger(t *testing.T) {
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
}

// setupTestCasbin 设置测试Casbin enforcer
func setupTestCasbin(t *testing.T) {
	// 使用内存模型初始化Casbin
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
}

// TestCasbinAuth_Authorized 测试授权访问
func TestCasbinAuth_Authorized(t *testing.T) {
	// 设置测试环境
	gin.SetMode(gin.TestMode)
	global.DB = setupCasbinTestDB(t)
	setupTestLogger(t)
	setupTestCasbin(t)

	// 创建测试角色
	role := &system.SysRole{
		RoleName: "Admin",
		RoleKey:  "admin",
		Status:   true,
	}
	global.DB.Create(role)

	// 添加Casbin策略
	_, err := global.CasbinEnforcer.AddPolicy("admin", "/api/v1/user", "GET")
	assert.NoError(t, err)

	// 创建测试路由
	router := gin.New()
	router.Use(func(c *gin.Context) {
		// 模拟JWT中间件设置的roleId
		c.Set("roleId", role.ID)
		c.Next()
	})
	router.Use(CasbinAuth())
	router.GET("/api/v1/user", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "success"})
	})

	// 发送测试请求
	req, _ := http.NewRequest("GET", "/api/v1/user", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 验证响应
	assert.Equal(t, 200, w.Code)
}

// TestCasbinAuth_Unauthorized 测试未授权访问返回403
func TestCasbinAuth_Unauthorized(t *testing.T) {
	// 设置测试环境
	gin.SetMode(gin.TestMode)
	global.DB = setupCasbinTestDB(t)
	setupTestLogger(t)
	setupTestCasbin(t)

	// 创建测试角色（没有权限）
	role := &system.SysRole{
		RoleName: "User",
		RoleKey:  "user",
		Status:   true,
	}
	global.DB.Create(role)

	// 不添加任何策略

	// 创建测试路由
	router := gin.New()
	router.Use(func(c *gin.Context) {
		// 模拟JWT中间件设置的roleId
		c.Set("roleId", role.ID)
		c.Next()
	})
	router.Use(CasbinAuth())
	router.GET("/api/v1/admin", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "success"})
	})

	// 发送测试请求
	req, _ := http.NewRequest("GET", "/api/v1/admin", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 验证响应 - 应该返回403
	assert.Equal(t, 200, w.Code) // 统一响应格式，HTTP状态码为200
	assert.Contains(t, w.Body.String(), "无权访问")
}

// TestCasbinAuth_MissingRoleId 测试缺少roleId
func TestCasbinAuth_MissingRoleId(t *testing.T) {
	// 设置测试环境
	gin.SetMode(gin.TestMode)
	global.DB = setupCasbinTestDB(t)
	setupTestLogger(t)
	setupTestCasbin(t)

	// 创建测试路由
	router := gin.New()
	// 不设置roleId
	router.Use(CasbinAuth())
	router.GET("/api/v1/user", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "success"})
	})

	// 发送测试请求
	req, _ := http.NewRequest("GET", "/api/v1/user", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 验证响应 - 应该返回401
	assert.Equal(t, 200, w.Code) // 统一响应格式，HTTP状态码为200
	assert.Contains(t, w.Body.String(), "未找到角色信息")
}

// TestCasbinAuth_RoleNotFound 测试角色不存在
func TestCasbinAuth_RoleNotFound(t *testing.T) {
	// 设置测试环境
	gin.SetMode(gin.TestMode)
	global.DB = setupCasbinTestDB(t)
	setupTestLogger(t)
	setupTestCasbin(t)

	// 创建测试路由
	router := gin.New()
	router.Use(func(c *gin.Context) {
		// 设置一个不存在的roleId
		c.Set("roleId", uint(9999))
		c.Next()
	})
	router.Use(CasbinAuth())
	router.GET("/api/v1/user", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "success"})
	})

	// 发送测试请求
	req, _ := http.NewRequest("GET", "/api/v1/user", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 验证响应 - 应该返回403
	assert.Equal(t, 200, w.Code) // 统一响应格式，HTTP状态码为200
	assert.Contains(t, w.Body.String(), "角色不存在")
}

// TestCasbinAuth_RESTfulPathMatching 测试RESTful路径匹配
func TestCasbinAuth_RESTfulPathMatching(t *testing.T) {
	// 设置测试环境
	gin.SetMode(gin.TestMode)
	global.DB = setupCasbinTestDB(t)
	setupTestLogger(t)
	setupTestCasbin(t)

	// 创建测试角色
	role := &system.SysRole{
		RoleName: "Admin",
		RoleKey:  "admin",
		Status:   true,
	}
	global.DB.Create(role)

	// 添加带参数的Casbin策略
	_, err := global.CasbinEnforcer.AddPolicy("admin", "/api/v1/user/:id", "GET")
	assert.NoError(t, err)

	// 创建测试路由
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("roleId", role.ID)
		c.Next()
	})
	router.Use(CasbinAuth())
	router.GET("/api/v1/user/:id", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "success"})
	})

	// 测试不同的ID参数
	testCases := []struct {
		path         string
		expectedCode int
	}{
		{"/api/v1/user/123", 200},
		{"/api/v1/user/456", 200},
		{"/api/v1/user/abc", 200},
	}

	for _, tc := range testCases {
		req, _ := http.NewRequest("GET", tc.path, nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, tc.expectedCode, w.Code, "Path: "+tc.path)
	}
}

// TestCasbinAuth_DifferentMethods 测试不同HTTP方法
func TestCasbinAuth_DifferentMethods(t *testing.T) {
	// 设置测试环境
	gin.SetMode(gin.TestMode)
	global.DB = setupCasbinTestDB(t)
	setupTestLogger(t)
	setupTestCasbin(t)

	// 创建测试角色
	role := &system.SysRole{
		RoleName: "Admin",
		RoleKey:  "admin",
		Status:   true,
	}
	global.DB.Create(role)

	// 只允许GET方法
	_, err := global.CasbinEnforcer.AddPolicy("admin", "/api/v1/user", "GET")
	assert.NoError(t, err)

	// 创建测试路由
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("roleId", role.ID)
		c.Next()
	})
	router.Use(CasbinAuth())
	router.GET("/api/v1/user", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "success"})
	})
	router.POST("/api/v1/user", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "success"})
	})

	// 测试GET请求 - 应该成功
	req, _ := http.NewRequest("GET", "/api/v1/user", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)
	assert.Contains(t, w.Body.String(), "success")

	// 测试POST请求 - 应该失败
	req, _ = http.NewRequest("POST", "/api/v1/user", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code) // 统一响应格式
	assert.Contains(t, w.Body.String(), "无权访问")
}

// TestCasbinAuth_RoleInheritance 测试角色继承
func TestCasbinAuth_RoleInheritance(t *testing.T) {
	// 设置测试环境
	gin.SetMode(gin.TestMode)
	global.DB = setupCasbinTestDB(t)
	setupTestLogger(t)
	setupTestCasbin(t)

	// 创建父角色
	parentRole := &system.SysRole{
		RoleName: "Admin",
		RoleKey:  "admin",
		Status:   true,
	}
	global.DB.Create(parentRole)

	// 创建子角色
	childRole := &system.SysRole{
		RoleName: "Manager",
		RoleKey:  "manager",
		Status:   true,
	}
	global.DB.Create(childRole)

	// 为父角色添加权限
	_, err := global.CasbinEnforcer.AddPolicy("admin", "/api/v1/admin", "GET")
	assert.NoError(t, err)

	// 为子角色添加自己的权限
	_, err = global.CasbinEnforcer.AddPolicy("manager", "/api/v1/manager", "GET")
	assert.NoError(t, err)

	// 设置角色继承关系：manager继承admin
	_, err = global.CasbinEnforcer.AddGroupingPolicy("manager", "admin")
	assert.NoError(t, err)

	// 创建测试路由
	router := gin.New()
	router.Use(func(c *gin.Context) {
		// 使用子角色
		c.Set("roleId", childRole.ID)
		c.Next()
	})
	router.Use(CasbinAuth())
	router.GET("/api/v1/admin", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "admin success"})
	})
	router.GET("/api/v1/manager", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "manager success"})
	})

	// 测试子角色访问父角色的资源 - 应该成功
	req, _ := http.NewRequest("GET", "/api/v1/admin", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)
	assert.Contains(t, w.Body.String(), "admin success")

	// 测试子角色访问自己的资源 - 应该成功
	req, _ = http.NewRequest("GET", "/api/v1/manager", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)
	assert.Contains(t, w.Body.String(), "manager success")
}
