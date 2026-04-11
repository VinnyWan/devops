<template>
  <el-breadcrumb separator="/">
    <el-breadcrumb-item :to="{ path: '/' }">首页</el-breadcrumb-item>
    <el-breadcrumb-item v-for="item in breadcrumbs" :key="item.path">
      {{ item.title }}
    </el-breadcrumb-item>
  </el-breadcrumb>
</template>

<script setup>
import { computed } from 'vue'
import { useRoute } from 'vue-router'

const route = useRoute()

const breadcrumbMap = {
  '/dashboard': '仪表盘',
  '/k8s/cluster': '集群管理',
  '/workload': '工作负载',
  '/system/user': '用户管理'
}

const breadcrumbs = computed(() => {
  const path = route.path
  const title = breadcrumbMap[path]
  return title ? [{ path, title }] : []
})
</script>
