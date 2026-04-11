<template>
  <div class="terminal" ref="terminalRef"></div>
</template>

<script setup>
import { ref, onMounted } from 'vue'

const props = defineProps({
  wsUrl: { type: String, required: true }
})

const terminalRef = ref(null)

onMounted(() => {
  // Terminal 实现需要 xterm.js 库
  // 这里提供基础结构,实际使用时需要安装 xterm 依赖
  console.log('Terminal WebSocket URL:', props.wsUrl)
})
</script>

<style scoped>
.terminal {
  width: 100%;
  height: 500px;
  background: #000;
  border-radius: 4px;
}
</style>
