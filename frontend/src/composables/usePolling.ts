import { ref, onMounted, onUnmounted } from 'vue'

// 轮询 composable，支持 AbortController 取消
export function usePolling(fn: (signal: AbortSignal) => Promise<void>, interval = 30000) {
  const loading = ref(false)
  let timer: ReturnType<typeof setInterval> | null = null
  let controller: AbortController | null = null

  async function execute() {
    controller?.abort()
    controller = new AbortController()
    loading.value = true
    try {
      await fn(controller.signal)
    } catch (e: unknown) {
      if (e instanceof DOMException && e.name === 'AbortError') return
      console.error(e)
    } finally {
      loading.value = false
    }
  }

  function start() {
    execute()
    timer = setInterval(execute, interval)
  }

  function stop() {
    controller?.abort()
    if (timer) {
      clearInterval(timer)
      timer = null
    }
  }

  onMounted(start)
  onUnmounted(stop)

  return { loading, execute, stop }
}
