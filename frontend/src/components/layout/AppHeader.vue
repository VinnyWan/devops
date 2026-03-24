<script setup lang="ts">
import { NDropdown, NButton } from 'naive-ui'
import { useAuth } from '@/composables/useAuth'
import { useRouter } from 'vue-router'
import AppBreadcrumb from './AppBreadcrumb.vue'

const router = useRouter()
const { user, logout } = useAuth()

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
  } else if (key === 'profile') {
    router.push('/profile')
  } else if (key === 'password') {
    router.push('/change-password')
  }
}
</script>

<template>
  <header class="app-header">
    <div class="header-left">
      <AppBreadcrumb />
    </div>

    <div class="header-right">
      <n-dropdown :options="userOptions" @select="handleUserAction">
        <n-button quaternary class="user-button">
          <template #icon>
            <svg
              viewBox="0 0 24 24"
              fill="none"
              xmlns="http://www.w3.org/2000/svg"
              class="user-icon"
            >
              <path
                d="M20 21V19C20 17.9391 19.5786 16.9217 18.8284 16.1716C18.0783 15.4214 17.0609 15 16 15H8C6.93913 15 5.92172 15.4214 5.17157 16.1716C4.42143 16.9217 4 17.9391 4 19V21"
                stroke="currentColor"
                stroke-width="2"
                stroke-linecap="round"
                stroke-linejoin="round"
              />
              <path
                d="M12 11C14.2091 11 16 9.20914 16 7C16 4.79086 14.2091 3 12 3C9.79086 3 8 4.79086 8 7C8 9.20914 9.79086 11 12 11Z"
                stroke="currentColor"
                stroke-width="2"
                stroke-linecap="round"
                stroke-linejoin="round"
              />
            </svg>
          </template>
          {{ user?.name || user?.username || '用户' }}
          <svg
            viewBox="0 0 24 24"
            fill="none"
            xmlns="http://www.w3.org/2000/svg"
            class="dropdown-icon"
          >
            <path
              d="M6 9L12 15L18 9"
              stroke="currentColor"
              stroke-width="2"
              stroke-linecap="round"
              stroke-linejoin="round"
            />
          </svg>
        </n-button>
      </n-dropdown>
    </div>
  </header>
</template>

<style scoped>
.app-header {
  position: fixed;
  top: 0;
  left: var(--sidebar-width);
  right: 0;
  height: var(--header-height);
  background: var(--header-bg);
  border-bottom: 1px solid var(--header-border);
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 var(--spacing-xl);
  z-index: var(--z-header);
  transition: left var(--transition-normal);
}

.header-left {
  display: flex;
  align-items: center;
}

.header-right {
  display: flex;
  align-items: center;
  gap: var(--spacing-lg);
}

.user-button {
  display: flex;
  align-items: center;
  gap: var(--spacing-sm);
  color: var(--text-primary);
  font-size: 14px;
}

.user-icon {
  width: 18px;
  height: 18px;
  color: var(--text-secondary);
}

.dropdown-icon {
  width: 14px;
  height: 14px;
  color: var(--text-muted);
}
</style>
