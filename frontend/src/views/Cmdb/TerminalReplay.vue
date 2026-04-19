<template>
  <div class="page-container" v-loading="loading">
    <div class="page-header">
      <div>
        <h3>终端会话回放</h3>
        <p class="page-subtitle">会话 ID: {{ sessionId }}</p>
      </div>
      <div class="actions">
        <el-button @click="goBack">返回列表</el-button>
        <el-button type="primary" @click="togglePlayback" :disabled="!recording.events.length">
          {{ playing ? '暂停' : '播放' }}
        </el-button>
        <el-button @click="restartPlayback" :disabled="!recording.events.length">重新开始</el-button>
        <el-select v-model="speed" style="width: 120px;" :disabled="!recording.events.length">
          <el-option label="1x" :value="1" />
          <el-option label="2x" :value="2" />
          <el-option label="4x" :value="4" />
        </el-select>
      </div>
    </div>

    <el-alert v-if="errorText" :title="errorText" type="error" :closable="false" style="margin-bottom: 16px;" />

    <el-descriptions :column="3" border class="meta-card">
      <el-descriptions-item label="主机名">{{ detail.hostName || '-' }}</el-descriptions-item>
      <el-descriptions-item label="主机 IP">{{ detail.hostIp || '-' }}</el-descriptions-item>
      <el-descriptions-item label="用户名">{{ detail.username || '-' }}</el-descriptions-item>
      <el-descriptions-item label="客户端 IP">{{ detail.clientIp || '-' }}</el-descriptions-item>
      <el-descriptions-item label="状态">
        <el-tag :type="statusTagType(detail.status)">{{ statusText(detail.status) }}</el-tag>
      </el-descriptions-item>
      <el-descriptions-item label="关闭原因/触发类型">{{ closeReasonText(detail) }}</el-descriptions-item>
      <el-descriptions-item label="文件大小">{{ formatFileSize(detail.fileSize) }}</el-descriptions-item>
      <el-descriptions-item label="开始时间">{{ formatTime(detail.startedAt) }}</el-descriptions-item>
      <el-descriptions-item label="结束时间">{{ formatTime(detail.finishedAt) }}</el-descriptions-item>
      <el-descriptions-item label="时长">{{ formatDuration(detail.duration) }}</el-descriptions-item>
    </el-descriptions>

    <div class="replay-status">
      <span>进度: {{ currentIndex }}/{{ recording.events.length }}</span>
      <span>播放时长: {{ currentPlaybackSeconds }}s / {{ totalPlaybackSeconds }}s</span>
      <span>录屏尺寸: {{ recording.width || '-' }} x {{ recording.height || '-' }}</span>
      <span>速度: {{ speed }}x</span>
    </div>

    <el-empty v-if="!loading && !errorText && !recording.events.length" description="暂无回放数据" />

    <Terminal v-else-if="!errorText" ref="terminalRef" :visible="true" readonly />
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, onBeforeUnmount, nextTick, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import Terminal from '@/components/K8s/Terminal.vue'
import { getTerminalSessionDetail, getTerminalRecording } from '@/api/cmdb/terminal'
import { formatTime } from '@/utils/format'

const route = useRoute()
const router = useRouter()
const sessionId = Number(route.params.id)

const loading = ref(false)
const playing = ref(false)
const speed = ref(1)
const currentIndex = ref(0)
const errorText = ref('')
const terminalRef = ref()
const detail = ref({})
const recording = reactive({ width: 0, height: 0, events: [] })
const totalPlaybackSeconds = computed(() => {
  if (!recording.events.length) return 0
  return Number(recording.events[recording.events.length - 1]?.time || 0)
})
const currentPlaybackSeconds = computed(() => {
  if (!recording.events.length || currentIndex.value <= 0) return 0
  const event = recording.events[Math.min(currentIndex.value - 1, recording.events.length - 1)]
  return Number(event?.time || 0)
})
let replayTimer = null

const clearReplayTimer = () => {
  if (replayTimer) {
    clearTimeout(replayTimer)
    replayTimer = null
  }
}

const formatDuration = (duration) => {
  const totalSeconds = Number(duration || 0)
  if (totalSeconds <= 0) return '0 秒'
  if (totalSeconds < 60) return `${totalSeconds} 秒`
  const minutes = Math.floor(totalSeconds / 60)
  const seconds = totalSeconds % 60
  return seconds ? `${minutes} 分 ${seconds} 秒` : `${minutes} 分`
}

