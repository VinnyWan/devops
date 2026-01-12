-- ========================================
-- 用户系统完整测试数据初始化SQL
-- ========================================
-- 包含所有用户相关表的测试数据：
-- 1. sys_department - 部门信息
-- 2. sys_post - 岗位信息
-- 3. sys_role - 角色信息
-- 4. sys_menu - 菜单权限
-- 5. sys_user - 用户信息
-- 6. user_roles - 用户角色关联
-- 7. role_menus - 角色菜单关联
-- 8. sys_login_log - 登录日志
-- 9. sys_operation_log - 操作日志
--
-- 使用方法: 
--   方法1: mysql -u root -p devops < scripts/init_user_test_data.sql
--   方法2: mysql -h 10.177.42.165 -P 3306 -u root -prootpassword devops < scripts/init_user_test_data.sql
-- ========================================

-- 设置字符集
SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

-- ========================================
-- 1. 清空现有数据（可选）
-- ========================================
-- 取消下面的注释以清空现有数据
-- TRUNCATE TABLE sys_operation_log;
-- TRUNCATE TABLE sys_login_log;
-- TRUNCATE TABLE role_menus;
-- TRUNCATE TABLE user_roles;
-- TRUNCATE TABLE sys_user;
-- TRUNCATE TABLE sys_menu;
-- TRUNCATE TABLE sys_role;
-- TRUNCATE TABLE sys_post;
-- TRUNCATE TABLE sys_department;

-- ========================================
-- 2. 插入部门数据 (sys_department)
-- ========================================

INSERT INTO `sys_department` (
    `id`,
    `dept_name`,
    `parent_id`,
    `sort`,
    `leader`,
    `phone`,
    `email`,
    `status`,
    `remark`,
    `created_at`,
    `updated_at`
) VALUES
-- 顶级部门
(1, '总公司', 0, 0, '张三', '13800000001', 'zhangsan@example.com', 1, '公司总部', NOW(), NOW()),

-- 二级部门
(2, '技术部', 1, 1, '李四', '13800000002', 'lisi@example.com', 1, '负责技术研发', NOW(), NOW()),
(3, '运维部', 1, 2, '王五', '13800000003', 'wangwu@example.com', 1, '负责系统运维', NOW(), NOW()),
(4, '产品部', 1, 3, '赵六', '13800000004', 'zhaoliu@example.com', 1, '负责产品规划', NOW(), NOW()),
(5, '市场部', 1, 4, '钱七', '13800000005', 'qianqi@example.com', 1, '负责市场推广', NOW(), NOW()),

-- 三级部门（技术部下）
(6, '后端开发组', 2, 1, '李四', '13800000002', 'backend@example.com', 1, '后端开发团队', NOW(), NOW()),
(7, '前端开发组', 2, 2, '周八', '13800000006', 'frontend@example.com', 1, '前端开发团队', NOW(), NOW()),
(8, '测试组', 2, 3, '吴九', '13800000007', 'qa@example.com', 1, '质量保证团队', NOW(), NOW()),

-- 三级部门（运维部下）
(9, 'DevOps组', 3, 1, '王五', '13800000003', 'devops@example.com', 1, 'DevOps团队', NOW(), NOW()),
(10, '安全组', 3, 2, '郑十', '13800000008', 'security@example.com', 1, '安全团队', NOW(), NOW());

-- ========================================
-- 3. 插入岗位数据 (sys_post)
-- ========================================

