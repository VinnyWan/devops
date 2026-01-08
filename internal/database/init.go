package database

import (
	"devops/internal/logger"
	"devops/models"
	"devops/utils"

	"go.uber.org/zap"
)

// InitData 初始化基础数据
func InitData() error {
	logger.Log.Info("开始初始化基础数据...")

	// 检查是否已有管理员用户
	var count int64
	Db.Model(&models.User{}).Count(&count)
	if count > 0 {
		logger.Log.Info("数据已存在，跳过初始化")
		return nil
	}

	// 创建默认管理员角色
	adminRole := models.Role{
		RoleName: "超级管理员",
		RoleKey:  "admin",
		Sort:     1,
		Status:   1,
		Remark:   "系统内置超级管理员角色",
	}
	if err := Db.Create(&adminRole).Error; err != nil {
		logger.Log.Error("创建管理员角色失败", zap.Error(err))
		return err
	}

	// 创建默认管理员用户
	hashedPassword, _ := utils.HashPassword("admin123")
	adminUser := models.User{
		Username: "admin",
		Password: hashedPassword,
		Nickname: "超级管理员",
		Email:    "admin@example.com",
		Phone:    "13800138000",
		Status:   1,
		Gender:   1,
		Remark:   "系统内置超级管理员",
	}
	if err := Db.Create(&adminUser).Error; err != nil {
		logger.Log.Error("创建管理员用户失败", zap.Error(err))
		return err
	}

	// 关联角色
	if err := Db.Model(&adminUser).Association("Roles").Append(&adminRole); err != nil {
		logger.Log.Error("关联角色失败", zap.Error(err))
		return err
	}

	// 创建默认部门
	dept := models.Department{
		DeptName: "总公司",
		ParentID: 0,
		Sort:     0,
		Leader:   "超级管理员",
		Phone:    "13800138000",
		Email:    "admin@example.com",
		Status:   1,
		Remark:   "顶级部门",
	}
	if err := Db.Create(&dept).Error; err != nil {
		logger.Log.Error("创建默认部门失败", zap.Error(err))
		return err
	}

	// 创建默认岗位
	post := models.Post{
		PostName: "董事长",
		PostCode: "ceo",
		Sort:     1,
		Status:   1,
		Remark:   "公司最高管理者",
	}
	if err := Db.Create(&post).Error; err != nil {
		logger.Log.Error("创建默认岗位失败", zap.Error(err))
		return err
	}

	// 创建系统菜单
	menus := []models.Menu{
		{
			MenuName:  "系统管理",
			ParentID:  0,
			Sort:      1,
			Path:      "/system",
			Component: "Layout",
			MenuType:  "M",
			Visible:   1,
			Status:    1,
			Icon:      "system",
			Remark:    "系统管理目录",
		},
		{
			MenuName:  "用户管理",
			ParentID:  1,
			Sort:      1,
			Path:      "user",
			Component: "system/user/index",
			MenuType:  "C",
			Visible:   1,
			Status:    1,
			Perms:     "system:user:list",
			Icon:      "user",
			Remark:    "用户管理菜单",
		},
		{
			MenuName:  "角色管理",
			ParentID:  1,
			Sort:      2,
			Path:      "role",
			Component: "system/role/index",
			MenuType:  "C",
			Visible:   1,
			Status:    1,
			Perms:     "system:role:list",
			Icon:      "peoples",
			Remark:    "角色管理菜单",
		},
		{
			MenuName:  "菜单管理",
			ParentID:  1,
			Sort:      3,
			Path:      "menu",
			Component: "system/menu/index",
			MenuType:  "C",
			Visible:   1,
			Status:    1,
			Perms:     "system:menu:list",
			Icon:      "tree-table",
			Remark:    "菜单管理菜单",
		},
		{
			MenuName:  "部门管理",
			ParentID:  1,
			Sort:      4,
			Path:      "dept",
			Component: "system/dept/index",
			MenuType:  "C",
			Visible:   1,
			Status:    1,
			Perms:     "system:dept:list",
			Icon:      "tree",
			Remark:    "部门管理菜单",
		},
		{
			MenuName:  "岗位管理",
			ParentID:  1,
			Sort:      5,
			Path:      "post",
			Component: "system/post/index",
			MenuType:  "C",
			Visible:   1,
			Status:    1,
			Perms:     "system:post:list",
			Icon:      "post",
			Remark:    "岗位管理菜单",
		},
		{
			MenuName:  "操作日志",
			ParentID:  1,
			Sort:      6,
			Path:      "operlog",
			Component: "system/operlog/index",
			MenuType:  "C",
			Visible:   1,
			Status:    1,
			Perms:     "system:operlog:list",
			Icon:      "form",
			Remark:    "操作日志菜单",
		},
		{
			MenuName:  "登录日志",
			ParentID:  1,
			Sort:      7,
			Path:      "loginlog",
			Component: "system/loginlog/index",
			MenuType:  "C",
			Visible:   1,
			Status:    1,
			Perms:     "system:loginlog:list",
			Icon:      "logininfor",
			Remark:    "登录日志菜单",
		},
	}

	for _, menu := range menus {
		if err := Db.Create(&menu).Error; err != nil {
			logger.Log.Error("创建菜单失败", zap.Error(err))
			return err
		}
	}

	logger.Log.Info("基础数据初始化完成")
	logger.Log.Info("默认管理员账号: admin, 密码: admin123")
	return nil
}
