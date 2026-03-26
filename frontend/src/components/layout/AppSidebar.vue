<script setup lang="ts">
import { computed, h, VNode } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { NMenu, NIcon } from 'naive-ui'
import type { MenuOption } from 'naive-ui'
import { useAppStore } from '@/stores/app'
import { useAuth } from '@/composables/useAuth'

interface AppMenuOption {
  label: string
  key: string
  permission?: string
  icon?: VNode
  children?: AppMenuOption[]
}

const router = useRouter()
const route = useRoute()
const appStore = useAppStore()
const { hasPermission } = useAuth()

const collapsed = computed(() => appStore.sidebarCollapsed)

// Icon components as inline SVG
const icons = {
  dashboard: h(
    'svg',
    { viewBox: '0 0 24 24', fill: 'none', xmlns: 'http://www.w3.org/2000/svg' },
    [
      h('path', {
        d: 'M3 13h8V3H3v10zm0 8h8v-6H3v6zm10 0h8V11h-8v10zm0-18v6h8V3h-8z',
        fill: 'currentColor',
      }),
    ]
  ),
  server: h(
    'svg',
    { viewBox: '0 0 24 24', fill: 'none', xmlns: 'http://www.w3.org/2000/svg' },
    [
      h('path', {
        d: 'M4 4h16v16H4V4zm2 2v12h12V6H6z',
        fill: 'currentColor',
      }),
    ]
  ),
  container: h(
    'svg',
    { viewBox: '0 0 24 24', fill: 'none', xmlns: 'http://www.w3.org/2000/svg' },
    [
      h('path', {
        d: 'M21 16.5c0 .38-.21.71-.53.88l-7.9 4.44c-.16.12-.36.18-.57.18-.21 0-.41-.06-.57-.18l-7.9-4.44A.991.991 0 013 16.5v-9c0-.38.21-.71.53-.88l7.9-4.44c.16-.12.36-.18.57-.18.21 0 .41.06.57.18l7.9 4.44c.32.17.53.5.53.88v9z',
        stroke: 'currentColor',
        'stroke-width': '2',
        'stroke-linecap': 'round',
        'stroke-linejoin': 'round',
        fill: 'none',
      }),
    ]
  ),
  alert: h(
    'svg',
    { viewBox: '0 0 24 24', fill: 'none', xmlns: 'http://www.w3.org/2000/svg' },
    [
      h('path', {
        d: 'M12 22c1.1 0 2-.9 2-2h-4c0 1.1.9 2 2 2zm6-6v-5c0-3.07-1.63-5.64-4.5-6.32V4c0-.83-.67-1.5-1.5-1.5s-1.5.67-1.5 1.5v.68C7.64 5.36 6 7.92 6 11v5l-2 2v1h16v-1l-2-2z',
        fill: 'currentColor',
      }),
    ]
  ),
  document: h(
    'svg',
    { viewBox: '0 0 24 24', fill: 'none', xmlns: 'http://www.w3.org/2000/svg' },
    [
      h('path', {
        d: 'M14 2H6c-1.1 0-2 .9-2 2v16c0 1.1.9 2 2 2h12c1.1 0 2-.9 2-2V8l-6-6zm4 18H6V4h7v5h5v11z',
        fill: 'currentColor',
      }),
    ]
  ),
  chart: h(
    'svg',
    { viewBox: '0 0 24 24', fill: 'none', xmlns: 'http://www.w3.org/2000/svg' },
    [
      h('path', {
        d: 'M3.5 18.5L9.5 12.5L13.5 16.5L22 6.92L20.59 5.5L13.5 13.5L9.5 9.5L2 17L3.5 18.5Z',
        fill: 'currentColor',
      }),
    ]
  ),
  storage: h(
    'svg',
    { viewBox: '0 0 24 24', fill: 'none', xmlns: 'http://www.w3.org/2000/svg' },
    [
      h('path', {
        d: 'M2 20h20v-4H2v4zm2-3h2v2H4v-2zM2 4v4h20V4H2zm2 3h2v2H4V7zm0 5h2v2H4v-2z',
        fill: 'currentColor',
      }),
    ]
  ),
  git: h(
    'svg',
    { viewBox: '0 0 24 24', fill: 'none', xmlns: 'http://www.w3.org/2000/svg' },
    [
      h('path', {
        d: 'M6 3v6h1V5.41l5.29 5.3.71-.71L7.41 4H11V3H6zm6 7.29l5.29-5.3H14V4h5v5h-1V5.41l-5.29 5.3.71.71.29-.3V17h-1v-6.71z',
        fill: 'currentColor',
      }),
    ]
  ),
  app: h(
    'svg',
    { viewBox: '0 0 24 24', fill: 'none', xmlns: 'http://www.w3.org/2000/svg' },
    [
      h('path', {
        d: 'M4 8h4V4H4v4zm6 12h4v-4h-4v4zm-6 0h4v-4H4v4zm0-6h4v-4H4v4zm6 0h4v-4h-4v4zm6-10v4h4V4h-4zm-6 4h4V4h-4v4zm6 6h4v-4h-4v4zm0 6h4v-4h-4v4z',
        fill: 'currentColor',
      }),
    ]
  ),
  shield: h(
    'svg',
    { viewBox: '0 0 24 24', fill: 'none', xmlns: 'http://www.w3.org/2000/svg' },
    [
      h('path', {
        d: 'M12 1L3 5v6c0 5.55 3.84 10.74 9 12 5.16-1.26 9-6.45 9-12V5l-9-4zm0 10.99l7-3.13V5.5L12 7.5 5 5.5v3.36l7 3.13z',
        fill: 'currentColor',
      }),
    ]
  ),
  api: h(
    'svg',
    { viewBox: '0 0 24 24', fill: 'none', xmlns: 'http://www.w3.org/2000/svg' },
    [
      h('path', {
        d: '13 13v8h8v-8h-8zM3 21h8v-8H3v8zM3 3v8h8V3H3zm13.66-1.31L11 7.34 16.66 13l5.66-5.66-5.66-5.65z',
        fill: 'currentColor',
      }),
    ]
  ),
  settings: h(
    'svg',
    { viewBox: '0 0 24 24', fill: 'none', xmlns: 'http://www.w3.org/2000/svg' },
    [
      h('path', {
        d: 'M19.14 12.94c.04-.31.06-.63.06-.94 0-.31-.02-.63-.06-.94l2.03-1.58c.18-.14.23-.41.12-.61l-1.92-3.32c-.12-.22-.37-.29-.59-.22l-2.39.96c-.5-.38-1.03-.7-1.62-.94l-.36-2.54c-.04-.24-.24-.41-.48-.41h-3.84c-.24 0-.43.17-.47.41l-.36 2.54c-.59.24-1.13.57-1.62.94l-2.39-.96c-.22-.08-.47 0-.59.22L5.16 8.87c-.12.21-.08.47.12.61l2.03 1.58c-.04.31-.07.64-.07.94s.02.63.06.94l-2.03 1.58c-.18.14-.23.41-.12.61l1.92 3.32c.12.22.37.29.59.22l2.39-.96c.5.38 1.03.7 1.62.94l.36 2.54c.05.24.24.41.48.41h3.84c.24 0 .44-.17.47-.41l.36-2.54c.59-.24 1.13-.56 1.62-.94l2.39.96c.22.08.47 0 .59-.22l1.92-3.32c.12-.22.07-.47-.12-.61l-2.01-1.58zM12 15.6c-1.98 0-3.6-1.62-3.6-3.6s1.62-3.6 3.6-3.6 3.6 1.62 3.6 3.6-1.62 3.6-3.6 3.6z',
        fill: 'currentColor',
      }),
    ]
  ),
}

