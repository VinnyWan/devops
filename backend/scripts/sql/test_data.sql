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
SET FOREIGN_KEY_CHECKS = 1;

-- 2. 创建部门 (多层级)
-- 总公司 (1) -> 研发部 (2), 运维部 (3), 产品部 (4), 人力资源部 (7)
-- 研发部 (2) -> 后端组 (5), 前端组 (6), 移动端组 (8)
-- 运维部 (3) -> 基础运维 (9), 应用运维 (10)
INSERT INTO `departments` (`id`, `name`, `parent_id`, `created_at`, `updated_at`) VALUES
(1, '总公司', NULL, NOW(), NOW()),
(2, '研发部', 1, NOW(), NOW()),
(3, '运维部', 1, NOW(), NOW()),
(4, '产品部', 1, NOW(), NOW()),
(5, '后端组', 2, NOW(), NOW()),
(6, '前端组', 2, NOW(), NOW()),
(7, '人力资源部', 1, NOW(), NOW()),
(8, '移动端组', 2, NOW(), NOW()),
(9, '基础运维', 3, NOW(), NOW()),
(10, '应用运维', 3, NOW(), NOW());

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
(18, '查看审计日志', 'audit', 'list', '查看操作审计日志', NOW(), NOW());

-- 4. 创建角色 (系统角色与自定义角色)
INSERT INTO `roles` (`id`, `name`, `description`, `type`, `created_at`, `updated_at`) VALUES
(1, 'SYSTEM_ADMIN', '系统管理员-拥有所有权限', 'system', NOW(), NOW()),
(2, 'READ_ONLY', '全局只读角色', 'system', NOW(), NOW()),
(3, 'DEPT_ADMIN', '部门管理员-管理部门内成员', 'custom', NOW(), NOW()),
(4, 'DEVELOPER', '研发人员-集群开发权限', 'custom', NOW(), NOW()),
(5, 'OPS_ENGINEER', '运维工程师-集群全权管理', 'custom', NOW(), NOW()),
(6, 'AUDITOR', '审计员-仅查看审计日志', 'custom', NOW(), NOW());

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

-- 7. 创建用户 (覆盖各种状态和类型)
-- 密码均为: 123456 (bcrypt: $2a$10$PXeuwlmYZmIGbyOlKM.SiOLRUdHr7.7cIFpcxx5cAfMLn9ZLxTR5i)
-- 注意: 这里的 Hash 是 123456 的正确 bcrypt 值
INSERT INTO `users` (`id`, `username`, `password`, `email`, `department_id`, `is_admin`, `status`, `created_at`, `updated_at`) VALUES
(1, 'admin', '$2a$10$PXeuwlmYZmIGbyOlKM.SiOLRUdHr7.7cIFpcxx5cAfMLn9ZLxTR5i', 'admin@example.com', 1, 1, 'active', NOW(), NOW()),
(2, 'dept_mgr', '$2a$10$PXeuwlmYZmIGbyOlKM.SiOLRUdHr7.7cIFpcxx5cAfMLn9ZLxTR5i', 'mgr@example.com', 2, 0, 'active', NOW(), NOW()),
(3, 'dev_01', '$2a$10$PXeuwlmYZmIGbyOlKM.SiOLRUdHr7.7cIFpcxx5cAfMLn9ZLxTR5i', 'dev01@example.com', 5, 0, 'active', NOW(), NOW()),
(4, 'dev_02', '$2a$10$PXeuwlmYZmIGbyOlKM.SiOLRUdHr7.7cIFpcxx5cAfMLn9ZLxTR5i', 'dev02@example.com', 6, 0, 'active', NOW(), NOW()),
(5, 'ops_01', '$2a$10$PXeuwlmYZmIGbyOlKM.SiOLRUdHr7.7cIFpcxx5cAfMLn9ZLxTR5i', 'ops01@example.com', 9, 0, 'active', NOW(), NOW()),
(6, 'auditor_user', '$2a$10$PXeuwlmYZmIGbyOlKM.SiOLRUdHr7.7cIFpcxx5cAfMLn9ZLxTR5i', 'audit@example.com', 7, 0, 'active', NOW(), NOW()),
(7, 'locked_user', '$2a$10$PXeuwlmYZmIGbyOlKM.SiOLRUdHr7.7cIFpcxx5cAfMLn9ZLxTR5i', 'locked@example.com', 4, 0, 'locked', NOW(), NOW()),
(8, 'pending_user', '$2a$10$PXeuwlmYZmIGbyOlKM.SiOLRUdHr7.7cIFpcxx5cAfMLn9ZLxTR5i', 'pending@example.com', 1, 0, 'pending', NOW(), NOW());

-- 8. 用户关联个体角色
-- admin (1) 拥有 SYSTEM_ADMIN (1)
INSERT INTO `user_roles` (`user_id`, `role_id`) VALUES (1, 1);
-- dept_mgr (2) 额外拥有 DEPT_ADMIN (3)
INSERT INTO `user_roles` (`user_id`, `role_id`) VALUES (2, 3);
-- dev_01 (3) 额外拥有 READ_ONLY (2) 权限
INSERT INTO `user_roles` (`user_id`, `role_id`) VALUES (3, 2);
