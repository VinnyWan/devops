import type { PageParams } from '@/types/api'
import type {
  Cluster,
  ClusterForm,
  WorkloadCounts,
  NetworkCounts,
  StorageCounts,
  EventListResponse,
} from '@/types/cluster'
import {
  k8sClusterCreatePost,
  k8sClusterDeletePost,
  k8sClusterDetailPost,
  k8sClusterEventsPost,
  k8sClusterHealthPost,
  k8sClusterListPost,
  k8sClusterStatsNetworkPost,
  k8sClusterStatsStoragePost,
  k8sClusterStatsWorkloadPost,
  k8sClusterUpdatePost,
} from '@/api/generated/cluster.api'
import { unwrapResponseData } from '@/api/service'

export async function getClusterList(params: PageParams = { page: 1, pageSize: 10 }) {
  const data = await unwrapResponseData<Cluster[] | null>(
    k8sClusterListPost(params as Parameters<typeof k8sClusterListPost>[0]),
  )
  return data ?? []
}

export function createCluster(data: ClusterForm) {
  return unwrapResponseData<unknown>(
    k8sClusterCreatePost(data as Parameters<typeof k8sClusterCreatePost>[0]),
  )
}

export function deleteCluster(id: number) {
  return unwrapResponseData<unknown>(
    k8sClusterDeletePost({ id } as Parameters<typeof k8sClusterDeletePost>[0]),
  )
}

export function updateCluster(data: Partial<ClusterForm> & { id: number }) {
  return unwrapResponseData<unknown>(
    k8sClusterUpdatePost(data as Parameters<typeof k8sClusterUpdatePost>[0]),
  )
}

export function checkClusterHealth(id: number) {
  return unwrapResponseData<{ healthy: boolean; status: string; error: string }>(
    k8sClusterHealthPost({ id } as Parameters<typeof k8sClusterHealthPost>[0]),
  )
}

export function getClusterDetail(id: number) {
  return unwrapResponseData<Cluster>(
    k8sClusterDetailPost({ id } as Parameters<typeof k8sClusterDetailPost>[0]),
  )
}

export function getClusterEvents(id: number, params?: PageParams & { keyword?: string }) {
  return unwrapResponseData<EventListResponse>(
    k8sClusterEventsPost({ id, ...params } as Parameters<typeof k8sClusterEventsPost>[0]),
  )
}

export function getWorkloadStats(id: number) {
  return unwrapResponseData<WorkloadCounts>(
    k8sClusterStatsWorkloadPost({ id } as Parameters<typeof k8sClusterStatsWorkloadPost>[0]),
  )
}

export function getNetworkStats(id: number) {
  return unwrapResponseData<NetworkCounts>(
    k8sClusterStatsNetworkPost({ id } as Parameters<typeof k8sClusterStatsNetworkPost>[0]),
  )
}

export function getStorageStats(id: number) {
  return unwrapResponseData<StorageCounts>(
    k8sClusterStatsStoragePost({ id } as Parameters<typeof k8sClusterStatsStoragePost>[0]),
  )
}
