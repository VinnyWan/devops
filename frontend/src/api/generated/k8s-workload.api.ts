import http from '@/api/index'
import type { ApiResponse } from '@/types/api'

// StatefulSet APIs
export function k8sK8sStatefulSetListPost(params: any) {
  return http.get<ApiResponse<any>>('/v1/k8s/statefulset/list', { params })
}

export function k8sK8sStatefulSetYamlPost(params: any) {
  return http.get<ApiResponse<any>>('/v1/k8s/statefulset/yaml', { params })
}

export function k8sK8sStatefulSetYamlUpdatePost(params: any, data: any) {
  return http.post<ApiResponse<any>>('/v1/k8s/statefulset/yaml/update', data, { params })
}

export function k8sK8sStatefulSetScalePost(data: any) {
  return http.post<ApiResponse<any>>('/v1/k8s/statefulset/scale', data)
}

export function k8sK8sStatefulSetRestartPost(data: any) {
  return http.post<ApiResponse<any>>('/v1/k8s/statefulset/restart', data)
}

export function k8sK8sStatefulSetDeletePost(data: any) {
  return http.post<ApiResponse<any>>('/v1/k8s/statefulset/delete', data)
}

// DaemonSet APIs
export function k8sK8sDaemonSetListPost(params: any) {
  return http.get<ApiResponse<any>>('/v1/k8s/daemonset/list', { params })
}

export function k8sK8sDaemonSetYamlPost(params: any) {
  return http.get<ApiResponse<any>>('/v1/k8s/daemonset/yaml', { params })
}

export function k8sK8sDaemonSetYamlUpdatePost(params: any, data: any) {
  return http.post<ApiResponse<any>>('/v1/k8s/daemonset/yaml/update', data, { params })
}

export function k8sK8sDaemonSetRestartPost(data: any) {
  return http.post<ApiResponse<any>>('/v1/k8s/daemonset/restart', data)
}

export function k8sK8sDaemonSetDeletePost(data: any) {
  return http.post<ApiResponse<any>>('/v1/k8s/daemonset/delete', data)
}
