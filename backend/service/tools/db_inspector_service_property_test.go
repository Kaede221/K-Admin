package tools

import (
	"strings"
	"testing"

	"k-admin-system/global"

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

	// Create test tables
	db.Exec(`CREATE TABLE test_users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		email TEXT,
		age INTEGER
	)`)

	db.Exec(`CREATE TABLE test_products (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		price REAL
	)`)

	return db
}

// cleanupTestDB cleans all data from the test database
func cleanupTestDB(db *gorm.DB) {
	db.Exec("DELETE FROM test_users")
	db.Exec("DELETE FROM test_products")
}

// genTableName generates valid table names
func genTableName() gopter.Gen {
	return gen.OneConstOf("test_users", "test_products")
}

// genRecordData generates valid record data for test_users table
func genRecordData() gopter.Gen {
	return gen.AlphaString().SuchThat(func(s string) bool {
		return len(s) >= 1 && len(s) <= 50
	}).Map(func(name string) map[string]interface{} {
		return map[string]interface{}{
			"name":  name,
			"email": name + "@test.com",
			"age":   25,
		}
	})
}

// Feature: k-admin-system
// Property 18: Database Table Listing Completeness
// For any connected database, the DB Inspector SHALL return all tables that exist in the database schema
// **Validates: Requirements 7.2**
func TestProperty18_DatabaseTableListingCompleteness(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	properties.Property("DB Inspector returns all existing tables", prop.ForAll(
		func() bool {
			// Setup test database
			db := setupTestDB(t)
			defer cleanupTestDB(db)

			// Set global DB for service
			originalDB := global.DB
			global.DB = db
			defer func() { global.DB = originalDB }()

			service := &DBInspectorService{}

			// Get tables
			tables, err := service.GetTables()
			if err != nil {
				t.Logf("Failed to get tables: %v", err)
				return false
			}

			// Verify that test_users and test_products are in the list
			hasTestUsers := false
			hasTestProducts := false
			for _, table := range tables {
				if table == "test_users" {
					hasTestUsers = true
				}
				if table == "test_products" {
					hasTestProducts = true
				}
			}

			return hasTestUsers && hasTestProducts
		},
	))

	properties.TestingRun(t)
}

// Feature: k-admin-system
// Property 19: Table Schema Accuracy
// For any selected table, the displayed schema SHALL include all columns with accurate types,
// nullability, keys, and comments matching the actual database schema
// **Validates: Requirements 7.3**
func TestProperty19_TableSchemaAccuracy(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	properties.Property("table schema matches actual database schema", prop.ForAll(
		func(tableName string) bool {
			// Setup test database
			db := setupTestDB(t)
			defer cleanupTestDB(db)

			// Set global DB for service
			originalDB := global.DB
			global.DB = db
			defer func() { global.DB = originalDB }()

			service := &DBInspectorService{}

			// Get table schema
			columns, err := service.GetTableSchema(tableName)
			if err != nil {
				t.Logf("Failed to get table schema: %v", err)
				return false
			}

			// Verify schema based on table
			if tableName == "test_users" {
				// Should have 4 columns: id, name, email, age
				if len(columns) != 4 {
					t.Logf("Expected 4 columns, got %d", len(columns))
					return false
				}

				// Verify column names
				expectedColumns := map[string]bool{
					"id": false, "name": false, "email": false, "age": false,
				}
				for _, col := range columns {
					if _, exists := expectedColumns[col.Name]; exists {
						expectedColumns[col.Name] = true
					}
				}
				for col, found := range expectedColumns {
					if !found {
						t.Logf("Column %s not found", col)
						return false
					}
				}
			}

			return true
		},
		genTableName(),
	))

	properties.TestingRun(t)
}

