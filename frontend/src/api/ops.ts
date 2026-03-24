import http from '@/api/index'
import { unwrapResponseData } from '@/api/service'
import type {
  AlertmanagerConfig,
  AlertHistory,
  AlertRule,
  AlertSilence,
  Application,
  AppTemplate,
  ApplicationDeployment,
  ApplicationTopology,
  ApplicationVersion,
  AuditLogRecord,
  HarborConfig,
  HarborProject,
  JenkinsConfig,
  ListData,
  LogEntry,
  NotificationChannel,
  Pipeline,
  PipelineLog,
  PipelineRun,
  PipelineTemplate,
  PrometheusConfig,
  QueryResult,
  RepositoryImage,
} from '@/types/ops'

export function getAlertRules(params: { keyword?: string } = {}) {
  return unwrapResponseData<ListData<AlertRule>>(http.get('/v1/alert/rules', { params }))
}

export function getAlertHistory(params: { status?: string; start?: string; end?: string } = {}) {
  return unwrapResponseData<ListData<AlertHistory>>(http.get('/v1/alert/history', { params }))
}

export function toggleAlertRule(data: { id: number; enabled: boolean }) {
  return unwrapResponseData<AlertRule>(http.post('/v1/alert/rule/toggle', data))
}

export function getAlertSilences(params: { ruleId?: number } = {}) {
  return unwrapResponseData<ListData<AlertSilence>>(http.get('/v1/alert/silences', { params }))
}

export function getAlertChannels(params: { type?: string } = {}) {
  return unwrapResponseData<ListData<NotificationChannel>>(
    http.get('/v1/alert/channels', { params }),
  )
}

export function getAlertConfig() {
  return unwrapResponseData<AlertmanagerConfig>(http.get('/v1/alert/config'))
}

export function searchLogs(
  params: {
    keyword?: string
    source?: string
    level?: string
    start?: string
    end?: string
    page?: number
    pageSize?: number
  } = {},
) {
  return unwrapResponseData<ListData<LogEntry>>(http.get('/v1/log/search', { params }))
}

export function queryMonitor(params: {
  metric: string
  start?: string
  end?: string
  step?: string
}) {
  return unwrapResponseData<QueryResult>(http.get('/v1/monitor/query', { params }))
}

export function getMonitorConfig() {
  return unwrapResponseData<PrometheusConfig>(http.get('/v1/monitor/config'))
}

export function getHarborProjects(params: { keyword?: string } = {}) {
  return unwrapResponseData<ListData<HarborProject>>(http.get('/v1/harbor/list', { params }))
}

export function getHarborImages(params: { projectName?: string; repository?: string } = {}) {
  return unwrapResponseData<ListData<RepositoryImage>>(http.get('/v1/harbor/images', { params }))
}

export function getHarborConfig() {
  return unwrapResponseData<HarborConfig>(http.get('/v1/harbor/config'))
}

export function getCICDPipelines(params: { status?: string; keyword?: string } = {}) {
  return unwrapResponseData<ListData<Pipeline>>(http.get('/v1/cicd/list', { params }))
}

export function getCICDLogs(params: { pipelineId?: number; stage?: string; limit?: number } = {}) {
  return unwrapResponseData<ListData<PipelineLog>>(http.get('/v1/cicd/logs', { params }))
}

export function getCICDTemplates(params: { keyword?: string } = {}) {
  return unwrapResponseData<ListData<PipelineTemplate>>(http.get('/v1/cicd/templates', { params }))
}

export function getCICDRuns(params: { pipelineId?: number; status?: string; limit?: number } = {}) {
  return unwrapResponseData<ListData<PipelineRun>>(http.get('/v1/cicd/runs', { params }))
}

export function getCICDConfig() {
  return unwrapResponseData<JenkinsConfig>(http.get('/v1/cicd/config'))
}

export function getApps() {
  return unwrapResponseData<Application[]>(http.get('/v1/app/list'))
}

export function getAppTemplates(params: { keyword?: string } = {}) {
  return unwrapResponseData<ListData<AppTemplate>>(http.get('/v1/app/template/list', { params }))
}

export function getAppDeployments(
  params: { appId?: number; environment?: string; limit?: number } = {},
) {
  return unwrapResponseData<ListData<ApplicationDeployment>>(
    http.get('/v1/app/deployment/list', { params }),
  )
}

export function getAppVersions(params: { appId?: number; limit?: number } = {}) {
  return unwrapResponseData<ListData<ApplicationVersion>>(
    http.get('/v1/app/version/list', { params }),
  )
}

export function getAppTopology(params: { appId?: number; environment?: string } = {}) {
  return unwrapResponseData<ApplicationTopology>(http.get('/v1/app/topology', { params }))
}

export function getAuditLogs(
  params: {
    userId?: number
    username?: string
    operation?: string
    resource?: string
    startAt?: string
    endAt?: string
    page?: number
    pageSize?: number
  } = {},
) {
  return unwrapResponseData<{
    list: AuditLogRecord[]
    total: number
    page: number
    pageSize: number
  }>(http.get('/v1/audit/list', { params }))
}

export function exportAuditLogs(
  params: {
    userId?: number
    username?: string
    operation?: string
    resource?: string
    startAt?: string
    endAt?: string
    limit?: number
  } = {},
) {
  return http.get<Blob>('/v1/audit/export', { params, responseType: 'blob' })
}

export function cleanupExpiredAuditLogs() {
  return unwrapResponseData<{ cleaned: number }>(http.post('/v1/audit/cleanup'))
}
