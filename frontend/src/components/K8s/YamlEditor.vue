<template>
  <div class="yaml-editor">
    <textarea ref="editorRef" v-model="content"></textarea>
  </div>
</template>

<script setup>
import { ref, watch } from 'vue'

const props = defineProps({
  modelValue: { type: String, default: '' },
  readonly: { type: Boolean, default: false }
})

const emit = defineEmits(['update:modelValue'])
const editorRef = ref(null)
const content = ref(props.modelValue)

watch(() => props.modelValue, (val) => {
  content.value = val
})

watch(content, (val) => {
  emit('update:modelValue', val)
})
</script>

<style scoped>
.yaml-editor textarea {
  width: 100%;
  min-height: 400px;
  font-family: monospace;
  padding: 12px;
  border: 1px solid #dcdfe6;
  border-radius: 4px;
}
</style>
