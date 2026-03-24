package utils

import (
	"fmt"
	"strings"
	"unicode"

	"golang.org/x/crypto/bcrypt"
)

// HashPassword 对密码进行哈希
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// CheckPassword 验证密码
func CheckPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// ValidatePasswordComplexity 校验密码复杂度
// 规则：最少 8 位，包含大写字母、小写字母、数字、特殊字符中的至少 3 种
func ValidatePasswordComplexity(password string) error {
	if len(password) < 8 {
		return fmt.Errorf("密码长度不能少于 8 位")
	}
	if len(password) > 128 {
		return fmt.Errorf("密码长度不能超过 128 位")
	}

	var hasUpper, hasLower, hasDigit, hasSpecial bool
	for _, ch := range password {
		switch {
		case unicode.IsUpper(ch):
			hasUpper = true
		case unicode.IsLower(ch):
			hasLower = true
		case unicode.IsDigit(ch):
			hasDigit = true
		case unicode.IsPunct(ch) || unicode.IsSymbol(ch):
			hasSpecial = true
		}
	}

	complexity := 0
	if hasUpper {
		complexity++
	}
	if hasLower {
		complexity++
	}
	if hasDigit {
		complexity++
	}
	if hasSpecial {
		complexity++
	}

	if complexity < 3 {
		var missing []string
		if !hasUpper {
			missing = append(missing, "大写字母")
		}
		if !hasLower {
			missing = append(missing, "小写字母")
		}
		if !hasDigit {
			missing = append(missing, "数字")
		}
		if !hasSpecial {
			missing = append(missing, "特殊字符")
		}
		return fmt.Errorf("密码需包含大写字母、小写字母、数字、特殊字符中的至少 3 种，当前缺少: %s", strings.Join(missing, "、"))
	}
	return nil
}
