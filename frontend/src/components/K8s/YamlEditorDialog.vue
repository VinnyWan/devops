<template>
  <el-dialog v-model="visible" :title="title" :width="width" destroy-on-close :top="top">
    <YamlEditor v-model="localValue" :readonly="readonly" :show-copy="showCopy" :min-height="minHeight" />
    <template #footer>
      <div class="dialog-footer">
        <slot name="footer-left" />
        <div class="dialog-footer-actions">
          <el-button @click="visible = false">取消</el-button>
          <el-button v-if="!readonly" type="primary" :loading="saving" @click="emitSave">保存</el-button>
        </div>
      </div>
    </template>
  </el-dialog>
</template>

<script setup>
import { computed } from 'vue'
import YamlEditor from './YamlEditor.vue'

const props = defineProps({
  modelValue: { type: Boolean, default: false },
  title: { type: String, default: 'YAML' },
  content: { type: String, default: '' },
  readonly: { type: Boolean, default: false },
  saving: { type: Boolean, default: false },
  width: { type: String, default: '900px' },
  top: { type: String, default: '3vh' },
  minHeight: { type: String, default: '600px' },
  showCopy: { type: Boolean, default: true }
})

const emit = defineEmits(['update:modelValue', 'update:content', 'save'])

const visible = computed({
  get: () => props.modelValue,
  set: (val) => emit('update:modelValue', val)
})

const localValue = computed({
  get: () => props.content,
  set: (val) => emit('update:content', val)
})

const emitSave = () => emit('save')
</script>

<style scoped>
.dialog-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
  width: 100%;
}
.dialog-footer-actions {
  display: flex;
  gap: 8px;
}
</style>
