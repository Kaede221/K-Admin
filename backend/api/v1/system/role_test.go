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

// setupTestDB 设置测试数据库
func setupRoleTestDB(t *testing.T) {
	var err error
	global.DB, err = gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to connect to test database: %v", err)
	}

	// 自动迁移
	err = global.DB.AutoMigrate(&system.SysRole{}, &system.SysUser{}, &system.SysMenu{})
	if err != nil {
		t.Fatalf("failed to migrate test database: %v", err)
	}
}

// TestCreateRole 测试创建角色
func TestCreateRole(t *testing.T) {
	setupRoleTestDB(t)
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		request        CreateRoleRequest
		expectedCode   int
		expectedStatus int
	}{
		{
			name: "成功创建角色",
			request: CreateRoleRequest{
				RoleName:  "测试角色",
				RoleKey:   "test_role",
				DataScope: "all",
				Sort:      1,
				Status:    true,
				Remark:    "测试角色备注",
			},
			expectedCode:   0,
			expectedStatus: http.StatusOK,
		},
		{
			name: "角色键重复",
			request: CreateRoleRequest{
				RoleName:  "测试角色2",
				RoleKey:   "test_role",
				DataScope: "all",
				Sort:      2,
				Status:    true,
			},
			expectedCode:   1,
			expectedStatus: http.StatusOK,
		},
		{
			name: "缺少必填字段",
			request: CreateRoleRequest{
				RoleName: "测试角色3",
				// 缺少 RoleKey
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
			c.Request = httptest.NewRequest("POST", "/api/v1/role", bytes.NewBuffer(body))
			c.Request.Header.Set("Content-Type", "application/json")

			api := RoleApi{}
			api.CreateRole(c)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.Equal(t, float64(tt.expectedCode), response["code"])
		})
	}
}

// TestUpdateRole 测试更新角色
func TestUpdateRole(t *testing.T) {
	setupRoleTestDB(t)
	gin.SetMode(gin.TestMode)

	// 创建测试角色
	role := &system.SysRole{
		RoleName:  "原始角色",
		RoleKey:   "original_role",
		DataScope: "all",
		Sort:      1,
		Status:    true,
	}
	global.DB.Create(role)

	tests := []struct {
		name           string
		request        UpdateRoleRequest
		expectedCode   int
		expectedStatus int
	}{
		{
			name: "成功更新角色",
			request: UpdateRoleRequest{
				ID:        role.ID,
				RoleName:  "更新后的角色",
				RoleKey:   "updated_role",
				DataScope: "dept",
				Sort:      2,
				Status:    false,
				Remark:    "更新备注",
			},
			expectedCode:   0,
			expectedStatus: http.StatusOK,
		},
		{
			name: "更新不存在的角色",
			request: UpdateRoleRequest{
				ID:       9999,
				RoleName: "不存在",
				RoleKey:  "not_exist",
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
			c.Request = httptest.NewRequest("PUT", "/api/v1/role", bytes.NewBuffer(body))
			c.Request.Header.Set("Content-Type", "application/json")

			api := RoleApi{}
			api.UpdateRole(c)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.Equal(t, float64(tt.expectedCode), response["code"])
		})
	}
}

// TestDeleteRole 测试删除角色
func TestDeleteRole(t *testing.T) {
	setupRoleTestDB(t)
	gin.SetMode(gin.TestMode)

	// 创建测试角色
	role1 := &system.SysRole{
		RoleName: "可删除角色",
		RoleKey:  "deletable_role",
	}
	global.DB.Create(role1)

	// 创建有关联用户的角色
	role2 := &system.SysRole{
		RoleName: "有用户的角色",
		RoleKey:  "role_with_users",
	}
	global.DB.Create(role2)

	user := &system.SysUser{
		Username: "testuser",
		Password: "password",
		RoleID:   role2.ID,
	}
	global.DB.Create(user)

	tests := []struct {
		name           string
		roleID         uint
		expectedCode   int
		expectedStatus int
	}{
		{
			name:           "成功删除角色",
			roleID:         role1.ID,
			expectedCode:   0,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "删除有关联用户的角色",
			roleID:         role2.ID,
			expectedCode:   1,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "删除不存在的角色",
			roleID:         9999,
			expectedCode:   1,
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			c.Request = httptest.NewRequest("DELETE", "/api/v1/role/1", nil)
			c.Params = gin.Params{{Key: "id", Value: strconv.FormatUint(uint64(tt.roleID), 10)}}

			api := RoleApi{}
			api.DeleteRole(c)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.Equal(t, float64(tt.expectedCode), response["code"])
		})
	}
}

// TestGetRoleList 测试获取角色列表
func TestGetRoleList(t *testing.T) {
	setupRoleTestDB(t)
	gin.SetMode(gin.TestMode)

	// 创建测试角色
	for i := 1; i <= 5; i++ {
		role := &system.SysRole{
			RoleName: "角色" + string(rune('0'+i)),
			RoleKey:  "role_" + string(rune('0'+i)),
			Sort:     i,
		}
		global.DB.Create(role)
	}

	tests := []struct {
		name           string
		page           int
		pageSize       int
		expectedCode   int
		expectedStatus int
		expectedCount  int
	}{
		{
			name:           "获取第一页",
			page:           1,
			pageSize:       3,
			expectedCode:   0,
			expectedStatus: http.StatusOK,
			expectedCount:  3,
		},
		{
			name:           "获取第二页",
			page:           2,
			pageSize:       3,
			expectedCode:   0,
			expectedStatus: http.StatusOK,
			expectedCount:  2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			c.Request = httptest.NewRequest("GET", "/api/v1/role/list?page="+strconv.Itoa(tt.page)+"&pageSize="+strconv.Itoa(tt.pageSize), nil)

			api := RoleApi{}
			api.GetRoleList(c)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.Equal(t, float64(tt.expectedCode), response["code"])

			if tt.expectedCode == 0 {
				data := response["data"].(map[string]interface{})
				list := data["list"].([]interface{})
				assert.Equal(t, tt.expectedCount, len(list))
			}
		})
	}
}

// TestAssignMenus 测试分配菜单权限
func TestAssignMenus(t *testing.T) {
	setupRoleTestDB(t)
	gin.SetMode(gin.TestMode)

	// 创建测试角色
	role := &system.SysRole{
		RoleName: "测试角色",
		RoleKey:  "test_role",
	}
	global.DB.Create(role)

	// 创建测试菜单
	menu1 := &system.SysMenu{
		Name: "菜单1",
		Path: "/menu1",
	}
	menu2 := &system.SysMenu{
		Name: "菜单2",
		Path: "/menu2",
	}
	global.DB.Create(menu1)
	global.DB.Create(menu2)

	tests := []struct {
		name           string
		request        AssignMenusRequest
		expectedCode   int
		expectedStatus int
	}{
		{
			name: "成功分配菜单",
			request: AssignMenusRequest{
				RoleID:  role.ID,
				MenuIDs: []uint{menu1.ID, menu2.ID},
			},
			expectedCode:   0,
			expectedStatus: http.StatusOK,
		},
		{
			name: "分配给不存在的角色",
			request: AssignMenusRequest{
				RoleID:  9999,
				MenuIDs: []uint{menu1.ID},
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
			c.Request = httptest.NewRequest("POST", "/api/v1/role/assign-menus", bytes.NewBuffer(body))
			c.Request.Header.Set("Content-Type", "application/json")

			api := RoleApi{}
			api.AssignMenus(c)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.Equal(t, float64(tt.expectedCode), response["code"])
		})
	}
}
