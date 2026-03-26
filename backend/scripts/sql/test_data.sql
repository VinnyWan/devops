-- 初始化数据脚本 (增强版 - 覆盖更多场景)

-- 1. 清理旧数据 (按照外键约束反向清理)
SET FOREIGN_KEY_CHECKS = 0;
TRUNCATE TABLE `user_roles`;
TRUNCATE TABLE `department_roles`;
TRUNCATE TABLE `role_permissions`;
TRUNCATE TABLE `users`;
TRUNCATE TABLE `roles`;
TRUNCATE TABLE `permissions`;
TRUNCATE TABLE `departments`;
TRUNCATE TABLE `tenants`;
SET FOREIGN_KEY_CHECKS = 1;

-- 1.1 创建默认租户
INSERT INTO `tenants` (`id`, `name`, `code`, `description`, `status`, `max_users`, `max_departments`, `max_roles`, `contact_name`, `contact_email`, `created_at`, `updated_at`) VALUES
(1, '默认租户', 'default', '系统默认租户', 'active', 1000, 100, 100, '系统管理员', 'admin@example.com', NOW(), NOW()),
(2, '研发中心', 'dev-center', '研发中心租户', 'active', 500, 50, 50, '研发主管', 'dev@example.com', NOW(), NOW());

-- 2. 创建部门 (多层级，支持租户)
-- 默认租户(1): 总公司 (1) -> 研发部 (2), 运维部 (3), 产品部 (4), 人力资源部 (7)
-- 研发部 (2) -> 后端组 (5), 前端组 (6), 移动端组 (8)
-- 运维部 (3) -> 基础运维 (9), 应用运维 (10)
INSERT INTO `departments` (`id`, `tenant_id`, `name`, `parent_id`, `created_at`, `updated_at`) VALUES
(1, 1, '总公司', NULL, NOW(), NOW()),
(2, 1, '研发部', 1, NOW(), NOW()),
(3, 1, '运维部', 1, NOW(), NOW()),
(4, 1, '产品部', 1, NOW(), NOW()),
(5, 1, '后端组', 2, NOW(), NOW()),
(6, 1, '前端组', 2, NOW(), NOW()),
(7, 1, '人力资源部', 1, NOW(), NOW()),
(8, 1, '移动端组', 2, NOW(), NOW()),
(9, 1, '基础运维', 3, NOW(), NOW()),
(10, 1, '应用运维', 3, NOW(), NOW()),
-- 研发中心租户(2)
(11, 2, '研发中心', NULL, NOW(), NOW()),
(12, 2, '前端组', 11, NOW(), NOW()),
(13, 2, '后端组', 11, NOW(), NOW());

-- 3. 创建权限 (覆盖全模块 CRUD)
INSERT INTO `permissions` (`id`, `name`, `resource`, `action`, `description`, `created_at`, `updated_at`) VALUES
-- 用户管理
(1, '查看用户', 'user', 'list', '查看用户列表', NOW(), NOW()),
(2, '创建用户', 'user', 'create', '创建新用户', NOW(), NOW()),
(3, '更新用户', 'user', 'update', '更新用户信息', NOW(), NOW()),
(4, '删除用户', 'user', 'delete', '删除用户', NOW(), NOW()),
-- 部门管理
(5, '查看部门', 'department', 'list', '查看部门列表', NOW(), NOW()),
(6, '创建部门', 'department', 'create', '创建新部门', NOW(), NOW()),
(7, '更新部门', 'department', 'update', '更新部门信息', NOW(), NOW()),
(8, '删除部门', 'department', 'delete', '删除部门', NOW(), NOW()),
-- 角色管理
(9, '查看角色', 'role', 'list', '查看角色列表', NOW(), NOW()),
(10, '创建角色', 'role', 'create', '创建新角色', NOW(), NOW()),
(11, '更新角色', 'role', 'update', '更新角色信息', NOW(), NOW()),
(12, '删除角色', 'role', 'delete', '删除角色', NOW(), NOW()),
-- 集群管理
(13, '查看集群', 'cluster', 'list', '查看集群列表', NOW(), NOW()),
(14, '创建集群', 'cluster', 'create', '创建新集群', NOW(), NOW()),
(15, '更新集群', 'cluster', 'update', '更新集群信息', NOW(), NOW()),
(16, '删除集群', 'cluster', 'delete', '删除集群', NOW(), NOW()),
-- 用户高级操作
(19, '重置密码', 'user', 'reset_password', '重置用户密码', NOW(), NOW()),
(20, '分配角色', 'user', 'assign_roles', '给用户分配角色', NOW(), NOW()),
(21, '锁定用户', 'user', 'lock', '锁定用户账号', NOW(), NOW()),
(22, '解锁用户', 'user', 'unlock', '解锁用户账号', NOW(), NOW()),
-- 权限管理
(17, '查看权限', 'permission', 'list', '查看权限列表', NOW(), NOW()),
-- 审计管理
(18, '查看审计日志', 'audit', 'list', '查看操作审计日志', NOW(), NOW()),
-- 应用管理
(23, '查看应用', 'app', 'list', '查看应用列表', NOW(), NOW()),
(24, '创建应用', 'app', 'create', '创建新应用', NOW(), NOW()),
(25, '更新应用', 'app', 'update', '更新应用信息', NOW(), NOW()),
(26, '删除应用', 'app', 'delete', '删除应用', NOW(), NOW()),
-- 告警管理
(27, '查看告警', 'alert', 'list', '查看告警列表', NOW(), NOW()),
(28, '创建告警', 'alert', 'create', '创建告警规则', NOW(), NOW()),
(29, '更新告警', 'alert', 'update', '更新告警规则', NOW(), NOW()),
(30, '删除告警', 'alert', 'delete', '删除告警规则', NOW(), NOW()),
-- 日志管理
(31, '查看日志', 'log', 'list', '查看日志列表', NOW(), NOW()),
-- 监控管理
(32, '查看监控', 'monitor', 'list', '查看监控数据', NOW(), NOW()),
-- Harbor管理
(33, '查看Harbor', 'harbor', 'list', '查看Harbor项目', NOW(), NOW()),
-- CI/CD管理
(34, '查看CI/CD', 'cicd', 'list', '查看CI/CD流水线', NOW(), NOW()),
(35, '创建CI/CD', 'cicd', 'create', '创建CI/CD流水线', NOW(), NOW()),
(36, '更新CI/CD', 'cicd', 'update', '更新CI/CD流水线', NOW(), NOW()),
(37, '删除CI/CD', 'cicd', 'delete', '删除CI/CD流水线', NOW(), NOW()),
-- 租户管理
(38, '查看租户', 'tenant', 'list', '查看租户列表', NOW(), NOW()),
(39, '创建租户', 'tenant', 'create', '创建新租户', NOW(), NOW()),
(40, '更新租户', 'tenant', 'update', '更新租户信息', NOW(), NOW()),
(41, '删除租户', 'tenant', 'delete', '删除租户', NOW(), NOW());

