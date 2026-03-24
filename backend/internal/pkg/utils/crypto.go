package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"devops-platform/config"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
)

var key []byte

// InitCrypto 初始化加密密钥
// release 模式下必须配置 crypto.secret，否则启动失败
func InitCrypto() error {
	secret := config.Cfg.GetString("crypto.secret")
	serverMode := config.Cfg.GetString("server.mode")

	if secret == "" {
		if serverMode == "release" {
			return fmt.Errorf("生产环境必须配置 crypto.secret，禁止使用默认密钥")
		}
		secret = "devops-platform-default-dev-key!"
		fmt.Println("Warning: 使用默认加密密钥，仅限开发环境")
	}

	// 确保密钥长度为32字节（AES-256）
	if len(secret) > 32 {
		key = []byte(secret[:32])
	} else if len(secret) < 32 {
		padded := make([]byte, 32)
		copy(padded, []byte(secret))
		key = padded
	} else {
		key = []byte(secret)
	}
	return nil
}

// Encrypt 加密数据
func Encrypt(text string) (string, error) {
	plaintext := []byte(text)
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	ciphertext := make([]byte, aes.BlockSize+len(plaintext))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", err
	}

	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], plaintext)

	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt 解密数据
func Decrypt(cryptoText string) (string, error) {
	ciphertext, err := base64.StdEncoding.DecodeString(cryptoText)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	if len(ciphertext) < aes.BlockSize {
		return "", errors.New("ciphertext too short")
	}
	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]

	stream := cipher.NewCFBDecrypter(block, iv)
	stream.XORKeyStream(ciphertext, ciphertext)

	return string(ciphertext), nil
}
