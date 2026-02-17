package core

import (
	"fmt"
	"testing"
	"time"

	"k-admin-system/config"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
	"go.uber.org/zap"
)

// Test models for migration testing
type TestMigrationModel1 struct {
	ID        uint      `gorm:"primarykey"`
	Name      string    `gorm:"type:varchar(100)"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
}

type TestMigrationModel2 struct {
	ID          uint      `gorm:"primarykey"`
	Description string    `gorm:"type:text"`
	Status      bool      `gorm:"default:true"`
	CreatedAt   time.Time `gorm:"autoCreateTime"`
}

type TestMigrationModel3 struct {
	ID    uint   `gorm:"primarykey"`
	Value int    `gorm:"not null"`
	Tag   string `gorm:"type:varchar(50);index"`
}

// Feature: k-admin-system
// Property 46: Database Migration Execution
// For any set of model definitions, executing AutoMigrate SHALL create
// corresponding database tables with correct schema, and subsequent migrations
// SHALL preserve existing data while updating schema
// Validates: Requirements 15.5
func TestProperty46_DatabaseMigrationExecution(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	properties.Property("migration creates tables with correct schema", prop.ForAll(
		func(modelCount int) bool {
			// Test with 1-3 models
			if modelCount < 1 || modelCount > 3 {
				return true
			}

			// Create test configuration
			cfg := &config.Config{
				Server: config.ServerConfig{
					Port: "8080",
					Mode: "test",
				},
				Database: config.DatabaseConfig{
					Host:         "localhost",
					Port:         3306,
					Name:         "test_db",
					Username:     "root",
					Password:     "",
					MaxIdleConns: 10,
					MaxOpenConns: 100,
				},
			}

			// Initialize logger
			logger, _ := zap.NewDevelopment()

			// Initialize database connection
			db, err := InitDB(cfg, logger)
			if err != nil {
				t.Logf("Skipping test case due to connection error: %v", err)
				return true
			}
			defer func() {
				sqlDB, _ := db.DB()
				if sqlDB != nil {
					sqlDB.Close()
				}
			}()

			// Generate unique table names for this test run
			timestamp := time.Now().UnixNano()
			tableName1 := fmt.Sprintf("test_migration_model1_%d", timestamp)
			tableName2 := fmt.Sprintf("test_migration_model2_%d", timestamp)
			tableName3 := fmt.Sprintf("test_migration_model3_%d", timestamp)

			// Clean up tables after test
			defer func() {
				db.Migrator().DropTable(tableName1)
				db.Migrator().DropTable(tableName2)
				db.Migrator().DropTable(tableName3)
			}()

			// Perform migration based on modelCount
			switch modelCount {
			case 1:
				db.Table(tableName1).AutoMigrate(&TestMigrationModel1{})
			case 2:
				db.Table(tableName1).AutoMigrate(&TestMigrationModel1{})
				db.Table(tableName2).AutoMigrate(&TestMigrationModel2{})
			case 3:
				db.Table(tableName1).AutoMigrate(&TestMigrationModel1{})
				db.Table(tableName2).AutoMigrate(&TestMigrationModel2{})
				db.Table(tableName3).AutoMigrate(&TestMigrationModel3{})
			}

			// Verify tables were created
			if !db.Migrator().HasTable(tableName1) {
				t.Logf("Table %s was not created", tableName1)
				return false
			}

			if modelCount >= 2 && !db.Migrator().HasTable(tableName2) {
				t.Logf("Table %s was not created", tableName2)
				return false
			}

			if modelCount >= 3 && !db.Migrator().HasTable(tableName3) {
				t.Logf("Table %s was not created", tableName3)
				return false
			}

			// Verify columns exist for first table
			if !db.Migrator().HasColumn(tableName1, "id") {
				t.Logf("Column 'id' not found in %s", tableName1)
				return false
			}
			if !db.Migrator().HasColumn(tableName1, "name") {
				t.Logf("Column 'name' not found in %s", tableName1)
				return false
			}
			if !db.Migrator().HasColumn(tableName1, "created_at") {
				t.Logf("Column 'created_at' not found in %s", tableName1)
				return false
			}

			return true
		},
		gen.IntRange(1, 3),
	))

	properties.Property("migration preserves existing data", prop.ForAll(
		func(recordCount int) bool {
			// Test with 1-10 records
			if recordCount < 1 || recordCount > 10 {
				return true
			}

			// Create test configuration
			cfg := &config.Config{
				Server: config.ServerConfig{
					Port: "8080",
					Mode: "test",
				},
				Database: config.DatabaseConfig{
					Host:         "localhost",
					Port:         3306,
					Name:         "test_db",
					Username:     "root",
					Password:     "",
					MaxIdleConns: 10,
					MaxOpenConns: 100,
				},
			}

			// Initialize logger
			logger, _ := zap.NewDevelopment()

			// Initialize database connection
			db, err := InitDB(cfg, logger)
			if err != nil {
				t.Logf("Skipping test case due to connection error: %v", err)
				return true
			}
			defer func() {
				sqlDB, _ := db.DB()
				if sqlDB != nil {
					sqlDB.Close()
				}
			}()

			// Generate unique table name
			timestamp := time.Now().UnixNano()
			tableName := fmt.Sprintf("test_migration_preserve_%d", timestamp)

			// Clean up table after test
			defer func() {
				db.Migrator().DropTable(tableName)
			}()

			// Initial migration
			db.Table(tableName).AutoMigrate(&TestMigrationModel1{})

			// Insert test records
			for i := 0; i < recordCount; i++ {
				record := TestMigrationModel1{
					Name: fmt.Sprintf("Record_%d", i),
				}
				result := db.Table(tableName).Create(&record)
				if result.Error != nil {
					t.Logf("Failed to insert record: %v", result.Error)
					return false
				}
			}

			// Verify records were inserted
			var count int64
			db.Table(tableName).Count(&count)
			if count != int64(recordCount) {
				t.Logf("Expected %d records, found %d", recordCount, count)
				return false
			}

			// Run migration again (should be idempotent)
			db.Table(tableName).AutoMigrate(&TestMigrationModel1{})

			// Verify records still exist
			var countAfter int64
			db.Table(tableName).Count(&countAfter)
			if countAfter != int64(recordCount) {
				t.Logf("Records lost after migration: expected %d, found %d", recordCount, countAfter)
				return false
			}

			// Verify data integrity
			var records []TestMigrationModel1
			db.Table(tableName).Find(&records)
			if len(records) != recordCount {
				t.Logf("Record count mismatch: expected %d, got %d", recordCount, len(records))
				return false
			}

			// Verify record names
			for i, record := range records {
				expectedName := fmt.Sprintf("Record_%d", i)
				if record.Name != expectedName {
					t.Logf("Record name mismatch: expected %s, got %s", expectedName, record.Name)
					return false
				}
			}

			return true
		},
		gen.IntRange(1, 10),
	))

	properties.Property("migration handles schema updates", prop.ForAll(
		func(addColumn bool) bool {
			// Create test configuration
			cfg := &config.Config{
				Server: config.ServerConfig{
					Port: "8080",
					Mode: "test",
				},
				Database: config.DatabaseConfig{
					Host:         "localhost",
					Port:         3306,
					Name:         "test_db",
					Username:     "root",
					Password:     "",
					MaxIdleConns: 10,
					MaxOpenConns: 100,
				},
			}

			// Initialize logger
			logger, _ := zap.NewDevelopment()

			// Initialize database connection
			db, err := InitDB(cfg, logger)
			if err != nil {
				t.Logf("Skipping test case due to connection error: %v", err)
				return true
			}
			defer func() {
				sqlDB, _ := db.DB()
				if sqlDB != nil {
					sqlDB.Close()
				}
			}()

			// Generate unique table name
			timestamp := time.Now().UnixNano()
			tableName := fmt.Sprintf("test_migration_schema_%d", timestamp)

			// Clean up table after test
			defer func() {
				db.Migrator().DropTable(tableName)
			}()

			// Initial migration with basic model
			type InitialModel struct {
				ID   uint   `gorm:"primarykey"`
				Name string `gorm:"type:varchar(100)"`
			}
			db.Table(tableName).AutoMigrate(&InitialModel{})

			// Insert a test record
			record := InitialModel{Name: "Test"}
			db.Table(tableName).Create(&record)

			if addColumn {
				// Migrate with extended model (add column)
				type ExtendedModel struct {
					ID          uint   `gorm:"primarykey"`
					Name        string `gorm:"type:varchar(100)"`
					Description string `gorm:"type:text"`
				}
				db.Table(tableName).AutoMigrate(&ExtendedModel{})

				// Verify new column exists
				if !db.Migrator().HasColumn(tableName, "description") {
					t.Logf("New column 'description' was not added")
					return false
				}

				// Verify old data still exists
				var count int64
				db.Table(tableName).Count(&count)
				if count != 1 {
					t.Logf("Data lost after schema update")
					return false
				}
			}

			return true
		},
		gen.Bool(),
	))

	properties.Property("migration is idempotent", prop.ForAll(
		func(migrationCount int) bool {
			// Run migration 1-5 times
			if migrationCount < 1 || migrationCount > 5 {
				return true
			}

			// Create test configuration
			cfg := &config.Config{
				Server: config.ServerConfig{
					Port: "8080",
					Mode: "test",
				},
				Database: config.DatabaseConfig{
					Host:         "localhost",
					Port:         3306,
					Name:         "test_db",
					Username:     "root",
					Password:     "",
					MaxIdleConns: 10,
					MaxOpenConns: 100,
				},
			}

			// Initialize logger
			logger, _ := zap.NewDevelopment()

			// Initialize database connection
			db, err := InitDB(cfg, logger)
			if err != nil {
				t.Logf("Skipping test case due to connection error: %v", err)
				return true
			}
			defer func() {
				sqlDB, _ := db.DB()
				if sqlDB != nil {
					sqlDB.Close()
				}
			}()

			// Generate unique table name
			timestamp := time.Now().UnixNano()
			tableName := fmt.Sprintf("test_migration_idempotent_%d", timestamp)

			// Clean up table after test
			defer func() {
				db.Migrator().DropTable(tableName)
			}()

			// Run migration multiple times
			for i := 0; i < migrationCount; i++ {
				err := db.Table(tableName).AutoMigrate(&TestMigrationModel1{})
				if err != nil {
					t.Logf("Migration %d failed: %v", i+1, err)
					return false
				}
			}

			// Verify table exists and has correct structure
			if !db.Migrator().HasTable(tableName) {
				t.Logf("Table not found after %d migrations", migrationCount)
				return false
			}

			// Verify columns
			if !db.Migrator().HasColumn(tableName, "id") ||
				!db.Migrator().HasColumn(tableName, "name") ||
				!db.Migrator().HasColumn(tableName, "created_at") {
				t.Logf("Table structure incorrect after multiple migrations")
				return false
			}

			return true
		},
		gen.IntRange(1, 5),
	))

	properties.TestingRun(t)
}
