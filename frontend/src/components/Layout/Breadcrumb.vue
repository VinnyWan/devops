<template>
  <el-breadcrumb separator="/">
    <el-breadcrumb-item :to="{ path: '/' }">首页</el-breadcrumb-item>
    <el-breadcrumb-item v-for="item in breadcrumbs" :key="item.path || item.title">
      <router-link v-if="item.to" :to="item.to">{{ item.title }}</router-link>
      <span v-else>{{ item.title }}</span>
    </el-breadcrumb-item>
  </el-breadcrumb>
</template>

<script setup>
import { computed } from 'vue'
import { useRoute } from 'vue-router'

const route = useRoute()

const breadcrumbMap = {
  '/dashboard': [{ title: '仪表盘' }],
  '/k8s/cluster': [{ title: '容器管理' }, { title: '集群管理', to: '/k8s/cluster' }],
  '/k8s/node': [{ title: '容器管理' }, { title: '节点管理', to: '/k8s/node' }],
  '/system/user': [{ title: '系统管理' }, { title: '用户管理' }]
}

const breadcrumbs = computed(() => {
  const path = route.path
  // Exact match
  if (breadcrumbMap[path]) {
    return breadcrumbMap[path]
  }
  // Prefix match for dynamic routes
  if (path.startsWith('/k8s/cluster/')) {
    return [
      { title: '容器管理' },
      { title: '集群管理', to: '/k8s/cluster' },
      { title: '集群详情' }
    ]
  }
  return []
})
</script>
