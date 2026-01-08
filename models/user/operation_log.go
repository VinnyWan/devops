package user

import (
	"time"

	"gorm.io/gorm"
)

// OperationLog 操作日志表
type OperationLog struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"createdAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	Module       string `gorm:"type:varchar(50);comment:模块名称" json:"module"`
	Type         string `gorm:"type:varchar(20);comment:操作类型" json:"type"`
	Title        string `gorm:"type:varchar(100);comment:操作标题" json:"title"`
	Method       string `gorm:"type:varchar(10);comment:请求方法" json:"method"`
	RequestURL   string `gorm:"type:varchar(255);comment:请求URL" json:"requestUrl"`
	RequestParam string `gorm:"type:text;comment:请求参数" json:"requestParam"`
	ResponseData string `gorm:"type:text;comment:响应数据" json:"responseData"`
	IP           string `gorm:"type:varchar(50);comment:操作IP" json:"ip"`
	Location     string `gorm:"type:varchar(100);comment:操作地点" json:"location"`
	Status       int    `gorm:"type:tinyint;comment:操作状态:1成功,2失败" json:"status"`
	ErrorMsg     string `gorm:"type:text;comment:错误信息" json:"errorMsg"`
	CostTime     int64  `gorm:"comment:耗时(毫秒)" json:"costTime"`
	OperatorID   uint   `gorm:"comment:操作人ID" json:"operatorId"`
	OperatorName string `gorm:"type:varchar(50);comment:操作人名称" json:"operatorName"`
}

// TableName 指定表名
func (OperationLog) TableName() string {
	return "sys_operation_log"
}
