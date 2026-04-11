<template>
  <div class="log-viewer">
    <div class="log-content" ref="logRef">
      <pre>{{ logs }}</pre>
    </div>
  </div>
</template>

<script setup>
import { ref, watch, nextTick } from 'vue'

const props = defineProps({
  logs: { type: String, default: '' },
  autoScroll: { type: Boolean, default: true }
})

const logRef = ref(null)

watch(() => props.logs, async () => {
  if (props.autoScroll) {
    await nextTick()
    if (logRef.value) {
      logRef.value.scrollTop = logRef.value.scrollHeight
    }
  }
})
</script>

<style scoped>
.log-viewer {
  background: #1e1e1e;
  border-radius: 4px;
  overflow: hidden;
}
.log-content {
  max-height: 500px;
  overflow-y: auto;
  padding: 12px;
}
.log-content pre {
  color: #d4d4d4;
  font-family: monospace;
  font-size: 13px;
  margin: 0;
  white-space: pre-wrap;
  word-break: break-all;
}
</style>
