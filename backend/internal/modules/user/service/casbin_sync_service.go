package service

import (
	"fmt"

	"devops-platform/internal/modules/user/model"
	"devops-platform/internal/modules/user/repository"

	"github.com/casbin/casbin/v3"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// userRoleBinding 用户-角色绑定关系
type userRoleBinding struct {
	UserID uint
	RoleID uint
}

// CasbinSyncService 负责 Casbin 策略同步
type CasbinSyncService struct {
	db             *gorm.DB
	enforcer       *casbin.SyncedEnforcer
	roleRepo       *repository.RoleRepo
	userRepo       *repository.UserRepo
	deptRepo       *repository.DepartmentRepo
	permissionRepo *repository.PermissionRepo
}

// NewCasbinSyncService 创建 CasbinSyncService 实例
// enforcer 由调用方通过 bootstrap.GetEnforcer() 传入，避免循环依赖
func NewCasbinSyncService(db *gorm.DB, enforcer *casbin.SyncedEnforcer) *CasbinSyncService {
	return &CasbinSyncService{
		db:             db,
		enforcer:       enforcer,
		roleRepo:       repository.NewRoleRepo(db),
		userRepo:       repository.NewUserRepo(db),
		deptRepo:       repository.NewDepartmentRepo(db),
		permissionRepo: repository.NewPermissionRepo(db),
	}
}

// SyncTenantPolicies 全量同步某租户的所有策略
// 清除旧策略 -> 同步角色权限(p) -> 同步用户角色(g)
func (s *CasbinSyncService) SyncTenantPolicies(tenantID uint) error {
	if s.enforcer == nil {
		return fmt.Errorf("casbin enforcer not initialized")
	}

	dom := fmt.Sprintf("%d", tenantID)

	// 清除该租户的所有策略 (sub, dom, obj, act) -> 按 dom 字段过滤
	if _, err := s.enforcer.RemoveFilteredPolicy(1, dom); err != nil {
		return fmt.Errorf("failed to remove policies for tenant %d: %w", tenantID, err)
	}
	// 清除该租户的所有角色绑定 (sub, role, dom) -> 按 dom 字段过滤
	if _, err := s.enforcer.RemoveFilteredGroupingPolicy(2, dom); err != nil {
		return fmt.Errorf("failed to remove grouping policies for tenant %d: %w", tenantID, err)
	}

	// 获取租户下所有角色（大页量获取，避免遗漏）
	roles, _, err := s.roleRepo.ListInTenant(tenantID, 1, 10000, "")
	if err != nil {
		return fmt.Errorf("failed to list roles for tenant %d: %w", tenantID, err)
	}

	// 同步每个角色的权限策略
	for _, role := range roles {
		if err := s.syncRolePermissionsToEnforcer(role, dom); err != nil {
			return fmt.Errorf("failed to sync role %d permissions: %w", role.ID, err)
		}
	}

	// 同步用户角色绑定
	if err := s.syncAllUserRoleBindings(tenantID, dom); err != nil {
		return fmt.Errorf("failed to sync user role bindings: %w", err)
	}

	return nil
}

// SyncUserRoles 同步单个用户角色
// 清除旧绑定 -> 添加新绑定
func (s *CasbinSyncService) SyncUserRoles(userID, tenantID uint, roleIDs []uint) error {
	if s.enforcer == nil {
		return fmt.Errorf("casbin enforcer not initialized")
	}

	dom := fmt.Sprintf("%d", tenantID)
	sub := fmt.Sprintf("%d", userID)

	// 清除该用户在租户下的所有角色绑定 (sub, role, dom)
	if _, err := s.enforcer.RemoveFilteredGroupingPolicy(0, sub, "", dom); err != nil {
		return fmt.Errorf("failed to remove user %d role bindings: %w", userID, err)
	}

	// 添加新的角色绑定
	for _, roleID := range roleIDs {
		roleSub := fmt.Sprintf("role:%d", roleID)
		if _, err := s.enforcer.AddGroupingPolicy(sub, roleSub, dom); err != nil {
			return fmt.Errorf("failed to add grouping policy for user %d role %d: %w", userID, roleID, err)
		}
	}

	return nil
}

// SyncRolePermissions 同步角色权限
// 清除旧策略 -> 添加新策略
func (s *CasbinSyncService) SyncRolePermissions(roleID, tenantID uint, permissions []model.Permission) error {
	if s.enforcer == nil {
		return fmt.Errorf("casbin enforcer not initialized")
	}

	dom := fmt.Sprintf("%d", tenantID)
	roleSub := fmt.Sprintf("role:%d", roleID)

	// 清除该角色的所有权限策略
	if _, err := s.enforcer.RemoveFilteredPolicy(0, roleSub, dom); err != nil {
		return fmt.Errorf("failed to remove policies for role %d: %w", roleID, err)
	}

	// 添加新的权限策略
	for _, perm := range permissions {
		if _, err := s.enforcer.AddPolicy(roleSub, dom, perm.Resource, perm.Action); err != nil {
			return fmt.Errorf("failed to add policy for role %d perm %d: %w", roleID, perm.ID, err)
		}
	}

	return nil
}

// SyncDepartmentRoles 同步部门角色到用户
// 获取部门角色 -> 获取部门用户 -> 添加绑定
func (s *CasbinSyncService) SyncDepartmentRoles(deptID, tenantID uint) error {
	if s.enforcer == nil {
		return fmt.Errorf("casbin enforcer not initialized")
	}

	dom := fmt.Sprintf("%d", tenantID)

	// 获取部门信息（含关联角色）
	dept, err := s.deptRepo.GetByIDInTenant(tenantID, deptID)
	if err != nil {
		return fmt.Errorf("failed to get department %d: %w", deptID, err)
	}

	if len(dept.Roles) == 0 {
		return nil
	}

	// 获取部门下的所有用户ID
	userIDs, err := s.userRepo.ListUserIDsByDepartmentIDInTenant(tenantID, deptID)
	if err != nil {
		return fmt.Errorf("failed to list users in department %d: %w", deptID, err)
	}

	if len(userIDs) == 0 {
		return nil
	}

	// 为每个用户添加部门角色绑定
	for _, uid := range userIDs {
		sub := fmt.Sprintf("%d", uid)
		for _, role := range dept.Roles {
			roleSub := fmt.Sprintf("role:%d", role.ID)
			if _, err := s.enforcer.AddGroupingPolicy(sub, roleSub, dom); err != nil {
				return fmt.Errorf("failed to add dept role binding user %d role %d: %w", uid, role.ID, err)
			}
		}
	}

	return nil
}

// -------------------------------------------------------------------
// 私有辅助方法
// -------------------------------------------------------------------

// syncRolePermissionsToEnforcer 将单个角色的权限写入 Casbin
func (s *CasbinSyncService) syncRolePermissionsToEnforcer(role model.Role, dom string) error {
	roleSub := fmt.Sprintf("role:%d", role.ID)

	if len(role.Permissions) == 0 {
		return nil
	}

	// 批量构建策略
	rules := make([][]string, 0, len(role.Permissions))
	for _, perm := range role.Permissions {
		rules = append(rules, []string{roleSub, dom, perm.Resource, perm.Action})
	}

	if len(rules) > 0 {
		if _, err := s.enforcer.AddPolicies(rules); err != nil {
			zap.L().Error("同步角色权限策略失败",
				zap.Uint("roleID", role.ID),
				zap.String("dom", dom),
				zap.Error(err),
			)
			return fmt.Errorf("failed to add policies for role %d: %w", role.ID, err)
		}
	}

	return nil
}

// syncAllUserRoleBindings 同步租户下所有用户的角色绑定
func (s *CasbinSyncService) syncAllUserRoleBindings(tenantID uint, dom string) error {
	// 使用 DB 直接查询，获取所有用户角色关联
	var userRoles []userRoleBinding
	err := s.db.Table("user_roles").
		Select("user_roles.user_id, user_roles.role_id").
		Joins("JOIN users ON users.id = user_roles.user_id").
		Joins("JOIN roles ON roles.id = user_roles.role_id").
		Where("users.tenant_id = ? AND (roles.tenant_id = ? OR roles.tenant_id IS NULL)", tenantID, tenantID).
		Find(&userRoles).Error
	if err != nil {
		return fmt.Errorf("failed to query user role bindings: %w", err)
	}

	// 批量构建角色绑定策略
	rules := make([][]string, 0, len(userRoles))
	for _, ur := range userRoles {
		sub := fmt.Sprintf("%d", ur.UserID)
		roleSub := fmt.Sprintf("role:%d", ur.RoleID)
		rules = append(rules, []string{sub, roleSub, dom})
	}

	// 处理部门角色绑定
	deptRoles, err := s.queryDepartmentUserRoles(tenantID)
	if err != nil {
		return fmt.Errorf("failed to query department role bindings: %w", err)
	}

	for _, dr := range deptRoles {
		sub := fmt.Sprintf("%d", dr.UserID)
		roleSub := fmt.Sprintf("role:%d", dr.RoleID)
		rules = append(rules, []string{sub, roleSub, dom})
	}

	if len(rules) > 0 {
		if _, err := s.enforcer.AddGroupingPolicies(rules); err != nil {
			return fmt.Errorf("failed to add grouping policies: %w", err)
		}
	}

	return nil
}

// queryDepartmentUserRoles 查询部门角色到用户的映射关系
func (s *CasbinSyncService) queryDepartmentUserRoles(tenantID uint) ([]userRoleBinding, error) {
	var bindings []userRoleBinding
	err := s.db.Table("department_roles").
		Select("users.id as user_id, department_roles.role_id").
		Joins("JOIN users ON users.primary_dept_id = department_roles.department_id").
		Joins("JOIN roles ON roles.id = department_roles.role_id").
		Where("users.tenant_id = ? AND (roles.tenant_id = ? OR roles.tenant_id IS NULL)", tenantID, tenantID).
		Find(&bindings).Error

	return bindings, err
}
