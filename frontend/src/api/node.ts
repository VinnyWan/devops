import type { ApiResponse } from '@/types/api'
import type { NodeInfo, Taint, NodeLabel } from '@/types/node'
import {
  k8sK8sNodeCordonPost,
  k8sK8sNodeDrainPost,
  k8sK8sNodeLabelsPost,
  k8sK8sNodesPost,
  k8sK8sNodeTaintsPost,
} from '@/api/generated/k8s-node.api'

export function getNodes(clusterId: number) {
  return k8sK8sNodesPost({ clusterId } as Parameters<
    typeof k8sK8sNodesPost
  >[0]) as unknown as Promise<{
    data: { code: number; message: string; data: NodeInfo[] }
  }>
}

export function cordonNode(clusterId: number, nodeName: string) {
  return k8sK8sNodeCordonPost({
    clusterId,
    name: nodeName,
    cordon: true,
  } as Parameters<typeof k8sK8sNodeCordonPost>[0]) as Promise<{ data: ApiResponse<unknown> }>
}

export function uncordonNode(clusterId: number, nodeName: string) {
  return k8sK8sNodeCordonPost({
    clusterId,
    name: nodeName,
    cordon: false,
  } as Parameters<typeof k8sK8sNodeCordonPost>[0]) as Promise<{ data: ApiResponse<unknown> }>
}

export function drainNode(clusterId: number, nodeName: string) {
  return k8sK8sNodeDrainPost({
    clusterId,
    name: nodeName,
  } as Parameters<typeof k8sK8sNodeDrainPost>[0]) as Promise<{ data: ApiResponse<unknown> }>
}

export function updateTaints(clusterId: number, nodeName: string, taints: Taint[]) {
  return k8sK8sNodeTaintsPost({
    clusterId,
    name: nodeName,
    taints,
  } as Parameters<typeof k8sK8sNodeTaintsPost>[0]) as Promise<{ data: ApiResponse<unknown> }>
}

export function updateLabels(clusterId: number, nodeName: string, labels: NodeLabel[]) {
  const labelMap = labels.reduce<Record<string, string>>((acc, current) => {
    acc[current.key] = current.value
    return acc
  }, {})
  return k8sK8sNodeLabelsPost({
    clusterId,
    name: nodeName,
    labels: labelMap,
  } as Parameters<typeof k8sK8sNodeLabelsPost>[0]) as Promise<{ data: ApiResponse<unknown> }>
}
