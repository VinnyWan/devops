-- ========================================
-- K8s完整测试数据初始化SQL
-- ========================================
-- 包含所有K8s相关表的测试数据：
-- 1. k8s_clusters - 集群信息
-- 2. k8s_cluster_accesses - 集群访问权限
-- 3. k8s_namespaces - 命名空间记录
-- 4. k8s_operation_logs - 操作日志
--
-- 使用方法: 
--   方法1: mysql -u root -p devops < scripts/init_k8s_test_data.sql
--   方法2: mysql -h 10.177.42.165 -P 3306 -u root -prootpassword devops < scripts/init_k8s_test_data.sql
-- ========================================

-- 设置字符集
SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

-- ========================================
-- 1. 清空现有数据（可选）
-- ========================================
-- 取消下面的注释以清空现有数据
-- TRUNCATE TABLE k8s_operation_logs;
-- TRUNCATE TABLE k8s_namespaces;
-- TRUNCATE TABLE k8s_cluster_accesses;
-- TRUNCATE TABLE k8s_clusters;

-- ========================================
-- 2. 插入集群测试数据 (k8s_clusters)
-- ========================================
-- 注意：dept_id 使用 1，请确保 departments 表中存在 id=1 的记录

-- 1. 本地开发集群
INSERT INTO `k8s_clusters` (
    `name`,
    `description`,
    `api_server`,
    `kube_config`,
    `version`,
    `import_method`,
    `import_status`,
    `cluster_status`,
    `status`,
    `dept_id`,
    `remark`,
    `created_at`,
    `updated_at`
) VALUES (
    '本地开发集群',
    '本地K3s开发测试集群，用于功能测试和开发',
    'https://127.0.0.1:6443',
    'apiVersion: v1
kind: Config
clusters:
- cluster:
    certificate-authority-data: LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUJkakNDQVIyZ0F3SUJBZ0lCQURBS0JnZ3Foa2pPUFFRREFqQWpNU0V3SHdZRFZRUUREQmhyTTNNdGMyVnkKZG1WeUxXTmhRREUzTXpZMU5EQTRNemN3SGhjTk1qWXdNVEF4TURFd05qRTNXaGNOTXpZd01EQTVNREV3TmpFMwpXakFqTVNFd0h3WURWUVFEREJock0zTXRjMlZ5ZG1WeUxXTmhRREUzTXpZMU5EQTRNemN3V1RBVEJnY3Foa2pPClBRSUJCZ2dxaGtqT1BRTUJCd05DQUFUTmVHbGtKZ3RXbVJOUlVycHU3M3lrcko4RlhVMlZFTGo2cS9TMHVBbkoKbFpDa1dXOVJJcTVMMXh2TktLV0lGUGprcW54UVRCZHRybWl1UE1DQzdXSU5vMEl3UURBT0JnTlZIUThCQWY4RQpCQU1DQXFRd0R3WURWUjBUQVFIL0JBVXdBd0VCL3pBZEJnTlZIUTRFRmdRVVo4SXRNTlRVT25qZ3M2MnFBdmhrCnFxcW5uRXd3Q2dZSUtvWkl6ajBFQXdJRFJ3QXdSQUlnV3pTQmRYUzJnVm1hZlZKSjRnRkJBZkNWOXphRFVzYkoKOUVhWDcwL0NsQkVDSUY4dVo2TWRLSjBQcFppbzE5VFpFU3VSYXNBV1FiSDRKZVdCZmh5L2ZzMgotLS0tLUVORCBDRVJUSUZJQ0FURS0tLS0t
    server: https://127.0.0.1:6443
  name: k3s-default
contexts:
- context:
    cluster: k3s-default
    user: k3s-default
  name: k3s-default
current-context: k3s-default
users:
- name: k3s-default
  user:
    client-certificate-data: LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUJrakNDQVRlZ0F3SUJBZ0lJZFNHc1Q0RzFMclF3Q2dZSUtvWkl6ajBFQXdJd0l6RWhNQjhHQTFVRUF3d1kKYXpOekxXTnNhV1Z1ZEMxallVQXhOek0yTlRRd09ETTNNQTR4RGpBTUJnTlZCQU1NQldGa2JXbHVNQjRYRFRJMgpNREV3TVRBeE1EWTVOME1YRFRJM01ERXdNVEF4TURZNU4xb3dNREVYTUJVR0ExVUVDaE1PYzNsemRHVnRPbTFoCmMzUmxjbk14RlRBVEJnTlZCQU1UREhONWMzUmxiVHBoWkdsdGFUQlpNQk1HQnlxR1NNNDlBZ0VHQ0NxR1NNNDkKQXdFSEEwSUFCRmxpc1hNbVRqbFFZV29FNTBFUTloR09LZnJMRFFPMFlmYWZHZ3Uyd080cWNsNnM1cXBJWUhaYgpYdnBmcGdIcjVGOUhOOWRMS0FWelduZnErbU9mS0JxalNEQkdNQTRHQTFVZER3RUIvd1FFQXdJRm9EQVRCZ05WCkhTVUVEREFLQmdnckJnRUZCUWNEQWpBZkJnTlZIU01FR0RBV2dCUnFuVk5QV1FkQUxYb25HbDZYa3RldUlwNGIKU1RBS0JnZ3Foa2pPUFFRREFnTkpBREJHQWlFQXhoV2Y5NjBPRHhQWlpIUHI5dkdFa0FYQldzWmUrVGRWL29NNApSeElNZ0FJRFFVWHRTY0VVTUg5NUNmU3YycUoxRzNiQ1Z4Y3UwRHI1cUUrcWc4UkRjQkU9Ci0tLS0tRU5EIENFUlRJRklDQVRFLS0tLS0KLS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUJkVENDQVJ5Z0F3SUJBZ0lCQURBS0JnZ3Foa2pPUFFRREFqQWpNU0V3SHdZRFZRUUREQmhyTTNNdFkyeHAKWlc1MExXTmhRREUzTXpZMU5EQTRNemN3SGhjTk1qWXdNVEF4TURFd05qRTNXaGNOTXpZd01EQTVNREV3TmpFMwpXakFqTVNFd0h3WURWUVFEREJock0zTXRZMnhwWlc1MExXTmhRREUzTXpZMU5EQTRNemN3V1RBVEJnY3Foa2pPClBRSUJCZ2dxaGtqT1BRTUJCd05DQUFTNlhOdERjUStNUlZsZ0lVY1piZEw3UkJOckJVMEpTT3pTZjdFd1p0bnUKYktQYmV4QWEyaFEvL1FoeE56cit4S2pUWFd0Yk5xU2xIL1JIT1dKRjNyS1hvMEl3UURBT0JnTlZIUThCQWY4RQpCQU1DQXFRd0R3WURWUjBUQVFIL0JBVXdBd0VCL3pBZEJnTlZIUTRFRmdRVWFwMVRUMWtIUUMxNkp4cGVsNUxYCnJpS2VHMGt3Q2dZSUtvWkl6ajBFQXdJRFJ3QXdSQUlnZXhWNU5tSU9YS3k4aU5rUWZsWHlXaHdVRCtxVU1Dc2oKaUg5aG9GRnZQMGdDSUNsMFdBMHZBK01XQ3U1MEs1cDFKQ2poaVFvczVqYlBqdEFLM2EvWnJ2YXMKLS0tLS1FTkQgQ0VSVElGSUNBVEUtLS0tLQo=
    client-key-data: LS0tLS1CRUdJTiBFQyBQUklWQVRFIEtFWS0tLS0tCk1IY0NBUUVFSUdJOW5xNThWaFF3UUQ5MXBkOUZlOEkzU2VsQmFVWXhEZ21BdC9BL2J4YU5vQW9HQ0NxR1NNNDkKQXdFSG9VUURRZ0FFV1dLeGN5Wk9PVkJoYWdUblFSRDJFWTRwK3NzTkE3Umg5cDhhQzdiWTdpcHlYcXptcWtoZwpkbHRlK2wrbUFldmtYMGMzMTBzb0JYTmFkK3I2WTU4b0dnPT0KLS0tLS1FTkQgRUMgUFJJVkFURSBLRVktLS0tLQo=',
    'v1.28.5+k3s1',
    'kubeconfig',
    'success',
    'unknown',
    1,
    1,
    '系统自动初始化的测试集群，可直接用于API测试',
    NOW(),
    NOW()
);

