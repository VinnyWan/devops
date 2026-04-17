<template>
  <el-container class="layout-container">
    <el-aside :width="isCollapse ? '64px' : '200px'" class="sidebar">
      <div class="logo">
        <span v-if="!isCollapse">运维平台</span>
      </div>
      <el-menu :default-active="$route.path" :collapse="isCollapse" :unique-opened="true" router>
        <el-menu-item index="/dashboard">
          <el-icon><HomeFilled /></el-icon>
          <span>仪表盘</span>
        </el-menu-item>
        <el-sub-menu index="k8s">
          <template #title>
            <el-icon><Grid /></el-icon>
            <span>容器管理</span>
          </template>
          <el-menu-item index="/k8s/cluster">集群管理</el-menu-item>
          <el-menu-item index="/k8s/node">节点管理</el-menu-item>
          <el-menu-item index="/k8s/namespace">命名空间</el-menu-item>
          <el-menu-item index="/k8s/workload">工作负载</el-menu-item>
          <el-menu-item index="/k8s/network">网络管理</el-menu-item>
          <el-menu-item index="/k8s/storage">存储管理</el-menu-item>
          <el-menu-item index="/k8s/config">配置管理</el-menu-item>
        </el-sub-menu>
        <el-sub-menu index="system">
          <template #title>
            <el-icon><Setting /></el-icon>
            <span>系统管理</span>
          </template>
          <el-menu-item index="/system/user">用户管理</el-menu-item>
          <el-menu-item index="/system/role">角色管理</el-menu-item>
          <el-menu-item index="/system/department">部门管理</el-menu-item>
          <el-menu-item index="/system/permission">权限管理</el-menu-item>
        </el-sub-menu>
        <el-sub-menu index="audit">
          <template #title>
            <el-icon><Notebook /></el-icon>
            <span>操作审计</span>
          </template>
          <el-menu-item index="/audit/operation">操作日志</el-menu-item>
          <el-menu-item index="/audit/login">登录日志</el-menu-item>
        </el-sub-menu>
        <el-sub-menu index="cmdb">
          <template #title>
            <el-icon><Monitor /></el-icon>
            <span>资产管理</span>
          </template>
          <el-menu-item index="/cmdb/hosts">主机管理</el-menu-item>
          <el-menu-item index="/cmdb/groups">分组管理</el-menu-item>
          <el-menu-item index="/cmdb/credentials">凭据管理</el-menu-item>
        </el-sub-menu>
      </el-menu>
    </el-aside>
    <el-container>
      <el-header>
        <div class="header-left">
          <el-icon class="menu-toggle" @click="toggleCollapse">
            <Expand v-if="isCollapse" />
            <Fold v-else />
          </el-icon>
          <Breadcrumb />
        </div>
        <div class="header-right">
          <span>{{ userStore.userInfo?.username }}</span>
          <el-button @click="handleLogout" link>退出</el-button>
        </div>
      </el-header>
      <el-main>
        <router-view />
      </el-main>
    </el-container>
  </el-container>
</template>

<script setup>
import { ref } from 'vue'
import { HomeFilled, Grid, Setting, Expand, Fold, Notebook, Monitor } from '@element-plus/icons-vue'
import { useUserStore } from '../../stores/user'
import { useRouter } from 'vue-router'
import Breadcrumb from './Breadcrumb.vue'

const userStore = useUserStore()
const router = useRouter()
const isCollapse = ref(false)

const toggleCollapse = () => {
  isCollapse.value = !isCollapse.value
}

const handleLogout = () => {
  userStore.logout()
  router.push('/login')
}
</script>

<style scoped>
.layout-container {
  height: 100vh;
}
.sidebar {
  background: #001529;
  color: #fff;
  transition: width var(--transition-base);
}
.logo {
  height: 60px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #fff;
  font-size: 18px;
  font-weight: 600;
}
.el-menu {
  background: #001529;
  border-right: none;
}
:deep(.el-menu-item) {
  color: rgba(255, 255, 255, 0.65);
}
:deep(.el-menu-item:hover),
:deep(.el-menu-item.is-active) {
  background: var(--color-primary) !important;
  color: #fff;
}
:deep(.el-sub-menu__title) {
  color: rgba(255, 255, 255, 0.65);
}
:deep(.el-sub-menu__title:hover) {
  background: rgba(255, 255, 255, 0.08);
}
:deep(.el-menu--inline) {
  background: #000c17 !important;
}
:deep(.el-menu--inline .el-menu-item) {
  background: #000c17 !important;
}
:deep(.el-menu--inline .el-menu-item:hover) {
  background: rgba(255, 255, 255, 0.08) !important;
}
.el-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  background: #fff;
  border-bottom: 1px solid var(--color-border);
  box-shadow: var(--shadow-sm);
}
.menu-toggle {
  font-size: 20px;
  cursor: pointer;
  transition: color var(--transition-fast);
}
.menu-toggle:hover {
  color: var(--color-primary);
}
.header-right {
  display: flex;
  gap: var(--spacing-md);
  align-items: center;
}
.header-left {
  display: flex;
  align-items: center;
  gap: 12px;
}
.el-main {
  background: var(--color-bg);
  padding: var(--spacing-lg);
}

@media (max-width: 768px) {
  .el-main {
    padding: var(--spacing-md);
  }
}
</style>
