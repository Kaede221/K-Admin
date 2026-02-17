package utils

import (
	"testing"

	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

// createTestEnforcer creates a test Casbin enforcer with in-memory adapter
func createTestEnforcer(t *testing.T) *casbin.Enforcer {
	// Define the model inline
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
	assert.NoError(t, err)

	enforcer, err := casbin.NewEnforcer(m)
	assert.NoError(t, err)

	return enforcer
}

func TestNewCasbinManager(t *testing.T) {
	enforcer := createTestEnforcer(t)
	logger := zap.NewNop()

	manager := NewCasbinManager(enforcer, logger)

	assert.NotNil(t, manager)
	assert.Equal(t, enforcer, manager.enforcer)
	assert.Equal(t, logger, manager.logger)
}

func TestCasbinManager_AddPolicy(t *testing.T) {
	enforcer := createTestEnforcer(t)
	logger := zap.NewNop()
	manager := NewCasbinManager(enforcer, logger)

	tests := []struct {
		name    string
		role    string
		path    string
		method  string
		wantErr bool
	}{
		{
			name:    "Add valid policy",
			role:    "admin",
			path:    "/api/v1/user",
			method:  "GET",
			wantErr: false,
		},
		{
			name:    "Add another valid policy",
			role:    "admin",
			path:    "/api/v1/user/:id",
			method:  "PUT",
			wantErr: false,
		},
		{
			name:    "Add duplicate policy",
			role:    "admin",
			path:    "/api/v1/user",
			method:  "GET",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := manager.AddPolicy(tt.role, tt.path, tt.method)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCasbinManager_Enforce(t *testing.T) {
	enforcer := createTestEnforcer(t)
	logger := zap.NewNop()
	manager := NewCasbinManager(enforcer, logger)

	// Add some policies
	err := manager.AddPolicy("admin", "/api/v1/user", "GET")
	assert.NoError(t, err)
	err = manager.AddPolicy("admin", "/api/v1/user/:id", "PUT")
	assert.NoError(t, err)
	err = manager.AddPolicy("user", "/api/v1/profile", "GET")
	assert.NoError(t, err)

	tests := []struct {
		name    string
		role    string
		path    string
		method  string
		want    bool
		wantErr bool
	}{
		{
			name:    "Admin can GET /api/v1/user",
			role:    "admin",
			path:    "/api/v1/user",
			method:  "GET",
			want:    true,
			wantErr: false,
		},
		{
			name:    "Admin can PUT /api/v1/user/123 (RESTful match)",
			role:    "admin",
			path:    "/api/v1/user/123",
			method:  "PUT",
			want:    true,
			wantErr: false,
		},
		{
			name:    "Admin cannot DELETE /api/v1/user",
			role:    "admin",
			path:    "/api/v1/user",
			method:  "DELETE",
			want:    false,
			wantErr: false,
		},
		{
			name:    "User can GET /api/v1/profile",
			role:    "user",
			path:    "/api/v1/profile",
			method:  "GET",
			want:    true,
			wantErr: false,
		},
		{
			name:    "User cannot GET /api/v1/user",
			role:    "user",
			path:    "/api/v1/user",
			method:  "GET",
			want:    false,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			allowed, err := manager.Enforce(tt.role, tt.path, tt.method)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, allowed)
			}
		})
	}
}

func TestCasbinManager_RemovePolicy(t *testing.T) {
	enforcer := createTestEnforcer(t)
	logger := zap.NewNop()
	manager := NewCasbinManager(enforcer, logger)

	// Add a policy first
	err := manager.AddPolicy("admin", "/api/v1/user", "GET")
	assert.NoError(t, err)

	tests := []struct {
		name    string
		role    string
		path    string
		method  string
		wantErr bool
	}{
		{
			name:    "Remove existing policy",
			role:    "admin",
			path:    "/api/v1/user",
			method:  "GET",
			wantErr: false,
		},
		{
			name:    "Remove non-existent policy",
			role:    "admin",
			path:    "/api/v1/user",
			method:  "POST",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := manager.RemovePolicy(tt.role, tt.path, tt.method)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCasbinManager_GetPoliciesForRole(t *testing.T) {
	enforcer := createTestEnforcer(t)
	logger := zap.NewNop()
	manager := NewCasbinManager(enforcer, logger)

	// Add multiple policies for admin role
	err := manager.AddPolicy("admin", "/api/v1/user", "GET")
	assert.NoError(t, err)
	err = manager.AddPolicy("admin", "/api/v1/user/:id", "PUT")
	assert.NoError(t, err)
	err = manager.AddPolicy("admin", "/api/v1/role", "POST")
	assert.NoError(t, err)

	// Add a policy for user role
	err = manager.AddPolicy("user", "/api/v1/profile", "GET")
	assert.NoError(t, err)

	tests := []struct {
		name      string
		role      string
		wantCount int
		wantErr   bool
	}{
		{
			name:      "Get policies for admin role",
			role:      "admin",
			wantCount: 3,
			wantErr:   false,
		},
		{
			name:      "Get policies for user role",
			role:      "user",
			wantCount: 1,
			wantErr:   false,
		},
		{
			name:      "Get policies for non-existent role",
			role:      "guest",
			wantCount: 0,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			policies, err := manager.GetPoliciesForRole(tt.role)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantCount, len(policies))

				// Verify all policies have the correct role
				for _, policy := range policies {
					assert.Equal(t, tt.role, policy[0])
				}
			}
		})
	}
}

func TestCasbinManager_UpdatePoliciesForRole(t *testing.T) {
	enforcer := createTestEnforcer(t)
	logger := zap.NewNop()
	manager := NewCasbinManager(enforcer, logger)

	// Add initial policies
	err := manager.AddPolicy("admin", "/api/v1/user", "GET")
	assert.NoError(t, err)
	err = manager.AddPolicy("admin", "/api/v1/user/:id", "PUT")
	assert.NoError(t, err)

	tests := []struct {
		name     string
		role     string
		policies [][]string
		wantErr  bool
	}{
		{
			name: "Update with new policies",
			role: "admin",
			policies: [][]string{
				{"admin", "/api/v1/role", "GET"},
				{"admin", "/api/v1/role/:id", "PUT"},
				{"admin", "/api/v1/menu", "POST"},
			},
			wantErr: false,
		},
		{
			name: "Update with invalid policy format",
			role: "admin",
			policies: [][]string{
				{"admin", "/api/v1/user"}, // Missing method
			},
			wantErr: true,
		},
		{
			name: "Update with role mismatch",
			role: "admin",
			policies: [][]string{
				{"user", "/api/v1/profile", "GET"}, // Wrong role
			},
			wantErr: true,
		},
		{
			name:     "Update with empty policies",
			role:     "admin",
			policies: [][]string{},
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := manager.UpdatePoliciesForRole(tt.role, tt.policies)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)

				// Verify the policies were updated correctly
				currentPolicies, err := manager.GetPoliciesForRole(tt.role)
				assert.NoError(t, err)
				assert.Equal(t, len(tt.policies), len(currentPolicies))
			}
		})
	}
}

