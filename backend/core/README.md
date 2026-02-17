# Core Package

This package contains core infrastructure components for the K-Admin system.

## Zap Logger

The logging system uses Uber's Zap logger with Lumberjack for log rotation.

### Features

- **Structured Logging**: JSON format for easy parsing and analysis
- **Log Rotation**: Automatic rotation based on file size and age
- **Environment-Specific Output**: 
  - Development/Test mode: Logs to both console and file
  - Production mode: Logs to file only
- **Configurable Log Levels**: debug, info, warn, error, fatal
- **Automatic Stack Traces**: Included for error-level and above
- **Caller Information**: Shows file and line number for each log entry

### Configuration

Configure logging in your `config.yaml`:

```yaml
logger:
  level: "info"              # Log level: debug, info, warn, error, fatal
  path: "./logs/app.log"     # Log file path
  max_size: 100              # Maximum size in megabytes before rotation
  max_age: 7                 # Maximum days to retain old log files
  max_backups: 3             # Maximum number of old log files to retain
  compress: true             # Compress rotated log files
```

### Usage

#### Initialization

The logger is initialized in `main.go` and stored in the global package:

```go
import (
    "k-admin-system/core"
    "k-admin-system/global"
)

func main() {
    // Initialize logger
    logger, err := core.InitLogger(cfg)
    if err != nil {
        log.Fatalf("Failed to initialize logger: %v", err)
    }
    global.Logger = logger
    defer core.SyncLogger(logger)
}
```

#### Logging Messages

Use the global logger instance throughout your application:

```go
import (
    "k-admin-system/global"
    "go.uber.org/zap"
)

// Info level
global.Logger.Info("User logged in",
    zap.String("username", "admin"),
    zap.String("ip", "192.168.1.1"),
)

// Debug level
global.Logger.Debug("Processing request",
    zap.String("method", "GET"),
    zap.String("path", "/api/users"),
)

// Warning level
global.Logger.Warn("Rate limit approaching",
    zap.String("user_id", "123"),
    zap.Int("requests", 95),
)

// Error level (includes stack trace)
global.Logger.Error("Database query failed",
    zap.Error(err),
    zap.String("query", "SELECT * FROM users"),
)

// Fatal level (logs and exits)
global.Logger.Fatal("Critical system failure",
    zap.Error(err),
)
```

#### Using Helper Functions

The package provides helper functions for convenience:

```go
import (
    "k-admin-system/core"
    "k-admin-system/global"
    "go.uber.org/zap"
)

core.LogInfo(global.Logger, "Operation completed", zap.Duration("elapsed", duration))
core.LogDebug(global.Logger, "Cache hit", zap.String("key", cacheKey))
core.LogWarn(global.Logger, "Deprecated API used", zap.String("endpoint", "/old/api"))
core.LogError(global.Logger, "Failed to send email", zap.Error(err))
```

### Log Output Format

Logs are written in JSON format for structured logging:

```json
{
  "level": "info",
  "timestamp": "2024-01-15T10:30:45.123+0800",
  "caller": "service/user_service.go:42",
  "msg": "User created successfully",
  "user_id": "12345",
  "username": "john_doe"
}
```

Error logs include stack traces:

```json
{
  "level": "error",
  "timestamp": "2024-01-15T10:31:12.456+0800",
  "caller": "service/user_service.go:78",
  "msg": "Failed to update user",
  "error": "database connection lost",
  "stacktrace": "k-admin-system/service.(*UserService).UpdateUser\n\t/path/to/user_service.go:78\n..."
}
```

### Log Rotation

Lumberjack automatically rotates log files based on:

- **Size**: When the log file reaches `max_size` megabytes
- **Age**: Old log files older than `max_age` days are deleted
- **Backups**: Only `max_backups` old log files are retained

Rotated files are named with timestamps: `app-2024-01-15T10-30-45.123.log`

If `compress` is enabled, rotated files are gzipped: `app-2024-01-15T10-30-45.123.log.gz`

### Best Practices

1. **Use Structured Fields**: Always use typed fields (zap.String, zap.Int, etc.) instead of string formatting
2. **Appropriate Log Levels**:
   - Debug: Detailed information for debugging
   - Info: General informational messages
   - Warn: Warning messages for potentially harmful situations
   - Error: Error messages for failures that don't stop the application
   - Fatal: Critical errors that require application shutdown
3. **Include Context**: Add relevant fields like user_id, request_id, etc.
4. **Avoid Sensitive Data**: Never log passwords, tokens, or personal information
5. **Sync Before Exit**: Always call `SyncLogger()` before application shutdown

### Testing

Run the logger tests:

```bash
go test -v ./core/...
```

Test coverage includes:
- Logger initialization with various configurations
- Log level parsing
- Helper functions
- Log file creation and rotation
- Environment-specific output behavior


## Database Connection (Gorm)

The database system uses Gorm ORM with MySQL driver, featuring connection pooling, automatic reconnection, and slow query logging.

