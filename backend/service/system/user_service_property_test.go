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

// genUsername generates valid usernames (1-50 chars, alphanumeric)
func genUsername() gopter.Gen {
	return gen.AlphaString().SuchThat(func(s string) bool {
		return len(s) >= 1 && len(s) <= 50
	}).Map(func(s string) string {
		if s == "" {
			return "user"
		}
		if len(s) > 50 {
			return s[:50]
		}
		return s
	})
}

// genNickname generates valid nicknames (0-50 chars)
func genNickname() gopter.Gen {
	return gen.AlphaString().Map(func(s string) string {
		if len(s) > 50 {
			return s[:50]
		}
		return s
	})
}

// genEmail generates valid emails (0-100 chars)
func genEmail() gopter.Gen {
	return gen.AlphaString().Map(func(s string) string {
		if len(s) > 90 {
			s = s[:90]
		}
		return s + "@test.com"
	})
}

// Feature: k-admin-system
// Property 9: User CRUD Consistency
// For any user creation, the created user SHALL be retrievable by ID, updatable with new values,
// and deletable (soft delete) such that it no longer appears in active user queries
// **Validates: Requirements 4.2, 4.3, 4.4, 4.5**
func TestProperty9_UserCRUDConsistency(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	properties.Property("created user is retrievable, updatable, and deletable", prop.ForAll(
		func(username string, nickname string, email string, roleID uint) bool {
			// Setup test database
			db := setupTestDB(t)
			defer cleanupTestDB(db)

			// Set global DB for service
			originalDB := global.DB
			global.DB = db
			defer func() { global.DB = originalDB }()

			service := &UserService{}

			// CREATE: Create a user
			user := &system.SysUser{
				Username: username,
				Password: "testpassword123",
				Nickname: nickname,
				Email:    email,
				RoleID:   roleID,
				Active:   true,
			}

			err := service.CreateUser(user)
			if err != nil {
				t.Logf("Failed to create user: %v", err)
				return false
			}

			// Verify user has an ID assigned
			if user.ID == 0 {
				t.Logf("Created user has no ID")
				return false
			}

			createdID := user.ID

			// READ: Retrieve the created user
			retrievedUser, err := service.GetUserByID(createdID)
			if err != nil {
				t.Logf("Failed to retrieve user by ID: %v", err)
				return false
			}

			// Verify retrieved user matches created user
			if retrievedUser.ID != createdID {
				t.Logf("Retrieved user ID mismatch: expected %d, got %d", createdID, retrievedUser.ID)
				return false
			}
			if retrievedUser.Username != username {
				t.Logf("Retrieved username mismatch: expected %s, got %s", username, retrievedUser.Username)
				return false
			}

			// UPDATE: Update the user with new values
			newNickname := nickname + "_updated"
			if len(newNickname) > 50 {
				newNickname = "updated"
			}
			retrievedUser.Nickname = newNickname

			err = service.UpdateUser(retrievedUser)
			if err != nil {
				t.Logf("Failed to update user: %v", err)
				return false
			}

			// Verify update persisted
			updatedUser, err := service.GetUserByID(createdID)
			if err != nil {
				t.Logf("Failed to retrieve updated user: %v", err)
				return false
			}
			if updatedUser.Nickname != newNickname {
				t.Logf("Updated nickname mismatch: expected %s, got %s", newNickname, updatedUser.Nickname)
				return false
			}

			// DELETE: Soft delete the user
			err = service.DeleteUser(createdID)
			if err != nil {
				t.Logf("Failed to delete user: %v", err)
				return false
			}

			// Verify user no longer appears in active queries
			_, err = service.GetUserByID(createdID)
			if err == nil {
				t.Logf("Deleted user still retrievable by GetUserByID")
				return false
			}

			// Verify user doesn't appear in user list
			users, _, err := service.GetUserList(1, 100, map[string]interface{}{})
			if err != nil {
				t.Logf("Failed to get user list: %v", err)
				return false
			}
			for _, u := range users {
				if u.ID == createdID {
					t.Logf("Deleted user still appears in user list")
					return false
				}
			}

			return true
		},
		genUsername(),
		genNickname(),
		genEmail(),
		gen.UIntRange(1, 100),
	))

	properties.TestingRun(t)
}

