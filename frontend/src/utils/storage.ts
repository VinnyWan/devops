const PREFIX = 'devops_'

export function getItem<T>(key: string): T | null {
  const raw = localStorage.getItem(PREFIX + key)
  if (!raw) return null
  try {
    return JSON.parse(raw) as T
  } catch {
    return null
  }
}

export function setItem(key: string, value: unknown): void {
  localStorage.setItem(PREFIX + key, JSON.stringify(value))
}

export function removeItem(key: string): void {
  localStorage.removeItem(PREFIX + key)
}
