package system

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"k-admin-system/global"
	"k-admin-system/model/system"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupMenuAPITestDB 设置测试数据库
func setupMenuAPITestDB(t *testing.T) {
	var err error
	global.DB, err = gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to connect to test database: %v", err)
	}

	// 自动迁移
	err = global.DB.AutoMigrate(&system.SysMenu{}, &system.SysRole{})
	if err != nil {
		t.Fatalf("failed to migrate test database: %v", err)
	}
}

// TestCreateMenu 测试创建菜单
func TestCreateMenu(t *testing.T) {
	setupMenuAPITestDB(t)
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		request        CreateMenuRequest
		expectedCode   int
		expectedStatus int
	}{
		{
			name: "成功创建菜单",
			request: CreateMenuRequest{
				Name:      "Dashboard",
				Path:      "/dashboard",
				Component: "views/dashboard/index",
				Sort:      1,
				Meta:      `{"icon":"dashboard","title":"Dashboard"}`,
			},
			expectedCode:   0,
			expectedStatus: http.StatusOK,
		},
		{
			name: "缺少必填字段",
			request: CreateMenuRequest{
				Path: "/test",
				// 缺少 Name
			},
			expectedCode:   1,
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			body, _ := json.Marshal(tt.request)
			c.Request = httptest.NewRequest("POST", "/api/v1/menu", bytes.NewBuffer(body))
			c.Request.Header.Set("Content-Type", "application/json")

			api := MenuApi{}
			api.CreateMenu(c)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.Equal(t, float64(tt.expectedCode), response["code"])
		})
	}
}

// TestUpdateMenu 测试更新菜单
func TestUpdateMenu(t *testing.T) {
	setupMenuAPITestDB(t)
	gin.SetMode(gin.TestMode)

	// 创建测试菜单
	menu := &system.SysMenu{
		Name:      "Original",
		Path:      "/original",
		Component: "views/original/index",
		Sort:      1,
	}
	global.DB.Create(menu)

	tests := []struct {
		name           string
		request        UpdateMenuRequest
		expectedCode   int
		expectedStatus int
	}{
		{
			name: "成功更新菜单",
			request: UpdateMenuRequest{
				ID:        menu.ID,
				Name:      "Updated",
				Path:      "/updated",
				Component: "views/updated/index",
				Sort:      2,
			},
			expectedCode:   0,
			expectedStatus: http.StatusOK,
		},
		{
			name: "更新不存在的菜单",
			request: UpdateMenuRequest{
				ID:   9999,
				Name: "NotExist",
				Path: "/notexist",
			},
			expectedCode:   1,
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			body, _ := json.Marshal(tt.request)
			c.Request = httptest.NewRequest("PUT", "/api/v1/menu", bytes.NewBuffer(body))
			c.Request.Header.Set("Content-Type", "application/json")

			api := MenuApi{}
			api.UpdateMenu(c)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.Equal(t, float64(tt.expectedCode), response["code"])
		})
	}
}

// TestDeleteMenu 测试删除菜单
func TestDeleteMenu(t *testing.T) {
	setupMenuAPITestDB(t)
	gin.SetMode(gin.TestMode)

	// 创建测试菜单
	menu1 := &system.SysMenu{
		Name: "Deletable",
		Path: "/deletable",
	}
	global.DB.Create(menu1)

	// 创建有子菜单的菜单
	menu2 := &system.SysMenu{
		Name: "Parent",
		Path: "/parent",
	}
	global.DB.Create(menu2)

	childMenu := &system.SysMenu{
		Name:     "Child",
		Path:     "/parent/child",
		ParentID: menu2.ID,
	}
	global.DB.Create(childMenu)

	tests := []struct {
		name           string
		menuID         uint
		expectedCode   int
		expectedStatus int
	}{
		{
			name:           "成功删除菜单",
			menuID:         menu1.ID,
			expectedCode:   0,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "删除有子菜单的菜单",
			menuID:         menu2.ID,
			expectedCode:   1,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "删除不存在的菜单",
			menuID:         9999,
			expectedCode:   1,
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			c.Request = httptest.NewRequest("DELETE", "/api/v1/menu/1", nil)
			c.Params = gin.Params{{Key: "id", Value: strconv.FormatUint(uint64(tt.menuID), 10)}}

			api := MenuApi{}
			api.DeleteMenu(c)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.Equal(t, float64(tt.expectedCode), response["code"])
		})
	}
}

