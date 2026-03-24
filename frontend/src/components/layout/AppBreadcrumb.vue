<script setup lang="ts">
import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { NBreadcrumb, NBreadcrumbItem } from 'naive-ui'

const route = useRoute()
const router = useRouter()

// Breadcrumb path mapping
const breadcrumbMap: Record<string, { label: string; parent?: string }> = {
  '/dashboard': { label: '仪表盘' },
  '/asset': { label: '资产管理' },
  '/cluster': { label: '集群管理', parent: 'container-management' },
  '/node': { label: '节点管理', parent: 'container-management' },
  '/namespace': { label: '命名空间', parent: 'container-management' },
  '/workload': { label: '工作负载', parent: 'container-management' },
  '/network': { label: '网络管理', parent: 'container-management' },
  '/storage': { label: '存储管理', parent: 'container-management' },
  '/config': { label: '配置管理', parent: 'container-management' },
  '/ops/alert': { label: '告警中心' },
  '/ops/log': { label: '日志检索' },
  '/ops/monitor': { label: '监控配置' },
  '/ops/harbor': { label: 'Harbor管理' },
  '/ops/cicd': { label: 'CI/CD流水线' },
  '/ops/app': { label: '应用管理' },
  '/ops/audit': { label: '审计日志' },
  '/system/users': { label: '用户管理', parent: 'system-management' },
  '/system/departments': { label: '部门管理', parent: 'system-management' },
  '/system/roles': { label: '角色管理', parent: 'system-management' },
  '/system/permissions': { label: '权限管理', parent: 'system-management' },
}

const parentLabels: Record<string, string> = {
  'container-management': '容器管理',
  'system-management': '系统管理',
}

const breadcrumbs = computed(() => {
  const path = route.path
  const current = breadcrumbMap[path]
  if (!current) {
    return [{ label: '首页', path: '/dashboard' }]
  }

  const items = []

  // Add parent if exists
  if (current.parent) {
    items.push({
      label: parentLabels[current.parent] || current.parent,
      path: null,
    })
  }

  // Add current item
  items.push({
    label: current.label,
    path: path,
  })

  return items
})

function navigateTo(path: string | null) {
  if (path) {
    router.push(path)
  }
}
</script>

<template>
  <n-breadcrumb class="app-breadcrumb">
    <n-breadcrumb-item @click="navigateTo('/dashboard')">
      <span class="breadcrumb-home">首页</span>
    </n-breadcrumb-item>
    <n-breadcrumb-item
      v-for="(item, index) in breadcrumbs"
      :key="index"
      @click="navigateTo(item.path)"
    >
      <span :class="{ 'breadcrumb-clickable': item.path }">
        {{ item.label }}
      </span>
    </n-breadcrumb-item>
  </n-breadcrumb>
</template>

<style scoped>
.app-breadcrumb {
  font-size: 14px;
}

.breadcrumb-home {
  color: var(--text-secondary);
}

.breadcrumb-clickable {
  color: var(--primary-color);
  cursor: pointer;
}

.breadcrumb-clickable:hover {
  text-decoration: underline;
}
</style>
