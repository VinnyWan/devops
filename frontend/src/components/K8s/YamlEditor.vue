<template>
  <div class="yaml-editor">
    <textarea
      class="yaml-textarea"
      :value="content"
      @input="onInput"
      :readonly="readonly"
      :style="{ minHeight: minHeight }"
    />
    <div class="yaml-actions" v-if="showCopy">
      <el-button size="small" @click="handleCopy">{{ copyText }}</el-button>
    </div>
  </div>
</template>

<script setup>
import { ref, watch } from 'vue'
import { ElMessage } from 'element-plus'

const props = defineProps({
  modelValue: { type: String, default: '' },
  readonly: { type: Boolean, default: false },
  showCopy: { type: Boolean, default: true },
  minHeight: { type: String, default: '400px' }
})

const emit = defineEmits(['update:modelValue'])

const content = ref(props.modelValue)
const copyText = ref('复制')

watch(() => props.modelValue, (val) => { content.value = val })

const onInput = (e) => {
  content.value = e.target.value
  emit('update:modelValue', content.value)
}

const handleCopy = async () => {
  try {
    await navigator.clipboard.writeText(content.value)
    copyText.value = '已复制'
    setTimeout(() => { copyText.value = '复制' }, 2000)
  } catch {
    ElMessage.error('复制失败')
  }
}
</script>

<style scoped>
.yaml-editor {
  position: relative;
}
.yaml-textarea {
  width: 100%;
  font-family: Consolas, 'Courier New', monospace;
  font-size: 13px;
  line-height: 1.5;
  padding: 12px;
  border: 1px solid #dcdfe6;
  border-radius: 4px;
  resize: vertical;
  box-sizing: border-box;
  background: #fafafa;
  color: #303133;
}
.yaml-textarea:focus {
  outline: none;
  border-color: #409eff;
}
.yaml-textarea:read-only {
  background: #f5f7fa;
  cursor: default;
}
.yaml-actions {
  margin-top: 8px;
  text-align: right;
}
</style>
