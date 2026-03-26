package bootstrap

import (
	"devops-platform/config"
	k8sModel "devops-platform/internal/modules/k8s/model"
	userModel "devops-platform/internal/modules/user/model"
	"devops-platform/internal/pkg/logger"
	"fmt"
	"time"

	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDB() error {
	host := config.Cfg.GetString("db.host")
	port := config.Cfg.GetString("db.port")
	dbname := config.Cfg.GetString("db.db")
	username := config.Cfg.GetString("db.username")
	password := config.Cfg.GetString("db.password")
	charset := config.Cfg.GetString("db.charset")
	maxIdle := config.Cfg.GetInt("db.maxIdle")
	maxOpen := config.Cfg.GetInt("db.maxOpen")

	// 校验必要参数，防止出现 dial tcp :0 错误
	if host == "" {
		return fmt.Errorf("database host is empty, please check your config")
	}
	if port == "" {
		// 尝试使用默认端口
		port = "3306"
		fmt.Println("Warning: database port is empty, using default 3306")
	}
	if dbname == "" {
		return fmt.Errorf("database name is empty, please check your config")
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=%s&parseTime=True&loc=Local&collation=utf8mb4_unicode_ci",
		username, password, host, port, dbname, charset)

	fmt.Printf("Connecting to database: %s@tcp(%s:%s)/%s\n", username, host, port, dbname)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("failed to connect database: %w", err)
	}

	// 获取底层的 sql.DB
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}

	// 设置连接池参数
	sqlDB.SetMaxIdleConns(maxIdle)
	sqlDB.SetMaxOpenConns(maxOpen)
	sqlDB.SetConnMaxLifetime(time.Hour)

	// 自动迁移数据表 (注意顺序：先迁移被依赖的表)
	err = db.AutoMigrate(
		&userModel.Tenant{},     // 租户表 (最优先，其他表可能关联)
		&userModel.Department{}, // 部门表
		&userModel.User{},
		&userModel.Role{},
		&userModel.Permission{},
		&userModel.AuditLog{}, // 审计日志
		&k8sModel.Cluster{},
	)
	if err != nil {
		return err
	}

	if err := ensureKeywordIndexes(db); err != nil {
		return err
	}

	DB = db

	if err := seedPermissions(db); err != nil {
		logger.Log.Warn("权限种子数据初始化失败（非致命）", zap.Error(err))
	}

	return nil
}

func ensureKeywordIndexes(db *gorm.DB) error {
	if db.Dialector == nil || db.Dialector.Name() != "mysql" {
		return nil
	}

	indexes := []struct {
		name  string
		table string
		ddl   string
	}{
		{
			name:  "idx_users_keyword_ft",
			table: "users",
			ddl:   "CREATE FULLTEXT INDEX idx_users_keyword_ft ON users (username, name, email)",
		},
		{
			name:  "idx_roles_keyword_ft",
			table: "roles",
			ddl:   "CREATE FULLTEXT INDEX idx_roles_keyword_ft ON roles (name, display_name, description)",
		},
		{
			name:  "idx_permissions_keyword_ft",
			table: "permissions",
			ddl:   "CREATE FULLTEXT INDEX idx_permissions_keyword_ft ON permissions (name, resource, action, description)",
		},
		{
			name:  "idx_departments_keyword_ft",
			table: "departments",
			ddl:   "CREATE FULLTEXT INDEX idx_departments_keyword_ft ON departments (name)",
		},
		{
			name:  "idx_clusters_keyword_ft",
			table: "clusters",
			ddl:   "CREATE FULLTEXT INDEX idx_clusters_keyword_ft ON clusters (name, url, remark, labels, status, env, k8s_version)",
		},
		{
			name:  "idx_users_dept_created_deleted",
			table: "users",
			ddl:   "CREATE INDEX idx_users_dept_created_deleted ON users (department_id, created_at, deleted_at)",
		},
		{
			name:  "idx_permissions_resource_deleted_action",
			table: "permissions",
			ddl:   "CREATE INDEX idx_permissions_resource_deleted_action ON permissions (resource, deleted_at, action)",
		},
		{
			name:  "idx_clusters_env_created_deleted",
			table: "clusters",
			ddl:   "CREATE INDEX idx_clusters_env_created_deleted ON clusters (env, created_at, deleted_at)",
		},
	}

	for _, index := range indexes {
		exists, err := mysqlIndexExists(db, index.table, index.name)
		if err != nil {
			return err
		}
		if exists {
			continue
		}
		if err := db.Exec(index.ddl).Error; err != nil {
			return fmt.Errorf("create index %s failed: %w", index.name, err)
		}
	}

	return nil
}