INSERT INTO `sys_post` (
    `id`,
    `post_name`,
    `post_code`,
    `sort`,
    `status`,
    `remark`,
    `created_at`,
    `updated_at`
) VALUES
(1, '董事长', 'chairman', 1, 1, '公司最高管理者', NOW(), NOW()),
(2, '总经理', 'general_manager', 2, 1, '公司运营负责人', NOW(), NOW()),
(3, '技术总监', 'cto', 3, 1, '技术部门负责人', NOW(), NOW()),
(4, '运维总监', 'ops_director', 4, 1, '运维部门负责人', NOW(), NOW()),
(5, '高级工程师', 'senior_engineer', 5, 1, '资深技术人员', NOW(), NOW()),
(6, '工程师', 'engineer', 6, 1, '普通技术人员', NOW(), NOW()),
(7, '初级工程师', 'junior_engineer', 7, 1, '初级技术人员', NOW(), NOW()),
(8, '产品经理', 'product_manager', 8, 1, '产品规划人员', NOW(), NOW()),
(9, '测试工程师', 'qa_engineer', 9, 1, '质量保证人员', NOW(), NOW()),
(10, '实习生', 'intern', 10, 1, '实习人员', NOW(), NOW());

-- ========================================
-- 4. 插入角色数据 (sys_role)
-- ========================================

INSERT INTO `sys_role` (
    `id`,
    `role_name`,
    `role_key`,
    `sort`,
    `status`,
    `remark`,
    `created_at`,
    `updated_at`
) VALUES
(1, '超级管理员', 'admin', 1, 1, '拥有所有权限', NOW(), NOW()),
(2, '系统管理员', 'system', 2, 1, '系统管理权限', NOW(), NOW()),
(3, '技术人员', 'tech', 3, 1, '技术相关权限', NOW(), NOW()),
(4, '运维人员', 'ops', 4, 1, '运维相关权限', NOW(), NOW()),
(5, '普通用户', 'user', 5, 1, '基础查看权限', NOW(), NOW()),
(6, '访客', 'guest', 6, 1, '只读权限', NOW(), NOW());

-- ========================================
-- 5. 插入菜单数据 (sys_menu)
-- ========================================

INSERT INTO `sys_menu` (
    `id`,
    `menu_name`,
    `parent_id`,
    `sort`,
    `path`,
    `component`,
    `menu_type`,
    `visible`,
    `status`,
    `perms`,
    `icon`,
    `remark`,
    `created_at`,
    `updated_at`
) VALUES
-- 一级菜单
(1, '系统管理', 0, 1, '/system', 'Layout', 'M', 1, 1, '', 'system', '系统管理目录', NOW(), NOW()),
(2, 'K8s管理', 0, 2, '/k8s', 'Layout', 'M', 1, 1, '', 'cloud', 'K8s集群管理', NOW(), NOW()),
(3, '监控中心', 0, 3, '/monitor', 'Layout', 'M', 1, 1, '', 'monitor', '系统监控目录', NOW(), NOW()),

-- 系统管理子菜单
(11, '用户管理', 1, 1, 'user', 'system/user/index', 'C', 1, 1, 'system:user:list', 'user', '用户管理菜单', NOW(), NOW()),
(12, '角色管理', 1, 2, 'role', 'system/role/index', 'C', 1, 1, 'system:role:list', 'peoples', '角色管理菜单', NOW(), NOW()),
(13, '菜单管理', 1, 3, 'menu', 'system/menu/index', 'C', 1, 1, 'system:menu:list', 'tree-table', '菜单管理菜单', NOW(), NOW()),
(14, '部门管理', 1, 4, 'dept', 'system/dept/index', 'C', 1, 1, 'system:dept:list', 'tree', '部门管理菜单', NOW(), NOW()),
(15, '岗位管理', 1, 5, 'post', 'system/post/index', 'C', 1, 1, 'system:post:list', 'post', '岗位管理菜单', NOW(), NOW()),

-- 用户管理按钮
(111, '用户查询', 11, 1, '', '', 'B', 1, 1, 'system:user:query', '', '用户查询按钮', NOW(), NOW()),
(112, '用户新增', 11, 2, '', '', 'B', 1, 1, 'system:user:add', '', '用户新增按钮', NOW(), NOW()),
(113, '用户修改', 11, 3, '', '', 'B', 1, 1, 'system:user:edit', '', '用户修改按钮', NOW(), NOW()),
(114, '用户删除', 11, 4, '', '', 'B', 1, 1, 'system:user:delete', '', '用户删除按钮', NOW(), NOW()),
(115, '重置密码', 11, 5, '', '', 'B', 1, 1, 'system:user:resetPwd', '', '重置密码按钮', NOW(), NOW()),