-- 2. 测试集群-1.27
INSERT INTO `k8s_clusters` (
    `name`,
    `description`,
    `api_server`,
    `kube_config`,
    `version`,
    `import_method`,
    `import_status`,
    `cluster_status`,
    `status`,
    `dept_id`,
    `remark`,
    `created_at`,
    `updated_at`
) VALUES (
    '测试集群-1.27',
    'Kubernetes 1.27版本测试集群',
    'https://test-k8s-1-27.example.com:6443',
    'apiVersion: v1
kind: Config
clusters:
- cluster:
    certificate-authority-data: LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUJkakNDQVIyZ0F3SUJBZ0lCQURBS0JnZ3Foa2pPUFFRREFqQWpNU0V3SHdZRFZRUUREQmhyTTNNdGMyVnkKZG1WeUxXTmhRREUzTXpZMU5EQTRNemN3SGhjTk1qWXdNVEF4TURFd05qRTNXaGNOTXpZd01EQTVNREV3TmpFMwpXakFqTVNFd0h3WURWUVFEREJock0zTXRjMlZ5ZG1WeUxXTmhRREUzTXpZMU5EQTRNemN3V1RBVEJnY3Foa2pPClBRSUJCZ2dxaGtqT1BRTUJCd05DQUFUTmVHbGtKZ3RXbVJOUlVycHU3M3lrcko4RlhVMlZFTGo2cS9TMHVBbkoKbFpDa1dXOVJJcTVMMXh2TktLV0lGUGprcW54UVRCZHRybWl1UE1DQzdXSU5vMEl3UURBT0JnTlZIUThCQWY4RQpCQU1DQXFRd0R3WURWUjBUQVFIL0JBVXdBd0VCL3pBZEJnTlZIUTRFRmdRVVo4SXRNTlRVT25qZ3M2MnFBdmhrCnFxcW5uRXd3Q2dZSUtvWkl6ajBFQXdJRFJ3QXdSQUlnV3pTQmRYUzJnVm1hZlZKSjRnRkJBZkNWOXphRFVzYkoKOUVhWDcwL0NsQkVDSUY4dVo2TWRLSjBQcFppbzE5VFpFU3VSYXNBV1FiSDRKZVdCZmh5L2ZzMgotLS0tLUVORCBDRVJUSUZJQ0FURS0tLS0t
    server: https://test-k8s-1-27.example.com:6443
  name: test-cluster
contexts:
- context:
    cluster: test-cluster
    user: test-user
  name: test-context
current-context: test-context
users:
- name: test-user
  user:
    client-certificate-data: LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUJrakNDQVRlZ0F3SUJBZ0lJZFNHc1Q0RzFMclF3Q2dZSUtvWkl6ajBFQXdJd0l6RWhNQjhHQTFVRUF3d1kKYXpOekxXTnNhV1Z1ZEMxallVQXhOek0yTlRRd09ETTNNQTR4RGpBTUJnTlZCQU1NQldGa2JXbHVNQjRYRFRJMgpNREV3TVRBeE1EWTVOME1YRFRJM01ERXdNVEF4TURZNU4xb3dNREVYTUJVR0ExVUVDaE1PYzNsemRHVnRPbTFoCmMzUmxjbk14RlRBVEJnTlZCQU1UREhONWMzUmxiVHBoWkdsdGFUQlpNQk1HQnlxR1NNNDlBZ0VHQ0NxR1NNNDkKQXdFSEEwSUFCRmxpc1hNbVRqbFFZV29FNTBFUTloR09LZnJMRFFPMFlmYWZHZ3Uyd080cWNsNnM1cXBJWUhaYgpYdnBmcGdIcjVGOUhOOWRMS0FWelduZnErbU9mS0JxalNEQkdNQTRHQTFVZER3RUIvd1FFQXdJRm9EQVRCZ05WCkhTVUVEREFLQmdnckJnRUZCUWNEQWpBZkJnTlZIU01FR0RBV2dCUnFuVk5QV1FkQUxYb25HbDZYa3RldUlwNGIKU1RBS0JnZ3Foa2pPUFFRREFnTkpBREJHQWlFQXhoV2Y5NjBPRHhQWlpIUHI5dkdFa0FYQldzWmUrVGRWL29NNApSeElNZ0FJRFFVWHRTY0VVTUg5NUNmU3YycUoxRzNiQ1Z4Y3UwRHI1cUUrcWc4UkRjQkU9Ci0tLS0tRU5EIENFUlRJRklDQVRFLS0tLS0KLS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUJkVENDQVJ5Z0F3SUJBZ0lCQURBS0JnZ3Foa2pPUFFRREFqQWpNU0V3SHdZRFZRUUREQmhyTTNNdFkyeHAKWlc1MExXTmhRREUzTXpZMU5EQTRNemN3SGhjTk1qWXdNVEF4TURFd05qRTNXaGNOTXpZd01EQTVNREV3TmpFMwpXakFqTVNFd0h3WURWUVFEREJock0zTXRZMnhwWlc1MExXTmhRREUzTXpZMU5EQTRNemN3V1RBVEJnY3Foa2pPClBRSUJCZ2dxaGtqT1BRTUJCd05DQUFTNlhOdERjUStNUlZsZ0lVY1piZEw3UkJOckJVMEpTT3pTZjdFd1p0bnUKYktQYmV4QWEyaFEvL1FoeE56cit4S2pUWFd0Yk5xU2xIL1JIT1dKRjNyS1hvMEl3UURBT0JnTlZIUThCQWY4RQpCQU1DQXFRd0R3WURWUjBUQVFIL0JBVXdBd0VCL3pBZEJnTlZIUTRFRmdRVWFwMVRUMWtIUUMxNkp4cGVsNUxYCnJpS2VHMGt3Q2dZSUtvWkl6ajBFQXdJRFJ3QXdSQUlnZXhWNU5tSU9YS3k4aU5rUWZsWHlXaHdVRCtxVU1Dc2oKaUg5aG9GRnZQMGdDSUNsMFdBMHZBK01XQ3U1MEs1cDFKQ2poaVFvczVqYlBqdEFLM2EvWnJ2YXMKLS0tLS1FTkQgQ0VSVElGSUNBVEUtLS0tLQo=
    client-key-data: LS0tLS1CRUdJTiBFQyBQUklWQVRFIEtFWS0tLS0tCk1IY0NBUUVFSUdJOW5xNThWaFF3UUQ5MXBkOUZlOEkzU2VsQmFVWXhEZ21BdC9BL2J4YU5vQW9HQ0NxR1NNNDkKQXdFSG9VUURRZ0FFV1dLeGN5Wk9PVkJoYWdUblFSRDJFWTRwK3NzTkE3Umg5cDhhQzdiWTdpcHlYcXptcWtoZwpkbHRlK2wrbUFldmtYMGMzMTBzb0JYTmFkK3I2WTU4b0dnPT0KLS0tLS1FTkQgRUMgUFJJVkFURSBLRVktLS0tLQo=',
    'v1.27.8',
    'kubeconfig',
    'failed',
    'unhealthy',
    0,
    1,
    '模拟不可访问的集群，用于测试失败场景',
    NOW(),
    NOW()
);

