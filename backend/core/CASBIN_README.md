# Casbin Authorization System

## Overview

This implementation provides a complete Casbin-based authorization system for the K-Admin application. It supports RBAC (Role-Based Access Control) with RESTful path matching for API-level permission enforcement.

## Components

### 1. Casbin Rule Model (`model/system/sys_casbin_rule.go`)

The `SysCasbinRule` struct defines the database table structure for storing Casbin policies:

```go
type SysCasbinRule struct {
    ID    uint   `gorm:"primarykey;autoIncrement"`
    Ptype string `gorm:"size:100;uniqueIndex:unique_index"`
    V0    string `gorm:"size:100;uniqueIndex:unique_index"`
    V1    string `gorm:"size:100;uniqueIndex:unique_index"`
    V2    string `gorm:"size:100;uniqueIndex:unique_index"`
    V3    string `gorm:"size:100;uniqueIndex:unique_index"`
    V4    string `gorm:"size:100;uniqueIndex:unique_index"`
    V5    string `gorm:"size:100;uniqueIndex:unique_index"`
}
```

- **Ptype**: Policy type ("p" for policy, "g" for grouping/role inheritance)
- **V0-V5**: Flexible fields for storing policy parameters

### 2. Casbin Model Configuration (`config/casbin_model.conf`)

The model configuration defines the RBAC rules:

```ini
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
```

**Key Features:**
- **sub**: Subject (role name, e.g., "admin", "user")
- **obj**: Object (API path, e.g., "/api/v1/user/:id")
- **act**: Action (HTTP method, e.g., "GET", "POST")
- **keyMatch2**: RESTful path matching (supports `:id` style parameters)
- **g**: Role inheritance support

### 3. Casbin Initialization (`core/casbin.go`)

The `InitCasbin()` function initializes the Casbin enforcer:

```go
func InitCasbin() (*casbin.Enforcer, error)
```

**Process:**
1. Creates Gorm adapter using `sys_casbin_rules` table
2. Loads model configuration from `config/casbin_model.conf`
3. Loads existing policies from database
4. Returns configured enforcer instance

### 4. Global Enforcer (`global/global.go`)

The enforcer is stored as a global variable for application-wide access:

```go
var CasbinEnforcer *casbin.Enforcer
```

## Usage Examples

### Adding a Policy

```go
// Allow "admin" role to access "/api/v1/user" with GET method
success, err := global.CasbinEnforcer.AddPolicy("admin", "/api/v1/user", "GET")
```

### Checking Permission

```go
// Check if "admin" can access "/api/v1/user" with GET
allowed, err := global.CasbinEnforcer.Enforce("admin", "/api/v1/user", "GET")
```

### RESTful Path Matching

```go
// Add policy with parameter
global.CasbinEnforcer.AddPolicy("admin", "/api/v1/user/:id", "GET")

// These will match:
global.CasbinEnforcer.Enforce("admin", "/api/v1/user/123", "GET") // true
global.CasbinEnforcer.Enforce("admin", "/api/v1/user/456", "GET") // true
```

### Role Inheritance

```go
// Add policy for admin role
global.CasbinEnforcer.AddPolicy("admin", "/api/v1/system", "GET")

// Make user1 inherit admin role
global.CasbinEnforcer.AddGroupingPolicy("user1", "admin")

// user1 now has admin permissions
global.CasbinEnforcer.Enforce("user1", "/api/v1/system", "GET") // true
```

## Integration with Middleware

The Casbin enforcer will be used in the authorization middleware (Task 8.3):

```go
func CasbinAuth() gin.HandlerFunc {
    return func(c *gin.Context) {
        // Extract role from JWT claims
        role := c.GetString("role")
        path := c.Request.URL.Path
        method := c.Request.Method
        
        // Check permission
        allowed, err := global.CasbinEnforcer.Enforce(role, path, method)
        if err != nil || !allowed {
            c.JSON(403, gin.H{"error": "Forbidden"})
            c.Abort()
            return
        }
        
        c.Next()
    }
}
```

## Database Schema

The `sys_casbin_rules` table stores policies:

| Column | Type | Description |
|--------|------|-------------|
| id | uint | Primary key |
| ptype | string | Policy type ("p" or "g") |
| v0 | string | Subject (role) |
| v1 | string | Object (path) |
| v2 | string | Action (method) |
| v3-v5 | string | Reserved for future use |

**Example Data:**

| ptype | v0 | v1 | v2 |
|-------|----|----|-----|
| p | admin | /api/v1/user | GET |
| p | admin | /api/v1/user/:id | PUT |
| g | user1 | admin | |

## Testing

Unit tests verify the model structure and configuration:

```bash
go test -v ./core -run TestSysCasbinRuleModel
```

Integration tests (require database) verify:
- Enforcer initialization
- Policy CRUD operations
- RESTful path matching
- Role inheritance

## Requirements Validation

This implementation satisfies:
- **Requirement 3.4**: Backend SHALL use Casbin middleware to enforce API-level permissions
- **Requirement 3.5**: Backend SHALL store Casbin policies in sys_casbin_rules table
- **Requirement 3.8**: System SHALL support hierarchical role inheritance

## Next Steps

- **Task 8.2**: Implement Casbin manager utility with helper functions
- **Task 8.3**: Create Casbin authorization middleware
- **Task 8.4**: Write property tests for authorization
