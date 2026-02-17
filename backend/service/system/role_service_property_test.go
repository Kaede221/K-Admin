package system

import (
	"fmt"
	"testing"

	"k-admin-system/global"
	"k-admin-system/model/system"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupRoleTestDB initializes an in-memory SQLite database for role testing
func setupRoleTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// Run migrations
	err = db.AutoMigrate(&system.SysRole{}, &system.SysUser{}, &system.SysMenu{})
	if err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}

	return db
}

// cleanupRoleTestDB cleans all data from the test database
func cleanupRoleTestDB(db *gorm.DB) {
	db.Exec("DELETE FROM sys_role_menus")
	db.Exec("DELETE FROM sys_roles")
	db.Exec("DELETE FROM sys_users")
	db.Exec("DELETE FROM sys_menus")
}

// genRoleName generates valid role names (1-50 chars)
func genRoleName() gopter.Gen {
	return gen.AlphaString().SuchThat(func(s string) bool {
		return len(s) >= 1 && len(s) <= 50
	}).Map(func(s string) string {
		if s == "" {
			return "role"
		}
		if len(s) > 50 {
			return s[:50]
		}
		return s
	})
}

// genRoleKey generates valid role keys (1-50 chars, alphanumeric)
func genRoleKey() gopter.Gen {
	return gen.AlphaString().SuchThat(func(s string) bool {
		return len(s) >= 1 && len(s) <= 50
	}).Map(func(s string) string {
		if s == "" {
			return "role_key"
		}
		if len(s) > 50 {
			return s[:50]
		}
		return s
	})
}

// Feature: k-admin-system
// Property 13: Role Deletion Protection
// For any role that has associated users, deletion attempts SHALL fail with an error
// indicating the role is in use
// **Validates: Requirements 5.6**
func TestProperty13_RoleDeletionProtection(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	properties.Property("role with users cannot be deleted", prop.ForAll(
		func(seed int) bool {
			// Generate unique values from seed
			roleName := fmt.Sprintf("Role%d", seed)
			roleKey := fmt.Sprintf("role_key_%d", seed)
			username := fmt.Sprintf("user%d", seed)
			// Setup test database
			db := setupRoleTestDB(t)
			defer cleanupRoleTestDB(db)

			// Set global DB for service
			originalDB := global.DB
			global.DB = db
			defer func() { global.DB = originalDB }()

			roleService := &RoleService{}
			userService := &UserService{}

			// Create a role
			role := &system.SysRole{
				RoleName:  roleName,
				RoleKey:   roleKey,
				DataScope: "all",
				Sort:      0,
				Status:    true,
			}

			err := roleService.CreateRole(role)
			if err != nil {
				t.Logf("Failed to create role: %v", err)
				return false
			}

			roleID := role.ID

			// Create a user associated with this role
			user := &system.SysUser{
				Username: username,
				Password: "password123",
				RoleID:   roleID,
				Active:   true,
			}

			err = userService.CreateUser(user)
			if err != nil {
				t.Logf("Failed to create user: %v", err)
				return false
			}

			// Attempt to delete the role
			err = roleService.DeleteRole(roleID)
			if err == nil {
				t.Logf("Role deletion should have failed but succeeded")
				return false
			}

			// Verify error message indicates role is in use
			if err.Error() != "cannot delete role with associated users" {
				t.Logf("Unexpected error message: %v", err)
				return false
			}

			// Verify role still exists
			retrievedRole, err := roleService.GetRoleByID(roleID)
			if err != nil {
				t.Logf("Role was deleted despite having users: %v", err)
				return false
			}
			if retrievedRole.ID != roleID {
				t.Logf("Retrieved role ID mismatch")
				return false
			}

			return true
		},
		gen.IntRange(1, 10000),
	))

	properties.Property("role without users can be deleted", prop.ForAll(
		func(seed int) bool {
			// Generate unique values from seed
			roleName := fmt.Sprintf("Role%d", seed)
			roleKey := fmt.Sprintf("role_key_%d", seed)
			// Setup test database
			db := setupRoleTestDB(t)
			defer cleanupRoleTestDB(db)

			// Set global DB for service
			originalDB := global.DB
			global.DB = db
			defer func() { global.DB = originalDB }()

			roleService := &RoleService{}

			// Create a role
			role := &system.SysRole{
				RoleName:  roleName,
				RoleKey:   roleKey,
				DataScope: "all",
				Sort:      0,
				Status:    true,
			}

			err := roleService.CreateRole(role)
			if err != nil {
				t.Logf("Failed to create role: %v", err)
				return false
			}

			roleID := role.ID

			// Delete the role (no users associated)
			err = roleService.DeleteRole(roleID)
			if err != nil {
				t.Logf("Failed to delete role without users: %v", err)
				return false
			}

			// Verify role no longer exists
			_, err = roleService.GetRoleByID(roleID)
			if err == nil {
				t.Logf("Deleted role still retrievable")
				return false
			}

			return true
		},
		gen.IntRange(1, 10000),
	))

	properties.TestingRun(t)
}