-- 3. 生产集群-示例
INSERT INTO `k8s_clusters` (
    `name`,
    `description`,
    `api_server`,
    `kube_config`,
    `version`,
    `import_method`,
    `import_status`,
    `cluster_status`,
    `status`,
    `dept_id`,
    `remark`,
    `created_at`,
    `updated_at`
) VALUES (
    '生产集群-示例',
    '生产环境K8s集群示例（仅供展示）',
    'https://prod-k8s.example.com:6443',
    'apiVersion: v1
kind: Config
clusters:
- cluster:
    certificate-authority-data: LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUJkakNDQVIyZ0F3SUJBZ0lCQURBS0JnZ3Foa2pPUFFRREFqQWpNU0V3SHdZRFZRUUREQmhyTTNNdGMyVnkKZG1WeUxXTmhRREUzTXpZMU5EQTRNemN3SGhjTk1qWXdNVEF4TURFd05qRTNXaGNOTXpZd01EQTVNREV3TmpFMwpXakFqTVNFd0h3WURWUVFEREJock0zTXRjMlZ5ZG1WeUxXTmhRREUzTXpZMU5EQTRNemN3V1RBVEJnY3Foa2pPClBRSUJCZ2dxaGtqT1BRTUJCd05DQUFUTmVHbGtKZ3RXbVJOUlVycHU3M3lrcko4RlhVMlZFTGo2cS9TMHVBbkoKbFpDa1dXOVJJcTVMMXh2TktLV0lGUGprcW54UVRCZHRybWl1UE1DQzdXSU5vMEl3UURBT0JnTlZIUThCQWY4RQpCQU1DQXFRd0R3WURWUjBUQVFIL0JBVXdBd0VCL3pBZEJnTlZIUTRFRmdRVVo4SXRNTlRVT25qZ3M2MnFBdmhrCnFxcW5uRXd3Q2dZSUtvWkl6ajBFQXdJRFJ3QXdSQUlnV3pTQmRYUzJnVm1hZlZKSjRnRkJBZkNWOXphRFVzYkoKOUVhWDcwL0NsQkVDSUY4dVo2TWRLSjBQcFppbzE5VFpFU3VSYXNBV1FiSDRKZVdCZmh5L2ZzMgotLS0tLUVORCBDRVJUSUZJQ0FURS0tLS0t
    server: https://prod-k8s.example.com:6443
  name: prod-cluster
contexts:
- context:
    cluster: prod-cluster
    user: prod-admin
  name: prod-context
current-context: prod-context
users:
- name: prod-admin
  user:
    client-certificate-data: LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUJrakNDQVRlZ0F3SUJBZ0lJZFNHc1Q0RzFMclF3Q2dZSUtvWkl6ajBFQXdJd0l6RWhNQjhHQTFVRUF3d1kKYXpOekxXTnNhV1Z1ZEMxallVQXhOek0yTlRRd09ETTNNQTR4RGpBTUJnTlZCQU1NQldGa2JXbHVNQjRYRFRJMgpNREV3TVRBeE1EWTVOME1YRFRJM01ERXdNVEF4TURZNU4xb3dNREVYTUJVR0ExVUVDaE1PYzNsemRHVnRPbTFoCmMzUmxjbk14RlRBVEJnTlZCQU1UREhONWMzUmxiVHBoWkdsdGFUQlpNQk1HQnlxR1NNNDlBZ0VHQ0NxR1NNNDkKQXdFSEEwSUFCRmxpc1hNbVRqbFFZV29FNTBFUTloR09LZnJMRFFPMFlmYWZHZ3Uyd080cWNsNnM1cXBJWUhaYgpYdnBmcGdIcjVGOUhOOWRMS0FWelduZnErbU9mS0JxalNEQkdNQTRHQTFVZER3RUIvd1FFQXdJRm9EQVRCZ05WCkhTVUVEREFLQmdnckJnRUZCUWNEQWpBZkJnTlZIU01FR0RBV2dCUnFuVk5QV1FkQUxYb25HbDZYa3RldUlwNGIKU1RBS0JnZ3Foa2pPUFFRREFnTkpBREJHQWlFQXhoV2Y5NjBPRHhQWlpIUHI5dkdFa0FYQldzWmUrVGRWL29NNApSeElNZ0FJRFFVWHRTY0VVTUg5NUNmU3YycUoxRzNiQ1Z4Y3UwRHI1cUUrcWc4UkRjQkU9Ci0tLS0tRU5EIENFUlRJRklDQVRFLS0tLS0KLS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUJkVENDQVJ5Z0F3SUJBZ0lCQURBS0JnZ3Foa2pPUFFRREFqQWpNU0V3SHdZRFZRUUREQmhyTTNNdFkyeHAKWlc1MExXTmhRREUzTXpZMU5EQTRNemN3SGhjTk1qWXdNVEF4TURFd05qRTNXaGNOTXpZd01EQTVNREV3TmpFMwpXakFqTVNFd0h3WURWUVFEREJock0zTXRZMnhwWlc1MExXTmhRREUzTXpZMU5EQTRNemN3V1RBVEJnY3Foa2pPClBRSUJCZ2dxaGtqT1BRTUJCd05DQUFTNlhOdERjUStNUlZsZ0lVY1piZEw3UkJOckJVMEpTT3pTZjdFd1p0bnUKYktQYmV4QWEyaFEvL1FoeE56cit4S2pUWFd0Yk5xU2xIL1JIT1dKRjNyS1hvMEl3UURBT0JnTlZIUThCQWY4RQpCQU1DQXFRd0R3WURWUjBUQVFIL0JBVXdBd0VCL3pBZEJnTlZIUTRFRmdRVWFwMVRUMWtIUUMxNkp4cGVsNUxYCnJpS2VHMGt3Q2dZSUtvWkl6ajBFQXdJRFJ3QXdSQUlnZXhWNU5tSU9YS3k4aU5rUWZsWHlXaHdVRCtxVU1Dc2oKaUg5aG9GRnZQMGdDSUNsMFdBMHZBK01XQ3U1MEs1cDFKQ2poaVFvczVqYlBqdEFLM2EvWnJ2YXMKLS0tLS1FTkQgQ0VSVElGSUNBVEUtLS0tLQo=
    client-key-data: LS0tLS1CRUdJTiBFQyBQUklWQVRFIEtFWS0tLS0tCk1IY0NBUUVFSUdJOW5xNThWaFF3UUQ5MXBkOUZlOEkzU2VsQmFVWXhEZ21BdC9BL2J4YU5vQW9HQ0NxR1NNNDkKQXdFSG9VUURRZ0FFV1dLeGN5Wk9PVkJoYWdUblFSRDJFWTRwK3NzTkE3Umg5cDhhQzdiWTdpcHlYcXptcWtoZwpkbHRlK2wrbUFldmtYMGMzMTBzb0JYTmFkK3I2WTU4b0dnPT0KLS0tLS1FTkQgRUMgUFJJVkFURSBLRVktLS0tLQo=',
    'v1.29.0',
    'kubeconfig',
    'success',
    'healthy',
    1,
    1,
    '示例集群，展示完整的字段信息',
    NOW(),
    NOW()
);

