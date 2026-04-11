<template>
  <div class="empty-state">
    <el-icon :size="64" color="#94A3B8">
      <component :is="icon" />
    </el-icon>
    <p class="empty-text">{{ text }}</p>
    <el-button v-if="buttonText" type="primary" @click="$emit('action')">
      {{ buttonText }}
    </el-button>
  </div>
</template>

<script setup>
import { Document } from '@element-plus/icons-vue'

defineProps({
  icon: { type: Object, default: () => Document },
  text: { type: String, default: '暂无数据' },
  buttonText: { type: String, default: '' }
})

defineEmits(['action'])
</script>

<style scoped>
.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: var(--spacing-xl);
  min-height: 300px;
}

.empty-text {
  margin: var(--spacing-md) 0;
  color: var(--color-text-secondary);
  font-size: 14px;
}
</style>
