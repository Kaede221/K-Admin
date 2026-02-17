package utils

import (
	"fmt"

	"github.com/casbin/casbin/v3"
	"go.uber.org/zap"
)

// CasbinManager provides utility functions for managing Casbin policies
type CasbinManager struct {
	enforcer *casbin.Enforcer
	logger   *zap.Logger
}

// NewCasbinManager creates a new CasbinManager instance
func NewCasbinManager(enforcer *casbin.Enforcer, logger *zap.Logger) *CasbinManager {
	return &CasbinManager{
		enforcer: enforcer,
		logger:   logger,
	}
}

// Enforce checks if a role has permission to access a path with a specific method
// Returns true if the permission is granted, false otherwise
func (cm *CasbinManager) Enforce(role, path, method string) (bool, error) {
	if cm.enforcer == nil {
		return false, fmt.Errorf("casbin enforcer is not initialized")
	}

	allowed, err := cm.enforcer.Enforce(role, path, method)
	if err != nil {
		cm.logger.Error("Failed to enforce policy",
			zap.String("role", role),
			zap.String("path", path),
			zap.String("method", method),
			zap.Error(err))
		return false, err
	}

	cm.logger.Debug("Policy enforcement result",
		zap.String("role", role),
		zap.String("path", path),
		zap.String("method", method),
		zap.Bool("allowed", allowed))

	return allowed, nil
}

// AddPolicy adds a new policy rule (role, path, method)
// Returns error if the policy already exists or if the operation fails
func (cm *CasbinManager) AddPolicy(role, path, method string) error {
	if cm.enforcer == nil {
		return fmt.Errorf("casbin enforcer is not initialized")
	}

	success, err := cm.enforcer.AddPolicy(role, path, method)
	if err != nil {
		cm.logger.Error("Failed to add policy",
			zap.String("role", role),
			zap.String("path", path),
			zap.String("method", method),
			zap.Error(err))
		return err
	}

	if !success {
		cm.logger.Warn("Policy already exists",
			zap.String("role", role),
			zap.String("path", path),
			zap.String("method", method))
		return fmt.Errorf("policy already exists")
	}

	cm.logger.Info("Policy added successfully",
		zap.String("role", role),
		zap.String("path", path),
		zap.String("method", method))

	return nil
}

// RemovePolicy removes an existing policy rule (role, path, method)
// Returns error if the policy doesn't exist or if the operation fails
func (cm *CasbinManager) RemovePolicy(role, path, method string) error {
	if cm.enforcer == nil {
		return fmt.Errorf("casbin enforcer is not initialized")
	}

	success, err := cm.enforcer.RemovePolicy(role, path, method)
	if err != nil {
		cm.logger.Error("Failed to remove policy",
			zap.String("role", role),
			zap.String("path", path),
			zap.String("method", method),
			zap.Error(err))
		return err
	}

	if !success {
		cm.logger.Warn("Policy does not exist",
			zap.String("role", role),
			zap.String("path", path),
			zap.String("method", method))
		return fmt.Errorf("policy does not exist")
	}

	cm.logger.Info("Policy removed successfully",
		zap.String("role", role),
		zap.String("path", path),
		zap.String("method", method))

	return nil
}

// GetPoliciesForRole retrieves all policies for a specific role
// Returns a slice of policy rules, where each rule is [role, path, method]
func (cm *CasbinManager) GetPoliciesForRole(role string) ([][]string, error) {
	if cm.enforcer == nil {
		return nil, fmt.Errorf("casbin enforcer is not initialized")
	}

	policies, err := cm.enforcer.GetFilteredPolicy(0, role)
	if err != nil {
		cm.logger.Error("Failed to get policies for role",
			zap.String("role", role),
			zap.Error(err))
		return nil, err
	}

	cm.logger.Debug("Retrieved policies for role",
		zap.String("role", role),
		zap.Int("count", len(policies)))

	return policies, nil
}

// UpdatePoliciesForRole updates all policies for a specific role
// This removes all existing policies for the role and adds the new ones
// The operation is performed within a transaction-like batch operation
func (cm *CasbinManager) UpdatePoliciesForRole(role string, policies [][]string) error {
	if cm.enforcer == nil {
		return fmt.Errorf("casbin enforcer is not initialized")
	}

	// Remove all existing policies for the role
	_, err := cm.enforcer.RemoveFilteredPolicy(0, role)
	if err != nil {
		cm.logger.Error("Failed to remove existing policies for role",
			zap.String("role", role),
			zap.Error(err))
		return fmt.Errorf("failed to remove existing policies: %w", err)
	}

	// Add new policies
	for _, policy := range policies {
		if len(policy) != 3 {
			cm.logger.Error("Invalid policy format",
				zap.String("role", role),
				zap.Any("policy", policy))
			return fmt.Errorf("invalid policy format: expected [role, path, method], got %v", policy)
		}

		// Verify the role matches
		if policy[0] != role {
			cm.logger.Error("Policy role mismatch",
				zap.String("expected_role", role),
				zap.String("policy_role", policy[0]))
			return fmt.Errorf("policy role mismatch: expected %s, got %s", role, policy[0])
		}

		// Convert []string to []interface{} for AddPolicy
		policyInterface := make([]interface{}, len(policy))
		for i, v := range policy {
			policyInterface[i] = v
		}

		_, err := cm.enforcer.AddPolicy(policyInterface...)
		if err != nil {
			cm.logger.Error("Failed to add policy during update",
				zap.String("role", role),
				zap.Any("policy", policy),
				zap.Error(err))
			return fmt.Errorf("failed to add policy: %w", err)
		}
	}

	cm.logger.Info("Policies updated successfully for role",
		zap.String("role", role),
		zap.Int("count", len(policies)))

	return nil
}

// GetEnforcer returns the underlying Casbin enforcer instance
// This is useful for advanced operations not covered by the manager
func (cm *CasbinManager) GetEnforcer() *casbin.Enforcer {
	return cm.enforcer
}
