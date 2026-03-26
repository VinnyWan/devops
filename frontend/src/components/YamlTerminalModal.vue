<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { NModal, NButton, NSpace, useMessage } from 'naive-ui'
import yaml from 'js-yaml'

const props = withDefaults(defineProps<{
  show: boolean
  title: string
  content: string
  loading?: boolean
  readonly?: boolean
  width?: string
}>(), {
  readonly: false,
  width: '90%'
})

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
    .replace(/^(\s*)([\w.-]+):/gm, '$1<span class="yaml-key">$2</span>:')
    .replace(/:\s*"([^"]*)"/g, ': "<span class="yaml-string">$1</span>"')
    .replace(/:\s*(true|false|null)/g, ': <span class="yaml-bool">$1</span>')
    .replace(/:\s*(\d+)/g, ': <span class="yaml-number">$1</span>')
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
  const scrollTop = (e.target as HTMLTextAreaElement | HTMLDivElement).scrollTop
  if (lineNumbersRef.value) {
    lineNumbersRef.value.scrollTop = scrollTop
  }
  if (highlightedRef.value) {
    highlightedRef.value.scrollTop = scrollTop
  }
}

watch(() => props.show, (newVal) => {
  if (newVal) {
    setTimeout(() => {
      if (textareaRef.value && lineNumbersRef.value) {
        lineNumbersRef.value.scrollTop = textareaRef.value.scrollTop
      }
    }, 0)
  }
})
</script>

<template>
  <NModal
    :show="show"
    @update:show="emit('update:show', $event)"
    :style="{ width: width, maxWidth: '1200px' }"
    :trap-focus="false"
    :content-style="{ padding: 0, background: '#1e1e1e' }"
    :bordered="false"
  >
    <div class="yaml-terminal">
      <div class="terminal-header">
        <span class="terminal-title">{{ readonly ? '查看' : '编辑' }}YAML - {{ title }}</span>
        <NSpace>
          <NButton v-if="!readonly" size="small" @click="validateYaml">校验格式</NButton>
          <NButton size="small" @click="copyToClipboard">复制</NButton>
          <NButton v-if="!readonly" size="small" type="primary" @click="emit('save')">保存</NButton>
          <NButton size="small" @click="emit('update:show', false)">关闭</NButton>
        </NSpace>
      </div>
      <div v-if="loading" class="terminal-loading">加载中...</div>
      <div v-else class="terminal-body">
        <div ref="lineNumbersRef" class="line-numbers">
          <div v-for="(_, i) in lines" :key="i" class="line-number">{{ i + 1 }}</div>
        </div>
        <div class="editor-container">
          <div ref="highlightedRef" class="code-highlight" :class="{ editable: !readonly }" v-html="highlightedContent"></div>
          <!-- 编辑模式 -->
          <textarea
            v-if="!readonly"
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
  overflow-y: auto;
  overflow-x: hidden;
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
  flex-shrink: 0;
}

.line-number {
  height: 20px;
}

.editor-container {
  flex: 1;
  position: relative;
  overflow: hidden;
  min-height: 300px;
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
  color: #d4d4d4;
  tab-size: 2;
}

.code-highlight.editable {
  pointer-events: none;
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
.code-highlight::-webkit-scrollbar,
.terminal-body::-webkit-scrollbar {
  width: 10px;
  height: 10px;
}

.code-editor::-webkit-scrollbar-track,
.code-highlight::-webkit-scrollbar-track,
.terminal-body::-webkit-scrollbar-track {
  background: #1e1e1e;
}

.code-editor::-webkit-scrollbar-thumb,
.code-highlight::-webkit-scrollbar-thumb,
.terminal-body::-webkit-scrollbar-thumb {
  background: #4a4a4a;
  border-radius: 5px;
}

.code-editor::-webkit-scrollbar-thumb:hover,
.code-highlight::-webkit-scrollbar-thumb:hover,
.terminal-body::-webkit-scrollbar-thumb:hover {
  background: #5a5a5a;
}

/* YAML 语法高亮 */
:deep(.yaml-key) {
  color: #9cdcfe;
  font-weight: 500;
}

:deep(.yaml-string) {
  color: #ce9178;
}

:deep(.yaml-number) {
  color: #b5cea8;
}

:deep(.yaml-bool) {
  color: #569cd6;
}

:deep(.yaml-comment) {
  color: #6a9955;
  font-style: italic;
}
</style>
