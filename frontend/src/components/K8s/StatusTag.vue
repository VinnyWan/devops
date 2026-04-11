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
    Succeeded: 'success',
    Ready: 'success',
    Pending: 'warning',
    Failed: 'danger',
    Unknown: 'info',
    Terminating: 'warning'
  }
  return statusMap[props.status] || 'info'
})
</script>