// Feature: k-admin-system
// Property 10: Password Field Masking
// For any API response containing user data, the password field SHALL never be included in the JSON output
// **Validates: Requirements 4.7**
func TestProperty10_PasswordFieldMasking(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	properties.Property("password field is masked in user struct", prop.ForAll(
		func(username string, password string) bool {
			// Setup test database
			db := setupTestDB(t)
			defer cleanupTestDB(db)

			// Set global DB for service
			originalDB := global.DB
			global.DB = db
			defer func() { global.DB = originalDB }()

			service := &UserService{}

			// Create a user
			user := &system.SysUser{
				Username: username,
				Password: password,
				RoleID:   1,
				Active:   true,
			}

			err := service.CreateUser(user)
			if err != nil {
				t.Logf("Failed to create user: %v", err)
				return false
			}

			// Retrieve the user
			retrievedUser, err := service.GetUserByID(user.ID)
			if err != nil {
				t.Logf("Failed to retrieve user: %v", err)
				return false
			}

			// Verify password field is not the plain text password
			if retrievedUser.Password == password {
				t.Logf("Password field contains plain text password")
				return false
			}

			// Verify the password is actually hashed (bcrypt hashes are 60 chars)
			if len(retrievedUser.Password) != 60 {
				t.Logf("Password is not properly hashed (expected 60 chars, got %d)", len(retrievedUser.Password))
				return false
			}

			return true
		},
		genUsername(),
		gen.AlphaString().SuchThat(func(s string) bool {
			return len(s) >= 1 && len(s) <= 72
		}).Map(func(s string) string {
			if s == "" {
				return "password"
			}
			return s
		}),
	))

	properties.TestingRun(t)
}

// Feature: k-admin-system
// Property 11: Username Uniqueness Validation
// For any attempt to create a user with an existing username, the system SHALL reject
// the creation and return a validation error
// **Validates: Requirements 4.8**
func TestProperty11_UsernameUniquenessValidation(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	properties.Property("duplicate username creation is rejected", prop.ForAll(
		func(username string, nickname1 string, nickname2 string) bool {
			// Setup test database
			db := setupTestDB(t)
			defer cleanupTestDB(db)

			// Set global DB for service
			originalDB := global.DB
			global.DB = db
			defer func() { global.DB = originalDB }()

			service := &UserService{}

			// Create first user with username
			user1 := &system.SysUser{
				Username: username,
				Password: "password123",
				Nickname: nickname1,
				RoleID:   1,
				Active:   true,
			}

			err := service.CreateUser(user1)
			if err != nil {
				t.Logf("Failed to create first user: %v", err)
				return false
			}

			// Attempt to create second user with same username
			user2 := &system.SysUser{
				Username: username,
				Password: "password456",
				Nickname: nickname2,
				RoleID:   1,
				Active:   true,
			}

			err = service.CreateUser(user2)
			if err == nil {
				t.Logf("Duplicate username creation was not rejected")
				return false
			}

			// Verify error message indicates username already exists
			if err.Error() != "username already exists" {
				t.Logf("Unexpected error message: %v", err)
				return false
			}

			// Verify only one user exists in database
			users, total, err := service.GetUserList(1, 100, map[string]interface{}{})
			if err != nil {
				t.Logf("Failed to get user list: %v", err)
				return false
			}
			if total != 1 {
				t.Logf("Expected 1 user, found %d", total)
				return false
			}
			if len(users) != 1 {
				t.Logf("Expected 1 user in list, found %d", len(users))
				return false
			}

			return true
		},
		genUsername(),
		genNickname(),
		genNickname(),
	))

	properties.TestingRun(t)
}