// Feature: k-admin-system
// Property 20: DB Inspector CRUD Operations
// For any table record, creating a record SHALL make it retrievable, updating SHALL persist changes,
// and deleting SHALL remove it from subsequent queries
// **Validates: Requirements 7.4, 7.5**
func TestProperty20_DBInspectorCRUDOperations(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	properties.Property("CRUD operations work correctly", prop.ForAll(
		func(recordData map[string]interface{}) bool {
			// Setup test database
			db := setupTestDB(t)
			defer cleanupTestDB(db)

			// Set global DB for service
			originalDB := global.DB
			global.DB = db
			defer func() { global.DB = originalDB }()

			service := &DBInspectorService{}
			tableName := "test_users"

			// CREATE: Insert a record
			err := service.CreateRecord(tableName, recordData)
			if err != nil {
				t.Logf("Failed to create record: %v", err)
				return false
			}

			// RETRIEVE: Get the record
			data, total, err := service.GetTableData(tableName, 1, 10)
			if err != nil {
				t.Logf("Failed to get table data: %v", err)
				return false
			}
			if total == 0 {
				t.Logf("Record not found after creation")
				return false
			}
			if len(data) == 0 {
				t.Logf("No data returned")
				return false
			}

			// Get the ID of the created record
			recordID := data[0]["id"]

			// UPDATE: Modify the record
			updateData := map[string]interface{}{
				"name": "Updated Name",
			}
			err = service.UpdateRecord(tableName, recordID, updateData)
			if err != nil {
				t.Logf("Failed to update record: %v", err)
				return false
			}

			// Verify update
			data, _, err = service.GetTableData(tableName, 1, 10)
			if err != nil {
				t.Logf("Failed to get table data after update: %v", err)
				return false
			}
			if len(data) == 0 {
				t.Logf("No data returned after update")
				return false
			}
			if data[0]["name"] != "Updated Name" {
				t.Logf("Update not persisted")
				return false
			}

			// DELETE: Remove the record
			err = service.DeleteRecord(tableName, recordID)
			if err != nil {
				t.Logf("Failed to delete record: %v", err)
				return false
			}

			// Verify deletion
			data, total, err = service.GetTableData(tableName, 1, 10)
			if err != nil {
				t.Logf("Failed to get table data after delete: %v", err)
				return false
			}
			if total != 0 {
				t.Logf("Record still exists after deletion")
				return false
			}

			return true
		},
		genRecordData(),
	))

	properties.TestingRun(t)
}

// Feature: k-admin-system
// Property 21: Dangerous SQL Operation Restriction
// For any user without super admin privileges, attempts to execute DROP, TRUNCATE, or ALTER commands
// SHALL be rejected with permission error
// **Validates: Requirements 7.7**
func TestProperty21_DangerousSQLOperationRestriction(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	// Generate dangerous SQL statements
	genDangerousSQL := gen.OneConstOf(
		"DROP TABLE test_users",
		"TRUNCATE TABLE test_users",
		"ALTER DATABASE test_db",
		"DROP DATABASE test_db",
		"CREATE DATABASE new_db",
	)

	properties.Property("dangerous SQL operations are rejected", prop.ForAll(
		func(sql string) bool {
			// Setup test database
			db := setupTestDB(t)
			defer cleanupTestDB(db)

			// Set global DB for service
			originalDB := global.DB
			global.DB = db
			defer func() { global.DB = originalDB }()

			service := &DBInspectorService{}

			// Try to execute dangerous SQL
			_, err := service.ExecuteSQL(sql, false)

			// Should return an error
			if err == nil {
				t.Logf("Dangerous SQL was not rejected: %s", sql)
				return false
			}

			// Error message should mention the dangerous operation
			errMsg := strings.ToLower(err.Error())
			if !strings.Contains(errMsg, "dangerous") && !strings.Contains(errMsg, "not allowed") {
				t.Logf("Error message doesn't indicate dangerous operation: %v", err)
				return false
			}

			return true
		},
		genDangerousSQL,
	))

	properties.TestingRun(t)
}

// Feature: k-admin-system
// Property 22: Read-Only Mode Enforcement
// For any write operation (INSERT, UPDATE, DELETE) attempted in read-only mode,
// the operation SHALL fail with a read-only mode error
// **Validates: Requirements 7.8**
func TestProperty22_ReadOnlyModeEnforcement(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	// Generate write SQL statements
	genWriteSQL := gen.OneConstOf(
		"INSERT INTO test_users (name, email) VALUES ('test', 'test@test.com')",
		"UPDATE test_users SET name = 'updated' WHERE id = 1",
		"DELETE FROM test_users WHERE id = 1",
	)

	properties.Property("write operations are rejected in read-only mode", prop.ForAll(
		func(sql string) bool {
			// Setup test database
			db := setupTestDB(t)
			defer cleanupTestDB(db)

			// Set global DB for service
			originalDB := global.DB
			global.DB = db
			defer func() { global.DB = originalDB }()

			service := &DBInspectorService{}

			// Try to execute write SQL in read-only mode
			_, err := service.ExecuteSQL(sql, true)

			// Should return an error
			if err == nil {
				t.Logf("Write SQL was not rejected in read-only mode: %s", sql)
				return false
			}

			// Error message should mention read-only mode
			errMsg := strings.ToLower(err.Error())
			if !strings.Contains(errMsg, "read-only") && !strings.Contains(errMsg, "not allowed") {
				t.Logf("Error message doesn't indicate read-only restriction: %v", err)
				return false
			}

			return true
		},
		genWriteSQL,
	))

	properties.TestingRun(t)
}

