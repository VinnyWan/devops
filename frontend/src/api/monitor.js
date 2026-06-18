import request from './request'

export const listPrometheusConfigs = (params) => request.get('/monitor/prometheus', { params })
export const getPrometheusConfig = (id) => request.get(`/monitor/prometheus/${id}`)
export const savePrometheusConfig = (data) => request.post('/monitor/prometheus', data)
export const updatePrometheusConfig = (id, data) => request.put(`/monitor/prometheus/${id}`, data)
export const deletePrometheusConfig = (id) => request.delete(`/monitor/prometheus/${id}`)
export const testPrometheusConnection = (data) => request.post('/monitor/prometheus/test', data)
export const queryHostMetrics = (params) => request.get('/monitor/host/metrics', { params })
export const queryPortStatus = (params) => request.get('/monitor/host/ports', { params })
export const queryAgentStatus = (params) => request.get('/monitor/agent/status', { params })
