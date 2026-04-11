package repository

import "errors"

var ErrTenantScopeRequired = errors.New("tenant scope is required")

func requireTenantScope(tenantID uint) error {
	if tenantID == 0 {
		return ErrTenantScopeRequired
	}
	return nil
}