-- K8s管理子菜单
(21, '集群管理', 2, 1, 'cluster', 'k8s/cluster/index', 'C', 1, 1, 'k8s:cluster:list', 'server', '集群管理菜单', NOW(), NOW()),
(22, '命名空间', 2, 2, 'namespace', 'k8s/namespace/index', 'C', 1, 1, 'k8s:namespace:list', 'component', '命名空间菜单', NOW(), NOW()),
(23, '工作负载', 2, 3, 'workload', 'k8s/workload/index', 'C', 1, 1, 'k8s:workload:list', 'nested', '工作负载菜单', NOW(), NOW()),

-- 集群管理按钮
(211, '集群查询', 21, 1, '', '', 'B', 1, 1, 'k8s:cluster:query', '', '集群查询按钮', NOW(), NOW()),
(212, '集群新增', 21, 2, '', '', 'B', 1, 1, 'k8s:cluster:add', '', '集群新增按钮', NOW(), NOW()),
(213, '集群修改', 21, 3, '', '', 'B', 1, 1, 'k8s:cluster:edit', '', '集群修改按钮', NOW(), NOW()),
(214, '集群删除', 21, 4, '', '', 'B', 1, 1, 'k8s:cluster:delete', '', '集群删除按钮', NOW(), NOW()),

-- 监控中心子菜单
(31, '登录日志', 3, 1, 'loginLog', 'monitor/loginLog/index', 'C', 1, 1, 'monitor:loginLog:list', 'logininfor', '登录日志菜单', NOW(), NOW()),
(32, '操作日志', 3, 2, 'operLog', 'monitor/operLog/index', 'C', 1, 1, 'monitor:operLog:list', 'form', '操作日志菜单', NOW(), NOW());

-- ========================================
-- 6. 插入用户数据 (sys_user)
-- ========================================
-- 密码统一为: admin123 (经过bcrypt加密后的值)

INSERT INTO `sys_user` (
    `id`,
    `username`,
    `password`,
    `nickname`,
    `email`,
    `phone`,
    `avatar`,
    `status`,
    `gender`,
    `dept_id`,
    `post_id`,
    `remark`,
    `created_at`,
    `updated_at`
) VALUES
-- 管理员用户
(1, 'admin', '$2a$10$7JB720yubVSZvUI0rEqK/.VqGOZTH.ulu33dHOiBE/TP8V38ZjQiK', '超级管理员', 'admin@example.com', '13800000001', '/avatar/admin.jpg', 1, 1, 1, 1, '系统超级管理员', NOW(), NOW()),
(2, 'system', '$2a$10$7JB720yubVSZvUI0rEqK/.VqGOZTH.ulu33dHOiBE/TP8V38ZjQiK', '系统管理员', 'system@example.com', '13800000010', '/avatar/system.jpg', 1, 1, 1, 2, '系统管理员账号', NOW(), NOW()),

-- 技术部用户
(3, 'lisi', '$2a$10$7JB720yubVSZvUI0rEqK/.VqGOZTH.ulu33dHOiBE/TP8V38ZjQiK', '李四', 'lisi@example.com', '13800000002', '/avatar/lisi.jpg', 1, 1, 2, 3, '技术总监', NOW(), NOW()),
(4, 'zhangsan_dev', '$2a$10$7JB720yubVSZvUI0rEqK/.VqGOZTH.ulu33dHOiBE/TP8V38ZjQiK', '张三', 'zhangsan@example.com', '13800000011', '/avatar/default.jpg', 1, 1, 6, 5, '后端高级工程师', NOW(), NOW()),
(5, 'wangwu_dev', '$2a$10$7JB720yubVSZvUI0rEqK/.VqGOZTH.ulu33dHOiBE/TP8V38ZjQiK', '王五', 'wangwu@example.com', '13800000012', '/avatar/default.jpg', 1, 1, 6, 6, '后端工程师', NOW(), NOW()),
(6, 'zhaoliu_fe', '$2a$10$7JB720yubVSZvUI0rEqK/.VqGOZTH.ulu33dHOiBE/TP8V38ZjQiK', '赵六', 'zhaoliu@example.com', '13800000013', '/avatar/default.jpg', 1, 1, 7, 6, '前端工程师', NOW(), NOW()),
(7, 'qianqi_qa', '$2a$10$7JB720yubVSZvUI0rEqK/.VqGOZTH.ulu33dHOiBE/TP8V38ZjQiK', '钱七', 'qianqi@example.com', '13800000014', '/avatar/default.jpg', 1, 2, 8, 9, '测试工程师', NOW(), NOW()),

