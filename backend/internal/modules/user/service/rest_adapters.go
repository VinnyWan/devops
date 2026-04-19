package service

import (
	"context"
	"errors"

	"devops-platform/internal/modules/user/model"
	"devops-platform/internal/pkg/utils"
)

// AdminCreateUser 管理员创建用户（系统管理接口）
func (s *UserService) AdminCreateUser(ctx context.Context, tenantID uint, operatorID uint, username, password, nickname, email, phone *string) (*model.User, error) {
	if username == nil || *username == "" {
		return nil, errors.New("username is required")
	}
	if password == nil || *password == "" {
		return nil, errors.New("password is required")
	}
	if err := utils.ValidatePasswordComplexity(*password); err != nil {
		return nil, err
	}

	// 检查用户名唯一性
	if existing, err := s.userRepo.GetByUsernameInTenant(tenantID, *username); err == nil && existing != nil {
		return nil, errors.New("username already exists")
	}

	hashedPassword, err := utils.HashPassword(*password)
	if err != nil {
		return nil, err
	}

	name := ""
	if nickname != nil {
		name = *nickname
	}
	emailVal := ""
	if email != nil {
		emailVal = *email
	}

	user := &model.User{
		TenantID:  &tenantID,
		Username:  *username,
		Password:  hashedPassword,
		Name:      name,
		Email:     emailVal,
		AuthType:  model.AuthTypeLocal,
		Status:    "active",
		IsAdmin:   false,
		IsLocked:  false,
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, err
	}
	return user, nil
}

// UpdateUserByMap 使用 map 更新用户（RESTful 路由适配）
func (s *UserService) UpdateUserByMap(ctx context.Context, tenantID uint, operatorID uint, userID uint, data map[string]interface{}) error {
	req := &UpdateUserRequest{ID: userID}

	if v, ok := data["username"].(string); ok {
		req.Username = &v
	}
	if v, ok := data["name"].(string); ok {
		req.Name = &v
	}
	if v, ok := data["email"].(string); ok {
		req.Email = &v
	}
	if v, ok := data["status"].(string); ok {
		req.Status = &v
	}
	if v, ok := data["primaryDeptId"].(float64); ok {
		id := uint(v)
		req.PrimaryDeptID = &id
	}

	return s.UpdateUserByRequest(ctx, tenantID, operatorID, req)
}

// UpdateRoleByMap 使用 map 更新角色（RESTful 路由适配）
func (s *RoleService) UpdateRoleByMap(tenantID uint, roleID uint, data map[string]interface{}) error {
	req := &UpdateRoleRequest{ID: roleID}

	if v, ok := data["name"].(string); ok {
		req.Name = v
	}
	if v, ok := data["displayName"].(string); ok {
		req.DisplayName = v
	}
	if v, ok := data["description"].(string); ok {
		req.Description = v
	}
	if v, ok := data["dataScope"].(string); ok {
		req.DataScope = &v
	}
	if v, ok := data["permissionIds"].([]interface{}); ok {
		for _, item := range v {
			if f, ok := item.(float64); ok {
				req.PermissionIDs = append(req.PermissionIDs, uint(f))
			}
		}
	}

	return s.UpdateRole(tenantID, req)
}

// UpdateByMap 使用 map 更新租户（RESTful 路由适配）
func (s *TenantService) UpdateByMap(tenantID uint, data map[string]interface{}) error {
	req := &UpdateTenantRequest{ID: tenantID}

	if v, ok := data["name"].(string); ok {
		req.Name = &v
	}
	if v, ok := data["description"].(string); ok {
		req.Description = &v
	}
	if v, ok := data["logo"].(string); ok {
		req.Logo = &v
	}
	if v, ok := data["status"].(string); ok {
		req.Status = &v
	}
	if v, ok := data["maxUsers"].(float64); ok {
		i := int(v)
		req.MaxUsers = &i
	}
	if v, ok := data["maxDepartments"].(float64); ok {
		i := int(v)
		req.MaxDepartments = &i
	}
	if v, ok := data["maxRoles"].(float64); ok {
		i := int(v)
		req.MaxRoles = &i
	}
	if v, ok := data["modules"].(string); ok {
		req.Modules = &v
	}
	if v, ok := data["contactName"].(string); ok {
		req.ContactName = &v
	}
	if v, ok := data["contactEmail"].(string); ok {
		req.ContactEmail = &v
	}
	if v, ok := data["contactPhone"].(string); ok {
		req.ContactPhone = &v
	}

	return s.Update(req)
}

// UpdateByMap 使用 map 更新部门（RESTful 路由适配）
func (s *DepartmentService) UpdateByMap(ctx context.Context, tenantID uint, operatorID uint, deptID uint, data map[string]interface{}) error {
	req := &UpdateDepartmentRequest{ID: deptID}

	if v, ok := data["name"].(string); ok {
		req.Name = v
	}
	if v, ok := data["parentId"].(float64); ok {
		id := uint(v)
		req.ParentID = &id
	}

	return s.Update(ctx, tenantID, operatorID, req)
}

// CreateByMap 使用 map 创建部门用户（RESTful 路由适配）
func (s *DepartmentUserService) CreateByMap(tenantID uint, operatorID uint, deptID uint, data map[string]interface{}) (*model.User, error) {
	req := &CreateDeptUserRequest{DepartmentID: deptID}

	if v, ok := data["username"].(string); ok {
		req.Username = v
	}
	if v, ok := data["password"].(string); ok {
		req.Password = v
	}
	if v, ok := data["email"].(string); ok {
		req.Email = v
	}
	if v, ok := data["name"].(string); ok {
		req.Name = v
	}
	if v, ok := data["status"].(string); ok {
		req.Status = v
	}

	return s.Create(tenantID, operatorID, req)
}

// UpdateByMap 使用 map 更新部门用户（RESTful 路由适配）
func (s *DepartmentUserService) UpdateByMap(tenantID uint, operatorID uint, userID uint, data map[string]interface{}) error {
	req := &UpdateDeptUserRequest{ID: userID}

	if v, ok := data["email"].(string); ok {
		req.Email = &v
	}
	if v, ok := data["name"].(string); ok {
		req.Name = &v
	}
	if v, ok := data["status"].(string); ok {
		req.Status = &v
	}

	return s.Update(tenantID, operatorID, req)
}