-- 4. 创建角色 (系统角色与自定义角色，支持租户)
-- 全局角色(tenant_id 为 NULL)可以被所有租户使用
INSERT INTO `roles` (`id`, `tenant_id`, `name`, `display_name`, `description`, `type`, `created_at`, `updated_at`) VALUES
(1, NULL, 'SYSTEM_ADMIN', '系统管理员', '系统管理员-拥有所有权限', 'system', NOW(), NOW()),
(2, NULL, 'READ_ONLY', '只读用户', '全局只读角色', 'system', NOW(), NOW()),
(3, NULL, 'DEPT_ADMIN', '部门管理员', '部门管理员-管理部门内成员', 'custom', NOW(), NOW()),
(4, NULL, 'DEVELOPER', '研发人员', '研发人员-集群开发权限', 'custom', NOW(), NOW()),
(5, NULL, 'OPS_ENGINEER', '运维工程师', '运维工程师-集群全权管理', 'custom', NOW(), NOW()),
(6, NULL, 'AUDITOR', '审计员', '审计员-仅查看审计日志', 'custom', NOW(), NOW()),
(7, NULL, 'TENANT_ADMIN', '租户管理员', '租户管理员-管理租户内所有资源', 'system', NOW(), NOW()),
-- 租户专属角色
(8, 2, 'DEV_CENTER_ADMIN', '研发中心管理员', '研发中心专属管理员', 'custom', NOW(), NOW());

-- 5. 角色关联权限
-- SYSTEM_ADMIN (ID:1): 所有权限 (1-18)
INSERT INTO `role_permissions` (`role_id`, `permission_id`) SELECT 1, id FROM permissions;

-- READ_ONLY (ID:2): 只有各模块的 list 权限
INSERT INTO `role_permissions` (`role_id`, `permission_id`) VALUES
(2, 1), (2, 5), (2, 9), (2, 13), (2, 17);

-- DEPT_ADMIN (ID:3): 用户和部门的 CRUD + 用户高级操作
INSERT INTO `role_permissions` (`role_id`, `permission_id`) VALUES
(3, 1), (3, 2), (3, 3), (3, 4), (3, 5), (3, 6), (3, 7), (3, 8),
(3, 19), (3, 20), (3, 21), (3, 22);

-- DEVELOPER (ID:4): 集群查看与创建更新
INSERT INTO `role_permissions` (`role_id`, `permission_id`) VALUES
(4, 13), (4, 14), (4, 15);

-- OPS_ENGINEER (ID:5): 集群所有权限
INSERT INTO `role_permissions` (`role_id`, `permission_id`) VALUES
(5, 13), (5, 14), (5, 15), (5, 16);