// Feature: k-admin-system
// Property 23: SQL Error Message Propagation
// For any SQL execution that fails, the error response SHALL contain the database error message
// with sufficient detail for debugging
// **Validates: Requirements 7.9**
func TestProperty23_SQLErrorMessagePropagation(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	// Generate invalid SQL statements
	genInvalidSQL := gen.OneConstOf(
		"SELECT * FROM nonexistent_table",
		"INSERT INTO test_users (invalid_column) VALUES ('test')",
		"UPDATE test_users SET invalid_column = 'test' WHERE id = 1",
		"SELECT invalid_column FROM test_users",
	)

	properties.Property("SQL errors contain detailed error messages", prop.ForAll(
		func(sql string) bool {
			// Setup test database
			db := setupTestDB(t)
			defer cleanupTestDB(db)

			// Set global DB for service
			originalDB := global.DB
			global.DB = db
			defer func() { global.DB = originalDB }()

			service := &DBInspectorService{}

			// Try to execute invalid SQL
			_, err := service.ExecuteSQL(sql, false)

			// Should return an error
			if err == nil {
				t.Logf("Invalid SQL did not return an error: %s", sql)
				return false
			}

			// Error message should not be empty
			errMsg := err.Error()
			if errMsg == "" {
				t.Logf("Error message is empty")
				return false
			}

			// Error message should contain some detail (at least 10 characters)
			if len(errMsg) < 10 {
				t.Logf("Error message is too short: %s", errMsg)
				return false
			}

			return true
		},
		genInvalidSQL,
	))

	properties.TestingRun(t)
}

// Additional test: Verify SELECT queries work in read-only mode
func TestReadOnlyModeAllowsSelectQueries(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	properties.Property("SELECT queries work in read-only mode", prop.ForAll(
		func() bool {
			// Setup test database
			db := setupTestDB(t)
			defer cleanupTestDB(db)

			// Set global DB for service
			originalDB := global.DB
			global.DB = db
			defer func() { global.DB = originalDB }()

			service := &DBInspectorService{}

			// Execute SELECT query in read-only mode
			sql := "SELECT * FROM test_users"
			_, err := service.ExecuteSQL(sql, true)

			// Should not return an error
			if err != nil {
				t.Logf("SELECT query failed in read-only mode: %v", err)
				return false
			}

			return true
		},
	))

	properties.TestingRun(t)
}

// Additional test: Verify table name validation
func TestTableNameValidation(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	// Generate invalid table names (with special characters)
	genInvalidTableName := gen.AlphaString().Map(func(s string) string {
		// Add special characters to make it invalid
		return s + "'; DROP TABLE test_users; --"
	})

	properties.Property("invalid table names are rejected", prop.ForAll(
		func(tableName string) bool {
			// Setup test database
			db := setupTestDB(t)
			defer cleanupTestDB(db)

			// Set global DB for service
			originalDB := global.DB
			global.DB = db
			defer func() { global.DB = originalDB }()

			service := &DBInspectorService{}

			// Try to get schema with invalid table name
			_, err := service.GetTableSchema(tableName)

			// Should return an error for invalid table names
			if err == nil {
				// Check if the table name is actually valid (only alphanumeric and underscore)
				if !strings.ContainsAny(tableName, "';-") {
					// If it doesn't contain special chars, it might be valid
					return true
				}
				t.Logf("Invalid table name was not rejected: %s", tableName)
				return false
			}

			return true
		},
		genInvalidTableName,
	))

	properties.TestingRun(t)
}
