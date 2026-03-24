// 日期格式化
export function formatDate(date: string | Date, fmt = 'YYYY-MM-DD HH:mm:ss'): string {
  const d = new Date(date)
  const map: Record<string, number> = {
    YYYY: d.getFullYear(),
    MM: d.getMonth() + 1,
    DD: d.getDate(),
    HH: d.getHours(),
    mm: d.getMinutes(),
    ss: d.getSeconds(),
  }
  return fmt.replace(/YYYY|MM|DD|HH|mm|ss/g, (match) => String(map[match]).padStart(2, '0'))
}

// 字节格式化
export function formatBytes(bytes: number, decimals = 2): string {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return `${parseFloat((bytes / Math.pow(k, i)).toFixed(decimals))} ${sizes[i]}`
}

// 百分比格式化
export function formatPercent(value: number, decimals = 1): string {
  return `${value.toFixed(decimals)}%`
}

// 相对时间
export function formatAge(dateStr: string): string {
  const diff = Date.now() - new Date(dateStr).getTime()
  const minutes = Math.floor(diff / 60000)
  if (minutes < 60) return `${minutes}m`
  const hours = Math.floor(minutes / 60)
  if (hours < 24) return `${hours}h`
  const days = Math.floor(hours / 24)
  return `${days}d`
}
