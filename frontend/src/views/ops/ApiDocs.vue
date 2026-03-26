<script setup lang="ts">
import { ref, onMounted, computed, watch } from 'vue'
import {
  NCard,
  NTabs,
  NTabPane,
  NInput,
  NSpace,
  NTag,
  NCollapse,
  NCollapseItem,
  NDescriptions,
  NDescriptionsItem,
  NCode,
  NEmpty,
  NButton,
  NBadge,
  useMessage,
} from 'naive-ui'

interface SwaggerPath {
  [method: string]: {
    summary: string
    description: string
    tags: string[]
    parameters?: Array<{
      name: string
      in: string
      type?: string
      description?: string
      required?: boolean
      schema?: { $ref?: string; type?: string }
    }>
    responses?: { [code: string]: { description: string } }
  }
}

interface SwaggerDoc {
  paths: { [path: string]: SwaggerPath }
  definitions?: { [name: string]: { properties?: { [key: string]: { type: string; description?: string } } } }
}

const message = useMessage()
const loading = ref(false)
const swaggerDoc = ref<SwaggerDoc | null>(null)
const keyword = ref('')
const selectedTag = ref('应用管理')
const expandedNames = ref<string[]>([])

const API_URL = 'http://localhost:8000/swagger/doc.json'

const methodColors: Record<string, string> = {
  get: 'success',
  post: 'info',
  put: 'warning',
  delete: 'error',
  patch: 'default',
}

async function fetchSwaggerDoc() {
  loading.value = true
  try {
    const res = await fetch(API_URL)
    if (!res.ok) throw new Error('获取文档失败')
    swaggerDoc.value = await res.json()
  } catch (e: any) {
    message.error(e.message || '加载 API 文档失败')
  } finally {
    loading.value = false
  }
}

const allTags = computed(() => {
  if (!swaggerDoc.value?.paths) return []
  const tags = new Set<string>()
  Object.values(swaggerDoc.value.paths).forEach((pathItem) => {
    Object.values(pathItem).forEach((op) => {
      op.tags?.forEach((t) => tags.add(t))
    })
  })
  return Array.from(tags)
})

const filteredApis = computed(() => {
  if (!swaggerDoc.value?.paths) return []
  const result: { path: string; method: string; info: SwaggerPath[string] }[] = []
  const kw = keyword.value.toLowerCase().trim()

  Object.entries(swaggerDoc.value.paths).forEach(([path, pathItem]) => {
    Object.entries(pathItem).forEach(([method, info]) => {
      if (!selectedTag.value || info.tags?.includes(selectedTag.value)) {
        if (!kw || path.toLowerCase().includes(kw) || info.summary?.toLowerCase().includes(kw)) {
          result.push({ path, method, info })
        }
      }
    })
  })
  return result
})

const apiCount = computed(() => filteredApis.value.length)

function getDefinitionName(ref: string) {
  return ref?.split('/')?.pop() || ''
}

function getDefinitionProps(ref: string) {
  const name = getDefinitionName(ref)
  return swaggerDoc.value?.definitions?.[name]?.properties || {}
}

watch(selectedTag, () => {
  expandedNames.value = []
})

onMounted(fetchSwaggerDoc)
</script>

<template>
  <n-card title="API 文档" :loading="loading">
    <template #header-extra>
      <n-space>
        <n-input
          v-model:value="keyword"
          placeholder="搜索接口..."
          clearable
          style="width: 200px"
        />
        <n-button type="primary" @click="fetchSwaggerDoc">刷新</n-button>
      </n-space>
    </template>

    <n-tabs v-model:value="selectedTag" type="line" animated>
      <n-tab-pane
        v-for="tag in allTags"
        :key="tag"
        :name="tag"
        :tab="tag"
      />
    </n-tabs>

    <n-space vertical style="margin-top: 16px">
      <n-space align="center">
        <span>共</span>
        <n-badge :value="apiCount" :max="999" type="info" />
        <span>个接口</span>
      </n-space>

      <n-empty v-if="filteredApis.length === 0" description="暂无接口" />

      <n-collapse v-else v-model:expanded-names="expandedNames">
        <n-collapse-item
          v-for="api in filteredApis"
          :key="`${api.path}-${api.method}`"
          :name="`${api.path}-${api.method}`"
        >
          <template #header>
            <n-space align="center">
              <n-tag :type="(methodColors[api.method] as 'success' | 'error' | 'warning' | 'default' | 'primary' | 'info') || 'default'" size="small" round>
                {{ api.method.toUpperCase() }}
              </n-tag>
              <span style="font-family: monospace">{{ api.path }}</span>
              <span style="color: #999; margin-left: 8px">{{ api.info.summary }}</span>
            </n-space>
          </template>

          <n-descriptions label-placement="left" :column="1" bordered size="small">
            <n-descriptions-item label="描述">
              {{ api.info.description || api.info.summary }}
            </n-descriptions-item>
            <n-descriptions-item label="标签">
              <n-space>
                <n-tag v-for="t in api.info.tags" :key="t" size="small">{{ t }}</n-tag>
              </n-space>
            </n-descriptions-item>
            <n-descriptions-item v-if="api.info.parameters?.length" label="参数">
              <n-descriptions label-placement="left" :column="1" size="small">
                <n-descriptions-item
                  v-for="p in api.info.parameters"
                  :key="p.name"
                  :label="p.name"
                >
                  <n-space align="center">
                    <n-tag size="small" :type="p.in === 'path' ? 'warning' : 'default'">
                      {{ p.in }}
                    </n-tag>
                    <span v-if="p.type || p.schema?.type">{{ p.type || p.schema?.type }}</span>
                    <n-tag v-if="p.required" type="error" size="small">必填</n-tag>
                    <span style="color: #666">{{ p.description }}</span>
                  </n-space>
                  <div v-if="p.schema?.$ref" style="margin-top: 8px">
                    <n-code :code="JSON.stringify(getDefinitionProps(p.schema.$ref), null, 2)" language="json" />
                  </div>
                </n-descriptions-item>
              </n-descriptions>
            </n-descriptions-item>
            <n-descriptions-item label="响应">
              <n-space vertical>
                <n-space
                  v-for="(resp, code) in api.info.responses"
                  :key="code"
                  align="center"
                >
                  <n-tag
                    :type="code.startsWith('2') ? 'success' : code.startsWith('4') ? 'warning' : 'error'"
                    size="small"
                  >
                    {{ code }}
                  </n-tag>
                  <span>{{ resp.description }}</span>
                </n-space>
              </n-space>
            </n-descriptions-item>
          </n-descriptions>
        </n-collapse-item>
      </n-collapse>
    </n-space>
  </n-card>
</template>
