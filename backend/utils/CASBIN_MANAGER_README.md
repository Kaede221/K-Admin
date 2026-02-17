# Casbin Manager Utility

## Overview

The `CasbinManager` provides a high-level utility interface for managing Casbin authorization policies in the K-Admin system. It wraps the Casbin enforcer with convenient methods for common policy operations and includes comprehensive logging.

## Features

- **Policy Enforcement**: Check if a role has permission to access a resource
- **Policy Management**: Add, remove, and update policies
- **Role-Based Queries**: Retrieve all policies for a specific role
- **Batch Updates**: Update all policies for a role in a single operation
- **Error Handling**: Comprehensive error checking and logging
- **Nil Safety**: Handles nil enforcer gracefully

## Usage

### Creating a Manager Instance

```go
import (
    "k-admin-system/global"
    "k-admin-system/utils"
)

// Create a manager using the global enforcer and logger
manager := utils.NewCasbinManager(global.CasbinEnforcer, global.Logger)
```

### Checking Permissions (Enforce)

```go
// Check if "admin" role can GET /api/v1/user
allowed, err := manager.Enforce("admin", "/api/v1/user", "GET")
if err != nil {
    // Handle error
}

if allowed {
    // Permission granted
} else {
    // Permission denied
}
```

**RESTful Path Matching:**
```go
// Add policy with parameter
manager.AddPolicy("admin", "/api/v1/user/:id", "PUT")

// These will match:
manager.Enforce("admin", "/api/v1/user/123", "PUT") // true
manager.Enforce("admin", "/api/v1/user/456", "PUT") // true
```

### Adding Policies

```go
// Add a new policy: admin can GET /api/v1/user
err := manager.AddPolicy("admin", "/api/v1/user", "GET")
if err != nil {
    // Handle error (e.g., policy already exists)
}
```

### Removing Policies

```go
// Remove a policy
err := manager.RemovePolicy("admin", "/api/v1/user", "GET")
if err != nil {
    // Handle error (e.g., policy doesn't exist)
}
```

### Getting Policies for a Role

```go
// Get all policies for the "admin" role
policies, err := manager.GetPoliciesForRole("admin")
if err != nil {
    // Handle error
}

// policies is [][]string, where each element is [role, path, method]
for _, policy := range policies {
    role := policy[0]
    path := policy[1]
    method := policy[2]
    fmt.Printf("Role: %s, Path: %s, Method: %s\n", role, path, method)
}
```

### Updating All Policies for a Role

```go
// Define new policies for the role
newPolicies := [][]string{
    {"admin", "/api/v1/user", "GET"},
    {"admin", "/api/v1/user/:id", "PUT"},
    {"admin", "/api/v1/role", "POST"},
}

// Update all policies for "admin" role
// This removes all existing policies and adds the new ones
err := manager.UpdatePoliciesForRole("admin", newPolicies)
if err != nil {
    // Handle error
}
```

### Accessing the Underlying Enforcer

For advanced operations not covered by the manager:

```go
enforcer := manager.GetEnforcer()

// Use enforcer directly for advanced operations
enforcer.AddGroupingPolicy("user1", "admin") // Role inheritance
```

## API Reference

### NewCasbinManager

```go
func NewCasbinManager(enforcer *casbin.Enforcer, logger *zap.Logger) *CasbinManager
```

Creates a new CasbinManager instance.

**Parameters:**
- `enforcer`: Casbin enforcer instance
- `logger`: Zap logger for logging operations

**Returns:** CasbinManager instance

---

### Enforce

```go
func (cm *CasbinManager) Enforce(role, path, method string) (bool, error)
```

Checks if a role has permission to access a path with a specific method.

**Parameters:**
- `role`: Role name (e.g., "admin", "user")
- `path`: API path (e.g., "/api/v1/user", "/api/v1/user/:id")
- `method`: HTTP method (e.g., "GET", "POST", "PUT", "DELETE")

**Returns:**
- `bool`: true if permission is granted, false otherwise
- `error`: Error if enforcement fails or enforcer is nil

---

### AddPolicy

```go
func (cm *CasbinManager) AddPolicy(role, path, method string) error
```