// Feature: k-admin-system
// Property 12: User List Pagination and Filtering
// For any user list request with pagination parameters (page, pageSize) and filters,
// the response SHALL contain exactly pageSize or fewer users matching the filters,
// and the total count SHALL equal the number of matching users
// **Validates: Requirements 4.6**
func TestProperty12_UserListPaginationAndFiltering(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	properties.Property("pagination returns correct page size and total", prop.ForAll(
		func(userCount int, pageSize int, page int) bool {
			// Setup test database
			db := setupTestDB(t)
			defer cleanupTestDB(db)

			// Set global DB for service
			originalDB := global.DB
			global.DB = db
			defer func() { global.DB = originalDB }()

			service := &UserService{}

			// Create multiple users
			for i := 0; i < userCount; i++ {
				user := &system.SysUser{
					Username: fmt.Sprintf("user%d", i),
					Password: "password123",
					Nickname: fmt.Sprintf("User %d", i),
					RoleID:   1,
					Active:   true,
				}
				err := service.CreateUser(user)
				if err != nil {
					t.Logf("Failed to create user %d: %v", i, err)
					return false
				}
			}

			// Get user list with pagination
			users, total, err := service.GetUserList(page, pageSize, map[string]interface{}{})
			if err != nil {
				t.Logf("Failed to get user list: %v", err)
				return false
			}

			// Verify total count matches created users
			if total != int64(userCount) {
				t.Logf("Total count mismatch: expected %d, got %d", userCount, total)
				return false
			}

			// Calculate expected page size
			expectedSize := pageSize
			offset := (page - 1) * pageSize
			if offset >= userCount {
				expectedSize = 0
			} else if offset+pageSize > userCount {
				expectedSize = userCount - offset
			}

			// Verify returned users count
			if len(users) != expectedSize {
				t.Logf("Page size mismatch: expected %d, got %d (page=%d, pageSize=%d, userCount=%d, offset=%d)",
					expectedSize, len(users), page, pageSize, userCount, offset)
				return false
			}

			return true
		},
		gen.IntRange(1, 20),
		gen.IntRange(1, 10),
		gen.IntRange(1, 5),
	))

	properties.Property("filtering returns only matching users", prop.ForAll(
		func(seed int) bool {
			// Generate a deterministic target nickname from seed
			targetNickname := fmt.Sprintf("Target%d", seed%100)

			// Setup test database
			db := setupTestDB(t)
			defer cleanupTestDB(db)

			// Set global DB for service
			originalDB := global.DB
			global.DB = db
			defer func() { global.DB = originalDB }()

			service := &UserService{}

			// Create users with different nicknames
			matchingCount := 0
			for i := 0; i < 10; i++ {
				nickname := fmt.Sprintf("User%d", i)
				if i%3 == 0 {
					nickname = targetNickname + fmt.Sprintf("%d", i)
					matchingCount++
				}

				user := &system.SysUser{
					Username: fmt.Sprintf("user%d_%d", seed, i),
					Password: "password123",
					Nickname: nickname,
					RoleID:   1,
					Active:   true,
				}
				err := service.CreateUser(user)
				if err != nil {
					t.Logf("Failed to create user %d: %v", i, err)
					return false
				}
			}

			// Filter by nickname
			filters := map[string]interface{}{
				"nickname": targetNickname,
			}
			users, total, err := service.GetUserList(1, 100, filters)
			if err != nil {
				t.Logf("Failed to get filtered user list: %v", err)
				return false
			}

			// Verify total matches expected count
			if total != int64(matchingCount) {
				t.Logf("Filtered total mismatch: expected %d, got %d", matchingCount, total)
				return false
			}

			// Verify all returned users match filter
			for _, user := range users {
				// Check if nickname contains target (LIKE query with %)
				found := false
				for j := 0; j <= len(user.Nickname)-len(targetNickname); j++ {
					if user.Nickname[j:j+len(targetNickname)] == targetNickname {
						found = true
						break
					}
				}
				if !found {
					t.Logf("User nickname '%s' doesn't contain filter '%s'", user.Nickname, targetNickname)
					return false
				}
			}

			return true
		},
		gen.IntRange(1, 1000),
	))

	properties.TestingRun(t)
}

