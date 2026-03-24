<script setup lang="ts">
import { computed } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import {
  NLayout,
  NLayoutSider,
  NLayoutHeader,
  NLayoutContent,
  NMenu,
  NDropdown,
  NSpace,
  NButton,
  NMessageProvider,
  NDialogProvider,
  NConfigProvider,
} from 'naive-ui'
import type { MenuOption } from 'naive-ui'
import { useAppStore } from '@/stores/app'
import { useAuth } from '@/composables/useAuth'

interface AppMenuOption {
  label: string
  key: string
  permission?: string
  children?: AppMenuOption[]
}

const router = useRouter()
const route = useRoute()
const appStore = useAppStore()
const { user, logout, hasPermission } = useAuth()
const collapsed = computed(() => appStore.sidebarCollapsed)

const allMenuOptions: AppMenuOption[] = [
  { label: '仪表盘', key: '/dashboard' },
  { label: '资产管理', key: '/asset' },
  {
    label: '容器管理',
    key: 'container-management',
    children: [
      { label: '集群管理', key: '/cluster', permission: 'cluster:list' },
      { label: '节点管理', key: '/node', permission: 'cluster:list' },
      { label: '命名空间', key: '/namespace', permission: 'cluster:list' },
      { label: '工作负载', key: '/workload', permission: 'cluster:list' },
      { label: '网络管理', key: '/network', permission: 'cluster:list' },
      { label: '存储管理', key: '/storage', permission: 'cluster:list' },
      { label: '配置管理', key: '/config', permission: 'cluster:list' },
    ],
  },
  { label: '告警中心', key: '/ops/alert', permission: 'alert:list' },
  { label: '日志检索', key: '/ops/log', permission: 'log:list' },
  { label: '监控配置', key: '/ops/monitor', permission: 'monitor:list' },
  { label: 'Harbor管理', key: '/ops/harbor', permission: 'harbor:list' },
  { label: 'CI/CD流水线', key: '/ops/cicd', permission: 'cicd:list' },
  { label: '应用管理', key: '/ops/app', permission: 'app:list' },
  { label: '审计日志', key: '/ops/audit', permission: 'audit:list' },
  {
    label: '系统管理',
    key: 'system',
    children: [
      { label: '用户管理', key: '/system/users', permission: 'user:list' },
      { label: '部门管理', key: '/system/departments', permission: 'department:list' },
      { label: '角色管理', key: '/system/roles', permission: 'role:list' },
      { label: '权限管理', key: '/system/permissions', permission: 'permission:list' },
    ],
  },
]

function filterMenuOptions(options: AppMenuOption[]): MenuOption[] {
  return options
    .map((item) => {
      if (item.permission && !hasPermission(item.permission)) {
        return null
      }
      const children = item.children ? filterMenuOptions(item.children) : undefined
      if (item.children && (!children || children.length === 0)) {
        return null
      }
      return {
        label: item.label,
        key: item.key,
        children,
      } as MenuOption
    })
    .filter((item): item is MenuOption => item !== null)
}

const menuOptions = computed(() => filterMenuOptions(allMenuOptions))
const activeKey = computed(() => route.path)

function handleMenuUpdate(key: string) {
  if (!key.startsWith('/')) {
    return
  }
  router.push(key)
}

const userOptions = [
  { label: '个人信息', key: 'profile' },
  { label: '修改密码', key: 'password' },
  { type: 'divider', key: 'd1' },
  { label: '退出登录', key: 'logout' },
]

async function handleUserAction(key: string) {
  if (key === 'logout') {
    await logout()
    router.push('/login')
  }
}
</script>

<template>
  <n-config-provider>
    <n-message-provider>
      <n-dialog-provider>
      <n-layout has-sider style="height: 100vh">
        <n-layout-sider
          bordered
          :collapsed="collapsed"
          collapse-mode="width"
          :collapsed-width="64"
          :width="220"
          show-trigger
          @collapse="appStore.toggleSidebar()"
          @expand="appStore.toggleSidebar()"
        >
          <div class="logo">
            <span v-if="!collapsed">DevOps</span>
            <span v-else>D</span>
          </div>
          <n-menu
            :collapsed="collapsed"
            :collapsed-width="64"
            :collapsed-icon-size="22"
            :options="menuOptions"
            :value="activeKey"
            :accordion="true"
            @update:value="handleMenuUpdate"
          />
        </n-layout-sider>
        <n-layout>
          <n-layout-header bordered style="height: 56px; padding: 0 24px">
            <n-space justify="end" align="center" style="height: 100%">
              <n-dropdown :options="userOptions" @select="handleUserAction">
                <n-button quaternary>
                  {{ user?.name || user?.username || '用户' }}
                </n-button>
              </n-dropdown>
            </n-space>
          </n-layout-header>
          <n-layout-content
            content-style="padding: 20px;"
            style="height: calc(100vh - 56px); overflow: auto"
          >
            <router-view />
          </n-layout-content>
        </n-layout>
      </n-layout>
    </n-dialog-provider>
  </n-message-provider>
  </n-config-provider>
</template>

<style scoped>
.logo {
  height: 56px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 18px;
  font-weight: 700;
  color: #18a058;
  border-bottom: 1px solid var(--n-border-color);
}
</style>
