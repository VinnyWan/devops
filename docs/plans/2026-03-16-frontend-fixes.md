# 前端页面修复实施计划

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**目标:** 修复工作负载和节点管理页面的5个UI/UX问题

**架构:** 前端Vue3组件修改，添加自定义YAML终端组件，修复分页和命名空间功能

**技术栈:** Vue 3, Naive UI, TypeScript, js-yaml

---

## 问题清单

1. ✅ 删除命名空间侧边栏（如果存在）
2. ✅ 修复分页功能（支持真实分页、手动输入页码、右对齐）
3. ✅ 修复节点管理aria-hidden错误
4. ✅ 修复工作负载命名空间下拉列表
5. ✅ 创建终端风格YAML编辑器

---

## Task 1: 检查并删除命名空间侧边栏

**目标:** 确认是否存在独立的命名空间侧边栏组件并删除

**Files:**
- Check: `frontend/src/views/workload/WorkloadList.vue`
- Check: `frontend/src/components/` (查找命名空间相关组件)

**Step 1: 检查WorkloadList.vue是否有侧边栏**

检查当前WorkloadList.vue的template结构，确认是否有侧边栏布局。

**Step 2: 确认修改范围**

如果没有侧边栏，此任务完成。如果有，记录需要删除的代码行。

**Step 3: 提交确认**

```bash
# 如果没有需要删除的内容
git commit --allow-empty -m "chore: 确认无命名空间侧边栏需要删除"
```

---

## Task 2: 修复工作负载页面分页功能

**目标:** 添加分页事件处理、手动输入页码、右对齐布局

**Files:**
- Modify: `frontend/src/views/workload/WorkloadList.vue:371-377`
- Modify: `frontend/src/views/workload/WorkloadList.vue:145-178` (fetchData函数)

**Step 1: 添加分页处理函数**

在script部分添加：

```typescript
function handlePageSizeChange(newSize: number) {
  pageSize.value = newSize
  page.value = 1
  fetchData()
}
```

位置：在 `confirmDelete` 函数之后（约325行）

**Step 2: 修改分页组件**

替换当前的NPagination（371-377行）：

```vue
<div style="display: flex; justify-content: flex-end; margin-top: 16px">
  <NPagination
    v-model:page="page"
    v-model:page-size="pageSize"
    :item-count="total"
    :page-sizes="[10, 20, 50, 100]"
    show-size-picker
    show-quick-jumper
    @update:page="fetchData"
    @update:page-size="handlePageSizeChange"
  />
</div>
```

**Step 3: 测试分页功能**

1. 启动开发服务器
2. 访问工作负载页面
3. 测试翻页、改变每页数量、手动输入页码
4. 验证每次操作都会调用fetchData

**Step 4: 提交**

```bash
git add frontend/src/views/workload/WorkloadList.vue
git commit -m "fix: 修复工作负载分页功能，支持手动输入页码和右对齐"
```

---

## Task 3: 修复节点管理分页功能

**目标:** 同样修复节点管理页面的分页

**Files:**
- Modify: `frontend/src/views/node/NodeList.vue:277-285`

**Step 1: 修改分页组件**

替换当前的NPagination（277-285行）：

```vue
<div style="display: flex; justify-content: flex-end; margin-top: 16px">
  <NPagination
    v-model:page="page"
    v-model:page-size="pageSize"
    :item-count="total"
    :page-sizes="[10, 20, 50, 100]"
    show-size-picker
    show-quick-jumper
    @update:page="handlePageChange"
    @update:page-size="handlePageSizeChange"
  />
</div>
```

**Step 2: 测试分页功能**

访问节点管理页面，测试分页操作

**Step 3: 提交**

```bash
git add frontend/src/views/node/NodeList.vue
git commit -m "fix: 修复节点管理分页功能，支持手动输入页码和右对齐"
```

---

## Task 4: 修复节点管理aria-hidden错误

**目标:** 解决NModal的focus trap导致的aria-hidden警告

**Files:**
- Modify: `frontend/src/views/node/NodeList.vue:289`

**Step 1: 添加trap-focus属性**

修改NModal配置：

```vue
<NModal
  v-model:show="showDetailModal"
  preset="card"
  title="节点详情"
  style="width: 900px; margin-top: 50px"
  :trap-focus="false"
  :block-scroll="true"
>
```

**Step 2: 测试**

1. 打开节点管理页面
2. 点击"详情"按钮
3. 检查浏览器控制台是否还有aria-hidden警告
4. 验证弹窗功能正常

**Step 3: 提交**

```bash
git add frontend/src/views/node/NodeList.vue
git commit -m "fix: 修复节点详情弹窗aria-hidden错误"
```

---

## Task 5: 增强命名空间下拉列表功能

**目标:** 添加错误处理和调试信息，确保命名空间列表正确加载

**Files:**
- Modify: `frontend/src/views/workload/WorkloadList.vue:131-143`

**Step 1: 增强fetchNamespaces函数**

替换现有函数（131-143行）：

```typescript
async function fetchNamespaces() {
  if (!currentClusterId.value) {
    console.warn('未选择集群，无法获取命名空间')
    return
  }
  try {
    const res = await k8sK8sNamespacesListPost({ clusterId: currentClusterId.value })
    const namespaces = res.data.data || []
    console.log('获取到的命名空间列表:', namespaces)
    namespaceOptions.value = [
      { label: '所有命名空间', value: 'all' },
      ...namespaces.map((ns: any) => ({ label: ns.name, value: ns.name }))
    ]
    if (namespaceOptions.value.length === 1) {
      message.warning('当前集群没有可用的命名空间')
    }
  } catch (error: any) {
    console.error('获取命名空间失败:', error)
    message.error('获取命名空间失败: ' + error.message)
  }
}
```

