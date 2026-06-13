package service

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"devops-platform/internal/modules/user/model"
	"devops-platform/internal/modules/user/repository"

	"gorm.io/gorm"
)

type ApiKeyService struct {
	repo *repository.ApiKeyRepo
	db   *gorm.DB
}

func NewApiKeyService(db *gorm.DB) *ApiKeyService {
	return &ApiKeyService{
		repo: repository.NewApiKeyRepo(db),
		db:   db,
	}
}

func (s *ApiKeyService) Create(userID, tenantID uint, req model.CreateApiKeyRequest) (*model.ApiKeyResponse, error) {
	key, hash, err := generateApiKey()
	if err != nil {
		return nil, fmt.Errorf("生成 API Key 失败: %w", err)
	}

	scopes := strings.Join(req.Scopes, ",")
	apiKey := &model.ApiKey{
		UserID:    userID,
		TenantID:  tenantID,
		Name:      req.Name,
		KeyHash:   hash,
		KeyPrefix: key[:11],
		Scopes:    scopes,
	}
	if req.ExpireDays > 0 {
		exp := time.Now().Add(time.Duration(req.ExpireDays) * 24 * time.Hour)
		apiKey.ExpiresAt = &exp
	}

	if err := s.repo.Create(apiKey); err != nil {
		return nil, err
	}

	return &model.ApiKeyResponse{
		ID:        apiKey.ID,
		Name:      apiKey.Name,
		KeyPrefix: apiKey.KeyPrefix,
		Scopes:    apiKey.Scopes,
		ExpiresAt: apiKey.ExpiresAt,
		CreatedAt: apiKey.CreatedAt,
		Key:       key,
	}, nil
}

func (s *ApiKeyService) List(userID, tenantID uint, req model.ListApiKeyRequest) ([]model.ApiKey, int64, error) {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}
	keys, total, err := s.repo.List(userID, tenantID, req.Keyword, req.Page, req.PageSize)
	if err != nil {
		return nil, 0, err
	}
	return keys, total, nil
}

func (s *ApiKeyService) Delete(id, userID, tenantID uint) error {
	return s.repo.Delete(id, userID, tenantID)
}

func (s *ApiKeyService) Validate(apiKey string) (*model.ApiKey, error) {
	hash := hashApiKey(apiKey)
	key, err := s.repo.FindByHash(hash)
	if err != nil {
		return nil, err
	}
	if key.ExpiresAt != nil && key.ExpiresAt.Before(time.Now()) {
		return nil, fmt.Errorf("API Key 已过期")
	}
	s.repo.UpdateLastUsed(key.ID)
	return key, nil
}

func generateApiKey() (raw string, hash string, err error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", "", err
	}
	raw = "ak_" + hex.EncodeToString(b)
	hash = hashApiKey(raw)
	return raw, hash, nil
}

func hashApiKey(key string) string {
	h := sha256.Sum256([]byte(key))
	return hex.EncodeToString(h[:])
}
