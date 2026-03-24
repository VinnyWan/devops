export interface ListData<T> {
  total: number
  items: T[]
}

export interface AlertRule {
  id: number
  name: string
  expr: string
  severity: string
  enabled: boolean
  cluster: string
  updatedAt: string
  description: string
}

export interface AlertHistory {
  id: number
  ruleId: number
  ruleName: string
  status: string
  severity: string
  summary: string
  startsAt: string
  endsAt: string
  cluster: string
  namespace: string
  instance: string
  fingerprint: string
}

export interface AlertSilence {
  id: number
  ruleId: number
  reason: string
  startsAt: string
  endsAt: string
  createdBy: string
  updatedAt: string
}

export interface NotificationChannel {
  id: number
  name: string
  type: string
  target: string
  enabled: boolean
  updatedAt: string
}

export interface AlertmanagerConfig {
  endpoint: string
  apiPath: string
  timeoutSeconds: number
  username: string
  password: string
  bearerToken: string
  tlsInsecureSkipVerify: boolean
  updatedAt: string
}

export interface LogEntry {
  id: number
  cluster: string
  namespace: string
  pod: string
  source: string
  level: string
  message: string
  createdAt: string
}

export interface QueryPoint {
  timestamp: string
  value: number
}

export interface QuerySeries {
  labels: Record<string, string>
  points: QueryPoint[]
}

export interface QueryResult {
  metric: string
  start: string
  end: string
  step: string
  series: QuerySeries[]
}

export interface PrometheusConfig {
  endpoint: string
  queryPath: string
  timeoutSeconds: number
  username: string
  password: string
  bearerToken: string
  tlsInsecureSkipVerify: boolean
  updatedAt: string
}

export interface HarborProject {
  id: number
  name: string
  public: boolean
  createdAt: string
  updatedAt: string
}

export interface RepositoryImage {
  id: number
  projectName: string
  repository: string
  tag: string
  digest: string
  size: number
  pushedAt: string
}

export interface HarborConfig {
  endpoint: string
  project: string
  username: string
  password: string
  robotToken: string
  timeoutSeconds: number
  tlsInsecureSkipVerify: boolean
  updatedAt: string
}

export interface Pipeline {
  id: number
  name: string
  status: string
  branch: string
  lastRunAt: string
  createdAt: string
  updatedAt: string
}

export interface PipelineLog {
  id: number
  pipelineId: number
  stage: string
  level: string
  message: string
  createdAt: string
}

export interface TemplateStage {
  name: string
  kind: string
  order: number
  parameters: Record<string, string>
}

export interface PipelineTemplate {
  id: number
  name: string
  description: string
  source: string
  stages: TemplateStage[]
  createdAt: string
  updatedAt: string
}

export interface PipelineRun {
  id: number
  pipelineId: number
  pipeline: string
  templateId: number
  template: string
  branch: string
  environment: string
  triggerType: string
  commitId: string
  operator: string
  status: string
  stages: TemplateStage[]
  createdAt: string
}

export interface JenkinsConfig {
  endpoint: string
  username: string
  apiToken: string
  defaultJob: string
  timeoutSeconds: number
  tlsInsecureSkipVerify: boolean
  updatedAt: string
}

export interface Application {
  id: number
  name: string
  namespace: string
  status: string
  createdAt: string
  updatedAt: string
}

export interface AppTemplate {
  id: number
  name: string
  type: string
  description: string
  environment: string[]
  variables: Record<string, string>
  createdAt: string
  updatedAt: string
}

export interface ApplicationDeployment {
  id: number
  appId: number
  appName: string
  templateId: number
  templateName: string
  cluster: string
  environment: string
  namespace: string
  version: string
  status: string
  operator: string
  variables: Record<string, string>
  createdAt: string
}

export interface ApplicationVersion {
  id: number
  appId: number
  version: string
  cluster: string
  environment: string
  image: string
  status: string
  operator: string
  createdAt: string
}

export interface TopologyNode {
  id: string
  name: string
  kind: string
  status: string
  cluster: string
  metadata: string
}

export interface TopologyEdge {
  from: string
  to: string
  kind: string
}

export interface ApplicationTopology {
  appId: number
  appName: string
  environment: string
  nodes: TopologyNode[]
  edges: TopologyEdge[]
  lastSyncTime: string
}

export interface AuditLogRecord {
  id: number
  userId: number
  username: string
  operation: string
  method: string
  path: string
  params: string
  result?: string
  errorMessage?: string
  ip: string
  status: number
  latency: number
  retentionDays: number
  requestAt?: string
  createdAt: string
}