-- 运维部用户
(8, 'sunba_ops', '$2a$10$7JB720yubVSZvUI0rEqK/.VqGOZTH.ulu33dHOiBE/TP8V38ZjQiK', '孙八', 'sunba@example.com', '13800000015', '/avatar/default.jpg', 1, 1, 3, 4, '运维总监', NOW(), NOW()),
(9, 'zhoujiu_devops', '$2a$10$7JB720yubVSZvUI0rEqK/.VqGOZTH.ulu33dHOiBE/TP8V38ZjQiK', '周九', 'zhoujiu@example.com', '13800000016', '/avatar/default.jpg', 1, 1, 9, 5, 'DevOps工程师', NOW(), NOW()),
(10, 'wushi_sec', '$2a$10$7JB720yubVSZvUI0rEqK/.VqGOZTH.ulu33dHOiBE/TP8V38ZjQiK', '吴十', 'wushi@example.com', '13800000017', '/avatar/default.jpg', 1, 1, 10, 5, '安全工程师', NOW(), NOW()),

-- 禁用/测试用户
(11, 'disabled_user', '$2a$10$7JB720yubVSZvUI0rEqK/.VqGOZTH.ulu33dHOiBE/TP8V38ZjQiK', '禁用用户', 'disabled@example.com', '13800000018', '/avatar/default.jpg', 2, 0, 1, 10, '测试禁用状态的用户', NOW(), NOW()),
(12, 'test_user', '$2a$10$7JB720yubVSZvUI0rEqK/.VqGOZTH.ulu33dHOiBE/TP8V38ZjQiK', '测试用户', 'test@example.com', '13800000019', '/avatar/default.jpg', 1, 0, 1, 10, '普通测试用户', NOW(), NOW());

-- ========================================
-- 7. 插入用户角色关联 (user_roles)
-- ========================================

INSERT INTO `user_roles` (`user_id`, `role_id`) VALUES
-- admin - 超级管理员
(1, 1),

-- system - 系统管理员
(2, 2),

-- lisi - 技术总监(技术人员)
(3, 3),

-- 后端开发(技术人员)
(4, 3),
(5, 3),

-- 前端开发(技术人员)
(6, 3),

-- 测试(普通用户)
(7, 5),

-- 运维总监(运维人员)
(8, 4),

-- DevOps工程师(运维人员)
(9, 4),

-- 安全工程师(运维人员 + 系统管理员)
(10, 2),
(10, 4),

-- 测试用户(访客)
(11, 6),
(12, 5);

-- ========================================
-- 8. 插入角色菜单关联 (role_menus)
-- ========================================

INSERT INTO `role_menus` (`role_id`, `menu_id`) VALUES
-- 超级管理员 - 拥有所有权限
(1, 1), (1, 2), (1, 3),
(1, 11), (1, 12), (1, 13), (1, 14), (1, 15),
(1, 111), (1, 112), (1, 113), (1, 114), (1, 115),
(1, 21), (1, 22), (1, 23),
(1, 211), (1, 212), (1, 213), (1, 214),
(1, 31), (1, 32),