// Feature: k-admin-system
// Property 14: Role Permission Assignment
// For any role with assigned menu IDs, querying the role's menus SHALL return exactly those menu IDs,
// and users with that role SHALL have access to those menus
// **Validates: Requirements 5.3, 5.4**
func TestProperty14_RolePermissionAssignment(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	properties.Property("assigned menus are retrievable", prop.ForAll(
		func(seed int, menuCount int) bool {
			// Generate unique values from seed
			roleName := fmt.Sprintf("Role%d", seed)
			roleKey := fmt.Sprintf("role_key_%d", seed)
			// Setup test database
			db := setupRoleTestDB(t)
			defer cleanupRoleTestDB(db)

			// Set global DB for service
			originalDB := global.DB
			global.DB = db
			defer func() { global.DB = originalDB }()

			roleService := &RoleService{}

			// Create a role
			role := &system.SysRole{
				RoleName:  roleName,
				RoleKey:   roleKey,
				DataScope: "all",
				Sort:      0,
				Status:    true,
			}

			err := roleService.CreateRole(role)
			if err != nil {
				t.Logf("Failed to create role: %v", err)
				return false
			}

			roleID := role.ID

			// Create menus
			menuIDs := make([]uint, menuCount)
			for i := 0; i < menuCount; i++ {
				menu := &system.SysMenu{
					ParentID:  0,
					Path:      fmt.Sprintf("/menu%d", i),
					Name:      fmt.Sprintf("Menu%d", i),
					Component: fmt.Sprintf("views/menu%d", i),
					Sort:      i,
					Meta:      `{"title":"Menu"}`,
					BtnPerms:  `[]`,
				}
				err := db.Create(menu).Error
				if err != nil {
					t.Logf("Failed to create menu %d: %v", i, err)
					return false
				}
				menuIDs[i] = menu.ID
			}

			// Assign menus to role
			err = roleService.AssignMenus(roleID, menuIDs)
			if err != nil {
				t.Logf("Failed to assign menus: %v", err)
				return false
			}

			// Retrieve role menus
			retrievedMenuIDs, err := roleService.GetRoleMenus(roleID)
			if err != nil {
				t.Logf("Failed to get role menus: %v", err)
				return false
			}

			// Verify retrieved menu IDs match assigned menu IDs
			if len(retrievedMenuIDs) != len(menuIDs) {
				t.Logf("Menu count mismatch: expected %d, got %d", len(menuIDs), len(retrievedMenuIDs))
				return false
			}

			// Create a map for easier lookup
			menuIDMap := make(map[uint]bool)
			for _, id := range menuIDs {
				menuIDMap[id] = true
			}

			// Verify all retrieved IDs are in the assigned set
			for _, id := range retrievedMenuIDs {
				if !menuIDMap[id] {
					t.Logf("Retrieved menu ID %d was not assigned", id)
					return false
				}
			}

			return true
		},
		gen.IntRange(1, 10000),
		gen.IntRange(0, 10),
	))

	properties.Property("menu assignment is idempotent", prop.ForAll(
		func(seed int, menuCount int) bool {
			// Generate unique values from seed
			roleName := fmt.Sprintf("Role%d", seed)
			roleKey := fmt.Sprintf("role_key_%d", seed)
			// Setup test database
			db := setupRoleTestDB(t)
			defer cleanupRoleTestDB(db)

			// Set global DB for service
			originalDB := global.DB
			global.DB = db
			defer func() { global.DB = originalDB }()

			roleService := &RoleService{}

			// Create a role
			role := &system.SysRole{
				RoleName:  roleName,
				RoleKey:   roleKey,
				DataScope: "all",
				Sort:      0,
				Status:    true,
			}

			err := roleService.CreateRole(role)
			if err != nil {
				t.Logf("Failed to create role: %v", err)
				return false
			}

			roleID := role.ID

			// Create menus
			menuIDs := make([]uint, menuCount)
			for i := 0; i < menuCount; i++ {
				menu := &system.SysMenu{
					ParentID:  0,
					Path:      fmt.Sprintf("/menu%d", i),
					Name:      fmt.Sprintf("Menu%d", i),
					Component: fmt.Sprintf("views/menu%d", i),
					Sort:      i,
					Meta:      `{"title":"Menu"}`,
					BtnPerms:  `[]`,
				}
				err := db.Create(menu).Error
				if err != nil {
					t.Logf("Failed to create menu %d: %v", i, err)
					return false
				}
				menuIDs[i] = menu.ID
			}

			// Assign menus twice
			err = roleService.AssignMenus(roleID, menuIDs)
			if err != nil {
				t.Logf("Failed to assign menus first time: %v", err)
				return false
			}

			err = roleService.AssignMenus(roleID, menuIDs)
			if err != nil {
				t.Logf("Failed to assign menus second time: %v", err)
				return false
			}

			// Retrieve role menus
			retrievedMenuIDs, err := roleService.GetRoleMenus(roleID)
			if err != nil {
				t.Logf("Failed to get role menus: %v", err)
				return false
			}

			// Verify no duplicates
			if len(retrievedMenuIDs) != len(menuIDs) {
				t.Logf("Menu count mismatch after double assignment: expected %d, got %d", len(menuIDs), len(retrievedMenuIDs))
				return false
			}

			return true
		},
		gen.IntRange(1, 10000),
		gen.IntRange(1, 10),
	))

	properties.Property("menu reassignment replaces old menus", prop.ForAll(
		func(seed int, firstCount int, secondCount int) bool {
			// Generate unique values from seed
			roleName := fmt.Sprintf("Role%d", seed)
			roleKey := fmt.Sprintf("role_key_%d", seed)
			// Setup test database
			db := setupRoleTestDB(t)
			defer cleanupRoleTestDB(db)

			// Set global DB for service
			originalDB := global.DB
			global.DB = db
			defer func() { global.DB = originalDB }()

			roleService := &RoleService{}

			// Create a role
			role := &system.SysRole{
				RoleName:  roleName,
				RoleKey:   roleKey,
				DataScope: "all",
				Sort:      0,
				Status:    true,
			}

			err := roleService.CreateRole(role)
			if err != nil {
				t.Logf("Failed to create role: %v", err)
				return false
			}

			roleID := role.ID

			// Create first set of menus
			firstMenuIDs := make([]uint, firstCount)
			for i := 0; i < firstCount; i++ {
				menu := &system.SysMenu{
					ParentID:  0,
					Path:      fmt.Sprintf("/menu_first_%d", i),
					Name:      fmt.Sprintf("MenuFirst%d", i),
					Component: fmt.Sprintf("views/menu_first_%d", i),
					Sort:      i,
					Meta:      `{"title":"Menu"}`,
					BtnPerms:  `[]`,
				}
				err := db.Create(menu).Error
				if err != nil {
					t.Logf("Failed to create first menu %d: %v", i, err)
					return false
				}
				firstMenuIDs[i] = menu.ID
			}

			// Assign first set of menus
			err = roleService.AssignMenus(roleID, firstMenuIDs)
			if err != nil {
				t.Logf("Failed to assign first menus: %v", err)
				return false
			}

			// Create second set of menus
			secondMenuIDs := make([]uint, secondCount)
			for i := 0; i < secondCount; i++ {
				menu := &system.SysMenu{
					ParentID:  0,
					Path:      fmt.Sprintf("/menu_second_%d", i),
					Name:      fmt.Sprintf("MenuSecond%d", i),
					Component: fmt.Sprintf("views/menu_second_%d", i),
					Sort:      i,
					Meta:      `{"title":"Menu"}`,
					BtnPerms:  `[]`,
				}
				err := db.Create(menu).Error
				if err != nil {
					t.Logf("Failed to create second menu %d: %v", i, err)
					return false
				}
				secondMenuIDs[i] = menu.ID
			}

			// Assign second set of menus (should replace first set)
			err = roleService.AssignMenus(roleID, secondMenuIDs)
			if err != nil {
				t.Logf("Failed to assign second menus: %v", err)
				return false
			}

			// Retrieve role menus
			retrievedMenuIDs, err := roleService.GetRoleMenus(roleID)
			if err != nil {
				t.Logf("Failed to get role menus: %v", err)
				return false
			}

			// Verify only second set of menus is assigned
			if len(retrievedMenuIDs) != len(secondMenuIDs) {
				t.Logf("Menu count mismatch: expected %d, got %d", len(secondMenuIDs), len(retrievedMenuIDs))
				return false
			}

			// Create a map for second menu IDs
			secondMenuIDMap := make(map[uint]bool)
			for _, id := range secondMenuIDs {
				secondMenuIDMap[id] = true
			}

			// Verify all retrieved IDs are from second set
			for _, id := range retrievedMenuIDs {
				if !secondMenuIDMap[id] {
					t.Logf("Retrieved menu ID %d is not from second set", id)
					return false
				}
			}

			// Verify no IDs from first set (unless they overlap with second set)
			firstMenuIDMap := make(map[uint]bool)
			for _, id := range firstMenuIDs {
				firstMenuIDMap[id] = true
			}

			for _, id := range retrievedMenuIDs {
				if firstMenuIDMap[id] && !secondMenuIDMap[id] {
					t.Logf("Retrieved menu ID %d is from first set but not second set", id)
					return false
				}
			}

			return true
		},
		gen.IntRange(1, 10000),
		gen.IntRange(1, 5),
		gen.IntRange(1, 5),
	))

	properties.TestingRun(t)
}
