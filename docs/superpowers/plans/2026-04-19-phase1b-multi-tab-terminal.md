# Multi-Tab Terminal Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Enable users to open multiple SSH sessions as tabs within a single browser window, each independently recorded.

**Architecture:** Frontend-only change. A new `MultiTabTerminal.vue` component wraps the existing `Terminal.vue` component with a tab bar. Each tab manages its own WebSocket connection and terminal instance. No backend changes needed — each tab creates a new WebSocket connection through the existing terminal bridge.

**Tech Stack:** Vue 3 + Element Plus + xterm.js (existing)

---

## File Structure

| Action | Path | Responsibility |
|--------|------|---------------|
| Create | `frontend/src/components/Cmdb/MultiTabTerminal.vue` | Tab container with terminal instances |
| Modify | `frontend/src/views/Cmdb/HostList.vue` | Replace single terminal dialog with multi-tab |

---

### Task 1: Create MultiTabTerminal Component

**Files:**
- Create: `frontend/src/components/Cmdb/MultiTabTerminal.vue`

- [ ] **Step 1: Create the multi-tab terminal wrapper**

Create `frontend/src/components/Cmdb/MultiTabTerminal.vue`:

```vue
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
```

- [ ] **Step 2: Verify component compiles**

Run: `cd frontend && npx vue-tsc --noEmit` (or just `npm run dev` to check)
Expected: no compile errors

- [ ] **Step 3: Commit**

```bash
cd frontend
git add src/components/Cmdb/MultiTabTerminal.vue
git commit -m "feat(cmdb): add MultiTabTerminal component with tab bar and keyboard shortcuts"
```

---

### Task 2: Integrate MultiTabTerminal into HostList

**Files:**
- Modify: `frontend/src/views/Cmdb/HostList.vue`

- [ ] **Step 1: Update HostList to use MultiTabTerminal**

In `frontend/src/views/Cmdb/HostList.vue`, make these changes:

1. Replace the import of `Terminal` with `MultiTabTerminal`:
```js
// Replace: import Terminal from '@/components/K8s/Terminal.vue'
import MultiTabTerminal from '@/components/Cmdb/MultiTabTerminal.vue'
```

2. Add a ref for the multi-tab terminal:
```js
const multiTabTerminal = ref(null)
```

3. Replace the existing terminal dialog logic. Find the `terminalDialog` ref and the `openTerminal` function, then replace them:

```js
const terminalDialog = ref(false)
const terminalHostId = ref(null)
const terminalWsUrl = ref('')
const terminalHostName = ref('')

const openTerminal = (row) => {
  const wsUrl = getTerminalConnectWsUrl(row.id)
  if (!terminalDialog.value) {
    terminalHostId.value = row.id
    terminalWsUrl.value = wsUrl
    terminalHostName.value = row.hostname || row.ip
    terminalDialog.value = true
  } else {
    // Dialog already open — add a new tab
    multiTabTerminal.value?.addTab(row.id, row.hostname || row.ip, wsUrl)
  }
}

const handleTerminalRequestConnect = () => {
  // Close dialog and re-open host selector
  terminalDialog.value = false
  // User picks a host from the table to add a new tab
}

const handleTerminalAllClosed = () => {
  terminalDialog.value = false
}
```

4. Replace the terminal dialog template. Find the existing `<el-dialog>` that contains `<Terminal>` and replace it with:

```html
<el-dialog
  v-model="terminalDialog"
  title="SSH 终端"
  width="80%"
  top="5vh"
  :close-on-click-modal="false"
  destroy-on-close
  class="terminal-dialog"
>
  <div style="height: 65vh;">
    <MultiTabTerminal
      ref="multiTabTerminal"
      :visible="terminalDialog"
      @request-connect="handleTerminalRequestConnect"
      @all-closed="handleTerminalAllClosed"
    />
  </div>
</el-dialog>
```

5. Add the `onMounted` hook to handle `terminalHostId` query param from dashboard "My Hosts" click:

```js
import { useRoute } from 'vue-router'
const route = useRoute()

onMounted(() => {
  if (route.query.terminalHostId) {
    const hostId = Number(route.query.terminalHostId)
    const host = hostList.value.find(h => h.id === hostId)
    if (host) {
      openTerminal(host)
    }
  }
})
```

6. After the dialog opens, auto-add the first tab using a `nextTick`:

```js
import { nextTick } from 'vue'

// In the openTerminal function, when opening dialog for first time:
watch(terminalDialog, async (val) => {
  if (val && terminalWsUrl.value && multiTabTerminal.value) {
    await nextTick()
    multiTabTerminal.value.addTab(
      terminalHostId.value,
      terminalHostName.value,
      terminalWsUrl.value
    )
  }
})
```

7. Add the import for `getTerminalConnectWsUrl`:
```js
import { getTerminalConnectWsUrl } from '@/api/cmdb/terminal'
```

8. Add dialog CSS:
```css
:deep(.terminal-dialog .el-dialog__body) {
  padding: 0;
  overflow: hidden;
}
```

- [ ] **Step 2: Test multi-tab terminal**

Run: `cd frontend && npm run dev`

Test flow:
1. Open `/cmdb/hosts`
2. Click "终端" on a host — dialog opens with one tab
3. Click "终端" on another host — new tab added
4. Switch between tabs — terminal content preserved
5. Close a tab via × button — remaining tabs stay
6. Close all tabs — dialog closes
7. Test keyboard shortcuts: Ctrl+Shift+T, Ctrl+Tab, Ctrl+W

- [ ] **Step 3: Commit**

```bash
cd frontend
git add src/views/Cmdb/HostList.vue
git commit -m "feat(cmdb): integrate multi-tab terminal into host list page"
```

---

## Self-Review Checklist

1. **Spec coverage:** Tab bar with hostname + close — Task 1. "+" button — Task 1. Keyboard shortcuts (Ctrl+Shift+T, Ctrl+Tab, Ctrl+W) — Task 1. Each tab own WebSocket — Task 1 (inherent from Terminal.vue). Backend no changes — confirmed.

2. **Placeholder scan:** No TBD/TODO. All code complete.

3. **Type consistency:** `addTab(hostId, hostname, wsUrl)` in expose matches usage in HostList. `getTerminalConnectWsUrl` imported from existing `api/cmdb/terminal.js`.
