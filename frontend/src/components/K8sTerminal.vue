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
  ws.onmessage = (e) => terminal?.write(e.data)
  ws.onclose = () => terminal?.write('\r\n连接已断开\r\n')

  terminal.onData((data) => {
    ws?.send(data)
  })

  window.addEventListener('resize', handleResize)
})

function handleResize() {
  fitAddon?.fit()
}

onUnmounted(() => {
  window.removeEventListener('resize', handleResize)
  ws?.close()
  terminal?.dispose()
})
</script>

<template>
  <div ref="termRef" style="height: 100%; min-height: 400px" />
</template>
