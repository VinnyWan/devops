import dayjs from 'dayjs'

export const formatTime = (t) => t ? dayjs(t).format('YYYY-MM-DD HH:mm:ss') : '-'

export const formatDuration = (seconds) => {
  const totalSeconds = Number(seconds || 0)
  if (totalSeconds <= 0) return '0 秒'
  if (totalSeconds < 60) return `${totalSeconds} 秒`
  const minutes = Math.floor(totalSeconds / 60)
  const secs = totalSeconds % 60
  return secs ? `${minutes} 分 ${secs} 秒` : `${minutes} 分`
}

export const formatFileSize = (bytes) => {
  const value = Number(bytes || 0)
  if (value <= 0) return '0 B'
  const units = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(value) / Math.log(1024))
  return (value / Math.pow(1024, i)).toFixed(1) + ' ' + units[i]
}

export const formatImages = (containers) => {
  if (!containers || !Array.isArray(containers)) return '-'
  return containers.map(c => c.image).join(', ')
}
