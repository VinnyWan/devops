# K8s 工作负载和节点管理增强计划

**日期**: 2026-03-16
**状态**: 设计阶段
**优先级**: 高

## 概述

本计划包含两个主要部分：
1. **节点管理页面修复** - 修复详情弹窗乱码、调整布局
2. **工作负载页面增强** - 支持三种资源类型（Deployment/StatefulSet/DaemonSet），添加完整的操作功能

## 一、节点管理页面修复

### 1.1 问题分析

**当前问题**:
- 节点详情弹窗出现乱码（需要诊断具体原因）
- 弹窗位置太靠下
- 集群选择器和刷新按钮布局不合理

**修复目标**:
- 诊断并修复乱码问题
- 调整弹窗位置到页面上方
- 将集群选择器移到刷新按钮旁边

### 1.2 实现步骤

#### Step 1: 诊断乱码问题
- 检查 `/v1/k8s/node/detail` API 返回的数据格式
- 验证 CPU/内存使用率数据是否正确解析
- 检查是否是 K8s 单位格式问题（如 1000m, 1Gi）

#### Step 2: 修复布局问题
**文件**: `frontend/src/views/node/NodeList.vue`

修改内容：
```vue
<!-- 修改前 (line 263-266) -->
<template #header-extra>
  <ClusterSelector @update:value="fetchNodes" />
</template>

<!-- 修改后 -->
<template #header-extra>
  <NSpace>
    <ClusterSelector @update:value="fetchNodes" />
    <NButton type="primary" @click="fetchNodes">刷新</NButton>
  </NSpace>
</template>

<!-- 移除独立的刷新按钮 (line 268) -->
```

#### Step 3: 调整弹窗位置
```vue
<!-- 修改前 (line 287) -->
<NModal v-model:show="showDetailModal" preset="card" title="节点详情" style="width: 900px">

<!-- 修改后 -->
<NModal
  v-model:show="showDetailModal"
  preset="card"
  title="节点详情"
  style="width: 900px; margin-top: 50px"
>
```

#### Step 4: 修复乱码（根据诊断结果）
可能的修复方案：
- 如果是单位问题：添加单位转换函数（如 `1000m` → `1 Core`）
- 如果是编码问题：检查 API 响应的 Content-Type
- 如果是数据格式问题：调整数据解析逻辑

---

## 二、工作负载页面增强

### 2.1 需求分析

**功能需求**:
1. 删除独立的命名空间页面路由
2. 在工作负载页面添加命名空间下拉框
3. 支持三种资源类型切换：Deployment / StatefulSet / DaemonSet
4. 显示完整的资源信息列
5. 添加操作按钮：伸缩、重启、YAML编辑、删除

**数据需求**:
- 名称、标签、容器组数量、Request/Limits（总和）、镜像、创建时间、更新时间

### 2.2 后端 API 开发

#### Phase 1: 创建 StatefulSet API

**文件**: `backend/internal/modules/k8s/api/statefulset.go`

需要实现的函数：
```go
// 列表
func ListStatefulSets(c *gin.Context)

// 详情
func GetStatefulSetDetail(c *gin.Context)

// 获取 YAML
func GetStatefulSetYAML(c *gin.Context)

// 更新 YAML
func UpdateStatefulSetYAML(c *gin.Context)

// 伸缩
func ScaleStatefulSet(c *gin.Context)

// 重启
func RestartStatefulSet(c *gin.Context)

// 删除
func DeleteStatefulSet(c *gin.Context)
```

#### Phase 2: 创建 DaemonSet API

**文件**: `backend/internal/modules/k8s/api/daemonset.go`

需要实现的函数：
```go
// 列表
func ListDaemonSets(c *gin.Context)

// 详情
func GetDaemonSetDetail(c *gin.Context)

// 获取 YAML
func GetDaemonSetYAML(c *gin.Context)

// 更新 YAML
func UpdateDaemonSetYAML(c *gin.Context)

// 重启
func RestartDaemonSet(c *gin.Context)

// 删除
func DeleteDaemonSet(c *gin.Context)

// 注意: DaemonSet 不支持伸缩操作
```

