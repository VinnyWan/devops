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
  wsUrl: { type: String, default: '' },
  visible: { type: Boolean, default: false },
  readonly: { type: Boolean, default: false }
})

const emit = defineEmits(['error'])

const terminalContainer = ref(null)
let terminal = null
let fitAddon = null
let ws = null
let initialized = false
let closeMessageShown = false

const resolveWsUrl = (rawUrl) => {
  if (!rawUrl) return ''
  if (rawUrl.startsWith('ws://') || rawUrl.startsWith('wss://')) {
    return rawUrl
  }
  if (rawUrl.startsWith('http://') || rawUrl.startsWith('https://')) {
    return rawUrl.replace(/^http/, 'ws')
  }

  const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
  const host = window.location.host
  return `${protocol}//${host}${rawUrl}`
}

const write = (content = '') => {
  terminal?.write(content)
}

const clear = () => {
  terminal?.clear()
}

const fit = () => {
  fitAddon?.fit()
}

const handleSocketMessage = (payload) => {
  if (!payload || typeof payload !== 'object') {
    write(String(payload || ''))
    return
  }

  const content = payload.data || ''
  switch (payload.operation) {
    case 'stdout':
    case 'stderr':
      write(content)
      break
    case 'error':
      write(`\r\n\x1b[31m${content}\x1b[0m\r\n`)
      emit('error', content || 'WebSocket connection error')
      break
    case 'closed':
      closeMessageShown = true
      write(`\r\n\x1b[33m${content || 'Connection closed.'}\x1b[0m\r\n`)
      break
    default:
      if (typeof content === 'string') {
        write(content)
      }
  }
}

const connectWebSocket = () => {
  if (!props.wsUrl) return

  closeMessageShown = false
  const wsFullUrl = resolveWsUrl(props.wsUrl)
  if (!wsFullUrl) return

  ws = new WebSocket(wsFullUrl)

  ws.onopen = () => {
    if (!props.readonly) {
      write('\x1b[32mConnected.\x1b[0m\r\n')
    }
  }

  ws.onmessage = (event) => {
    if (typeof event.data !== 'string') {
      return
    }

    try {
      handleSocketMessage(JSON.parse(event.data))
    } catch {
      write(event.data)
    }
  }

  ws.onerror = () => {
    write('\r\n\x1b[31mConnection error.\x1b[0m\r\n')
    emit('error', 'WebSocket connection error')
  }

  ws.onclose = (event) => {
    if (closeMessageShown) {
      return
    }
    const reason = event.reason || 'Connection closed.'
    write(`\r\n\x1b[33m${reason}\x1b[0m\r\n`)
  }
}

const initTerminal = () => {
  if (initialized || !terminalContainer.value) return
  initialized = true

  terminal = new Terminal({
    cursorBlink: !props.readonly,
    disableStdin: props.readonly,
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
    fit()
  })

  terminal.onData((data) => {
    if (props.readonly) return
    if (ws && ws.readyState === WebSocket.OPEN) {
      ws.send(JSON.stringify({ operation: 'stdin', data }))
    }
  })

  terminal.onResize(({ cols, rows }) => {
    if (ws && ws.readyState === WebSocket.OPEN) {
      ws.send(JSON.stringify({ operation: 'resize', cols, rows }))
    }
  })

  if (props.wsUrl) {
    connectWebSocket()
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
  fitAddon = null
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

defineExpose({
  fit,
  write,
  clear
})
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