// TestGetMenuTree 测试获取菜单树
func TestGetMenuTree(t *testing.T) {
	setupMenuAPITestDB(t)
	gin.SetMode(gin.TestMode)

	// 创建层级菜单
	parent := &system.SysMenu{
		Name: "System",
		Path: "/system",
		Sort: 1,
	}
	global.DB.Create(parent)

	child1 := &system.SysMenu{
		Name:     "Users",
		Path:     "/system/users",
		ParentID: parent.ID,
		Sort:     1,
	}
	child2 := &system.SysMenu{
		Name:     "Roles",
		Path:     "/system/roles",
		ParentID: parent.ID,
		Sort:     2,
	}
	global.DB.Create(child1)
	global.DB.Create(child2)

	// 创建角色并分配菜单
	role := &system.SysRole{
		RoleName: "Admin",
		RoleKey:  "admin",
	}
	global.DB.Create(role)
	global.DB.Model(role).Association("Menus").Append([]system.SysMenu{*parent, *child1, *child2})

	tests := []struct {
		name           string
		roleID         uint
		expectedCode   int
		expectedStatus int
		checkTree      bool
	}{
		{
			name:           "获取所有菜单树",
			roleID:         0,
			expectedCode:   0,
			expectedStatus: http.StatusOK,
			checkTree:      true,
		},
		{
			name:           "根据角色获取菜单树",
			roleID:         role.ID,
			expectedCode:   0,
			expectedStatus: http.StatusOK,
			checkTree:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			url := "/api/v1/menu/tree"
			if tt.roleID > 0 {
				url += "?roleId=" + strconv.FormatUint(uint64(tt.roleID), 10)
			}
			c.Request = httptest.NewRequest("GET", url, nil)

			api := MenuApi{}
			api.GetMenuTree(c)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.Equal(t, float64(tt.expectedCode), response["code"])

			if tt.checkTree && tt.expectedCode == 0 {
				data := response["data"].([]interface{})
				assert.Greater(t, len(data), 0, "Should have menu items")
			}
		})
	}
}

// TestGetAllMenus 测试获取所有菜单
func TestGetAllMenus(t *testing.T) {
	setupMenuAPITestDB(t)
	gin.SetMode(gin.TestMode)

	// 创建测试菜单
	for i := 1; i <= 3; i++ {
		menu := &system.SysMenu{
			Name: "Menu" + strconv.Itoa(i),
			Path: "/menu" + strconv.Itoa(i),
			Sort: i,
		}
		global.DB.Create(menu)
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Request = httptest.NewRequest("GET", "/api/v1/menu/all", nil)

	api := MenuApi{}
	api.GetAllMenus(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, float64(0), response["code"])

	data := response["data"].([]interface{})
	assert.Equal(t, 3, len(data), "Should have 3 menus")
}

// TestMenuHierarchy 测试菜单层级结构
func TestMenuHierarchy(t *testing.T) {
	setupMenuAPITestDB(t)
	gin.SetMode(gin.TestMode)

	// 创建三层菜单结构
	level1 := &system.SysMenu{
		Name: "Level1",
		Path: "/level1",
		Sort: 1,
	}
	global.DB.Create(level1)

	level2 := &system.SysMenu{
		Name:     "Level2",
		Path:     "/level1/level2",
		ParentID: level1.ID,
		Sort:     1,
	}
	global.DB.Create(level2)

	level3 := &system.SysMenu{
		Name:     "Level3",
		Path:     "/level1/level2/level3",
		ParentID: level2.ID,
		Sort:     1,
	}
	global.DB.Create(level3)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Request = httptest.NewRequest("GET", "/api/v1/menu/tree", nil)

	api := MenuApi{}
	api.GetMenuTree(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, float64(0), response["code"])

	// 验证树结构
	data := response["data"].([]interface{})
	assert.Equal(t, 1, len(data), "Should have 1 root menu")

	rootMenu := data[0].(map[string]interface{})
	assert.Equal(t, "Level1", rootMenu["name"])

	children := rootMenu["children"].([]interface{})
	assert.Equal(t, 1, len(children), "Level1 should have 1 child")

	level2Menu := children[0].(map[string]interface{})
	assert.Equal(t, "Level2", level2Menu["name"])

	grandchildren := level2Menu["children"].([]interface{})
	assert.Equal(t, 1, len(grandchildren), "Level2 should have 1 child")

	level3Menu := grandchildren[0].(map[string]interface{})
	assert.Equal(t, "Level3", level3Menu["name"])
}

// TestMenuSorting 测试菜单排序
func TestMenuSorting(t *testing.T) {
	setupMenuAPITestDB(t)
	gin.SetMode(gin.TestMode)

	// 创建不同排序的菜单
	menu3 := &system.SysMenu{Name: "Menu3", Path: "/menu3", Sort: 3}
	menu1 := &system.SysMenu{Name: "Menu1", Path: "/menu1", Sort: 1}
	menu2 := &system.SysMenu{Name: "Menu2", Path: "/menu2", Sort: 2}

	global.DB.Create(menu3)
	global.DB.Create(menu1)
	global.DB.Create(menu2)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Request = httptest.NewRequest("GET", "/api/v1/menu/all", nil)

	api := MenuApi{}
	api.GetAllMenus(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	data := response["data"].([]interface{})
	assert.Equal(t, 3, len(data))

	// 验证排序
	firstMenu := data[0].(map[string]interface{})
	assert.Equal(t, "Menu1", firstMenu["name"])

	secondMenu := data[1].(map[string]interface{})
	assert.Equal(t, "Menu2", secondMenu["name"])

	thirdMenu := data[2].(map[string]interface{})
	assert.Equal(t, "Menu3", thirdMenu["name"])
}