function renderIcon(icon: VNode) {
  return () =>
    h(
      NIcon,
      { size: 18 },
      {
        default: () => icon,
      }
    )
}

const allMenuOptions: AppMenuOption[] = [
  {
    label: '仪表盘',
    key: '/dashboard',
    icon: icons.dashboard,
  },
  {
    label: '资产管理',
    key: '/asset',
    icon: icons.server,
  },
  {
    label: '容器管理',
    key: 'container-management',
    icon: icons.container,
    children: [
      {
        label: '集群管理',
        key: '/cluster',
        permission: 'cluster:list',
      },
      {
        label: '节点管理',
        key: '/node',
        permission: 'cluster:list',
      },
      {
        label: '工作负载',
        key: '/workload',
        permission: 'cluster:list',
      },
      {
        label: '网络管理',
        key: '/network',
        permission: 'cluster:list',
      },
      {
        label: '存储管理',
        key: '/storage',
        permission: 'cluster:list',
      },
      {
        label: '配置管理',
        key: '/config',
        permission: 'cluster:list',
      },
    ],
  },
  {
    label: '告警中心',
    key: '/ops/alert',
    permission: 'alert:list',
    icon: icons.alert,
  },
  {
    label: '日志检索',
    key: '/ops/log',
    permission: 'log:list',
    icon: icons.document,
  },
  {
    label: '监控配置',
    key: '/ops/monitor',
    permission: 'monitor:list',
    icon: icons.chart,
  },
  {
    label: 'Harbor管理',
    key: '/ops/harbor',
    permission: 'harbor:list',
    icon: icons.storage,
  },
  {
    label: 'CI/CD流水线',
    key: '/ops/cicd',
    permission: 'cicd:list',
    icon: icons.git,
  },
  {
    label: '应用管理',
    key: '/ops/app',
    permission: 'app:list',
    icon: icons.app,
  },
  {
    label: '审计日志',
    key: '/ops/audit',
    permission: 'audit:list',
    icon: icons.shield,
  },
  {
    label: 'API 文档',
    key: '/ops/api-docs',
    icon: icons.api,
  },
  {
    label: '系统管理',
    key: 'system-management',
    icon: icons.settings,
    children: [
      {
        label: '用户管理',
        key: '/system/users',
        permission: 'user:list',
      },
      {
        label: '部门管理',
        key: '/system/departments',
        permission: 'department:list',
      },
      {
        label: '角色管理',
        key: '/system/roles',
        permission: 'role:list',
      },
      {
        label: '权限管理',
        key: '/system/permissions',
        permission: 'permission:list',
      },
    ],
  },
]