-- 4. 待导入集群
INSERT INTO `k8s_clusters` (
    `name`,
    `description`,
    `api_server`,
    `kube_config`,
    `version`,
    `import_method`,
    `import_status`,
    `cluster_status`,
    `status`,
    `dept_id`,
    `remark`,
    `created_at`,
    `updated_at`
) VALUES (
    '待导入集群',
    '正在导入中的集群',
    'https://pending-k8s.example.com:6443',
    'apiVersion: v1
kind: Config
clusters:
- cluster:
    certificate-authority-data: LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUJkakNDQVIyZ0F3SUJBZ0lCQURBS0JnZ3Foa2pPUFFRREFqQWpNU0V3SHdZRFZRUUREQmhyTTNNdGMyVnkKZG1WeUxXTmhRREUzTXpZMU5EQTRNemN3SGhjTk1qWXdNVEF4TURFd05qRTNXaGNOTXpZd01EQTVNREV3TmpFMwpXakFqTVNFd0h3WURWUVFEREJock0zTXRjMlZ5ZG1WeUxXTmhRREUzTXpZMU5EQTRNemN3V1RBVEJnY3Foa2pPClBRSUJCZ2dxaGtqT1BRTUJCd05DQUFUTmVHbGtKZ3RXbVJOUlVycHU3M3lrcko4RlhVMlZFTGo2cS9TMHVBbkoKbFpDa1dXOVJJcTVMMXh2TktLV0lGUGprcW54UVRCZHRybWl1UE1DQzdXSU5vMEl3UURBT0JnTlZIUThCQWY4RQpCQU1DQXFRd0R3WURWUjBUQVFIL0JBVXdBd0VCL3pBZEJnTlZIUTRFRmdRVVo4SXRNTlRVT25qZ3M2MnFBdmhrCnFxcW5uRXd3Q2dZSUtvWkl6ajBFQXdJRFJ3QXdSQUlnV3pTQmRYUzJnVm1hZlZKSjRnRkJBZkNWOXphRFVzYkoKOUVhWDcwL0NsQkVDSUY4dVo2TWRLSjBQcFppbzE5VFpFU3VSYXNBV1FiSDRKZVdCZmh5L2ZzMgotLS0tLUVORCBDRVJUSUZJQ0FURS0tLS0t
    server: https://pending-k8s.example.com:6443
  name: pending-cluster
contexts:
- context:
    cluster: pending-cluster
    user: pending-user
  name: pending-context
current-context: pending-context
users:
- name: pending-user
  user:
    client-certificate-data: LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUJrakNDQVRlZ0F3SUJBZ0lJZFNHc1Q0RzFMclF3Q2dZSUtvWkl6ajBFQXdJd0l6RWhNQjhHQTFVRUF3d1kKYXpOekxXTnNhV1Z1ZEMxallVQXhOek0yTlRRd09ETTNNQTR4RGpBTUJnTlZCQU1NQldGa2JXbHVNQjRYRFRJMgpNREV3TVRBeE1EWTVOME1YRFRJM01ERXdNVEF4TURZNU4xb3dNREVYTUJVR0ExVUVDaE1PYzNsemRHVnRPbTFoCmMzUmxjbk14RlRBVEJnTlZCQU1UREhONWMzUmxiVHBoWkdsdGFUQlpNQk1HQnlxR1NNNDlBZ0VHQ0NxR1NNNDkKQXdFSEEwSUFCRmxpc1hNbVRqbFFZV29FNTBFUTloR09LZnJMRFFPMFlmYWZHZ3Uyd080cWNsNnM1cXBJWUhaYgpYdnBmcGdIcjVGOUhOOWRMS0FWelduZnErbU9mS0JxalNEQkdNQTRHQTFVZER3RUIvd1FFQXdJRm9EQVRCZ05WCkhTVUVEREFLQmdnckJnRUZCUWNEQWpBZkJnTlZIU01FR0RBV2dCUnFuVk5QV1FkQUxYb25HbDZYa3RldUlwNGIKU1RBS0JnZ3Foa2pPUFFRREFnTkpBREJHQWlFQXhoV2Y5NjBPRHhQWlpIUHI5dkdFa0FYQldzWmUrVGRWL29NNApSeElNZ0FJRFFVWHRTY0VVTUg5NUNmU3YycUoxRzNiQ1Z4Y3UwRHI1cUUrcWc4UkRjQkU9Ci0tLS0tRU5EIENFUlRJRklDQVRFLS0tLS0KLS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUJkVENDQVJ5Z0F3SUJBZ0lCQURBS0JnZ3Foa2pPUFFRREFqQWpNU0V3SHdZRFZRUUREQmhyTTNNdFkyeHAKWlc1MExXTmhRREUzTXpZMU5EQTRNemN3SGhjTk1qWXdNVEF4TURFd05qRTNXaGNOTXpZd01EQTVNREV3TmpFMwpXakFqTVNFd0h3WURWUVFEREJock0zTXRZMnhwWlc1MExXTmhRREUzTXpZMU5EQTRNemN3V1RBVEJnY3Foa2pPClBRSUJCZ2dxaGtqT1BRTUJCd05DQUFTNlhOdERjUStNUlZsZ0lVY1piZEw3UkJOckJVMEpTT3pTZjdFd1p0bnUKYktQYmV4QWEyaFEvL1FoeE56cit4S2pUWFd0Yk5xU2xIL1JIT1dKRjNyS1hvMEl3UURBT0JnTlZIUThCQWY4RQpCQU1DQXFRd0R3WURWUjBUQVFIL0JBVXdBd0VCL3pBZEJnTlZIUTRFRmdRVWFwMVRUMWtIUUMxNkp4cGVsNUxYCnJpS2VHMGt3Q2dZSUtvWkl6ajBFQXdJRFJ3QXdSQUlnZXhWNU5tSU9YS3k4aU5rUWZsWHlXaHdVRCtxVU1Dc2oKaUg5aG9GRnZQMGdDSUNsMFdBMHZBK01XQ3U1MEs1cDFKQ2poaVFvczVqYlBqdEFLM2EvWnJ2YXMKLS0tLS1FTkQgQ0VSVElGSUNBVEUtLS0tLQo=
    client-key-data: LS0tLS1CRUdJTiBFQyBQUklWQVRFIEtFWS0tLS0tCk1IY0NBUUVFSUdJOW5xNThWaFF3UUQ5MXBkOUZlOEkzU2VsQmFVWXhEZ21BdC9BL2J4YU5vQW9HQ0NxR1NNNDkKQXdFSG9VUURRZ0FFV1dLeGN5Wk9PVkJoYWdUblFSRDJFWTRwK3NzTkE3Umg5cDhhQzdiWTdpcHlYcXptcWtoZwpkbHRlK2wrbUFldmtYMGMzMTBzb0JYTmFkK3I2WTU4b0dnPT0KLS0tLS1FTkQgRUMgUFJJVkFURSBLRVktLS0tLQo=',
    '',
    'kubeconfig',
    'importing',
    'unknown',
    1,
    1,
    '模拟导入过程中的集群状态',
    NOW(),
    NOW()
);

