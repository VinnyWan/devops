<template>
  <el-tag :type="tagType" size="small">{{ status }}</el-tag>
</template>

<script setup>
import { computed } from 'vue'

const props = defineProps({
  status: { type: String, required: true }
})

const tagType = computed(() => {
  const statusMap = {
    Running: 'success',
    Available: 'success',
    Succeeded: 'success',
    Completed: 'success',
    Ready: 'success',
    Scheduled: '',
    Active: '',
    Progressing: 'warning',
    Pending: 'warning',
    Suspended: 'warning',
    Terminating: 'warning',
    Unavailable: 'danger',
    Degraded: 'danger',
    Failed: 'danger',
    Unknown: 'info'
  }
  return statusMap[props.status] || 'info'
})
</script>
