package bootstrap

import (
	"devops-platform/config"
	cmdbModel "devops-platform/internal/modules/cmdb/model"
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
		&userModel.UserDepartment{},  // 用户-部门多对多
		&userModel.FieldPermission{}, // 字段级权限
		&userModel.AuditLog{},        // 审计日志
		&userModel.LoginLog{},        // 登录日志
		&k8sModel.Cluster{},
		&cmdbModel.Host{},
		&cmdbModel.HostGroup{},
		&cmdbModel.Credential{},
		&cmdbModel.TerminalSession{},
		&cmdbModel.SessionTag{},
		&cmdbModel.HostPermission{},
		&cmdbModel.CloudAccount{},
		&cmdbModel.CloudResource{},
		&cmdbModel.FileOperationLog{},
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

	// 迁移 users.department_id -> user_departments（在 AutoMigrate 之后、种子数据之前）
	if err := migrateUserDepartments(db); err != nil {
		logger.Log.Warn("用户部门迁移失败（非致命）", zap.Error(err))
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
		createdCount := 0

		// --- API 权限种子 ---
		apiPermissions := []userModel.Permission{
			{Name: "查看用户", Type: userModel.PermissionTypeAPI, Resource: "user", Action: "list"},
			{Name: "创建用户", Type: userModel.PermissionTypeAPI, Resource: "user", Action: "create"},
			{Name: "更新用户", Type: userModel.PermissionTypeAPI, Resource: "user", Action: "update"},
			{Name: "删除用户", Type: userModel.PermissionTypeAPI, Resource: "user", Action: "delete"},
			{Name: "重置密码", Type: userModel.PermissionTypeAPI, Resource: "user", Action: "reset_password"},
			{Name: "分配角色", Type: userModel.PermissionTypeAPI, Resource: "user", Action: "assign_roles"},
			{Name: "锁定用户", Type: userModel.PermissionTypeAPI, Resource: "user", Action: "lock"},
			{Name: "解锁用户", Type: userModel.PermissionTypeAPI, Resource: "user", Action: "unlock"},
			{Name: "查看部门", Type: userModel.PermissionTypeAPI, Resource: "department", Action: "list"},
			{Name: "创建部门", Type: userModel.PermissionTypeAPI, Resource: "department", Action: "create"},
			{Name: "更新部门", Type: userModel.PermissionTypeAPI, Resource: "department", Action: "update"},
			{Name: "删除部门", Type: userModel.PermissionTypeAPI, Resource: "department", Action: "delete"},
			{Name: "查看角色", Type: userModel.PermissionTypeAPI, Resource: "role", Action: "list"},
			{Name: "创建角色", Type: userModel.PermissionTypeAPI, Resource: "role", Action: "create"},
			{Name: "更新角色", Type: userModel.PermissionTypeAPI, Resource: "role", Action: "update"},
			{Name: "删除角色", Type: userModel.PermissionTypeAPI, Resource: "role", Action: "delete"},
			{Name: "查看集群", Type: userModel.PermissionTypeAPI, Resource: "cluster", Action: "list"},
			{Name: "创建集群", Type: userModel.PermissionTypeAPI, Resource: "cluster", Action: "create"},
			{Name: "更新集群", Type: userModel.PermissionTypeAPI, Resource: "cluster", Action: "update"},
			{Name: "删除集群", Type: userModel.PermissionTypeAPI, Resource: "cluster", Action: "delete"},
			{Name: "查看权限", Type: userModel.PermissionTypeAPI, Resource: "permission", Action: "list"},
			{Name: "创建权限", Type: userModel.PermissionTypeAPI, Resource: "permission", Action: "create"},
			{Name: "更新权限", Type: userModel.PermissionTypeAPI, Resource: "permission", Action: "update"},
			{Name: "删除权限", Type: userModel.PermissionTypeAPI, Resource: "permission", Action: "delete"},
			{Name: "查看审计日志", Type: userModel.PermissionTypeAPI, Resource: "audit", Action: "list"},
		// 应用管理权限
		{Name: "查看应用", Type: userModel.PermissionTypeAPI, Resource: "app", Action: "list"},
		{Name: "创建应用", Type: userModel.PermissionTypeAPI, Resource: "app", Action: "create"},
		{Name: "更新应用", Type: userModel.PermissionTypeAPI, Resource: "app", Action: "update"},
		{Name: "删除应用", Type: userModel.PermissionTypeAPI, Resource: "app", Action: "delete"},
		// 告警管理权限
		{Name: "查看告警", Type: userModel.PermissionTypeAPI, Resource: "alert", Action: "list"},
		{Name: "创建告警规则", Type: userModel.PermissionTypeAPI, Resource: "alert", Action: "create"},
		{Name: "更新告警规则", Type: userModel.PermissionTypeAPI, Resource: "alert", Action: "update"},
		{Name: "删除告警规则", Type: userModel.PermissionTypeAPI, Resource: "alert", Action: "delete"},
		// 日志管理权限
		{Name: "查看日志", Type: userModel.PermissionTypeAPI, Resource: "log", Action: "list"},
		// 监控管理权限
		{Name: "查看监控", Type: userModel.PermissionTypeAPI, Resource: "monitor", Action: "list"},
		// Harbor 管理权限
		{Name: "查看Harbor", Type: userModel.PermissionTypeAPI, Resource: "harbor", Action: "list"},
		// CI/CD 管理权限
		{Name: "查看CI/CD", Type: userModel.PermissionTypeAPI, Resource: "cicd", Action: "list"},
		{Name: "创建CI/CD", Type: userModel.PermissionTypeAPI, Resource: "cicd", Action: "create"},
		{Name: "更新CI/CD", Type: userModel.PermissionTypeAPI, Resource: "cicd", Action: "update"},
		{Name: "删除CI/CD", Type: userModel.PermissionTypeAPI, Resource: "cicd", Action: "delete"},
		// 租户管理权限
			{Name: "查看租户", Type: userModel.PermissionTypeAPI, Resource: "tenant", Action: "list"},
			{Name: "创建租户", Type: userModel.PermissionTypeAPI, Resource: "tenant", Action: "create"},
			{Name: "更新租户", Type: userModel.PermissionTypeAPI, Resource: "tenant", Action: "update"},
			{Name: "删除租户", Type: userModel.PermissionTypeAPI, Resource: "tenant", Action: "delete"},
			// CMDB 资产管理权限
			{Name: "查看主机列表", Type: userModel.PermissionTypeAPI, Resource: "cmdb:host", Action: "list"},
			{Name: "查看主机详情", Type: userModel.PermissionTypeAPI, Resource: "cmdb:host", Action: "get"},
			{Name: "创建主机", Type: userModel.PermissionTypeAPI, Resource: "cmdb:host", Action: "create"},
			{Name: "更新主机", Type: userModel.PermissionTypeAPI, Resource: "cmdb:host", Action: "update"},
			{Name: "删除主机", Type: userModel.PermissionTypeAPI, Resource: "cmdb:host", Action: "delete"},
			{Name: "测试主机连接", Type: userModel.PermissionTypeAPI, Resource: "cmdb:host", Action: "test"},
			{Name: "主机管理（管理员）", Type: userModel.PermissionTypeAPI, Resource: "cmdb:host", Action: "admin"},
			{Name: "查看分组", Type: userModel.PermissionTypeAPI, Resource: "cmdb:group", Action: "list"},
			{Name: "创建分组", Type: userModel.PermissionTypeAPI, Resource: "cmdb:group", Action: "create"},
			{Name: "更新分组", Type: userModel.PermissionTypeAPI, Resource: "cmdb:group", Action: "update"},
			{Name: "删除分组", Type: userModel.PermissionTypeAPI, Resource: "cmdb:group", Action: "delete"},
			{Name: "查看凭据", Type: userModel.PermissionTypeAPI, Resource: "cmdb:credential", Action: "list"},
			{Name: "创建凭据", Type: userModel.PermissionTypeAPI, Resource: "cmdb:credential", Action: "create"},
			{Name: "更新凭据", Type: userModel.PermissionTypeAPI, Resource: "cmdb:credential", Action: "update"},
			{Name: "删除凭据", Type: userModel.PermissionTypeAPI, Resource: "cmdb:credential", Action: "delete"},
			{Name: "连接终端", Type: userModel.PermissionTypeAPI, Resource: "cmdb:terminal", Action: "connect"},
			{Name: "查看终端会话", Type: userModel.PermissionTypeAPI, Resource: "cmdb:terminal", Action: "list"},
			{Name: "查看终端详情", Type: userModel.PermissionTypeAPI, Resource: "cmdb:terminal", Action: "get"},
			{Name: "回放终端录像", Type: userModel.PermissionTypeAPI, Resource: "cmdb:terminal", Action: "replay"},
			// CMDB 权限配置
			{Name: "查看权限配置", Type: userModel.PermissionTypeAPI, Resource: "cmdb:permission", Action: "list"},
			{Name: "授予权限", Type: userModel.PermissionTypeAPI, Resource: "cmdb:permission", Action: "create"},
			{Name: "更新权限", Type: userModel.PermissionTypeAPI, Resource: "cmdb:permission", Action: "update"},
			{Name: "删除权限", Type: userModel.PermissionTypeAPI, Resource: "cmdb:permission", Action: "delete"},
			// 云账号管理
			{Name: "查看云账号", Type: userModel.PermissionTypeAPI, Resource: "cmdb:cloud", Action: "list"},
			{Name: "查看云账号详情", Type: userModel.PermissionTypeAPI, Resource: "cmdb:cloud", Action: "get"},
			{Name: "添加云账号", Type: userModel.PermissionTypeAPI, Resource: "cmdb:cloud", Action: "create"},
			{Name: "更新云账号", Type: userModel.PermissionTypeAPI, Resource: "cmdb:cloud", Action: "update"},
			{Name: "删除云账号", Type: userModel.PermissionTypeAPI, Resource: "cmdb:cloud", Action: "delete"},
			{Name: "同步云资源", Type: userModel.PermissionTypeAPI, Resource: "cmdb:cloud", Action: "sync"},
	}

	for _, perm := range apiPermissions {
		var existing userModel.Permission
		result := db.Where("resource = ? AND action = ?", perm.Resource, perm.Action).First(&existing)
		if result.Error == gorm.ErrRecordNotFound {
			if err := db.Create(&perm).Error; err != nil {
				return fmt.Errorf("创建权限 %s:%s 失败: %w", perm.Resource, perm.Action, err)
			}
			createdCount++
		}
	}

	// --- 菜单权限种子 ---
	// 先插入顶级菜单，获取 ID 后再插入子菜单
	menuSystem := userModel.Permission{
		Name: "系统管理", Type: userModel.PermissionTypeMenu, Resource: "system", Action: "view",
		Path: "/system", Icon: "Setting", Sort: 100,
	}
	var existingSystem userModel.Permission
	if db.Where("resource = ? AND action = ?", menuSystem.Resource, menuSystem.Action).First(&existingSystem).Error == gorm.ErrRecordNotFound {
		if err := db.Create(&menuSystem).Error; err != nil {
			return fmt.Errorf("创建菜单权限 %s:%s 失败: %w", menuSystem.Resource, menuSystem.Action, err)
		}
		createdCount++
	} else {
		menuSystem.ID = existingSystem.ID
	}

	// 系统管理下的子菜单
	subMenus := []userModel.Permission{
		{Name: "用户管理", Type: userModel.PermissionTypeMenu, Resource: "user", Action: "view", Path: "/system/user", Icon: "User", Sort: 101, ParentID: &menuSystem.ID},
		{Name: "角色管理", Type: userModel.PermissionTypeMenu, Resource: "role", Action: "view", Path: "/system/role", Icon: "Lock", Sort: 102, ParentID: &menuSystem.ID},
		{Name: "权限管理", Type: userModel.PermissionTypeMenu, Resource: "permission", Action: "view", Path: "/system/permission", Icon: "Key", Sort: 103, ParentID: &menuSystem.ID},
		{Name: "部门管理", Type: userModel.PermissionTypeMenu, Resource: "department", Action: "view", Path: "/system/department", Icon: "OfficeBuilding", Sort: 104, ParentID: &menuSystem.ID},
	}
	// 注意：子菜单的 resource 与 API 权重复（如 user:view），需用 type=menu 区分
	for _, menu := range subMenus {
		var existing userModel.Permission
		result := db.Where("type = ? AND resource = ? AND action = ?", userModel.PermissionTypeMenu, menu.Resource, menu.Action).First(&existing)
		if result.Error == gorm.ErrRecordNotFound {
			if err := db.Create(&menu).Error; err != nil {
				return fmt.Errorf("创建菜单权限 %s:%s 失败: %w", menu.Resource, menu.Action, err)
			}
			createdCount++
		}
	}

	// 租户管理菜单（顶级）
	menuTenant := userModel.Permission{
		Name: "租户管理", Type: userModel.PermissionTypeMenu, Resource: "tenant", Action: "view",
		Path: "/platform/tenant", Icon: "House", Sort: 200,
	}
	var existingTenant userModel.Permission
	if db.Where("type = ? AND resource = ? AND action = ?", userModel.PermissionTypeMenu, menuTenant.Resource, menuTenant.Action).First(&existingTenant).Error == gorm.ErrRecordNotFound {
		if err := db.Create(&menuTenant).Error; err != nil {
			return fmt.Errorf("创建菜单权限 %s:%s 失败: %w", menuTenant.Resource, menuTenant.Action, err)
		}
		createdCount++
	}

	// --- 按钮权限种子 ---
	buttonPermissions := []userModel.Permission{
		{Name: "创建用户", Type: userModel.PermissionTypeButton, Resource: "user", Action: "create_btn"},
		{Name: "编辑用户", Type: userModel.PermissionTypeButton, Resource: "user", Action: "update_btn"},
		{Name: "删除用户", Type: userModel.PermissionTypeButton, Resource: "user", Action: "delete_btn"},
		{Name: "创建角色", Type: userModel.PermissionTypeButton, Resource: "role", Action: "create_btn"},
		{Name: "编辑角色", Type: userModel.PermissionTypeButton, Resource: "role", Action: "update_btn"},
		{Name: "删除角色", Type: userModel.PermissionTypeButton, Resource: "role", Action: "delete_btn"},
	}
	for _, btn := range buttonPermissions {
		var existing userModel.Permission
		result := db.Where("type = ? AND resource = ? AND action = ?", userModel.PermissionTypeButton, btn.Resource, btn.Action).First(&existing)
		if result.Error == gorm.ErrRecordNotFound {
			if err := db.Create(&btn).Error; err != nil {
				return fmt.Errorf("创建按钮权限 %s:%s 失败: %w", btn.Resource, btn.Action, err)
			}
			createdCount++
		}
	}

	if createdCount > 0 {
		logger.Log.Info(fmt.Sprintf("权限种子数据初始化完成，新增 %d 条记录", createdCount))
	}
	return nil
}

// migrateUserDepartments 将旧表 users.department_id 迁移到 user_departments 表
// 幂等设计：可重复执行不出错
func migrateUserDepartments(db *gorm.DB) error {
	// 检查 users 表是否还有 department_id 列
	var colCount int64
	err := db.Raw(
		`SELECT COUNT(1) FROM information_schema.columns WHERE table_schema = DATABASE() AND table_name = 'users' AND column_name = 'department_id'`,
	).Scan(&colCount).Error
	if err != nil {
		return fmt.Errorf("检查 users.department_id 列失败: %w", err)
	}
	if colCount == 0 {
		// 旧列已不存在，无需迁移
		return nil
	}

	// 查询需要迁移的记录：department_id 非空 且 user_departments 中尚无对应记录
	type migrateRow struct {
		UserID       uint `gorm:"column:id"`
		DepartmentID uint `gorm:"column:department_id"`
		TenantID     *uint `gorm:"column:tenant_id"`
	}
	var rows []migrateRow
	err = db.Table("users").
		Select("id, department_id, tenant_id").
		Where("department_id IS NOT NULL AND department_id > 0").
		Find(&rows).Error
	if err != nil {
		return fmt.Errorf("查询待迁移用户部门数据失败: %w", err)
	}
	if len(rows) == 0 {
		return nil
	}

	migratedCount := 0
	for _, row := range rows {
		// 检查 user_departments 中是否已有该记录（幂等）
		var existCount int64
		db.Table("user_departments").
			Where("user_id = ? AND dept_id = ?", row.UserID, row.DepartmentID).
			Count(&existCount)
		if existCount > 0 {
			continue
		}

		// 插入 user_departments 记录
		ud := userModel.UserDepartment{
			UserID:    row.UserID,
			DeptID:    row.DepartmentID,
			IsPrimary: true,
		}
		if err := db.Create(&ud).Error; err != nil {
			logger.Log.Warn("迁移用户部门关联失败，跳过",
				zap.Uint("userID", row.UserID),
				zap.Uint("deptID", row.DepartmentID),
				zap.Error(err))
			continue
		}
		migratedCount++
	}

	// 将 users.department_id 的值复制到 users.primary_dept_id
	err = db.Table("users").
		Where("department_id IS NOT NULL AND department_id > 0").
		Where("primary_dept_id IS NULL").
		Update("primary_dept_id", gorm.Expr("department_id")).Error
	if err != nil {
		return fmt.Errorf("回填 users.primary_dept_id 失败: %w", err)
	}

	if migratedCount > 0 {
		logger.Log.Info(fmt.Sprintf("用户部门迁移完成，迁移 %d 条记录", migratedCount))
	}
	return nil
}