-- ========================================
-- 3. 插入集群访问权限数据 (k8s_cluster_accesses)
-- ========================================
-- 假设角色表中存在：role_id=1 是超级管理员，role_id=2 是普通用户

INSERT INTO `k8s_cluster_accesses` (
    `cluster_id`,
    `role_id`,
    `access_type`,
    `namespaces`,
    `created_at`,
    `updated_at`
) VALUES
-- 超级管理员对本地开发集群有admin权限，可访问所有namespace
(1, 1, 'admin', '', NOW(), NOW()),

-- 普通用户对本地开发集群有readonly权限，只能访问default和dev namespace
(1, 2, 'readonly', '["default","dev","test"]', NOW(), NOW()),

-- 超级管理员对生产集群有admin权限
(3, 1, 'admin', '', NOW(), NOW()),

-- 普通用户对生产集群有readonly权限，只能访问指定namespace
(3, 2, 'readonly', '["default","production"]', NOW(), NOW());

-- ========================================
-- 4. 插入命名空间记录数据 (k8s_namespaces)
-- ========================================

INSERT INTO `k8s_namespaces` (
    `cluster_id`,
    `name`,
    `labels`,
    `annotations`,
    `status`,
    `created_at`,
    `updated_at`
) VALUES
-- 本地开发集群的namespace
(1, 'default', '{"kubernetes.io/metadata.name":"default"}', '{}', 'Active', NOW(), NOW()),
(1, 'kube-system', '{"kubernetes.io/metadata.name":"kube-system"}', '{}', 'Active', NOW(), NOW()),
(1, 'dev', '{"env":"development","team":"backend"}', '{"description":"开发环境"}', 'Active', NOW(), NOW()),
(1, 'test', '{"env":"test","team":"qa"}', '{"description":"测试环境"}', 'Active', NOW(), NOW()),