func mysqlIndexExists(db *gorm.DB, tableName, indexName string) (bool, error) {
	var count int64
	err := db.Raw(
		`SELECT COUNT(1) 
FROM information_schema.statistics 
WHERE table_schema = DATABASE() AND table_name = ? AND index_name = ?`,
		tableName,
		indexName,
	).Scan(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// seedPermissions 幂等初始化权限种子数据，已存在的记录会跳过
func seedPermissions(db *gorm.DB) error {
	permissions := []userModel.Permission{
		{Name: "查看用户", Resource: "user", Action: "list", Description: "查看用户列表"},
		{Name: "创建用户", Resource: "user", Action: "create", Description: "创建新用户"},
		{Name: "更新用户", Resource: "user", Action: "update", Description: "更新用户信息"},
		{Name: "删除用户", Resource: "user", Action: "delete", Description: "删除用户"},
		{Name: "重置密码", Resource: "user", Action: "reset_password", Description: "重置用户密码"},
		{Name: "分配角色", Resource: "user", Action: "assign_roles", Description: "给用户分配角色"},
		{Name: "锁定用户", Resource: "user", Action: "lock", Description: "锁定用户账号"},
		{Name: "解锁用户", Resource: "user", Action: "unlock", Description: "解锁用户账号"},
		{Name: "查看部门", Resource: "department", Action: "list", Description: "查看部门列表"},
		{Name: "创建部门", Resource: "department", Action: "create", Description: "创建新部门"},
		{Name: "更新部门", Resource: "department", Action: "update", Description: "更新部门信息"},
		{Name: "删除部门", Resource: "department", Action: "delete", Description: "删除部门"},
		{Name: "查看角色", Resource: "role", Action: "list", Description: "查看角色列表"},
		{Name: "创建角色", Resource: "role", Action: "create", Description: "创建新角色"},
		{Name: "更新角色", Resource: "role", Action: "update", Description: "更新角色信息"},
		{Name: "删除角色", Resource: "role", Action: "delete", Description: "删除角色"},
		{Name: "查看集群", Resource: "cluster", Action: "list", Description: "查看集群列表"},
		{Name: "创建集群", Resource: "cluster", Action: "create", Description: "创建新集群"},
		{Name: "更新集群", Resource: "cluster", Action: "update", Description: "更新集群信息"},
		{Name: "删除集群", Resource: "cluster", Action: "delete", Description: "删除集群"},
		{Name: "查看权限", Resource: "permission", Action: "list", Description: "查看权限列表"},
		{Name: "创建权限", Resource: "permission", Action: "create", Description: "创建权限"},
		{Name: "更新权限", Resource: "permission", Action: "update", Description: "更新权限"},
		{Name: "删除权限", Resource: "permission", Action: "delete", Description: "删除权限"},
		{Name: "查看审计日志", Resource: "audit", Action: "list", Description: "查看操作审计日志"},
		// 应用管理权限
		{Name: "查看应用", Resource: "app", Action: "list", Description: "查看应用列表"},
		{Name: "创建应用", Resource: "app", Action: "create", Description: "创建新应用"},
		{Name: "更新应用", Resource: "app", Action: "update", Description: "更新应用信息"},
		{Name: "删除应用", Resource: "app", Action: "delete", Description: "删除应用"},
		// 告警管理权限
		{Name: "查看告警", Resource: "alert", Action: "list", Description: "查看告警列表"},
		{Name: "创建告警规则", Resource: "alert", Action: "create", Description: "创建告警规则"},
		{Name: "更新告警规则", Resource: "alert", Action: "update", Description: "更新告警规则"},
		{Name: "删除告警规则", Resource: "alert", Action: "delete", Description: "删除告警规则"},
		// 日志管理权限
		{Name: "查看日志", Resource: "log", Action: "list", Description: "查看日志列表"},
		// 监控管理权限
		{Name: "查看监控", Resource: "monitor", Action: "list", Description: "查看监控数据"},
		// Harbor 管理权限
		{Name: "查看Harbor", Resource: "harbor", Action: "list", Description: "查看Harbor项目"},
		// CI/CD 管理权限
		{Name: "查看CI/CD", Resource: "cicd", Action: "list", Description: "查看CI/CD流水线"},
		{Name: "创建CI/CD", Resource: "cicd", Action: "create", Description: "创建CI/CD流水线"},
		{Name: "更新CI/CD", Resource: "cicd", Action: "update", Description: "更新CI/CD流水线"},
		{Name: "删除CI/CD", Resource: "cicd", Action: "delete", Description: "删除CI/CD流水线"},
		// 租户管理权限
		{Name: "查看租户", Resource: "tenant", Action: "list", Description: "查看租户列表"},
		{Name: "创建租户", Resource: "tenant", Action: "create", Description: "创建新租户"},
		{Name: "更新租户", Resource: "tenant", Action: "update", Description: "更新租户信息"},
		{Name: "删除租户", Resource: "tenant", Action: "delete", Description: "删除租户"},
	}

	createdCount := 0
	for _, perm := range permissions {
		var existing userModel.Permission
		result := db.Where("resource = ? AND action = ?", perm.Resource, perm.Action).First(&existing)
		if result.Error == gorm.ErrRecordNotFound {
			if err := db.Create(&perm).Error; err != nil {
				return fmt.Errorf("创建权限 %s:%s 失败: %w", perm.Resource, perm.Action, err)
			}
			createdCount++
		}
	}

	if createdCount > 0 {
		logger.Log.Info(fmt.Sprintf("权限种子数据初始化完成，新增 %d 条记录", createdCount))
	}
	return nil
}