const formatFileSize = (size) => {
  const value = Number(size || 0)
  if (value <= 0) return '0 B'
  if (value < 1024) return `${value} B`
  if (value < 1024 * 1024) return `${(value / 1024).toFixed(2)} KB`
  return `${(value / 1024 / 1024).toFixed(2)} MB`
}

const statusTagType = (val) => ({
  active: 'success',
  closed: 'info',
  interrupted: 'danger',
  idle_timeout: 'warning',
  max_duration: 'warning'
}[val] || 'info')
const statusText = (val) => ({
  active: '活跃',
  closed: '已关闭',
  interrupted: '已中断',
  idle_timeout: '空闲超时',
  max_duration: '时长超限'
}[val] || val || '-')
const closeReasonFallback = (status) => ({
  active: '会话进行中',
  closed: '用户主动关闭或连接正常结束',
  interrupted: '连接异常中断',
  idle_timeout: '空闲超时自动断开',
  max_duration: '会话时长超限自动断开'
}[status] || '-')
const closeReasonText = (session) => session?.closeReason || closeReasonFallback(session?.status)

const renderNextEvent = () => {
  if (!playing.value || currentIndex.value >= recording.events.length) {
    playing.value = false
    clearReplayTimer()
    return
  }

  const event = recording.events[currentIndex.value]
  if (event.type === 'o' || event.type === 'stdout' || event.type === 'stderr') {
    terminalRef.value?.write(event.data || '')
  }
  currentIndex.value += 1

  if (currentIndex.value >= recording.events.length) {
    playing.value = false
    clearReplayTimer()
    return
  }

  const currentTime = Number(event.time || 0)
  const nextTime = Number(recording.events[currentIndex.value].time || 0)
  const delay = Math.max(0, (nextTime - currentTime) * 1000 / speed.value)
  replayTimer = setTimeout(renderNextEvent, delay)
}

const startPlayback = () => {
  if (!recording.events.length || currentIndex.value >= recording.events.length) {
    return
  }
  clearReplayTimer()
  playing.value = true
  if (currentIndex.value === 0) {
    terminalRef.value?.clear()
  }
  renderNextEvent()
}

const pausePlayback = () => {
  playing.value = false
  clearReplayTimer()
}

const togglePlayback = () => {
  if (playing.value) {
    pausePlayback()
  } else {
    startPlayback()
  }
}

const restartPlayback = () => {
  pausePlayback()
  currentIndex.value = 0
  terminalRef.value?.clear()
  startPlayback()
}

const resetPlaybackState = () => {
  pausePlayback()
  currentIndex.value = 0
  terminalRef.value?.clear()
}

const resetReplayData = () => {
  recording.width = 0
  recording.height = 0
  recording.events = []
  detail.value = {}
}

const handleResize = () => {
  terminalRef.value?.fit()
}

const loadData = async () => {
  loading.value = true
  errorText.value = ''
  resetPlaybackState()
  resetReplayData()
  try {
    const [detailRes, recordingRes] = await Promise.all([
      getTerminalSessionDetail({ id: sessionId }),
      getTerminalRecording({ id: sessionId })
    ])
    detail.value = detailRes.data || {}
    recording.width = recordingRes.data?.width || 0
    recording.height = recordingRes.data?.height || 0
    recording.events = recordingRes.data?.events || []
    currentIndex.value = 0
    await nextTick()
    terminalRef.value?.clear()
    terminalRef.value?.fit()
  } catch (error) {
    resetReplayData()
    errorText.value = error.response?.data?.message || error.message || '加载终端回放失败'
  } finally {
    loading.value = false
  }
}

const goBack = () => {
  router.push('/cmdb/terminal/sessions')
}

onMounted(() => {
  window.addEventListener('resize', handleResize)
  loadData()
})

onBeforeUnmount(() => {
  pausePlayback()
  window.removeEventListener('resize', handleResize)
})
</script>

<style scoped>
.page-container { background: #fff; border-radius: 4px; padding: 24px; }
.page-header { display: flex; justify-content: space-between; align-items: flex-start; gap: 16px; margin-bottom: 16px; }
.page-header h3 { margin: 0; font-size: 18px; font-weight: 500; }
.page-subtitle { margin: 8px 0 0; color: #909399; }
.actions { display: flex; gap: 8px; align-items: center; flex-wrap: wrap; }
.meta-card { margin-bottom: 16px; }
.replay-status { display: flex; gap: 24px; flex-wrap: wrap; color: #606266; margin-bottom: 12px; }
</style>
