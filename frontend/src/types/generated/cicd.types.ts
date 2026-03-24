export type CICDCicdConfigPostParams = Record<string, never>

export type CICDCicdConfigPostRequest = Record<string, never>

export type CICDCicdConfigPostResponse = Record<string, unknown>

export type CICDCicdConfigUpsertPostParams = Record<string, never>

export type CICDCicdConfigUpsertPostRequest = Record<string, unknown>

export type CICDCicdConfigUpsertPostResponse = Record<string, unknown>

export type CICDCicdListPostParams = { status?: string; keyword?: string }

export type CICDCicdListPostRequest = Record<string, never>

export type CICDCicdListPostResponse = Record<string, unknown>

export type CICDCicdLogsPostParams = { pipelineId?: number; stage?: string; limit?: number }

export type CICDCicdLogsPostRequest = Record<string, never>

export type CICDCicdLogsPostResponse = Record<string, unknown>

export type CICDCicdOrchestrationPreviewPostParams = {
  pipelineId?: number
  templateId?: number
  environment?: string
}

export type CICDCicdOrchestrationPreviewPostRequest = Record<string, never>

export type CICDCicdOrchestrationPreviewPostResponse = Record<string, unknown>

export type CICDCicdRunsPostParams = { pipelineId?: number; status?: string; limit?: number }

export type CICDCicdRunsPostRequest = Record<string, never>

export type CICDCicdRunsPostResponse = Record<string, unknown>

export type CICDCicdStatusPostParams = { status?: string; keyword?: string }

export type CICDCicdStatusPostRequest = Record<string, never>

export type CICDCicdStatusPostResponse = Record<string, unknown>

export type CICDCicdTemplateSavePostParams = Record<string, never>

export type CICDCicdTemplateSavePostRequest = Record<string, unknown>

export type CICDCicdTemplateSavePostResponse = Record<string, unknown>

export type CICDCicdTemplatesPostParams = { keyword?: string }

export type CICDCicdTemplatesPostRequest = Record<string, never>

export type CICDCicdTemplatesPostResponse = Record<string, unknown>

export type CICDCicdTriggerPostParams = Record<string, never>

export type CICDCicdTriggerPostRequest = Record<string, unknown>

export type CICDCicdTriggerPostResponse = Record<string, unknown>