Adds a new policy rule.

**Parameters:**
- `role`: Role name
- `path`: API path
- `method`: HTTP method

**Returns:**
- `error`: Error if policy already exists, operation fails, or enforcer is nil

---

### RemovePolicy

```go
func (cm *CasbinManager) RemovePolicy(role, path, method string) error
```

Removes an existing policy rule.

**Parameters:**
- `role`: Role name
- `path`: API path
- `method`: HTTP method

**Returns:**
- `error`: Error if policy doesn't exist, operation fails, or enforcer is nil

---

### GetPoliciesForRole

```go
func (cm *CasbinManager) GetPoliciesForRole(role string) ([][]string, error)
```

Retrieves all policies for a specific role.

**Parameters:**
- `role`: Role name

**Returns:**
- `[][]string`: Slice of policy rules, where each rule is [role, path, method]
- `error`: Error if operation fails or enforcer is nil

---

### UpdatePoliciesForRole

```go
func (cm *CasbinManager) UpdatePoliciesForRole(role string, policies [][]string) error
```

Updates all policies for a specific role. This removes all existing policies for the role and adds the new ones.

**Parameters:**
- `role`: Role name
- `policies`: New policies to set, where each policy is [role, path, method]

**Returns:**
- `error`: Error if operation fails, policy format is invalid, role mismatch, or enforcer is nil

**Validation:**
- Each policy must have exactly 3 elements: [role, path, method]
- The role in each policy must match the specified role parameter

---

### GetEnforcer

```go
func (cm *CasbinManager) GetEnforcer() *casbin.Enforcer
```

Returns the underlying Casbin enforcer instance for advanced operations.

**Returns:** Casbin enforcer instance

## Integration with Role Service

The CasbinManager is designed to be used in the RoleService for managing API permissions:

```go
type RoleService struct {
    db            *gorm.DB
    casbinManager *utils.CasbinManager
}

func (s *RoleService) AssignAPIs(roleID uint, policies [][]string) error {
    // Get role key
    role, err := s.GetRoleByID(roleID)
    if err != nil {
        return err
    }

    // Update policies using CasbinManager
    return s.casbinManager.UpdatePoliciesForRole(role.RoleKey, policies)
}

func (s *RoleService) GetRoleAPIs(roleID uint) ([][]string, error) {
    // Get role key
    role, err := s.GetRoleByID(roleID)
    if err != nil {
        return nil, err
    }

    // Get policies using CasbinManager
    return s.casbinManager.GetPoliciesForRole(role.RoleKey)
}
```

## Error Handling

The CasbinManager provides detailed error messages for various scenarios:

- **Nil Enforcer**: Returns error with message "casbin enforcer is not initialized"
- **Duplicate Policy**: Returns error with message "policy already exists"
- **Non-existent Policy**: Returns error with message "policy does not exist"
- **Invalid Format**: Returns error describing the format issue
- **Role Mismatch**: Returns error when policy role doesn't match expected role

All errors are logged with appropriate context using the Zap logger.

## Logging

The CasbinManager logs all operations at different levels:

- **Debug**: Policy enforcement results, policy retrieval
- **Info**: Successful policy additions, removals, and updates
- **Warn**: Duplicate policies, non-existent policies
- **Error**: Operation failures, invalid formats, role mismatches

## Testing

Comprehensive unit tests are provided in `casbin_test.go`:

```bash
# Run all Casbin manager tests
go test -v ./utils -run TestCasbinManager

# Run specific test
go test -v ./utils -run TestCasbinManager_Enforce
```

Tests cover:
- Policy addition and removal
- Permission enforcement with RESTful path matching
- Policy retrieval for roles
- Batch policy updates
- Error handling for nil enforcer
- Invalid input validation

## Requirements Validation

This implementation satisfies:
- **Requirement 3.4**: Backend SHALL use Casbin middleware to enforce API-level permissions
- **Requirement 5.4**: System SHALL support assigning API permissions to roles via Casbin policies

## Next Steps

- **Task 8.3**: Create Casbin authorization middleware using this manager
- **Task 8.4**: Write property tests for authorization