-- 系统管理员 - 系统管理相关
(2, 1), (2, 3),
(2, 11), (2, 12), (2, 13), (2, 14), (2, 15),
(2, 111), (2, 112), (2, 113), (2, 114), (2, 115),
(2, 31), (2, 32),

-- 技术人员 - K8s管理权限
(3, 2),
(3, 21), (3, 22), (3, 23),
(3, 211), (3, 212), (3, 213),

-- 运维人员 - K8s管理 + 监控
(4, 2), (4, 3),
(4, 21), (4, 22), (4, 23),
(4, 211), (4, 212), (4, 213), (4, 214),
(4, 31), (4, 32),

-- 普通用户 - 基础查看权限
(5, 1), (5, 2),
(5, 11), (5, 14),
(5, 111),
(5, 21), (5, 22),
(5, 211),

-- 访客 - 只读权限
(6, 1),
(6, 11), (6, 14),
(6, 111);

-- ========================================
-- 9. 插入登录日志 (sys_login_log)
-- ========================================

INSERT INTO `sys_login_log` (
    `username`,
    `ip`,
    `location`,
    `browser`,
    `os`,
    `status`,
    `message`,
    `login_time`,
    `created_at`
) VALUES
-- 成功的登录记录
('admin', '127.0.0.1', '本地', 'Chrome 120', 'macOS', 1, '登录成功', NOW() - INTERVAL 1 HOUR, NOW() - INTERVAL 1 HOUR),
('admin', '127.0.0.1', '本地', 'Chrome 120', 'macOS', 1, '登录成功', NOW() - INTERVAL 2 HOUR, NOW() - INTERVAL 2 HOUR),
('lisi', '192.168.1.100', '北京', 'Firefox 121', 'Windows 10', 1, '登录成功', NOW() - INTERVAL 3 HOUR, NOW() - INTERVAL 3 HOUR),
('zhangsan_dev', '192.168.1.101', '上海', 'Chrome 120', 'Windows 11', 1, '登录成功', NOW() - INTERVAL 4 HOUR, NOW() - INTERVAL 4 HOUR),
('wangwu_dev', '192.168.1.102', '深圳', 'Edge 120', 'Windows 10', 1, '登录成功', NOW() - INTERVAL 5 HOUR, NOW() - INTERVAL 5 HOUR),
('sunba_ops', '192.168.1.103', '杭州', 'Safari 17', 'macOS', 1, '登录成功', NOW() - INTERVAL 6 HOUR, NOW() - INTERVAL 6 HOUR),
('zhoujiu_devops', '192.168.1.104', '成都', 'Chrome 120', 'Linux', 1, '登录成功', NOW() - INTERVAL 8 HOUR, NOW() - INTERVAL 8 HOUR),

-- 失败的登录记录
('admin', '203.0.113.1', '未知', 'Chrome 120', 'Windows 10', 2, '密码错误', NOW() - INTERVAL 12 HOUR, NOW() - INTERVAL 12 HOUR),
('test_user', '203.0.113.2', '未知', 'Chrome 120', 'Windows 10', 2, '用户不存在', NOW() - INTERVAL 1 DAY, NOW() - INTERVAL 1 DAY),
('admin', '203.0.113.3', '未知', 'Chrome 120', 'macOS', 2, '验证码错误', NOW() - INTERVAL 2 DAY, NOW() - INTERVAL 2 DAY),

-- 历史登录记录
('admin', '127.0.0.1', '本地', 'Chrome 119', 'macOS', 1, '登录成功', NOW() - INTERVAL 3 DAY, NOW() - INTERVAL 3 DAY),
('system', '192.168.1.200', '北京', 'Firefox 120', 'Windows 10', 1, '登录成功', NOW() - INTERVAL 4 DAY, NOW() - INTERVAL 4 DAY),
('lisi', '192.168.1.100', '北京', 'Firefox 121', 'Windows 10', 1, '登录成功', NOW() - INTERVAL 5 DAY, NOW() - INTERVAL 5 DAY),
('disabled_user', '192.168.1.105', '广州', 'Chrome 120', 'Windows 10', 2, '账号已被禁用', NOW() - INTERVAL 7 DAY, NOW() - INTERVAL 7 DAY);

