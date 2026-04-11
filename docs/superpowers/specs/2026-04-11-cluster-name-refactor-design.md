# 重构：clusterId → clusterName

**日期**: 2026-04-11
**状态**: 已批准

## 目标
将 K8s API 的集群标识参数从 `clusterId`（uint 数据库 ID）改为 `clusterName`（string 集群名称），使前后端传递更直观。

## 变更范围

### 后端（28 文件，~400 处）
- `api/common.go`: `getClusterIDFromQuery()` → `getClusterNameFromQuery()`
- 所有 Request struct: `ClusterID uint` → `ClusterName string`
- 所有 Service 方法: 接收 `clusterName string`，内部通过 name 查 DB 获取 config
- Swagger 注释同步更新

### 前端（6 文件，~33 处）
- `ClusterSelector.vue`: v-model 绑定 `cluster.name` 而非 `cluster.id`
- 所有 K8s 页面: 传参 `clusterName` 替代 `clusterId`

### 不变
- 数据库模型不变
- 路由路径不变
