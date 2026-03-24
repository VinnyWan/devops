import type { Deployment, Pod, Service, Ingress, ConfigMap } from '@/types/k8s'
import {
  k8sK8sConfigmapListPost,
  k8sK8sDeploymentListPost,
  k8sK8sIngressListPost,
  k8sK8sPodListPost,
  k8sK8sServiceListPost,
} from '@/api/generated/k8s-resource.api'

export function getDeployments(clusterId: number, namespace?: string) {
  return k8sK8sDeploymentListPost({
    clusterId,
    namespace: namespace || 'default',
  } as Parameters<typeof k8sK8sDeploymentListPost>[0]) as unknown as Promise<{
    data: { code: number; message: string; data: Deployment[] }
  }>
}

export function getPods(clusterId: number, namespace?: string) {
  return k8sK8sPodListPost({
    clusterId,
    namespace: namespace || 'default',
  } as Parameters<typeof k8sK8sPodListPost>[0]) as unknown as Promise<{
    data: { code: number; message: string; data: Pod[] }
  }>
}

export function getServices(clusterId: number, namespace?: string) {
  return k8sK8sServiceListPost({
    clusterId,
    namespace: namespace || 'default',
  } as Parameters<typeof k8sK8sServiceListPost>[0]) as unknown as Promise<{
    data: { code: number; message: string; data: Service[] }
  }>
}

export function getIngresses(clusterId: number, namespace?: string) {
  return k8sK8sIngressListPost({
    clusterId,
    namespace: namespace || 'default',
  } as Parameters<typeof k8sK8sIngressListPost>[0]) as unknown as Promise<{
    data: { code: number; message: string; data: Ingress[] }
  }>
}

export function getConfigMaps(clusterId: number, namespace?: string) {
  return k8sK8sConfigmapListPost({
    clusterId,
    namespace: namespace || 'default',
  } as Parameters<typeof k8sK8sConfigmapListPost>[0]) as unknown as Promise<{
    data: { code: number; message: string; data: ConfigMap[] }
  }>
}