-- AUDITOR (ID:6): 审计查看
INSERT INTO `role_permissions` (`role_id`, `permission_id`) VALUES
(6, 18);

-- 6. 部门关联角色 (默认继承)
-- 研发部(2) 及其子部门继承 DEVELOPER(4)
INSERT INTO `department_roles` (`department_id`, `role_id`) VALUES (2, 4);
-- 运维部(3) 及其子部门继承 OPS_ENGINEER(5)
INSERT INTO `department_roles` (`department_id`, `role_id`) VALUES (3, 5);

-- 7. 创建用户 (覆盖各种状态和类型，支持租户)
-- 密码均为: admin@2026 (bcrypt hash)
INSERT INTO `users` (`id`, `tenant_id`, `username`, `password`, `email`, `department_id`, `is_admin`, `status`, `created_at`, `updated_at`) VALUES
-- 默认租户用户
(1, 1, 'admin', '$2a$10$D3icwPuI7rPaCxOCuuyPFecYbqGy2vJOT40vCy/Qhd6Zz2RU0ufxC', 'admin@example.com', 1, 1, 'active', NOW(), NOW()),
(2, 1, 'dept_mgr', '$2a$10$D3icwPuI7rPaCxOCuuyPFecYbqGy2vJOT40vCy/Qhd6Zz2RU0ufxC', 'mgr@example.com', 2, 0, 'active', NOW(), NOW()),
(3, 1, 'dev_01', '$2a$10$D3icwPuI7rPaCxOCuuyPFecYbqGy2vJOT40vCy/Qhd6Zz2RU0ufxC', 'dev01@example.com', 5, 0, 'active', NOW(), NOW()),
(4, 1, 'dev_02', '$2a$10$D3icwPuI7rPaCxOCuuyPFecYbqGy2vJOT40vCy/Qhd6Zz2RU0ufxC', 'dev02@example.com', 6, 0, 'active', NOW(), NOW()),
(5, 1, 'ops_01', '$2a$10$D3icwPuI7rPaCxOCuuyPFecYbqGy2vJOT40vCy/Qhd6Zz2RU0ufxC', 'ops01@example.com', 9, 0, 'active', NOW(), NOW()),
(6, 1, 'auditor_user', '$2a$10$D3icwPuI7rPaCxOCuuyPFecYbqGy2vJOT40vCy/Qhd6Zz2RU0ufxC', 'audit@example.com', 7, 0, 'active', NOW(), NOW()),
(7, 1, 'locked_user', '$2a$10$D3icwPuI7rPaCxOCuuyPFecYbqGy2vJOT40vCy/Qhd6Zz2RU0ufxC', 'locked@example.com', 4, 0, 'locked', NOW(), NOW()),
(8, 1, 'pending_user', '$2a$10$D3icwPuI7rPaCxOCuuyPFecYbqGy2vJOT40vCy/Qhd6Zz2RU0ufxC', 'pending@example.com', 1, 0, 'pending', NOW(), NOW()),
-- 研发中心租户用户
(9, 2, 'dev_center_admin', '$2a$10$D3icwPuI7rPaCxOCuuyPFecYbqGy2vJOT40vCy/Qhd6Zz2RU0ufxC', 'dev_admin@example.com', 11, 1, 'active', NOW(), NOW()),
(10, 2, 'dev_fe', '$2a$10$D3icwPuI7rPaCxOCuuyPFecYbqGy2vJOT40vCy/Qhd6Zz2RU0ufxC', 'fe@example.com', 12, 0, 'active', NOW(), NOW()),
(11, 2, 'dev_be', '$2a$10$D3icwPuI7rPaCxOCuuyPFecYbqGy2vJOT40vCy/Qhd6Zz2RU0ufxC', 'be@example.com', 13, 0, 'active', NOW(), NOW());

-- 8. 用户关联个体角色
-- admin (1) 拥有 SYSTEM_ADMIN (1)
INSERT INTO `user_roles` (`user_id`, `role_id`) VALUES (1, 1);
-- dept_mgr (2) 额外拥有 DEPT_ADMIN (3)
INSERT INTO `user_roles` (`user_id`, `role_id`) VALUES (2, 3);
-- dev_01 (3) 额外拥有 READ_ONLY (2) 权限
INSERT INTO `user_roles` (`user_id`, `role_id`) VALUES (3, 2);
-- 研发中心管理员 (9) 拥有 TENANT_ADMIN (7) 和租户专属角色 (8)
INSERT INTO `user_roles` (`user_id`, `role_id`) VALUES (9, 7), (9, 8);
-- 前端开发 (10) 拥有 DEVELOPER (4)
INSERT INTO `user_roles` (`user_id`, `role_id`) VALUES (10, 4);
-- 后端开发 (11) 拥有 DEVELOPER (4)
INSERT INTO `user_roles` (`user_id`, `role_id`) VALUES (11, 4);
