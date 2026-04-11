-- 数据范围测试场景补充数据
-- 使用方式:
-- 1. 先执行 backend/scripts/sql/test_data.sql
-- 2. 再执行本文件，补充“用户授权 / 部门授权 / 部门树范围 / 全租户范围 / 跨租户隔离”场景

SET FOREIGN_KEY_CHECKS = 0;
DELETE FROM `user_roles` WHERE `user_id` IN (101, 102, 103, 104) OR `role_id` IN (101, 102, 103, 104, 105);
DELETE FROM `department_roles` WHERE `role_id` IN (101, 102, 103, 104, 105);
DELETE FROM `role_permissions` WHERE `role_id` IN (101, 102, 103, 104, 105);
DELETE FROM `users` WHERE `id` IN (101, 102, 103, 104);
DELETE FROM `roles` WHERE `id` IN (101, 102, 103, 104, 105);
SET FOREIGN_KEY_CHECKS = 1;

-- 1. 创建测试角色
-- 101: 用户直绑，部门树范围，验证“自定义角色也可拥有 department_tree”
-- 102: 用户直绑，仅本人部门，验证“self_department 只放行当前部门”
-- 103: 用户直绑，全租户范围，验证“tenant 只覆盖当前租户，不跨租户”
-- 104: 租户 2 的用户直绑，部门树范围，验证“相同权限在不同租户下仍被 tenant_id 隔离”
-- 105: 部门绑定角色，验证“给部门加权限后，部门成员直接生效”
INSERT INTO `roles`
(`id`, `tenant_id`, `name`, `display_name`, `description`, `type`, `data_scope`, `created_at`, `updated_at`)
VALUES
(101, 1, 'RND_SCOPE_MANAGER', '研发范围管理员', '用户直绑-本人部门树范围测试角色', 'custom', 'department_tree', NOW(), NOW()),
(102, 1, 'OPS_SELF_VIEWER', '运维本部门观察员', '用户直绑-仅本人部门范围测试角色', 'custom', 'self_department', NOW(), NOW()),
(103, 1, 'TENANT_SCOPE_AUDITOR', '租户范围审计员', '用户直绑-全租户范围测试角色', 'custom', 'tenant', NOW(), NOW()),
(104, 2, 'DEV_CENTER_SCOPE_MANAGER', '研发中心范围管理员', '租户2-本人部门树范围测试角色', 'custom', 'department_tree', NOW(), NOW()),
(105, 2, 'FE_DEPT_GRANTED', '前端部门授权角色', '部门绑定-仅本人部门范围测试角色', 'custom', 'self_department', NOW(), NOW());

-- 2. 角色绑定权限
INSERT INTO `role_permissions` (`role_id`, `permission_id`) VALUES
-- RND_SCOPE_MANAGER
(101, 1), (101, 3), (101, 5),
-- OPS_SELF_VIEWER
(102, 1), (102, 5),
-- TENANT_SCOPE_AUDITOR
(103, 1), (103, 5), (103, 9), (103, 18),
-- DEV_CENTER_SCOPE_MANAGER
(104, 1), (104, 2), (104, 5),
-- FE_DEPT_GRANTED
(105, 1), (105, 5);

-- 3. 创建测试用户
-- 101 在默认租户研发部(2)，可访问 2/5/6/8，不能访问 3/4/7/9/10，更不能访问租户 2
-- 102 在默认租户基础运维(9)，只能访问部门 9
-- 103 在默认租户产品部(4)，可访问租户 1 全部数据，但不能访问租户 2
-- 104 在研发中心租户根部门(11)，可访问 11/12/13
INSERT INTO `users`
(`id`, `tenant_id`, `username`, `password`, `email`, `department_id`, `is_admin`, `status`, `created_at`, `updated_at`)
VALUES
(101, 1, 'scope_rd_mgr', '$2a$10$D3icwPuI7rPaCxOCuuyPFecYbqGy2vJOT40vCy/Qhd6Zz2RU0ufxC', 'scope_rd_mgr@example.com', 2, 0, 'active', NOW(), NOW()),
(102, 1, 'scope_ops_self', '$2a$10$D3icwPuI7rPaCxOCuuyPFecYbqGy2vJOT40vCy/Qhd6Zz2RU0ufxC', 'scope_ops_self@example.com', 9, 0, 'active', NOW(), NOW()),
(103, 1, 'scope_tenant_audit', '$2a$10$D3icwPuI7rPaCxOCuuyPFecYbqGy2vJOT40vCy/Qhd6Zz2RU0ufxC', 'scope_tenant_audit@example.com', 4, 0, 'active', NOW(), NOW()),
(104, 2, 'scope_dev_center_mgr', '$2a$10$D3icwPuI7rPaCxOCuuyPFecYbqGy2vJOT40vCy/Qhd6Zz2RU0ufxC', 'scope_dev_center_mgr@example.com', 11, 0, 'active', NOW(), NOW());

-- 4. 用户直绑角色
INSERT INTO `user_roles` (`user_id`, `role_id`) VALUES
(101, 101),
(102, 102),
(103, 103),
(104, 104);

-- 5. 部门绑定角色
-- 将 FE_DEPT_GRANTED 绑定给研发中心租户前端组(12)
-- 现有用户 dev_fe(10) 没有 user:list 直接角色，但执行完本脚本后会通过部门拿到 user:list / department:list
INSERT INTO `department_roles` (`department_id`, `role_id`) VALUES
(12, 105);

-- 6. 场景说明
-- 场景 A: scope_rd_mgr(101) 具备 user:list + department_tree，可验证默认租户研发部树范围
-- 场景 B: scope_ops_self(102) 具备 user:list + self_department，可验证只能访问基础运维(9)
-- 场景 C: scope_tenant_audit(103) 具备 tenant 范围，可验证默认租户全量可见但不跨租户
-- 场景 D: scope_dev_center_mgr(104) 具备租户 2 的 department_tree，可验证 tenant 维度隔离
-- 场景 E: dev_fe(10) 通过部门 12 绑定角色 105 获取权限，可验证“给部门加权限即可生效”
