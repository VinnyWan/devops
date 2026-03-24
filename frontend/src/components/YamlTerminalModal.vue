<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { NModal, NButton, NSpace, useMessage } from 'naive-ui'
import yaml from 'js-yaml'

const props = defineProps<{
  show: boolean
  title: string
  content: string
  loading?: boolean
}>()

const emit = defineEmits<{
  'update:show': [value: boolean]
  'update:content': [value: string]
  'save': []
}>()

const message = useMessage()
const textareaRef = ref<HTMLTextAreaElement>()
const lineNumbersRef = ref<HTMLDivElement>()
const highlightedRef = ref<HTMLDivElement>()

const lines = computed(() => props.content.split('\n'))

const highlightedContent = computed(() => {
  return props.content
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/^(\s*)([\w-]+):/gm, '$1<span class="yaml-key">$2</span>:')
    .replace(/:\s*([^\n]+)/g, ': <span class="yaml-value">$1</span>')
    .replace(/(#[^\n]*)/g, '<span class="yaml-comment">$1</span>')
})

function copyToClipboard() {
  navigator.clipboard.writeText(props.content)
  message.success('已复制到剪贴板')
}

function validateYaml() {
  try {
    yaml.load(props.content)
    message.success('YAML格式正确')
  } catch (e: any) {
    message.error('YAML格式错误: ' + e.message)
  }
}

function handleInput(e: Event) {
  emit('update:content', (e.target as HTMLTextAreaElement).value)
}

function handleScroll(e: Event) {
  if (lineNumbersRef.value && highlightedRef.value) {
    lineNumbersRef.value.scrollTop = (e.target as HTMLTextAreaElement).scrollTop
    highlightedRef.value.scrollTop = (e.target as HTMLTextAreaElement).scrollTop
  }
}

watch(() => props.show, (newVal) => {
  if (newVal && textareaRef.value && lineNumbersRef.value) {
    lineNumbersRef.value.scrollTop = textareaRef.value.scrollTop
  }
})
</script>

<template>
  <NModal
    :show="show"
    @update:show="emit('update:show', $event)"
    style="width: 90%; max-width: 1200px"
    :trap-focus="false"
    :content-style="{ padding: 0, background: '#1e1e1e' }"
    :bordered="false"
  >
    <div class="yaml-terminal">
      <div class="terminal-header">
        <span class="terminal-title">编辑YAML - {{ title }}</span>
        <NSpace>
          <NButton size="small" @click="validateYaml">校验格式</NButton>
          <NButton size="small" @click="copyToClipboard">复制</NButton>
          <NButton size="small" type="primary" @click="emit('save')">保存</NButton>
          <NButton size="small" @click="emit('update:show', false)">关闭</NButton>
        </NSpace>
      </div>
      <div v-if="loading" class="terminal-loading">加载中...</div>
      <div v-else class="terminal-body">
        <div ref="lineNumbersRef" class="line-numbers">
          <div v-for="(_, i) in lines" :key="i" class="line-number">{{ i + 1 }}</div>
        </div>
        <div class="editor-container">
          <div ref="highlightedRef" class="code-highlight" v-html="highlightedContent"></div>
          <textarea
            ref="textareaRef"
            :value="content"
            @input="handleInput"
            @scroll="handleScroll"
            class="code-editor"
            spellcheck="false"
          />
        </div>
      </div>
    </div>
  </NModal>
</template>

<style scoped>
.yaml-terminal {
  background: #1e1e1e;
  border: 2px solid #ffffff;
  border-radius: 8px;
  overflow: hidden;
}

.terminal-header {
  background: #ffffff;
  color: #333;
  padding: 12px 16px;
  display: flex;
  justify-content: space-between;
  align-items: center;
  border-bottom: 1px solid #e0e0e0;
}

.terminal-title {
  font-family: 'Monaco', 'Menlo', 'Consolas', monospace;
  font-size: 14px;
  font-weight: 500;
}

.terminal-loading {
  background: #1e1e1e;
  color: #d4d4d4;
  padding: 40px;
  text-align: center;
  font-family: 'Monaco', 'Menlo', 'Consolas', monospace;
}

.terminal-body {
  background: #1e1e1e;
  display: flex;
  max-height: 70vh;
  overflow: hidden;
}

.line-numbers {
  background: #252526;
  color: #858585;
  padding: 16px 8px;
  text-align: right;
  font-family: 'Monaco', 'Menlo', 'Consolas', monospace;
  font-size: 13px;
  line-height: 20px;
  user-select: none;
  overflow-y: hidden;
}

.line-number {
  height: 20px;
}

.editor-container {
  flex: 1;
  position: relative;
  overflow: hidden;
}

.code-highlight {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  padding: 16px;
  font-family: 'Monaco', 'Menlo', 'Consolas', monospace;
  font-size: 13px;
  line-height: 20px;
  white-space: pre;
  overflow-y: auto;
  pointer-events: none;
  color: #d4d4d4;
  tab-size: 2;
}

.code-editor {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: transparent;
  color: transparent;
  caret-color: #d4d4d4;
  border: none;
  outline: none;
  padding: 16px;
  font-family: 'Monaco', 'Menlo', 'Consolas', monospace;
  font-size: 13px;
  line-height: 20px;
  resize: none;
  overflow-y: auto;
  white-space: pre;
  tab-size: 2;
}

.code-editor::-webkit-scrollbar,
.code-highlight::-webkit-scrollbar {
  width: 10px;
  height: 10px;
}

.code-editor::-webkit-scrollbar-track,
.code-highlight::-webkit-scrollbar-track {
  background: #ffffff;
}

.code-editor::-webkit-scrollbar-thumb,
.code-highlight::-webkit-scrollbar-thumb {
  background: #ffffff;
  border-radius: 5px;
  border: 2px solid #e0e0e0;
}

.code-editor::-webkit-scrollbar-thumb:hover,
.code-highlight::-webkit-scrollbar-thumb:hover {
  background: #f5f5f5;
}

:deep(.yaml-key) {
  color: #9cdcfe;
  font-weight: 500;
}

:deep(.yaml-value) {
  color: #ce9178;
}

:deep(.yaml-comment) {
  color: #6a9955;
  font-style: italic;
}
</style>
