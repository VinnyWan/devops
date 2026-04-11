package service

import (
	"errors"
	"sync"

	"devops-platform/internal/modules/app/model"
	"devops-platform/internal/modules/app/repository"
)

// EnumService 枚举服务
type EnumService struct {
	repo *repository.AppRepo
	mu   sync.RWMutex
}

// NewEnumService 创建枚举服务
func NewEnumService() *EnumService {
	return NewEnumServiceWithRepo(repository.NewAppRepo())
}

func NewEnumServiceWithRepo(repo *repository.AppRepo) *EnumService {
	if repo == nil {
		repo = repository.NewAppRepo()
	}
	return &EnumService{
		repo: repo,
	}
}

// ListEnums 获取枚举列表
func (s *EnumService) ListEnums(enumType string) []model.Enum {
	return s.repo.ListEnums(enumType)
}

// ListAllEnums 获取所有枚举（包括禁用的）
func (s *EnumService) ListAllEnums(enumType string) []model.Enum {
	return s.repo.ListAllEnums(enumType)
}

// GetEnumByID 根据ID获取枚举
func (s *EnumService) GetEnumByID(id uint) (model.Enum, error) {
	enum, ok := s.repo.GetEnumByID(id)
	if !ok {
		return model.Enum{}, errors.New("枚举不存在")
	}
	return enum, nil
}

// SaveEnum 保存枚举
func (s *EnumService) SaveEnum(req model.EnumRequest) (model.Enum, error) {
	// 验证枚举类型
	validTypes := s.repo.GetEnumTypes()
	isValidType := false
	for _, t := range validTypes {
		if t == req.EnumType {
			isValidType = true
			break
		}
	}
	if !isValidType {
		return model.Enum{}, errors.New("无效的枚举类型")
	}

	// 验证必填字段
	if req.EnumKey == "" {
		return model.Enum{}, errors.New("枚举键不能为空")
	}
	if req.EnumValue == "" {
		return model.Enum{}, errors.New("枚举值不能为空")
	}

	enum := model.Enum{
		ID:        req.ID,
		EnumType:  req.EnumType,
		EnumKey:   req.EnumKey,
		EnumValue: req.EnumValue,
		SortOrder: req.SortOrder,
		IsActive:  req.IsActive,
	}

	return s.repo.SaveEnum(enum), nil
}

// DeleteEnum 删除枚举
func (s *EnumService) DeleteEnum(id uint) error {
	// 获取枚举信息
	enum, ok := s.repo.GetEnumByID(id)
	if !ok {
		return errors.New("枚举不存在")
	}

	// 检查是否被使用
	if s.repo.CheckEnumUsage(enum.EnumType, enum.EnumKey) {
		return errors.New("该枚举值正在被使用，无法删除")
	}

	if !s.repo.DeleteEnum(id) {
		return errors.New("删除失败")
	}
	return nil
}

// GetEnumTypes 获取所有枚举类型
func (s *EnumService) GetEnumTypes() []string {
	return s.repo.GetEnumTypes()
}

// GetEnumsByType 按类型获取枚举，返回前端友好的格式
func (s *EnumService) GetEnumsByType() map[string]interface{} {
	result := make(map[string]interface{})

	types := s.repo.GetEnumTypes()
	for _, t := range types {
		enums := s.repo.ListEnums(t)
		options := make([]map[string]interface{}, 0, len(enums))
		for _, e := range enums {
			options = append(options, map[string]interface{}{
				"label": e.EnumValue,
				"value": e.EnumKey,
			})
		}
		result[t] = options
	}

	return result
}
