<script setup lang="ts">
import { computed } from 'vue'
import {
  NLayout,
  NLayoutContent,
  NMessageProvider,
  NDialogProvider,
  NConfigProvider,
  darkTheme,
} from 'naive-ui'
import { useAppStore } from '@/stores/app'
import AppSidebar from '@/components/layout/AppSidebar.vue'
import AppHeader from '@/components/layout/AppHeader.vue'

import '@/assets/styles/variables.css'

const appStore = useAppStore()
const collapsed = computed(() => appStore.sidebarCollapsed)

// Custom theme overrides
const themeOverrides = {
  common: {
    primaryColor: '#3b82f6',
    primaryColorHover: '#2563eb',
    primaryColorPressed: '#1d4ed8',
    borderRadius: '6px',
    fontFamily:
      'Inter, -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif',
  },
  Select: {
    peers: {
      InternalSelection: {
        borderColor: '#e2e8f0',
        borderColorHover: '#3b82f6',
        borderColorFocus: '#3b82f6',
      },
    },
  },
}

// Dark theme overrides for sidebar
const darkThemeOverrides = {
  common: {
    primaryColor: '#4f46e5',
    primaryColorHover: '#6366f1',
    primaryColorPressed: '#4338ca',
  },
  Menu: {
    color: 'transparent',
    itemColorActive: '#4f46e5',
    itemColorActiveHover: '#6366f1',
    itemTextColor: '#94a3b8',
    itemTextColorHover: '#ffffff',
    itemTextColorActive: '#ffffff',
    itemTextColorChildActive: '#ffffff',
    itemIconColor: '#94a3b8',
    itemIconColorHover: '#ffffff',
    itemIconColorActive: '#ffffff',
    arrowColor: '#94a3b8',
  },
}
</script>

<template>
  <n-config-provider :theme-overrides="themeOverrides">
    <n-message-provider>
      <n-dialog-provider>
        <div class="app-layout">
          <!-- Sidebar -->
          <n-config-provider :theme="darkTheme" :theme-overrides="darkThemeOverrides">
            <AppSidebar />
          </n-config-provider>

          <!-- Header -->
          <AppHeader />

          <!-- Main Content Area -->
          <div class="app-main" :class="{ 'app-main--collapsed': collapsed }">
            <n-layout-content class="app-content">
              <router-view />
            </n-layout-content>
          </div>
        </div>
      </n-dialog-provider>
    </n-message-provider>
  </n-config-provider>
</template>

<style scoped>
.app-layout {
  display: flex;
  min-height: 100vh;
  background: var(--content-bg);
}

.app-main {
  flex: 1;
  margin-left: var(--sidebar-width);
  transition: margin-left var(--transition-normal);
  display: flex;
  flex-direction: column;
  min-height: 100vh;
}

.app-main--collapsed {
  margin-left: var(--sidebar-collapsed-width);
}

.app-content {
  flex: 1;
  margin-top: var(--header-height);
  background: var(--content-bg);
  overflow: auto;
}
</style>
