package service

import (
	"context"

	"devops-platform/internal/modules/user/model"
	"devops-platform/internal/modules/user/repository"

	"gorm.io/gorm"
)

type FieldFilterService struct {
	fieldPermRepo *repository.FieldPermissionRepo
}

func NewFieldFilterService(db *gorm.DB) *FieldFilterService {
	return &FieldFilterService{
		fieldPermRepo: repository.NewFieldPermissionRepo(db),
	}
}

// FilterFields 根据角色过滤响应字段
// hidden -> 移除字段，readonly -> 标记字段
func (s *FieldFilterService) FilterFields(ctx context.Context, input map[string]interface{}, resource string, roleIDs []uint) map[string]interface{} {
	result := make(map[string]interface{})
	for k, v := range input {
		result[k] = v
	}

	fieldPerms, err := s.fieldPermRepo.GetByRoleIDs(roleIDs)
	if err != nil {
		return result
	}

	rules := make(map[string]model.FieldAction)
	for _, fp := range fieldPerms {
		if fp.Resource != resource {
			continue
		}
		existing, exists := rules[fp.FieldName]
		if !exists {
			rules[fp.FieldName] = fp.Action
			continue
		}
		// hidden > readonly > visible 优先级
		if existing == model.FieldActionReadonly && fp.Action == model.FieldActionHidden {
			rules[fp.FieldName] = fp.Action
		}
	}

	for fieldName, action := range rules {
		switch action {
		case model.FieldActionHidden:
			delete(result, fieldName)
		case model.FieldActionReadonly:
			result[fieldName+"_readonly"] = true
		}
	}
	return result
}