-- 生产集群的namespace
(3, 'default', '{"kubernetes.io/metadata.name":"default"}', '{}', 'Active', NOW(), NOW()),
(3, 'kube-system', '{"kubernetes.io/metadata.name":"kube-system"}', '{}', 'Active', NOW(), NOW()),
(3, 'production', '{"env":"production","team":"ops"}', '{"description":"生产环境","monitoring":"enabled"}', 'Active', NOW(), NOW()),
(3, 'staging', '{"env":"staging","team":"devops"}', '{"description":"预发布环境"}', 'Active', NOW(), NOW());

-- ========================================
-- 5. 插入操作日志数据 (k8s_operation_logs)
-- ========================================
-- 假设用户表中存在：user_id=1 是admin用户

INSERT INTO `k8s_operation_logs` (
    `cluster_id`,
    `user_id`,
    `username`,
    `operation`,
    `resource`,
    `namespace`,
    `name`,
    `result`,
    `message`,
    `ip`,
    `created_at`
) VALUES
-- 成功的操作日志
(1, 1, 'admin', 'create', 'deployment', 'dev', 'nginx-deployment', 'success', '成功创建Deployment', '127.0.0.1', NOW() - INTERVAL 2 DAY),
(1, 1, 'admin', 'create', 'service', 'dev', 'nginx-service', 'success', '成功创建Service', '127.0.0.1', NOW() - INTERVAL 2 DAY),
(1, 1, 'admin', 'update', 'deployment', 'dev', 'nginx-deployment', 'success', '更新Deployment副本数为3', '127.0.0.1', NOW() - INTERVAL 1 DAY),
(1, 1, 'admin', 'list', 'pod', 'dev', '', 'success', '查看Pod列表', '127.0.0.1', NOW() - INTERVAL 1 DAY),
(1, 1, 'admin', 'get', 'pod', 'dev', 'nginx-deployment-7d6c8f9b-xyz', 'success', '查看Pod详情', '127.0.0.1', NOW() - INTERVAL 12 HOUR),