### Features

- **Connection Pooling**: Configurable pool size for optimal performance
- **Automatic Reconnection**: Handles connection loss gracefully
- **Slow Query Logging**: Logs queries exceeding 200ms threshold
- **Custom Logger Integration**: Gorm queries logged through Zap
- **Environment-Specific Query Logging**:
  - Debug mode: All queries logged
  - Test mode: Only warnings and errors
  - Release mode: Only errors

### Configuration

Configure database connection in your `config.yaml`:

```yaml
database:
  host: "localhost"
  port: 3306
  name: "k_admin"
  username: "root"
  password: ""
  max_idle_conns: 10    # Maximum idle connections in pool
  max_open_conns: 100   # Maximum open connections to database
```

### Usage

#### Initialization

The database is initialized in `main.go` and stored in the global package:

```go
import (
    "k-admin-system/core"
    "k-admin-system/global"
)

func main() {
    // Initialize database
    db, err := core.InitDB(cfg, logger)
    if err != nil {
        logger.Fatal("Failed to initialize database", zap.Error(err))
    }
    global.DB = db
}
```

#### Using the Database

Access the global database instance throughout your application:

```go
import (
    "k-admin-system/global"
    "k-admin-system/model"
)

// Create
user := &model.SysUser{Username: "admin", Password: "hashed_password"}
result := global.DB.Create(user)

// Read
var user model.SysUser
global.DB.First(&user, 1) // Find by primary key
global.DB.Where("username = ?", "admin").First(&user)

// Update
global.DB.Model(&user).Update("active", true)
global.DB.Model(&user).Updates(map[string]interface{}{"active": true, "email": "new@example.com"})

// Delete (soft delete if model has DeletedAt field)
global.DB.Delete(&user, 1)

// Query with pagination
var users []model.SysUser
global.DB.Limit(10).Offset(0).Find(&users)
```

### Connection Pool Management

The connection pool is configured with:

- **MaxIdleConns**: Maximum number of idle connections (default: 10)
- **MaxOpenConns**: Maximum number of open connections (default: 100)
- **ConnMaxLifetime**: Maximum time a connection can be reused (1 hour)

These settings ensure:
- Efficient resource usage
- Prevention of connection exhaustion
- Automatic cleanup of stale connections

### Slow Query Logging

Queries exceeding 200ms are automatically logged with:

```json
{
  "level": "warn",
  "msg": "Slow query detected",
  "elapsed": "250ms",
  "threshold": "200ms",
  "sql": "SELECT * FROM sys_users WHERE ...",
  "rows": 1
}
```

### Query Logging by Environment

**Debug Mode** (server.mode: debug):
- All queries logged at DEBUG level
- Includes execution time, SQL, and rows affected

**Test Mode** (server.mode: test):
- Only slow queries and errors logged
- Reduces noise during testing

**Release Mode** (server.mode: release):
- Only errors logged
- Minimal performance impact

### Error Handling

Database errors are automatically logged:

```json
{
  "level": "error",
  "msg": "Database query error",
  "error": "Error 1062: Duplicate entry 'admin' for key 'username'",
  "elapsed": "5ms",
  "sql": "INSERT INTO sys_users ...",
  "rows": 0
}
```

### Automatic Reconnection

The MySQL driver automatically handles:
- Connection loss detection
- Reconnection attempts
- Connection pool refresh

No manual intervention required for transient network issues.

### Testing

Run the database tests:

```bash
# Run all core tests
go test -v ./core/...

# Run only database tests
go test -v ./core -run TestGormLogger
```

**Note**: Database connection tests require a running MySQL instance and are skipped by default. Run manually with a test database configured.

### Best Practices

1. **Use Transactions**: Wrap multiple operations in transactions for data consistency
2. **Avoid N+1 Queries**: Use Preload or Joins for related data
3. **Index Properly**: Ensure frequently queried columns are indexed
4. **Monitor Slow Queries**: Review slow query logs regularly
5. **Connection Pool Sizing**: Adjust based on application load
6. **Use Prepared Statements**: Gorm uses prepared statements by default for security

### Example: Transaction Usage

```go
import (
    "k-admin-system/global"
    "gorm.io/gorm"
)

func CreateUserWithRole(user *model.SysUser, roleID uint) error {
    return global.DB.Transaction(func(tx *gorm.DB) error {
        // Create user
        if err := tx.Create(user).Error; err != nil {
            return err
        }
        
        // Assign role
        if err := tx.Model(user).Update("role_id", roleID).Error; err != nil {
            return err
        }
        
        return nil
    })
}
```

### Example: Preloading Related Data

```go
// Avoid N+1 queries by preloading relationships
var users []model.SysUser
global.DB.Preload("Role").Find(&users)

// Each user now has Role data loaded
for _, user := range users {
    fmt.Printf("User: %s, Role: %s\n", user.Username, user.Role.RoleName)
}
```
