# 节点管理功能增强实现计划

## 概述
为节点管理页面添加5个核心功能：详情查看、污点管理、标签管理、禁止调度、驱逐节点。

## 后端API状态
✅ 所有API已实现，无需后端开发

## 实现方案
**方案1：单文件实现**（已选择）
- 所有功能在 NodeList.vue 中实现
- 使用 NModal 组件创建弹窗
- 代码量约500-600行

## 实现步骤

### 阶段1：添加操作列按钮
**文件：** `frontend/src/views/node/NodeList.vue`

**任务：**
1. 在columns中添加操作列
2. 使用NButton和NSpace创建5个操作按钮
3. 添加对应的点击事件处理函数

**预期结果：** 每行节点显示5个操作按钮

---

### 阶段2：实现节点详情弹窗
**API：** `k8sK8sNodeDetailPost`

**任务：**
1. 创建详情弹窗状态和数据变量
2. 实现fetchNodeDetail函数调用API
3. 显示CPU/内存使用率（百分比）
4. 显示Pod使用率
5. 实现Pod列表（分页）

**数据结构：**
```typescript
interface NodeDetail {
  cpuUsage: string
  cpuCapacity: string
  memoryUsage: string
  memoryCapacity: string
  podCount: number
  podCapacity: number
  pods: Pod[]
}
```

---

### 阶段3：实现污点管理弹窗
**API：** `k8sK8sNodeTaintsPost`

**任务：**
1. 显示当前节点污点列表
2. 添加污点表单（Key、Value、Effect下拉）
3. 删除污点功能
4. 提交更新到后端

**污点效果选项：**
- NoSchedule（禁止调度）
- NoExecute（禁止执行）
- PreferNoSchedule（尽量避免调度）

---

### 阶段4：实现标签管理弹窗
**API：** `k8sK8sNodeLabelsPost`

**任务：**
1. 显示当前节点标签列表
2. 添加标签表单（Key、Value）
3. 删除标签功能
4. 提交更新到后端

---

### 阶段5：实现禁止调度功能
**API：** `k8sK8sNodeCordonPost`

**任务：**
1. 检查节点当前调度状态
2. 点击按钮切换状态（Cordon/Uncordon）
3. 显示成功/失败消息

---

### 阶段6：实现驱逐节点功能
**API：** `k8sK8sNodeDrainPost`

**任务：**
1. 创建驱逐确认弹窗
2. 提供高级选项：
   - 强制驱逐（force）
   - 忽略DaemonSet（ignoreDaemonSets）
   - 删除本地数据（deleteLocalData）
   - 优雅期限（gracePeriodSeconds）
3. 提交驱逐请求

---

## 验收标准

### 功能验收
- [ ] 详情弹窗显示CPU/内存/Pod使用率
- [ ] 详情弹窗Pod列表支持分页
- [ ] 污点管理可添加/删除污点
- [ ] 标签管理可添加/删除标签
- [ ] 禁止调度按钮正常工作
- [ ] 驱逐节点提供高级选项

### 技术验收
- [ ] 所有API调用正确传参
- [ ] 错误处理完善
- [ ] 成功/失败消息提示
- [ ] 操作后刷新列表

## 风险与注意事项

1. **并发更新冲突**
   - 后端使用Retry机制处理
   - 前端显示友好错误提示

2. **驱逐节点风险**
   - 必须显示确认弹窗
   - 提示用户操作不可逆

3. **污点/标签格式验证**
   - Key必须符合K8s命名规范
   - Value可以为空

## 预估工作量
- 开发时间：2-3小时
- 测试时间：1小时
- 总计：3-4小时
