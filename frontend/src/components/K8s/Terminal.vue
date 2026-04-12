<template>
  <div ref="terminalContainer" class="terminal-container"></div>
</template>

<script setup>
import { ref, onMounted, onBeforeUnmount, watch, nextTick } from 'vue'
import { Terminal } from '@xterm/xterm'
import { FitAddon } from '@xterm/addon-fit'
import { WebLinksAddon } from '@xterm/addon-web-links'
import '@xterm/xterm/css/xterm.css'

const props = defineProps({
  wsUrl: { type: String, required: true },
  visible: { type: Boolean, default: false }
})

const emit = defineEmits(['error'])

const terminalContainer = ref(null)
let terminal = null
let fitAddon = null
let ws = null
let initialized = false

const initTerminal = () => {
  if (initialized || !terminalContainer.value) return
  initialized = true

  terminal = new Terminal({
    cursorBlink: true,
    fontSize: 14,
    fontFamily: 'Consolas, "Courier New", monospace',
    theme: {
      background: '#1e1e1e',
      foreground: '#d4d4d4',
      cursor: '#d4d4d4'
    }
  })

  fitAddon = new FitAddon()
  terminal.loadAddon(fitAddon)
  terminal.loadAddon(new WebLinksAddon())

  terminal.open(terminalContainer.value)
  nextTick(() => {
    fitAddon.fit()
  })

  terminal.onData((data) => {
    if (ws && ws.readyState === WebSocket.OPEN) {
      const msg = JSON.stringify({ operation: 'stdin', data: data })
      ws.send(msg)
    }
  })

  terminal.onResize(({ cols, rows }) => {
    if (ws && ws.readyState === WebSocket.OPEN) {
      ws.send(JSON.stringify({ operation: 'resize', cols, rows }))
    }
  })

  connectWebSocket()
}

const connectWebSocket = () => {
  if (!props.wsUrl) return

  const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
  const host = window.location.host
  const wsFullUrl = `${protocol}//${host}${props.wsUrl}`

  ws = new WebSocket(wsFullUrl)

  ws.onopen = () => {
    terminal.write('\x1b[32mConnected.\x1b[0m\r\n')
  }

  ws.onmessage = (event) => {
    const data = JSON.parse(event.data)
    if (data.operation === 'stdout' || data.operation === 'stderr') {
      terminal.write(data.data)
    }
  }

  ws.onerror = () => {
    terminal.write('\r\n\x1b[31mConnection error.\x1b[0m\r\n')
    emit('error', 'WebSocket connection error')
  }

  ws.onclose = () => {
    terminal.write('\r\n\x1b[33mConnection closed.\x1b[0m\r\n')
  }
}

const cleanup = () => {
  if (ws) {
    ws.close()
    ws = null
  }
  if (terminal) {
    terminal.dispose()
    terminal = null
  }
  initialized = false
}

watch(() => props.visible, (val) => {
  if (val) {
    nextTick(() => {
      initTerminal()
    })
  } else {
    cleanup()
  }
})

onMounted(() => {
  if (props.visible) {
    nextTick(() => {
      initTerminal()
    })
  }
})

onBeforeUnmount(() => {
  cleanup()
})

defineExpose({ fit: () => fitAddon?.fit() })
</script>

<style scoped>
.terminal-container {
  width: 100%;
  height: 60vh;
  background: #1e1e1e;
  border-radius: 4px;
  padding: 4px;
}
</style>
