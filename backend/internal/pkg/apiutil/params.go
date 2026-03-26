package apiutil

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

// ParsePagination 解析分页参数
func ParsePagination(c *gin.Context) (page, pageSize int) {
	page, _ = strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ = strconv.Atoi(c.DefaultQuery("page_size", "10"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}
	return
}

// ParseIDFromQuery 从查询参数解析ID
func ParseIDFromQuery(c *gin.Context, paramName string) (uint, error) {
	idStr := c.Query(paramName)
	id, err := strconv.ParseUint(idStr, 10, 32)
	return uint(id), err
}

// ParseIDFromParam 从路径参数解析ID
func ParseIDFromParam(c *gin.Context, paramName string) (uint, error) {
	idStr := c.Param(paramName)
	id, err := strconv.ParseUint(idStr, 10, 32)
	return uint(id), err
}