function convertToMenuOption(item: AppMenuOption): MenuOption | null {
  if (item.permission && !hasPermission(item.permission)) {
    return null
  }

  const children = item.children
    ?.map(convertToMenuOption)
    .filter((child): child is MenuOption => child !== null)

  if (item.children && (!children || children.length === 0)) {
    return null
  }

  return {
    label: item.label,
    key: item.key,
    icon: item.icon ? renderIcon(item.icon) : undefined,
    children:
      children && children.length > 0
        ? children.map((child) => ({
            ...child,
            icon: undefined, // Remove icons from sub-items for cleaner look
          }))
        : undefined,
  }
}

const menuOptions = computed(() => {
  return allMenuOptions
    .map(convertToMenuOption)
    .filter((item): item is MenuOption => item !== null)
})

const activeKey = computed(() => route.path)

function handleMenuUpdate(key: string) {
  if (!key.startsWith('/')) {
    return
  }
  router.push(key)
}
</script>

<template>
  <aside class="app-sidebar" :class="{ 'app-sidebar--collapsed': collapsed }">
    <div class="sidebar-logo">
      <div class="logo-icon">
        <svg viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
          <path
            d="M12 2L2 7L12 12L22 7L12 2Z"
            stroke="currentColor"
            stroke-width="2"
            stroke-linecap="round"
            stroke-linejoin="round"
          />
          <path
            d="M2 17L12 22L22 17"
            stroke="currentColor"
            stroke-width="2"
            stroke-linecap="round"
            stroke-linejoin="round"
          />
          <path
            d="M2 12L12 17L22 12"
            stroke="currentColor"
            stroke-width="2"
            stroke-linecap="round"
            stroke-linejoin="round"
          />
        </svg>
      </div>
      <span v-if="!collapsed" class="logo-text">DevOps</span>
    </div>

    <div class="sidebar-menu">
      <n-menu
        :collapsed="collapsed"
        :collapsed-width="64"
        :collapsed-icon-size="18"
        :options="menuOptions"
        :value="activeKey"
        :accordion="true"
        :indent="20"
        @update:value="handleMenuUpdate"
      />
    </div>

    <div class="sidebar-toggle" @click="appStore.toggleSidebar()">
      <svg
        viewBox="0 0 24 24"
        fill="none"
        xmlns="http://www.w3.org/2000/svg"
        :class="{ 'toggle-icon--rotated': collapsed }"
      >
        <path
          d="M15 18L9 12L15 6"
          stroke="currentColor"
          stroke-width="2"
          stroke-linecap="round"
          stroke-linejoin="round"
        />
      </svg>
    </div>
  </aside>
