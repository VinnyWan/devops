package jwt

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrTokenExpired = errors.New("token has expired")
	ErrTokenInvalid = errors.New("token is invalid")
)

type Claims struct {
	UserID     uint   `json:"user_id"`
	Username   string `json:"username"`
	TenantID   uint   `json:"tenant_id"`
	TenantCode string `json:"tenant_code"`
	TokenType  string `json:"token_type"` // "access" or "refresh"
	jwt.RegisteredClaims
}

type Manager struct {
	secret       []byte
	accessExpiry time.Duration
	refreshExpiry time.Duration
}

func NewManager(secret string, accessExpiry, refreshExpiry time.Duration) *Manager {
	if secret == "" {
		b := make([]byte, 32)
		rand.Read(b)
		secret = hex.EncodeToString(b)
	}
	return &Manager{
		secret:        []byte(secret),
		accessExpiry:  accessExpiry,
		refreshExpiry: refreshExpiry,
	}
}

var defaultManager *Manager

func InitDefault(secret string) {
	defaultManager = NewManager(secret, 1*time.Hour, 7*24*time.Hour)
}

func Default() *Manager {
	if defaultManager == nil {
		defaultManager = NewManager("", 1*time.Hour, 7*24*time.Hour)
	}
	return defaultManager
}

func (m *Manager) GenerateAccessToken(userID uint, username string, tenantID uint, tenantCode string) (string, error) {
	return m.generate(userID, username, tenantID, tenantCode, "access", m.accessExpiry)
}

func (m *Manager) GenerateRefreshToken(userID uint, username string, tenantID uint, tenantCode string) (string, error) {
	return m.generate(userID, username, tenantID, tenantCode, "refresh", m.refreshExpiry)
}

func (m *Manager) generate(userID uint, username string, tenantID uint, tenantCode string, tokenType string, expiry time.Duration) (string, error) {
	now := time.Now()
	claims := &Claims{
		UserID:     userID,
		Username:   username,
		TenantID:   tenantID,
		TenantCode: tenantCode,
		TokenType:  tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(expiry)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(m.secret)
}

func (m *Manager) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return m.secret, nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrTokenExpired
		}
		return nil, ErrTokenInvalid
	}
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrTokenInvalid
	}
	return claims, nil
}

func (m *Manager) RefreshAccessToken(refreshToken string) (string, error) {
	claims, err := m.ValidateToken(refreshToken)
	if err != nil {
		return "", err
	}
	if claims.TokenType != "refresh" {
		return "", ErrTokenInvalid
	}
	return m.GenerateAccessToken(claims.UserID, claims.Username, claims.TenantID, claims.TenantCode)
}
