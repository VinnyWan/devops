import { ref } from 'vue'
import { useMessage } from 'naive-ui'
import http from '@/api/index'
import type { ApiResponse } from '@/types/api'

export interface UseYamlModalOptions {
  /** 获取 YAML 成功后的回调 */
  onOpenSuccess?: () => void
  /** 保存 YAML 成功后的回调 */
  onSaveSuccess?: () => void
  /** 默认是否只读 */
  readonly?: boolean
}

/**
 * YAML 弹窗组合式函数
 * 配合 YamlTerminalModal 组件使用
 */
export function useYamlModal(options: UseYamlModalOptions = {}) {
  const message = useMessage()

  // 状态
  const show = ref(false)
  const content = ref('')
  const loading = ref(false)
  const readonly = ref(options.readonly ?? false)
  const title = ref('')
  const currentResource = ref<{
    type: string
    namespace: string
    name: string
    clusterId?: number
  } | null>(null)

  /**
   * 打开 YAML 弹窗（使用通用 API）
   */
  async function open(
    clusterId: number,
    resourceType: string,
    namespace: string,
    name: string,
    isReadonly: boolean = false
  ) {
    currentResource.value = { type: resourceType, namespace, name, clusterId }
    title.value = `${resourceTypeNames[resourceType] || resourceType} / ${name}`
    readonly.value = isReadonly
    loading.value = true
    show.value = true
    content.value = ''

    try {
      const res = await http.get<ApiResponse<{ yaml: string }>>(
        '/api/v1/k8s/resource/yaml',
        {
          params: {
            clusterId,
            resourceType,
            namespace: namespace || undefined,
            name
          }
        }
      )
      content.value = res.data?.data?.yaml || ''
      options.onOpenSuccess?.()
    } catch (error: any) {
      message.error('获取 YAML 失败: ' + (error.message || '未知错误'))
      content.value = ''
    } finally {
      loading.value = false
    }
  }

  /**
   * 打开 YAML 弹窗（使用自定义获取函数）
   */
  async function openWithFetch(
    fetchFn: () => Promise<string>,
    info: { type: string; namespace: string; name: string },
    isReadonly: boolean = false
  ) {
    currentResource.value = info
    title.value = `${resourceTypeNames[info.type] || info.type} / ${info.name}`
    readonly.value = isReadonly
    loading.value = true
    show.value = true
    content.value = ''

    try {
      content.value = await fetchFn()
      options.onOpenSuccess?.()
    } catch (error: any) {
      message.error('获取 YAML 失败: ' + (error.message || '未知错误'))
      content.value = ''
    } finally {
      loading.value = false
    }
  }

  /**
   * 保存 YAML（使用自定义保存函数）
   */
  async function save(saveFn: (yaml: string) => Promise<void>) {
    if (!content.value) {
      message.warning('YAML 内容为空')
      return
    }

    loading.value = true
    try {
      await saveFn(content.value)
      message.success('保存成功')
      show.value = false
      options.onSaveSuccess?.()
    } catch (error: any) {
      message.error('保存失败: ' + (error.message || '未知错误'))
    } finally {
      loading.value = false
    }
  }

  /**
   * 关闭弹窗
   */
  function close() {
    show.value = false
    content.value = ''
    currentResource.value = null
  }

  return {
    // 状态 - 直接返回 ref，组件可以直接绑定
    show,
    content,
    loading,
    title,
    readonly,
    currentResource,

    // 方法
    open,
    openWithFetch,
    save,
    close
  }
}

/**
 * 资源类型映射（用于显示友好名称）
 */
export const resourceTypeNames: Record<string, string> = {
  pod: 'Pod',
  deployment: 'Deployment',
  statefulset: 'StatefulSet',
  daemonset: 'DaemonSet',
  service: 'Service',
  ingress: 'Ingress',
  configmap: 'ConfigMap',
  secret: 'Secret',
  pvc: 'PersistentVolumeClaim',
  pv: 'PersistentVolume',
  namespace: 'Namespace',
  node: 'Node',
  job: 'Job',
  cronjob: 'CronJob'
}