-- ========================================
-- 10. 插入操作日志 (sys_operation_log)
-- ========================================

INSERT INTO `sys_operation_log` (
    `module`,
    `type`,
    `title`,
    `method`,
    `request_url`,
    `request_param`,
    `response_data`,
    `ip`,
    `location`,
    `status`,
    `error_msg`,
    `cost_time`,
    `operator_id`,
    `operator_name`,
    `created_at`
) VALUES
-- 用户管理操作
('用户管理', 'CREATE', '新增用户', 'POST', '/api/system/users', '{"username":"test_user","nickname":"测试用户"}', '{"code":200,"msg":"操作成功"}', '127.0.0.1', '本地', 1, '', 125, 1, 'admin', NOW() - INTERVAL 1 HOUR),
('用户管理', 'UPDATE', '修改用户', 'PUT', '/api/system/users/12', '{"nickname":"测试用户更新"}', '{"code":200,"msg":"操作成功"}', '127.0.0.1', '本地', 1, '', 89, 1, 'admin', NOW() - INTERVAL 2 HOUR),
('用户管理', 'DELETE', '删除用户', 'DELETE', '/api/system/users/999', '{}', '{"code":500,"msg":"用户不存在"}', '127.0.0.1', '本地', 2, '用户不存在', 45, 1, 'admin', NOW() - INTERVAL 3 HOUR),
('用户管理', 'QUERY', '查询用户列表', 'GET', '/api/system/users?page=1&pageSize=10', '{}', '{"code":200,"data":{"total":12}}', '127.0.0.1', '本地', 1, '', 156, 2, 'system', NOW() - INTERVAL 4 HOUR),

-- 角色管理操作
('角色管理', 'CREATE', '新增角色', 'POST', '/api/system/roles', '{"roleName":"测试角色","roleKey":"test"}', '{"code":200,"msg":"操作成功"}', '127.0.0.1', '本地', 1, '', 98, 2, 'system', NOW() - INTERVAL 5 HOUR),
('角色管理', 'UPDATE', '修改角色权限', 'PUT', '/api/system/roles/3', '{"menus":[1,2,3]}', '{"code":200,"msg":"操作成功"}', '127.0.0.1', '本地', 1, '', 234, 2, 'system', NOW() - INTERVAL 6 HOUR),

-- 部门管理操作
('部门管理', 'CREATE', '新增部门', 'POST', '/api/system/departments', '{"deptName":"研发中心","parentId":2}', '{"code":200,"msg":"操作成功"}', '192.168.1.100', '北京', 1, '', 67, 1, 'admin', NOW() - INTERVAL 8 HOUR),
('部门管理', 'UPDATE', '修改部门', 'PUT', '/api/system/departments/6', '{"leader":"新领导"}', '{"code":200,"msg":"操作成功"}', '192.168.1.100', '北京', 1, '', 54, 3, 'lisi', NOW() - INTERVAL 10 HOUR),

-- K8s集群操作
('K8s管理', 'CREATE', '新增集群', 'POST', '/api/k8s/clusters', '{"name":"测试集群","apiServer":"https://test.k8s.com"}', '{"code":200,"msg":"操作成功"}', '192.168.1.104', '成都', 1, '', 456, 9, 'zhoujiu_devops', NOW() - INTERVAL 12 HOUR),
('K8s管理', 'UPDATE', '更新集群', 'PUT', '/api/k8s/clusters/1', '{"description":"更新描述"}', '{"code":200,"msg":"操作成功"}', '192.168.1.103', '杭州', 1, '', 234, 8, 'sunba_ops', NOW() - INTERVAL 1 DAY),
('K8s管理', 'DELETE', '删除集群', 'DELETE', '/api/k8s/clusters/2', '{}', '{"code":200,"msg":"操作成功"}', '192.168.1.103', '杭州', 1, '', 189, 8, 'sunba_ops', NOW() - INTERVAL 2 DAY),
('K8s管理', 'QUERY', '查询集群列表', 'GET', '/api/k8s/clusters?page=1', '{}', '{"code":200,"data":{"total":4}}', '192.168.1.104', '成都', 1, '', 123, 9, 'zhoujiu_devops', NOW() - INTERVAL 3 DAY),