-- 失败的操作日志
(1, 1, 'admin', 'delete', 'deployment', 'production', 'critical-app', 'failed', '权限不足：无法删除生产环境资源', '127.0.0.1', NOW() - INTERVAL 6 HOUR),
(2, 1, 'admin', 'create', 'deployment', 'default', 'test-app', 'failed', '集群连接失败', '127.0.0.1', NOW() - INTERVAL 3 HOUR),

-- 生产集群操作日志
(3, 1, 'admin', 'create', 'deployment', 'production', 'api-server', 'success', '部署API服务', '10.0.1.100', NOW() - INTERVAL 5 DAY),
(3, 1, 'admin', 'create', 'service', 'production', 'api-server-svc', 'success', '创建API服务的Service', '10.0.1.100', NOW() - INTERVAL 5 DAY),
(3, 1, 'admin', 'create', 'ingress', 'production', 'api-ingress', 'success', '配置Ingress路由', '10.0.1.100', NOW() - INTERVAL 5 DAY),
(3, 1, 'admin', 'update', 'deployment', 'production', 'api-server', 'success', '滚动更新到v1.2.0版本', '10.0.1.100', NOW() - INTERVAL 3 DAY),
(3, 1, 'admin', 'create', 'configmap', 'production', 'app-config', 'success', '创建应用配置', '10.0.1.100', NOW() - INTERVAL 2 DAY),
(3, 1, 'admin', 'create', 'secret', 'production', 'db-credentials', 'success', '创建数据库密钥', '10.0.1.100', NOW() - INTERVAL 2 DAY),
(3, 1, 'admin', 'list', 'pod', 'production', '', 'success', '查看生产环境Pod状态', '10.0.1.100', NOW() - INTERVAL 1 HOUR),

