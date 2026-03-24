// K8s 资源相关类型
export interface K8sEvent {
  type: string
  reason: string
  message: string
  source: string
  involvedObject: string
  count: number
  firstTimestamp: string
  lastTimestamp: string
}

export interface Pod {
  name: string
  namespace: string
  status: string
  nodeName: string
  podIP: string
  restarts: number
  age: string
  containers: Container[]
}

export interface Container {
  name: string
  image: string
  ready: boolean
  restartCount: number
  state: string
}

export interface Deployment {
  name: string
  namespace: string
  replicas: number
  readyReplicas: number
  updatedReplicas: number
  availableReplicas: number
  age: string
}

export interface Service {
  name: string
  namespace: string
  type: string
  clusterIP: string
  ports: string
  age: string
}

export interface Ingress {
  name: string
  namespace: string
  hosts: string
  address: string
  ports: string
  age: string
}

export interface ConfigMap {
  name: string
  namespace: string
  dataCount: number
  age: string
}