-- 系统配置操作
('系统配置', 'UPDATE', '修改系统配置', 'PUT', '/api/system/config', '{"key":"site_name","value":"DevOps平台"}', '{"code":200,"msg":"操作成功"}', '127.0.0.1', '本地', 1, '', 78, 1, 'admin', NOW() - INTERVAL 4 DAY),

-- 失败的操作
('用户管理', 'DELETE', '批量删除用户', 'DELETE', '/api/system/users/batch', '{"ids":[1,2]}', '{"code":500,"msg":"不能删除管理员"}', '127.0.0.1', '本地', 2, '不能删除管理员账号', 67, 2, 'system', NOW() - INTERVAL 5 DAY),
('K8s管理', 'CREATE', '新增集群', 'POST', '/api/k8s/clusters', '{"name":"重复集群"}', '{"code":500,"msg":"集群名称已存在"}', '192.168.1.104', '成都', 2, '集群名称已存在', 145, 9, 'zhoujiu_devops', NOW() - INTERVAL 6 DAY),

-- 最近的操作
('用户管理', 'QUERY', '查询用户详情', 'GET', '/api/system/users/1', '{}', '{"code":200,"data":{"username":"admin"}}', '127.0.0.1', '本地', 1, '', 45, 1, 'admin', NOW() - INTERVAL 30 MINUTE),
('部门管理', 'QUERY', '查询部门树', 'GET', '/api/system/departments/tree', '{}', '{"code":200,"data":[]}', '127.0.0.1', '本地', 1, '', 89, 1, 'admin', NOW() - INTERVAL 10 MINUTE);

-- 恢复外键检查
SET FOREIGN_KEY_CHECKS = 1;

-- ========================================
-- 11. 数据验证和统计
-- ========================================

-- 查询部门列表
SELECT 
    '部门信息' as table_name,
    id,
    dept_name,
    parent_id,
    leader,
    status
FROM sys_department
ORDER BY parent_id, sort;

-- 查询岗位列表
SELECT 
    '岗位信息' as table_name,
    id,
    post_name,
    post_code,
    status
FROM sys_post
ORDER BY sort;

-- 查询角色列表
SELECT 
    '角色信息' as table_name,
    id,
    role_name,
    role_key,
    status
FROM sys_role
ORDER BY sort;

-- 查询用户列表（前10个）
SELECT 
    '用户信息' as table_name,
    id,
    username,
    nickname,
    email,
    dept_id,
    post_id,
    status
FROM sys_user
ORDER BY id
LIMIT 10;

-- 查询用户角色关联
SELECT 
    '用户角色关联' as table_name,
    u.username,
    r.role_name
FROM user_roles ur
JOIN sys_user u ON ur.user_id = u.id
JOIN sys_role r ON ur.role_id = r.id
ORDER BY u.id;

-- 查询最近登录日志（最新10条）
SELECT 
    '登录日志（最新10条）' as table_name,
    username,
    ip,
    location,
    status,
    message,
    login_time
FROM sys_login_log
ORDER BY login_time DESC
LIMIT 10;

-- 查询最近操作日志（最新10条）
SELECT 
    '操作日志（最新10条）' as table_name,
    module,
    title,
    method,
    operator_name,
    status,
    cost_time,
    created_at
FROM sys_operation_log
ORDER BY created_at DESC
LIMIT 10;

-- ========================================
-- 12. 统计信息
-- ========================================

