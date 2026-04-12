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
		&userModel.Tenant{},         // 租户表 (最优先，其他表可能关联)
		&userModel.Department{},     // 部门表
		&userModel.User{},
		&userModel.Role{},
		&userModel.Permission{},
		&userModel.UserDepartment{}, // 用户-部门多对多
		&userModel.FieldPermission{}, // 字段级权限
		&userModel.AuditLog{},       // 审计日志
		&k8sModel.Cluster{},
	)
	if err != nil {
		return err
	}

	if err := ensureKeywordIndexes(db); err != nil {
		return err
	}

	if err := ensureDefaultTenantAndBackfill(db); err != nil {
		return err
	}
	if err := ensureRoleDataScopes(db); err != nil {
		return err
	}

	DB = db

	if err := seedPermissions(db); err != nil {
		logger.Log.Warn("权限种子数据初始化失败（非致命）", zap.Error(err))
	}

	return nil
}

func ensureDefaultTenantAndBackfill(db *gorm.DB) error {
	const defaultTenantCode = "default"
	const defaultTenantName = "默认租户"

	var tenant userModel.Tenant
	err := db.Where("code = ?", defaultTenantCode).First(&tenant).Error
	if err != nil {
		if err != gorm.ErrRecordNotFound {
			return err
		}
		tenant = userModel.Tenant{
			Name:           defaultTenantName,
			Code:           defaultTenantCode,
			Description:    "系统默认租户",
			Status:         "active",
			MaxUsers:       1000,
			MaxDepartments: 100,
			MaxRoles:       100,
		}
		if err := db.Create(&tenant).Error; err != nil {
			return fmt.Errorf("create default tenant failed: %w", err)
		}
	}

	// 保留 roles.tenant_id 为空的全局系统角色语义，避免被错误回填成默认租户角色。
	for _, table := range []string{"users", "departments", "clusters"} {
		if err := db.Table(table).
			Where("tenant_id IS NULL").
			Update("tenant_id", tenant.ID).Error; err != nil {
			return fmt.Errorf("backfill %s.tenant_id failed: %w", table, err)
		}
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

func ensureRoleDataScopes(db *gorm.DB) error {
	if err := db.Model(&userModel.Role{}).
		Where("COALESCE(data_scope, '') = ''").
		Update("data_scope", string(userModel.DataScopeSelfDepartment)).Error; err != nil {
		return fmt.Errorf("backfill role data_scope failed: %w", err)
	}

	for _, roleName := range []string{"SYSTEM_ADMIN", "TENANT_ADMIN", "DEPT_ADMIN", "READ_ONLY"} {
		dataScope, ok := userModel.DefaultRoleDataScope(roleName)
		if !ok {
			continue
		}
		if err := db.Model(&userModel.Role{}).
			Where("name = ? AND data_scope <> ?", roleName, dataScope).
			Update("data_scope", string(dataScope)).Error; err != nil {
			return fmt.Errorf("backfill role %s data_scope failed: %w", roleName, err)
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
		{Name: "查看用户", Resource: "user", Action: "list"},
		{Name: "创建用户", Resource: "user", Action: "create"},
		{Name: "更新用户", Resource: "user", Action: "update"},
		{Name: "删除用户", Resource: "user", Action: "delete"},
		{Name: "重置密码", Resource: "user", Action: "reset_password"},
		{Name: "分配角色", Resource: "user", Action: "assign_roles"},
		{Name: "锁定用户", Resource: "user", Action: "lock"},
		{Name: "解锁用户", Resource: "user", Action: "unlock"},
		{Name: "查看部门", Resource: "department", Action: "list"},
		{Name: "创建部门", Resource: "department", Action: "create"},
		{Name: "更新部门", Resource: "department", Action: "update"},
		{Name: "删除部门", Resource: "department", Action: "delete"},
		{Name: "查看角色", Resource: "role", Action: "list"},
		{Name: "创建角色", Resource: "role", Action: "create"},
		{Name: "更新角色", Resource: "role", Action: "update"},
		{Name: "删除角色", Resource: "role", Action: "delete"},
		{Name: "查看集群", Resource: "cluster", Action: "list"},
		{Name: "创建集群", Resource: "cluster", Action: "create"},
		{Name: "更新集群", Resource: "cluster", Action: "update"},
		{Name: "删除集群", Resource: "cluster", Action: "delete"},
		{Name: "查看权限", Resource: "permission", Action: "list"},
		{Name: "创建权限", Resource: "permission", Action: "create"},
		{Name: "更新权限", Resource: "permission", Action: "update"},
		{Name: "删除权限", Resource: "permission", Action: "delete"},
		{Name: "查看审计日志", Resource: "audit", Action: "list"},
		// 应用管理权限
		{Name: "查看应用", Resource: "app", Action: "list"},
		{Name: "创建应用", Resource: "app", Action: "create"},
		{Name: "更新应用", Resource: "app", Action: "update"},
		{Name: "删除应用", Resource: "app", Action: "delete"},
		// 告警管理权限
		{Name: "查看告警", Resource: "alert", Action: "list"},
		{Name: "创建告警规则", Resource: "alert", Action: "create"},
		{Name: "更新告警规则", Resource: "alert", Action: "update"},
		{Name: "删除告警规则", Resource: "alert", Action: "delete"},
		// 日志管理权限
		{Name: "查看日志", Resource: "log", Action: "list"},
		// 监控管理权限
		{Name: "查看监控", Resource: "monitor", Action: "list"},
		// Harbor 管理权限
		{Name: "查看Harbor", Resource: "harbor", Action: "list"},
		// CI/CD 管理权限
		{Name: "查看CI/CD", Resource: "cicd", Action: "list"},
		{Name: "创建CI/CD", Resource: "cicd", Action: "create"},
		{Name: "更新CI/CD", Resource: "cicd", Action: "update"},
		{Name: "删除CI/CD", Resource: "cicd", Action: "delete"},
		// 租户管理权限
		{Name: "查看租户", Resource: "tenant", Action: "list"},
		{Name: "创建租户", Resource: "tenant", Action: "create"},
		{Name: "更新租户", Resource: "tenant", Action: "update"},
		{Name: "删除租户", Resource: "tenant", Action: "delete"},
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