#### Phase 3: 注册路由

**文件**: `backend/routers/v1/k8s.go`

在 `registerCluster` 函数中添加：
```go
// StatefulSet
g.GET("/statefulset/list", listPermission, api.ListStatefulSets)
g.GET("/statefulset/detail", listPermission, api.GetStatefulSetDetail)
g.GET("/statefulset/yaml", listPermission, api.GetStatefulSetYAML)
g.POST("/statefulset/yaml/update", updatePermission,
    middleware.SetAuditOperation("YAML更新StatefulSet"),
    api.UpdateStatefulSetYAML)
g.POST("/statefulset/restart", updatePermission,
    middleware.SetAuditOperation("重启StatefulSet"),
    api.RestartStatefulSet)
g.POST("/statefulset/scale", updatePermission,
    middleware.SetAuditOperation("扩缩容StatefulSet"),
    api.ScaleStatefulSet)
g.POST("/statefulset/delete", deletePermission,
    middleware.SetAuditOperation("删除StatefulSet"),
    api.DeleteStatefulSet)

// DaemonSet
g.GET("/daemonset/list", listPermission, api.ListDaemonSets)
g.GET("/daemonset/detail", listPermission, api.GetDaemonSetDetail)
g.GET("/daemonset/yaml", listPermission, api.GetDaemonSetYAML)
g.POST("/daemonset/yaml/update", updatePermission,
    middleware.SetAuditOperation("YAML更新DaemonSet"),
    api.UpdateDaemonSetYAML)
g.POST("/daemonset/restart", updatePermission,
    middleware.SetAuditOperation("重启DaemonSet"),
    api.RestartDaemonSet)
g.POST("/daemonset/delete", deletePermission,
    middleware.SetAuditOperation("删除DaemonSet"),
    api.DeleteDaemonSet)
```

#### Phase 4: 增强 Deployment API

**文件**: `backend/internal/modules/k8s/api/deployment.go`

修改 `ListDeployments` 函数，确保返回：
- labels (map[string]string)
- podCount (ready/desired)
- totalRequests (CPU/Memory)
- totalLimits (CPU/Memory)
- images ([]string - 所有容器镜像)
- creationTimestamp
- updateTimestamp

---

## 三、实施顺序

### Phase 1: 节点管理修复（1-2小时）
1. 诊断乱码问题
2. 修复布局和弹窗位置
3. 测试验证

### Phase 2: 后端 API 开发（4-6小时）
1. 创建 StatefulSet API（2小时）
2. 创建 DaemonSet API（2小时）
3. 增强 Deployment API（1小时）
4. 测试所有 API（1小时）

### Phase 3: 前端类型和 API（1小时）
1. 生成 TypeScript 类型定义
2. 创建 API 函数

### Phase 4: 前端页面重构（3-4小时）
1. 删除命名空间路由
2. 重构工作负载页面结构
3. 实现资源类型切换
4. 实现命名空间选择
5. 实现数据表格和列定义
6. 实现操作功能（伸缩、重启、YAML、删除）

### Phase 5: 测试和优化（2小时）
1. 功能测试
2. UI/UX 优化
3. 错误处理完善

**总计**: 约 11-15 小时

---

## 四、验收标准

### 4.1 节点管理
- [ ] 详情弹窗无乱码，数据显示正确
- [ ] 弹窗位置在页面上方
- [ ] 集群选择器和刷新按钮在同一行

### 4.2 工作负载
- [ ] 命名空间路由已删除
- [ ] 工作负载页面有命名空间下拉框
- [ ] 支持三种资源类型切换
- [ ] 表格显示所有必需列
- [ ] 伸缩功能正常（Deployment/StatefulSet）
- [ ] 重启功能正常（所有类型）
- [ ] YAML 编辑和保存功能正常
- [ ] 删除功能有二次确认且正常工作
- [ ] 默认显示所有命名空间的资源
- [ ] 选择特定命名空间后只显示该命名空间的资源