-- 部门统计
SELECT 
    '=== 部门统计 ===' as summary,
    COUNT(*) as total_depts,
    SUM(CASE WHEN parent_id = 0 THEN 1 ELSE 0 END) as top_level_depts,
    SUM(CASE WHEN status = 1 THEN 1 ELSE 0 END) as active_depts
FROM sys_department;

-- 用户统计
SELECT 
    '=== 用户统计 ===' as summary,
    COUNT(*) as total_users,
    SUM(CASE WHEN status = 1 THEN 1 ELSE 0 END) as active_users,
    SUM(CASE WHEN status = 2 THEN 1 ELSE 0 END) as disabled_users,
    SUM(CASE WHEN gender = 1 THEN 1 ELSE 0 END) as male_users,
    SUM(CASE WHEN gender = 2 THEN 1 ELSE 0 END) as female_users
FROM sys_user;

-- 角色统计
SELECT 
    '=== 角色统计 ===' as summary,
    COUNT(*) as total_roles,
    SUM(CASE WHEN status = 1 THEN 1 ELSE 0 END) as active_roles
FROM sys_role;

-- 菜单统计
SELECT 
    '=== 菜单统计 ===' as summary,
    COUNT(*) as total_menus,
    SUM(CASE WHEN menu_type = 'M' THEN 1 ELSE 0 END) as directory_count,
    SUM(CASE WHEN menu_type = 'C' THEN 1 ELSE 0 END) as menu_count,
    SUM(CASE WHEN menu_type = 'B' THEN 1 ELSE 0 END) as button_count
FROM sys_menu;

-- 登录日志统计
SELECT 
    '=== 登录日志统计 ===' as summary,
    COUNT(*) as total_logins,
    SUM(CASE WHEN status = 1 THEN 1 ELSE 0 END) as success_logins,
    SUM(CASE WHEN status = 2 THEN 1 ELSE 0 END) as failed_logins
FROM sys_login_log;

-- 操作日志统计
SELECT 
    '=== 操作日志统计 ===' as summary,
    COUNT(*) as total_operations,
    SUM(CASE WHEN status = 1 THEN 1 ELSE 0 END) as success_operations,
    SUM(CASE WHEN status = 2 THEN 1 ELSE 0 END) as failed_operations,
    AVG(cost_time) as avg_cost_time
FROM sys_operation_log;

-- 按模块统计操作
SELECT 
    '=== 按模块统计操作 ===' as summary,
    module,
    COUNT(*) as operation_count,
    AVG(cost_time) as avg_cost_time
FROM sys_operation_log
GROUP BY module
ORDER BY operation_count DESC;

-- 数据导入完成提示
SELECT '
========================================
✅ 用户系统测试数据导入完成！
========================================
数据概览：
- 10 个部门（3级树形结构）
- 10 个岗位
- 6 个角色
- 29 个菜单（含按钮权限）
- 12 个用户（含禁用用户）
- 14 条用户角色关联
- 48 条角色菜单关联
- 14 条登录日志
- 17 条操作日志

测试账号说明（密码统一为: admin123）：
1. admin - 超级管理员（拥有所有权限）
2. system - 系统管理员（系统管理权限）
3. lisi - 技术总监（技术相关权限）
4. zhangsan_dev - 后端高级工程师
5. wangwu_dev - 后端工程师
6. zhaoliu_fe - 前端工程师
7. qianqi_qa - 测试工程师
8. sunba_ops - 运维总监（运维权限）
9. zhoujiu_devops - DevOps工程师
10. wushi_sec - 安全工程师（运维+系统管理）
11. disabled_user - 禁用用户（用于测试）
12. test_user - 普通测试用户

现在可以测试登录：
curl -X POST "http://127.0.0.1:8000/api/auth/login" \
  -H "Content-Type: application/json" \
  -d "{\"username\":\"admin\",\"password\":\"admin123\"}"

查询用户列表：
curl -X GET "http://127.0.0.1:8000/api/system/users?page=1&pageSize=10" \
  -H "Authorization: Bearer YOUR_TOKEN"
========================================
' as message;
