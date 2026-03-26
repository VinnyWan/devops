<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { Terminal } from '@xterm/xterm'
import { FitAddon } from '@xterm/addon-fit'
import '@xterm/xterm/css/xterm.css'

const props = defineProps<{
  wsUrl: string
}>()

const termRef = ref<HTMLDivElement>()
let terminal: Terminal | null = null
let fitAddon: FitAddon | null = null
let ws: WebSocket | null = null

onMounted(() => {
  if (!termRef.value) return

  terminal = new Terminal({
    cursorBlink: true,
    fontSize: 14,
    fontFamily: 'Menlo, Monaco, Consolas, monospace',
    theme: {
      background: '#1e1e1e',
      foreground: '#d4d4d4',
    },
  })

  fitAddon = new FitAddon()
  terminal.loadAddon(fitAddon)
  terminal.open(termRef.value)
  fitAddon.fit()

  // WebSocket 连接
  ws = new WebSocket(props.wsUrl)

  ws.onopen = () => {
    terminal?.writeln('\x1b[32m连接已建立\x1b[0m')
  }

  ws.onmessage = (e) => {
    try {
      const message = JSON.parse(e.data)
      if (message.operation === 'stdout' && message.data) {
        terminal?.write(message.data)
      } else if (message.operation === 'error') {
        terminal?.writeln(`\x1b[31m错误: ${message.data}\x1b[0m`)
      }
    } catch {
      // 如果不是JSON格式，直接写入
      terminal?.write(e.data)
    }
  }

  ws.onclose = () => {
    terminal?.writeln('\r\n\x1b[33m连接已断开\x1b[0m')
  }

  ws.onerror = (error) => {
    console.error('WebSocket error:', error)
    terminal?.writeln('\x1b[31m连接错误\x1b[0m')
  }

  // 用户输入时发送到后端
  terminal.onData((data) => {
    if (ws && ws.readyState === WebSocket.OPEN) {
      ws.send(JSON.stringify({
        operation: 'stdin',
        data: data
      }))
    }
  })

  // 终端resize时发送尺寸变化
  terminal.onResize(({ cols, rows }) => {
    if (ws && ws.readyState === WebSocket.OPEN) {
      ws.send(JSON.stringify({
        operation: 'resize',
        cols: cols,
        rows: rows
      }))
    }
  })

  window.addEventListener('resize', handleResize)
})

function handleResize() {
  fitAddon?.fit()
}

onUnmounted(() => {
  window.removeEventListener('resize', handleResize)
  if (ws) {
    ws.close()
  }
  if (terminal) {
    terminal.dispose()
  }
})
</script>

<template>
  <div ref="termRef" style="height: 100%; min-height: 400px" />
</template>
