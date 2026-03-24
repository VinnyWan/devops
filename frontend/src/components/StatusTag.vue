<script setup lang="ts">
import { computed } from 'vue'
import { NTag } from 'naive-ui'

const props = defineProps<{ status: string }>()

const typeMap: Record<string, 'success' | 'warning' | 'error' | 'info'> = {
  healthy: 'success',
  Ready: 'success',
  Running: 'success',
  Active: 'success',
  unhealthy: 'error',
  NotReady: 'error',
  Failed: 'error',
  unknown: 'warning',
  Warning: 'warning',
  Pending: 'warning',
}

const tagType = computed(() => typeMap[props.status] || 'warning')
</script>

<template>
  <n-tag :type="tagType" size="small">{{ status }}</n-tag>
</template>