-- 最近的操作
(1, 1, 'admin', 'exec', 'pod', 'dev', 'nginx-deployment-7d6c8f9b-xyz', 'success', '进入容器执行命令', '127.0.0.1', NOW() - INTERVAL 30 MINUTE),
(3, 1, 'admin', 'get', 'deployment', 'production', 'api-server', 'success', '查看Deployment状态', '10.0.1.100', NOW() - INTERVAL 10 MINUTE);

-- 恢复外键检查
SET FOREIGN_KEY_CHECKS = 1;

-- ========================================
-- 6. 数据验证和统计
-- ========================================

-- 查询集群列表
SELECT 
    '集群信息' as table_name,
    id,
    name,
    version,
    import_method,
    import_status,
    cluster_status,
    status,
    created_at
FROM k8s_clusters
ORDER BY id;

-- 查询集群访问权限
SELECT 
    '集群访问权限' as table_name,
    id,
    cluster_id,
    role_id,
    access_type,
    namespaces
FROM k8s_cluster_accesses
ORDER BY cluster_id, role_id;

-- 查询命名空间列表
SELECT 
    '命名空间' as table_name,
    id,
    cluster_id,
    name,
    status,
    created_at
FROM k8s_namespaces
ORDER BY cluster_id, name;

-- 查询最近的操作日志（最新10条）
SELECT 
    '操作日志（最新10条）' as table_name,
    id,
    cluster_id,
    username,
    operation,
    resource,
    namespace,
    name,
    result,
    created_at
FROM k8s_operation_logs
ORDER BY created_at DESC
LIMIT 10;

-- ========================================
-- 7. 统计信息
-- ========================================

-- 集群统计
SELECT 
    '=== 集群统计 ===' as summary,
    COUNT(*) as total_clusters,
    SUM(CASE WHEN import_status = 'success' THEN 1 ELSE 0 END) as success_count,
    SUM(CASE WHEN import_status = 'failed' THEN 1 ELSE 0 END) as failed_count,
    SUM(CASE WHEN import_status = 'importing' THEN 1 ELSE 0 END) as importing_count,
    SUM(CASE WHEN cluster_status = 'healthy' THEN 1 ELSE 0 END) as healthy_count,
    SUM(CASE WHEN cluster_status = 'unhealthy' THEN 1 ELSE 0 END) as unhealthy_count,
    SUM(CASE WHEN status = 1 THEN 1 ELSE 0 END) as enabled_count
FROM k8s_clusters;

-- 访问权限统计
SELECT 
    '=== 访问权限统计 ===' as summary,
    COUNT(*) as total_accesses,
    SUM(CASE WHEN access_type = 'admin' THEN 1 ELSE 0 END) as admin_count,
    SUM(CASE WHEN access_type = 'readonly' THEN 1 ELSE 0 END) as readonly_count
FROM k8s_cluster_accesses;

-- 命名空间统计
SELECT 
    '=== 命名空间统计 ===' as summary,
    COUNT(*) as total_namespaces,
    COUNT(DISTINCT cluster_id) as clusters_with_namespaces,
    SUM(CASE WHEN status = 'Active' THEN 1 ELSE 0 END) as active_count
FROM k8s_namespaces;

-- 操作日志统计
SELECT 
    '=== 操作日志统计 ===' as summary,
    COUNT(*) as total_operations,
    SUM(CASE WHEN result = 'success' THEN 1 ELSE 0 END) as success_operations,
    SUM(CASE WHEN result = 'failed' THEN 1 ELSE 0 END) as failed_operations,
    COUNT(DISTINCT resource) as resource_types
FROM k8s_operation_logs;

-- 按资源类型统计操作
SELECT 
    '=== 按资源类型统计 ===' as summary,
    resource,
    COUNT(*) as operation_count,
    SUM(CASE WHEN result = 'success' THEN 1 ELSE 0 END) as success_count
FROM k8s_operation_logs
GROUP BY resource
ORDER BY operation_count DESC;

-- 按操作类型统计
SELECT 
    '=== 按操作类型统计 ===' as summary,
    operation,
    COUNT(*) as count,
    SUM(CASE WHEN result = 'success' THEN 1 ELSE 0 END) as success_count
FROM k8s_operation_logs
GROUP BY operation
ORDER BY count DESC;

-- 数据导入完成提示
SELECT '
========================================
✅ K8s测试数据导入完成！
========================================
数据概览：
- 4 个集群
- 4 条访问权限记录
- 8 个命名空间
- 16 条操作日志

测试数据说明：
1. 集群ID=1: 本地开发集群（成功导入，状态未知）
2. 集群ID=2: 测试集群-1.27（导入失败，不健康）
3. 集群ID=3: 生产集群示例（成功导入，健康）
4. 集群ID=4: 待导入集群（导入中）

现在可以通过API测试：
curl -X GET "http://127.0.0.1:8000/api/k8s/clusters?page=1&pageSize=10" \
  -H "Authorization: Bearer YOUR_TOKEN"
========================================
' as message;
