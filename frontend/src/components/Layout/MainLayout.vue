<template>
  <el-container class="layout-container">
    <el-aside :width="isCollapse ? '64px' : '220px'" class="sidebar">
      <div class="logo">
        <div class="logo-icon" v-if="isCollapse">S</div>
        <span v-if="!isCollapse" class="logo-text">SRE Platform</span>
      </div>
      <el-menu :default-active="activeMenu" :collapse="isCollapse" :unique-opened="true" router>
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
        <el-sub-menu index="cmdb">
          <template #title>
            <el-icon><Monitor /></el-icon>
            <span>资产管理</span>
          </template>
          <el-menu-item index="/cmdb/hosts">主机管理</el-menu-item>
          <el-menu-item index="/cmdb/groups">分组管理</el-menu-item>
          <el-menu-item index="/cmdb/credentials">凭据管理</el-menu-item>
          <el-menu-item index="/cmdb/terminal/sessions">终端审计</el-menu-item>
          <el-menu-item index="/cmdb/permissions">权限配置</el-menu-item>
          <el-menu-item index="/cmdb/cloud-accounts">云账号</el-menu-item>
          <el-menu-item index="/cmdb/files">文件管理</el-menu-item>
        </el-sub-menu>
        <el-sub-menu index="audit">
          <template #title>
            <el-icon><Notebook /></el-icon>
            <span>操作审计</span>
          </template>
          <el-menu-item index="/audit/operation">操作日志</el-menu-item>
          <el-menu-item index="/audit/login">登录日志</el-menu-item>
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
          <span class="username">{{ userStore.userInfo?.username }}</span>
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
import { computed, ref } from 'vue'
import { HomeFilled, Grid, Setting, Expand, Fold, Notebook, Monitor } from '@element-plus/icons-vue'
import { useUserStore } from '../../stores/user'
import { useRoute, useRouter } from 'vue-router'
import Breadcrumb from './Breadcrumb.vue'

const userStore = useUserStore()
const route = useRoute()
const router = useRouter()
const isCollapse = ref(false)
// Normalize active menu paths so dynamic detail pages keep their parent nav highlighted.
const activeMenu = computed(() => route.meta.activeMenu || route.path)

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

/* Sidebar — Light Blue-Gray Theme */
.sidebar {
  background: var(--color-bg-muted);
  border-right: 1px solid var(--color-border);
  transition: width var(--transition-base);
  overflow-x: hidden;
  overflow-y: auto;
  display: flex;
  flex-direction: column;
}

.logo {
  height: 56px;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 0 var(--spacing-md);
  border-bottom: 1px solid var(--color-border);
  flex-shrink: 0;
}

.logo-icon {
  width: 32px;
  height: 32px;
  border-radius: var(--radius-sm);
  background: var(--color-primary);
  color: white;
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: 700;
  font-size: 14px;
}

.logo-text {
  font-size: 16px;
  font-weight: 700;
  color: var(--color-text);
  letter-spacing: -0.02em;
}

/* Menu — Light Blue-Gray */
.el-menu {
  background: transparent;
  border-right: none;
  padding: var(--spacing-xs);
}

:deep(.el-menu-item) {
  color: var(--color-text-secondary);
  border-radius: var(--radius-sm);
  margin: 2px 0;
  height: 40px;
  line-height: 40px;
  transition: all var(--transition-fast);
}

:deep(.el-menu-item:hover) {
  background: var(--color-bg-white) !important;
  color: var(--color-primary);
}

:deep(.el-menu-item.is-active) {
  background: var(--color-bg-white) !important;
  color: var(--color-primary);
  font-weight: 600;
  border-left: 3px solid var(--color-primary);
  padding-left: 17px;
  box-shadow: var(--shadow-xs);
}

:deep(.el-sub-menu__title) {
  color: var(--color-text);
  border-radius: var(--radius-sm);
  margin: 2px 0;
  height: 40px;
  line-height: 40px;
}

:deep(.el-sub-menu__title:hover) {
  background: var(--color-bg-white);
  color: var(--color-primary);
}

:deep(.el-menu--inline) {
  background: transparent !important;
  padding-left: 8px;
}

:deep(.el-menu--inline .el-menu-item) {
  background: transparent !important;
  font-size: 13px;
  height: 36px;
  line-height: 36px;
}

:deep(.el-menu--inline .el-menu-item:hover) {
  background: var(--color-bg-white) !important;
}

:deep(.el-menu--inline .el-menu-item.is-active) {
  background: var(--color-bg-white) !important;
  border-left: 3px solid var(--color-primary);
  padding-left: 17px;
  box-shadow: var(--shadow-xs);
}

/* Header */
.el-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  background: var(--color-bg-white);
  border-bottom: 1px solid var(--color-border-light);
  height: 56px;
  padding: 0 var(--spacing-lg);
}

.menu-toggle {
  font-size: 18px;
  cursor: pointer;
  color: var(--color-text-secondary);
  transition: color var(--transition-fast);
  padding: var(--spacing-xs);
  border-radius: var(--radius-xs);
}
.menu-toggle:hover {
  color: var(--color-primary);
  background: var(--color-bg-hover);
}

.header-right {
  display: flex;
  gap: var(--spacing-sm);
  align-items: center;
}

.username {
  font-size: var(--font-size-sm);
  color: var(--color-text-secondary);
  font-weight: 500;
}

.header-left {
  display: flex;
  align-items: center;
  gap: var(--spacing-sm);
}

/* Main Content */
.el-main {
  background: var(--color-bg);
  padding: var(--spacing-lg);
  overflow-y: auto;
}

@media (max-width: 768px) {
  .el-main {
    padding: var(--spacing-md);
  }
  .sidebar {
    position: fixed;
    z-index: 100;
    height: 100vh;
    box-shadow: var(--shadow-lg);
  }
}
</style>