**Step 2: 测试**

1. 打开工作负载页面
2. 选择集群
3. 检查命名空间下拉列表是否显示多个选项
4. 查看控制台日志确认数据获取

**Step 3: 提交**

```bash
git add frontend/src/views/workload/WorkloadList.vue
git commit -m "fix: 增强命名空间下拉列表错误处理和调试信息"
```

---

## Task 6: 安装YAML校验依赖

**目标:** 安装js-yaml用于YAML格式校验

**Files:**
- Modify: `frontend/package.json`

**Step 1: 安装依赖**

```bash
cd frontend
npm install js-yaml
npm install -D @types/js-yaml
```

**Step 2: 提交**

```bash
git add frontend/package.json frontend/package-lock.json
git commit -m "chore: 添加js-yaml依赖用于YAML校验"
```

---

## Task 7: 创建YamlTerminalModal组件

**目标:** 创建终端风格的YAML编辑器组件

**Files:**
- Create: `frontend/src/components/YamlTerminalModal.vue`

**Component Requirements:**
- 白边黑底终端风格
- 左侧行号显示
- 顶部标题栏显示"编辑YAML - 资源名称"
- 校验、复制、保存、关闭按钮
- 支持滚动
- 可编辑textarea

**Step 1: 创建组件文件**

创建 `frontend/src/components/YamlTerminalModal.vue`（见Task 8详细实现）

---

## Task 8: 实现YamlTerminalModal组件代码

**目标:** 编写完整的组件实现

**Files:**
- Write: `frontend/src/components/YamlTerminalModal.vue`

**Implementation:** 组件包含script、template、style三部分

**Step 1: 提交**

```bash
git add frontend/src/components/YamlTerminalModal.vue
git commit -m "feat: 创建终端风格YAML编辑器组件"
```

---

## Task 9: 在WorkloadList中集成YamlTerminalModal

**目标:** 替换现有的YAML弹窗为新的终端风格组件

**Files:**
- Modify: `frontend/src/views/workload/WorkloadList.vue`

**Step 1: 导入组件**

在script setup顶部添加（约第4行）：

```typescript
import YamlTerminalModal from '@/components/YamlTerminalModal.vue'
```

**Step 2: 替换YAML Modal**

删除现有的YAML Modal（399-409行），替换为：

```vue
<YamlTerminalModal
  v-model:show="showYamlModal"
  v-model:content="yamlContent"
  :title="`${resourceType} / ${currentWorkload?.name || ''}`"
  :loading="yamlLoading"
  @save="saveYaml"
/>
```

**Step 3: 测试**

1. 打开工作负载页面
2. 点击任意资源的"YAML"按钮
3. 验证终端风格显示
4. 测试复制、校验、保存功能

**Step 4: 提交**

```bash
git add frontend/src/views/workload/WorkloadList.vue
git commit -m "feat: 集成终端风格YAML编辑器到工作负载页面"
```

---

## 验收标准

### 1. 命名空间侧边栏
- [ ] 确认无独立的命名空间侧边栏或已删除

### 2. 分页功能
- [ ] 工作负载页面分页右对齐
- [ ] 支持手动输入页码（show-quick-jumper）
- [ ] 翻页时正确调用fetchData
- [ ] 改变每页数量时重置到第1页
- [ ] 节点管理页面分页同样修复

### 3. aria-hidden错误
- [ ] 打开节点详情弹窗无控制台警告
- [ ] 弹窗功能正常

### 4. 命名空间下拉列表
- [ ] 选择集群后显示该集群的所有命名空间
- [ ] 默认选项"所有命名空间"存在
- [ ] 选择特定命名空间后只显示该命名空间的资源
- [ ] 错误时显示友好提示

### 5. YAML终端编辑器
- [ ] 白边黑底终端风格
- [ ] 左侧显示行号
- [ ] 标题显示"编辑YAML - 资源类型/资源名称"
- [ ] 复制按钮功能正常
- [ ] 校验按钮能检测YAML语法错误
- [ ] 保存按钮能更新资源
- [ ] 支持上下滚动
- [ ] 编辑器可编辑

---

## 实施顺序

1. Task 1: 检查命名空间侧边栏（5分钟）
2. Task 2: 修复工作负载分页（15分钟）
3. Task 3: 修复节点管理分页（10分钟）
4. Task 4: 修复aria-hidden错误（5分钟）
5. Task 5: 增强命名空间功能（10分钟）
6. Task 6: 安装依赖（5分钟）
7. Task 7-8: 实现YamlTerminalModal（30分钟）
8. Task 9: 集成到WorkloadList（15分钟）
9. 测试和验收（20分钟）

**总计:** 约2小时

---

## 风险和注意事项

1. **js-yaml依赖**: 确保package.json正确更新
2. **样式兼容性**: 终端风格在不同浏览器中测试
3. **滚动同步**: 行号和代码区域滚动需要同步处理
4. **性能**: 大型YAML文件可能影响编辑器性能
5. **后端API**: 确保命名空间API返回正确格式的数据

---

## 后续优化（可选）

1. 添加YAML语法高亮
2. 使用Monaco Editor替代textarea获得更好体验
3. 添加YAML自动格式化功能
4. 支持YAML diff对比
5. 添加撤销/重做功能