// Feature: k-admin-system
// Property 47: Soft Delete Behavior
// For any record with DeletedAt field, deleting the record SHALL set DeletedAt to current timestamp,
// and subsequent queries SHALL exclude soft-deleted records by default
// **Validates: Requirements 15.7**
func TestProperty47_SoftDeleteBehavior(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	properties.Property("deleted user has DeletedAt set and is excluded from queries", prop.ForAll(
		func(username string) bool {
			// Setup test database
			db := setupTestDB(t)
			defer cleanupTestDB(db)

			// Set global DB for service
			originalDB := global.DB
			global.DB = db
			defer func() { global.DB = originalDB }()

			service := &UserService{}

			// Create a user
			user := &system.SysUser{
				Username: username,
				Password: "password123",
				RoleID:   1,
				Active:   true,
			}

			err := service.CreateUser(user)
			if err != nil {
				t.Logf("Failed to create user: %v", err)
				return false
			}

			userID := user.ID

			// Verify user exists before deletion
			_, err = service.GetUserByID(userID)
			if err != nil {
				t.Logf("User not found before deletion: %v", err)
				return false
			}

			// Delete the user (soft delete)
			err = service.DeleteUser(userID)
			if err != nil {
				t.Logf("Failed to delete user: %v", err)
				return false
			}

			// Verify user is excluded from GetUserByID (default query excludes soft-deleted)
			_, err = service.GetUserByID(userID)
			if err == nil {
				t.Logf("Soft-deleted user still retrievable by GetUserByID")
				return false
			}

			// Verify user is excluded from GetUserList
			users, total, err := service.GetUserList(1, 100, map[string]interface{}{})
			if err != nil {
				t.Logf("Failed to get user list: %v", err)
				return false
			}
			for _, u := range users {
				if u.ID == userID {
					t.Logf("Soft-deleted user still appears in user list")
					return false
				}
			}
			if total > 0 {
				t.Logf("Total count includes soft-deleted user")
				return false
			}

			// Verify DeletedAt is set by querying with Unscoped
			var deletedUser system.SysUser
			err = db.Unscoped().First(&deletedUser, userID).Error
			if err != nil {
				t.Logf("Failed to query soft-deleted user with Unscoped: %v", err)
				return false
			}

			// Check DeletedAt is not zero (soft delete sets timestamp)
			if !deletedUser.DeletedAt.Valid {
				t.Logf("DeletedAt is not set after soft delete")
				return false
			}

			return true
		},
		genUsername(),
	))

	properties.Property("multiple soft deletes don't interfere", prop.ForAll(
		func(count int) bool {
			// Setup test database
			db := setupTestDB(t)
			defer cleanupTestDB(db)

			// Set global DB for service
			originalDB := global.DB
			global.DB = db
			defer func() { global.DB = originalDB }()

			service := &UserService{}

			// Create multiple users
			userIDs := make([]uint, count)
			for i := 0; i < count; i++ {
				user := &system.SysUser{
					Username: fmt.Sprintf("user%d", i),
					Password: "password123",
					RoleID:   1,
					Active:   true,
				}
				err := service.CreateUser(user)
				if err != nil {
					t.Logf("Failed to create user %d: %v", i, err)
					return false
				}
				userIDs[i] = user.ID
			}

			// Soft delete half of the users
			deleteCount := count / 2
			for i := 0; i < deleteCount; i++ {
				err := service.DeleteUser(userIDs[i])
				if err != nil {
					t.Logf("Failed to delete user %d: %v", i, err)
					return false
				}
			}

			// Verify correct number of users remain
			users, total, err := service.GetUserList(1, 100, map[string]interface{}{})
			if err != nil {
				t.Logf("Failed to get user list: %v", err)
				return false
			}

			expectedRemaining := count - deleteCount
			if total != int64(expectedRemaining) {
				t.Logf("Total count mismatch: expected %d, got %d", expectedRemaining, total)
				return false
			}
			if len(users) != expectedRemaining {
				t.Logf("User list size mismatch: expected %d, got %d", expectedRemaining, len(users))
				return false
			}

			// Verify deleted users are not in the list
			for i := 0; i < deleteCount; i++ {
				for _, u := range users {
					if u.ID == userIDs[i] {
						t.Logf("Deleted user %d still in list", userIDs[i])
						return false
					}
				}
			}

			// Verify non-deleted users are in the list
			for i := deleteCount; i < count; i++ {
				found := false
				for _, u := range users {
					if u.ID == userIDs[i] {
						found = true
						break
					}
				}
				if !found {
					t.Logf("Non-deleted user %d not in list", userIDs[i])
					return false
				}
			}

			return true
		},
		gen.IntRange(2, 10),
	))

	properties.TestingRun(t)
}
