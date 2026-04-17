package service

import (
	"devops-platform/internal/modules/cmdb/model"
	"devops-platform/internal/modules/cmdb/repository"
	"devops-platform/internal/pkg/utils"
	"errors"

	"gorm.io/gorm"
)

type CredentialService struct {
	repo *repository.CredentialRepo
}

func NewCredentialService(db *gorm.DB) *CredentialService {
	return &CredentialService{repo: repository.NewCredentialRepo(db)}
}

type CredentialCreateRequest struct {
	Name        string `json:"name" binding:"required"`
	Type        string `json:"type" binding:"required,oneof=password key"`
	Username    string `json:"username" binding:"required"`
	Password    string `json:"password"`
	PrivateKey  string `json:"privateKey"`
	Passphrase  string `json:"passphrase"`
	Description string `json:"description"`
}

type CredentialUpdateRequest struct {
	ID          uint   `json:"id" binding:"required"`
	Type        string `json:"type" binding:"omitempty,oneof=password key"`
	Name        string `json:"name"`
	Username    string `json:"username"`
	Password    string `json:"password"`
	PrivateKey  string `json:"privateKey"`
	Passphrase  string `json:"passphrase"`
	Description string `json:"description"`
}

func (s *CredentialService) normalizePage(page, pageSize int) (int, int) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}
	return page, pageSize
}

func (s *CredentialService) ListInTenant(tenantID uint, page, pageSize int, keyword string) ([]model.Credential, int64, error) {
	page, pageSize = s.normalizePage(page, pageSize)
	return s.repo.ListInTenant(tenantID, page, pageSize, keyword)
}

func (s *CredentialService) GetByIDInTenant(tenantID uint, id uint) (*model.Credential, error) {
	cred, err := s.repo.GetByIDInTenant(tenantID, id)
	if err != nil {
		return nil, err
	}
	// 不返回敏感字段
	cred.Password = ""
	cred.PrivateKey = ""
	cred.Passphrase = ""
	return cred, nil
}

func (s *CredentialService) CreateInTenant(tenantID uint, req *CredentialCreateRequest) (*model.Credential, error) {
	cred := &model.Credential{
		Name:        req.Name,
		Type:        req.Type,
		Username:    req.Username,
		Description: req.Description,
	}

	// 加密敏感字段
	if req.Type == "password" {
		if req.Password == "" {
			return nil, errors.New("密码不能为空")
		}
		encrypted, err := utils.Encrypt(req.Password)
		if err != nil {
			return nil, errors.New("密码加密失败")
		}
		cred.Password = encrypted
	} else if req.Type == "key" {
		if req.PrivateKey == "" {
			return nil, errors.New("私钥不能为空")
		}
		encryptedKey, err := utils.Encrypt(req.PrivateKey)
		if err != nil {
			return nil, errors.New("私钥加密失败")
		}
		cred.PrivateKey = encryptedKey

		if req.Passphrase != "" {
			encryptedPass, err := utils.Encrypt(req.Passphrase)
			if err != nil {
				return nil, errors.New("密钥密码加密失败")
			}
			cred.Passphrase = encryptedPass
		}
	}

	if err := s.repo.CreateInTenant(tenantID, cred); err != nil {
		return nil, err
	}

	// 返回时不包含敏感字段
	cred.Password = ""
	cred.PrivateKey = ""
	cred.Passphrase = ""
	return cred, nil
}

func (s *CredentialService) UpdateInTenant(tenantID uint, req *CredentialUpdateRequest) (*model.Credential, error) {
	cred, err := s.repo.GetByIDInTenant(tenantID, req.ID)
	if err != nil {
		return nil, err
	}

	if req.Type != "" && req.Type != cred.Type {
		switch req.Type {
		case "password":
			if req.Password == "" {
				return nil, errors.New("切换为密码类型时必须填写密码")
			}
			cred.PrivateKey = ""
			cred.Passphrase = ""
		case "key":
			if req.PrivateKey == "" {
				return nil, errors.New("切换为密钥类型时必须填写私钥")
			}
			cred.Password = ""
			cred.Passphrase = ""
		}
	}
	if req.Type != "" {
		cred.Type = req.Type
	}
	if req.Name != "" {
		cred.Name = req.Name
	}
	if req.Username != "" {
		cred.Username = req.Username
	}
	if req.Password != "" {
		encrypted, err := utils.Encrypt(req.Password)
		if err != nil {
			return nil, errors.New("密码加密失败")
		}
		cred.Password = encrypted
		if cred.Type == "password" {
			cred.PrivateKey = ""
			cred.Passphrase = ""
		}
	}
	if req.PrivateKey != "" {
		encryptedKey, err := utils.Encrypt(req.PrivateKey)
		if err != nil {
			return nil, errors.New("私钥加密失败")
		}
		cred.PrivateKey = encryptedKey
		if cred.Type == "key" {
			cred.Password = ""
		}
	}
	if req.Passphrase != "" {
		encryptedPass, err := utils.Encrypt(req.Passphrase)
		if err != nil {
			return nil, errors.New("密钥密码加密失败")
		}
		cred.Passphrase = encryptedPass
	}
	if req.Description != "" {
		cred.Description = req.Description
	}

	if err := s.repo.UpdateInTenant(tenantID, cred); err != nil {
		return nil, err
	}

	cred.Password = ""
	cred.PrivateKey = ""
	cred.Passphrase = ""
	return cred, nil
}

func (s *CredentialService) DeleteInTenant(tenantID uint, id uint) error {
	// 检查是否被主机引用
	referenced, err := s.repo.IsReferencedByHosts(id)
	if err != nil {
		return err
	}
	if referenced {
		return errors.New("该凭据正在被主机使用，无法删除")
	}

	return s.repo.DeleteInTenant(tenantID, id)
}

// GetDecryptedInTenant 获取租户内的解密凭据（仅内部使用，用于 SSH 连接）
func (s *CredentialService) GetDecryptedInTenant(tenantID uint, id uint) (*model.Credential, error) {
	return s.repo.GetByIDInTenant(tenantID, id)
}
