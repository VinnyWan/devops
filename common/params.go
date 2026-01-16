package common

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// RequireUintQuery 读取必填的 uint 查询参数，失败时写入 400 响应并返回 false。
func RequireUintQuery(c *gin.Context, name string) (uint, bool) {
	value := strings.TrimSpace(c.Query(name))
	if value == "" {
		BadRequest(c, fmt.Sprintf("参数%s不能为空", name))
		return 0, false
	}

	parsed, err := strconv.ParseUint(value, 10, 64)
	if err != nil {
		BadRequest(c, fmt.Sprintf("参数%s格式错误", name))
		return 0, false
	}

	return uint(parsed), true
}

// OptionalUintQuery 读取可选的 uint 查询参数。
func OptionalUintQuery(c *gin.Context, name string) (uint, bool, error) {
	value := strings.TrimSpace(c.Query(name))
	if value == "" {
		return 0, false, nil
	}

	parsed, err := strconv.ParseUint(value, 10, 64)
	if err != nil {
		return 0, false, fmt.Errorf("参数%s格式错误", name)
	}

	return uint(parsed), true, nil
}

// ParsePageParams 解析分页参数并限制最大 pageSize（maxPageSize<=0 表示不限制）。
func ParsePageParams(c *gin.Context, defaultPage, defaultPageSize, maxPageSize int) (int, int, error) {
	pageStr := strings.TrimSpace(c.DefaultQuery("page", strconv.Itoa(defaultPage)))
	pageSizeStr := strings.TrimSpace(c.DefaultQuery("pageSize", strconv.Itoa(defaultPageSize)))

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		return 0, 0, fmt.Errorf("page必须为正整数")
	}

	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil || pageSize < 1 {
		return 0, 0, fmt.Errorf("pageSize必须为正整数")
	}

	if maxPageSize > 0 && pageSize > maxPageSize {
		pageSize = maxPageSize
	}

	return page, pageSize, nil
}