func TestCasbinManager_GetEnforcer(t *testing.T) {
	enforcer := createTestEnforcer(t)
	logger := zap.NewNop()
	manager := NewCasbinManager(enforcer, logger)

	returnedEnforcer := manager.GetEnforcer()
	assert.Equal(t, enforcer, returnedEnforcer)
}

func TestCasbinManager_NilEnforcer(t *testing.T) {
	logger := zap.NewNop()
	manager := &CasbinManager{
		enforcer: nil,
		logger:   logger,
	}

	// Test all methods with nil enforcer
	t.Run("Enforce with nil enforcer", func(t *testing.T) {
		_, err := manager.Enforce("admin", "/api/v1/user", "GET")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not initialized")
	})

	t.Run("AddPolicy with nil enforcer", func(t *testing.T) {
		err := manager.AddPolicy("admin", "/api/v1/user", "GET")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not initialized")
	})

	t.Run("RemovePolicy with nil enforcer", func(t *testing.T) {
		err := manager.RemovePolicy("admin", "/api/v1/user", "GET")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not initialized")
	})

	t.Run("GetPoliciesForRole with nil enforcer", func(t *testing.T) {
		_, err := manager.GetPoliciesForRole("admin")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not initialized")
	})

	t.Run("UpdatePoliciesForRole with nil enforcer", func(t *testing.T) {
		err := manager.UpdatePoliciesForRole("admin", [][]string{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not initialized")
	})
}
