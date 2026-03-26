import type { Deployment, Pod, Service, Ingress, ConfigMap } from '@/types/k8s'
import {
  k8sK8sConfigmapListPost,
  k8sK8sDeploymentListPost,
  k8sK8sIngressListPost,
  k8sK8sPodListPost,
  k8sK8sServiceListPost,
} from '@/api/generated/k8s-resource.api'
import { unwrapResponseData } from './service'

export function getDeployments(clusterId: number, namespace?: string) {
  return unwrapResponseData<Deployment[]>(
    k8sK8sDeploymentListPost({ clusterId, namespace: namespace || 'default' })
  )
}

export function getPods(clusterId: number, namespace?: string) {
  return unwrapResponseData<Pod[]>(
    k8sK8sPodListPost({ clusterId, namespace: namespace || 'default' })
  )
}

export function getServices(clusterId: number, namespace?: string) {
  return unwrapResponseData<Service[]>(
    k8sK8sServiceListPost({ clusterId, namespace: namespace || 'default' })
  )
}

export function getIngresses(clusterId: number, namespace?: string) {
  return unwrapResponseData<Ingress[]>(
    k8sK8sIngressListPost({ clusterId, namespace: namespace || 'default' })
  )
}

export function getConfigMaps(clusterId: number, namespace?: string) {
  return unwrapResponseData<ConfigMap[]>(
    k8sK8sConfigmapListPost({ clusterId, namespace: namespace || 'default' })
  )
}