</template>

<style scoped>
.app-sidebar {
  position: fixed;
  top: 0;
  left: 0;
  width: var(--sidebar-width);
  height: 100vh;
  background: var(--sidebar-bg);
  display: flex;
  flex-direction: column;
  z-index: var(--z-sidebar);
  transition: width var(--transition-normal);
}

.app-sidebar--collapsed {
  width: var(--sidebar-collapsed-width);
}

.sidebar-logo {
  height: var(--header-height);
  display: flex;
  align-items: center;
  padding: 0 var(--spacing-lg);
  gap: var(--spacing-md);
  border-bottom: 1px solid rgba(255, 255, 255, 0.1);
}

.logo-icon {
  width: 28px;
  height: 28px;
  color: #ffffff;
  flex-shrink: 0;
}

.logo-icon svg {
  width: 100%;
  height: 100%;
}

.logo-text {
  font-size: 18px;
  font-weight: 700;
  color: var(--sidebar-text);
  white-space: nowrap;
  overflow: hidden;
}

.sidebar-menu {
  flex: 1;
  overflow-y: auto;
  overflow-x: hidden;
  padding: var(--spacing-md) 0;
}

.sidebar-toggle {
  height: 40px;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  border-top: 1px solid rgba(255, 255, 255, 0.1);
  color: var(--sidebar-text-muted);
  transition: all var(--transition-fast);
}

.sidebar-toggle:hover {
  background: var(--sidebar-hover-bg);
  color: var(--sidebar-text);
}

.sidebar-toggle svg {
  width: 18px;
  height: 18px;
  transition: transform var(--transition-normal);
}

.toggle-icon--rotated {
  transform: rotate(180deg);
}

/* Override Naive UI Menu Styles */
:deep(.n-menu) {
  background: transparent !important;
}

:deep(.n-menu-item) {
  margin: 2px 8px;
  border-radius: var(--radius-sm);
}

:deep(.n-menu-item-content) {
  padding: 0 12px !important;
  height: 40px !important;
}

:deep(.n-menu-item-content:hover) {
  background: var(--sidebar-hover-bg) !important;
}

:deep(.n-menu-item-content--selected) {
  background: var(--sidebar-active-bg) !important;
}

:deep(.n-menu-item-content--selected::before) {
  opacity: 0 !important;
}

:deep(.n-menu-item-content__icon) {
  color: var(--sidebar-text-muted) !important;
}

:deep(.n-menu-item-content--selected .n-menu-item-content__icon) {
  color: var(--sidebar-text) !important;
}

:deep(.n-menu-item-content-header) {
  color: var(--sidebar-text-muted) !important;
  font-size: 14px !important;
}

:deep(.n-menu-item-content--selected .n-menu-item-content-header) {
  color: var(--sidebar-text) !important;
  font-weight: 500 !important;
}

/* Submenu styles */
:deep(.n-submenu-children) {
  background: rgba(0, 0, 0, 0.15) !important;
  margin: 0 8px;
  border-radius: var(--radius-sm);
}

:deep(.n-submenu-children .n-menu-item-content) {
  padding-left: 40px !important;
  height: 36px !important;
}
</style>
