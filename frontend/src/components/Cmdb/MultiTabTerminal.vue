<template>
  <div class="multi-tab-terminal">
    <!-- Tab Bar -->
    <div class="tab-bar">
      <div
        v-for="(tab, index) in tabs"
        :key="tab.id"
        class="tab-item"
        :class="{ active: index === activeIndex }"
        @click="activeIndex = index"
      >
        <span class="tab-name" :title="tab.hostname">{{ tab.hostname }}</span>
        <span class="tab-close" @click.stop="closeTab(index)">×</span>
      </div>
      <div class="tab-add" @click="$emit('requestConnect')">+</div>
    </div>

    <!-- Terminal Panels -->
    <div class="terminal-panels">
      <div
        v-for="(tab, index) in tabs"
        :key="tab.id"
        v-show="index === activeIndex"
        class="terminal-panel"
      >
        <Terminal
          :ref="el => setTerminalRef(el, index)"
          :ws-url="tab.wsUrl"
          :visible="index === activeIndex && visible"
        />
      </div>
    </div>

    <!-- Status Bar -->
    <div class="status-bar">
      <span>{{ tabs.length }} 个会话 · 录像中</span>
      <span class="shortcuts">Ctrl+Shift+T 新建 | Ctrl+Tab 切换 | Ctrl+W 关闭</span>
    </div>
  </div>
</template>

<script setup>
import { ref, watch, onMounted, onBeforeUnmount, nextTick } from 'vue'
import Terminal from '@/components/K8s/Terminal.vue'

const props = defineProps({
  visible: { type: Boolean, default: false }
})

const emit = defineEmits(['requestConnect', 'allClosed'])

const tabs = ref([])
const activeIndex = ref(-1)
const terminalRefs = ref({})

let tabIdCounter = 0

const setTerminalRef = (el, index) => {
  if (el) {
    terminalRefs.value[index] = el
  }
}

const addTab = (hostId, hostname, wsUrl) => {
  const id = ++tabIdCounter
  tabs.value.push({ id, hostId, hostname, wsUrl })
  activeIndex.value = tabs.value.length - 1
}

const closeTab = (index) => {
  const ref = terminalRefs.value[index]
  if (ref && typeof ref.clear === 'function') {
    ref.clear()
  }
  tabs.value.splice(index, 1)
  delete terminalRefs.value[index]

  // Rebuild refs map after splice
  const newRefs = {}
  tabs.value.forEach((_, i) => {
    if (terminalRefs.value[i + 1] !== undefined) {
      newRefs[i] = terminalRefs.value[i + 1]
    }
  })
  terminalRefs.value = newRefs

  if (tabs.value.length === 0) {
    activeIndex.value = -1
    emit('allClosed')
  } else if (activeIndex.value >= tabs.value.length) {
    activeIndex.value = tabs.value.length - 1
  } else if (activeIndex.value > index) {
    activeIndex.value--
  }
}

const closeCurrentTab = () => {
  if (activeIndex.value >= 0) {
    closeTab(activeIndex.value)
  }
}

const switchToNextTab = () => {
  if (tabs.value.length > 0) {
    activeIndex.value = (activeIndex.value + 1) % tabs.value.length
  }
}

const handleKeydown = (e) => {
  if (!props.visible) return
  if (e.ctrlKey && e.shiftKey && e.key === 'T') {
    e.preventDefault()
    emit('requestConnect')
  } else if (e.ctrlKey && e.key === 'Tab') {
    e.preventDefault()
    switchToNextTab()
  } else if (e.ctrlKey && e.key === 'w') {
    e.preventDefault()
    closeCurrentTab()
  }
}

onMounted(() => {
  window.addEventListener('keydown', handleKeydown)
})

onBeforeUnmount(() => {
  window.removeEventListener('keydown', handleKeydown)
})

// Re-fit terminal when switching tabs
watch(activeIndex, async () => {
  await nextTick()
  const ref = terminalRefs.value[activeIndex.value]
  if (ref && typeof ref.fit === 'function') {
    setTimeout(() => ref.fit(), 50)
  }
})

defineExpose({ addTab, closeCurrentTab })
</script>

<style scoped>
.multi-tab-terminal {
  display: flex;
  flex-direction: column;
  height: 100%;
  background: #1e1e1e;
}

.tab-bar {
  display: flex;
  background: #252526;
  border-bottom: 1px solid #3c3c3c;
  overflow-x: auto;
  flex-shrink: 0;
}

.tab-item {
  display: flex;
  align-items: center;
  padding: 6px 12px;
  cursor: pointer;
  border-right: 1px solid #3c3c3c;
  font-size: 12px;
  color: #969696;
  white-space: nowrap;
  min-width: 100px;
  max-width: 160px;
  transition: background 0.15s;
}

.tab-item:hover {
  background: #2d2d2d;
}

.tab-item.active {
  background: #1e1e1e;
  color: #ffffff;
  border-bottom: 2px solid #409EFF;
}

.tab-name {
  overflow: hidden;
  text-overflow: ellipsis;
  flex: 1;
}

.tab-close {
  margin-left: 8px;
  color: #666;
  font-size: 14px;
  border-radius: 3px;
  width: 18px;
  height: 18px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.tab-close:hover {
  background: #c14545;
  color: #fff;
}

.tab-add {
  padding: 6px 14px;
  cursor: pointer;
  color: #aaa;
  font-size: 16px;
  transition: background 0.15s;
}

.tab-add:hover {
  background: #3c3c3c;
  color: #fff;
}

.terminal-panels {
  flex: 1;
  overflow: hidden;
  position: relative;
}

.terminal-panel {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
}

.status-bar {
  display: flex;
  justify-content: space-between;
  padding: 4px 12px;
  background: #252526;
  border-top: 1px solid #3c3c3c;
  font-size: 11px;
  color: #666;
  flex-shrink: 0;
}

.shortcuts {
  color: #555;
}
</style>
