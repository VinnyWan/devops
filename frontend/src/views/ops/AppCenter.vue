<script setup lang="ts">
import { h, ref, onMounted, computed } from 'vue'
import {
  NCard,
  NDataTable,
  NButton,
  NSpace,
  NInput,
  NSelect,
  NTabs,
  NTabPane,
  NForm,
  NFormItem,
  NInputNumber,
  NModal,
  useMessage,
} from 'naive-ui'
import {
  getApps,
  getAppTemplates,
  getAppDeployments,
  getAppVersions,
  getAppTopology,
  getAppConfig,
  saveAppConfig,
  getBuildConfig,
  saveBuildConfig,
  getDeployConfig,
  saveDeployConfig,
  getTechStackConfig,
  saveTechStackConfig,
  getEnumOptions,
} from '@/api/ops'
import type {
  AppTemplate,
  Application,
  ApplicationDeployment,
  ApplicationVersion,
  AppConfig,
  BuildConfig,
  DeployConfig,
  TechStackConfig,
  EnumOptions,
} from '@/types/ops'

const message = useMessage()
const loading = ref(false)
const configLoading = ref(false)
const keyword = ref('')
const environment = ref('')
const apps = ref<Application[]>([])
const templates = ref<AppTemplate[]>([])
const deployments = ref<ApplicationDeployment[]>([])
const versions = ref<ApplicationVersion[]>([])
const topologySummary = ref('未查询')
const selectedAppId = ref<number | null>(null)
const showConfigModal = computed({
  get: () => selectedAppId.value !== null,
  set: (val: boolean) => { if (!val) selectedAppId.value = null },
})
const enumOptions = ref<EnumOptions | null>(null)

// 配置表单
const appConfigForm = ref<AppConfig>({
  appId: 0,
  name: '',
  owner: '',
  developers: '',
  testers: '',
  gitAddress: '',
  appState: '',
  language: '',
  description: '',
  domain: '',
})

const buildConfigForm = ref<BuildConfig>({
  appId: 0,
  buildEnv: '',
  buildTool: '',
  buildConfig: '',
  customConfig: '',
  dockerfile: '',
})

const deployConfigForm = ref<DeployConfig>({
  appId: 0,
  servicePort: 8080,
  cpuRequest: '500m',
  cpuLimit: '1',
  memoryRequest: '512Mi',
  memoryLimit: '1Gi',
  environment: '',
  envVars: '',
})

const techStackConfigForm = ref<TechStackConfig>({
  appId: 0,
  name: '',
  language: '',
  version: '',
  baseImage: '',
  buildImage: '',
  runtimeImage: '',
})

// 下拉选项
const cpuOptions = computed(() =>
  enumOptions.value?.cpuOptions?.map(v => ({ label: v, value: v })) || []
)
const memoryOptions = computed(() =>
  enumOptions.value?.memoryOptions?.map(v => ({ label: v, value: v })) || []
)
const appStateOptions = computed(() => enumOptions.value?.appStates || [])
const buildEnvOptions = computed(() => enumOptions.value?.buildEnvs || [])
const languageOptions = computed(() => enumOptions.value?.languages || [])
const environmentOptions = computed(() => enumOptions.value?.environments || [])

const appColumns = [
  { title: '应用ID', key: 'id', width: 90 },
  { title: '应用名', key: 'name', width: 180 },
  { title: '命名空间', key: 'namespace', width: 140 },
  { title: '状态', key: 'status', width: 100 },
  { title: '更新时间', key: 'updatedAt', width: 180 },
  {
    title: '操作',
    key: 'actions',
    width: 120,
    render: (row: Application) =>
      h(
        NButton,
        {
          size: 'small',
          type: 'primary',
          onClick: () => selectApp(row),
        },
        { default: () => '配置管理' }
      ),
  },
]

const templateColumns = [
  { title: '模板ID', key: 'id', width: 90 },
  { title: '模板名', key: 'name', width: 180 },
  { title: '类型', key: 'type', width: 120 },
  { title: '描述', key: 'description', ellipsis: { tooltip: true } },
]

const deploymentColumns = [
  { title: '记录ID', key: 'id', width: 90 },
  { title: '应用', key: 'appName', width: 160 },
  { title: '集群', key: 'cluster', width: 120 },
  { title: '环境', key: 'environment', width: 120 },
  { title: '版本', key: 'version', width: 130 },
  { title: '状态', key: 'status', width: 100 },
  { title: '时间', key: 'createdAt', width: 180 },
]

const versionColumns = [
  { title: '版本ID', key: 'id', width: 90 },
  { title: '版本', key: 'version', width: 120 },
  { title: '镜像', key: 'image', ellipsis: { tooltip: true } },
  { title: '环境', key: 'environment', width: 120 },
  { title: '状态', key: 'status', width: 100 },
  { title: '时间', key: 'createdAt', width: 180 },
]

async function fetchData() {
  loading.value = true
  try {
    const [appsRes, templateRes, deploymentRes, versionRes, enumRes] = await Promise.all([
      getApps(),
      getAppTemplates({ keyword: keyword.value.trim() || undefined }),
      getAppDeployments({ environment: environment.value.trim() || undefined, limit: 50 }),
      getAppVersions({ limit: 50 }),
      getEnumOptions(),
    ])
    apps.value = appsRes
    templates.value = templateRes.items
    deployments.value = deploymentRes.items
    versions.value = versionRes.items
    enumOptions.value = enumRes
    const firstApp = apps.value[0]
    if (firstApp) {
      const topology = await getAppTopology({
        appId: firstApp.id,
        environment: environment.value.trim() || undefined,
      })
      topologySummary.value = `${topology.appName} 节点 ${topology.nodes.length} / 边 ${topology.edges.length}`
    } else {
      topologySummary.value = '暂无应用'
    }
  } finally {
    loading.value = false
  }
}

function selectApp(app: Application) {
  selectedAppId.value = app.id
  loadAppConfigs(app.id)
}

async function loadAppConfigs(appId: number) {
  configLoading.value = true
  try {
    const [appCfg, buildCfg, deployCfg, techCfg] = await Promise.all([
      getAppConfig(appId).catch(() => null),
      getBuildConfig(appId).catch(() => null),
      getDeployConfig(appId).catch(() => null),
      getTechStackConfig(appId).catch(() => null),
    ])
    if (appCfg) appConfigForm.value = { ...appCfg }
    else {
      appConfigForm.value = {
        appId,
        name: apps.value.find(a => a.id === appId)?.name || '',
        owner: '',
        developers: '',
        testers: '',
        gitAddress: '',
        appState: '',
        language: '',
        description: '',
        domain: '',
      }
    }
    if (buildCfg) buildConfigForm.value = { ...buildCfg }
    else {
      buildConfigForm.value = {
        appId,
        buildEnv: '',
        buildTool: '',
        buildConfig: '',
        customConfig: '',
        dockerfile: '',
      }
    }
    if (deployCfg) deployConfigForm.value = { ...deployCfg }
    else {
      deployConfigForm.value = {
        appId,
        servicePort: 8080,
        cpuRequest: '500m',
        cpuLimit: '1',
        memoryRequest: '512Mi',
        memoryLimit: '1Gi',
        environment: '',
        envVars: '',
      }
    }
    if (techCfg) techStackConfigForm.value = { ...techCfg }
    else {
      techStackConfigForm.value = {
        appId,
        name: '',
        language: '',
        version: '',
        baseImage: '',
        buildImage: '',
        runtimeImage: '',
      }
    }
  } catch (e: any) {
    message.error(e.message || '加载配置失败')
  } finally {
    configLoading.value = false
  }
}

async function handleSaveAppConfig() {
  try {
    await saveAppConfig(appConfigForm.value)
    message.success('保存应用配置成功')
  } catch (e: any) {
    message.error(e.message || '保存失败')
  }
}

async function handleSaveBuildConfig() {
  try {
    await saveBuildConfig(buildConfigForm.value)
    message.success('保存构建配置成功')
  } catch (e: any) {
    message.error(e.message || '保存失败')
  }
}

async function handleSaveDeployConfig() {
  try {
    await saveDeployConfig(deployConfigForm.value)
    message.success('保存部署配置成功')
  } catch (e: any) {
    message.error(e.message || '保存失败')
  }
}

async function handleSaveTechStackConfig() {
  try {
    await saveTechStackConfig(techStackConfigForm.value)
    message.success('保存技术栈配置成功')
  } catch (e: any) {
    message.error(e.message || '保存失败')
  }
}

onMounted(fetchData)
</script>

<template>
  <n-card :title="`应用管理（拓扑：${topologySummary}）`">
    <template #header-extra>
      <n-space>
        <n-input
          v-model:value="keyword"
          placeholder="模板关键词"
          clearable
          style="width: 180px"
          @keyup.enter="fetchData"
        />
        <n-input
          v-model:value="environment"
          placeholder="环境过滤"
          clearable
          style="width: 160px"
          @keyup.enter="fetchData"
        />
        <n-button type="primary" @click="fetchData">查询</n-button>
      </n-space>
    </template>
    <n-tabs type="line" animated>
      <n-tab-pane name="apps" tab="应用">
        <n-data-table
          :columns="appColumns"
          :data="apps"
          :loading="loading"
          :bordered="false"
          :row-key="(row: Application) => row.id"
        />
      </n-tab-pane>
      <n-tab-pane name="templates" tab="模板">
        <n-data-table
          :columns="templateColumns"
          :data="templates"
          :loading="loading"
          :bordered="false"
          :row-key="(row: AppTemplate) => row.id"
        />
      </n-tab-pane>
      <n-tab-pane name="deployments" tab="部署记录">
        <n-data-table
          :columns="deploymentColumns"
          :data="deployments"
          :loading="loading"
          :bordered="false"
          :row-key="(row: ApplicationDeployment) => row.id"
        />
      </n-tab-pane>
      <n-tab-pane name="versions" tab="版本">
        <n-data-table
          :columns="versionColumns"
          :data="versions"
          :loading="loading"
          :bordered="false"
          :row-key="(row: ApplicationVersion) => row.id"
        />
      </n-tab-pane>
    </n-tabs>
  </n-card>

  <!-- 应用配置弹窗 -->
  <n-modal
    v-model:show="showConfigModal"
    preset="card"
    title="应用配置管理"
    style="width: 900px; max-width: calc(100vw - 32px)"
    :mask-closable="true"
  >
    <n-tabs type="line" animated>
      <n-tab-pane name="app-config" tab="应用配置">
        <n-form label-placement="left" label-width="100">
          <n-form-item label="应用名称">
            <n-input v-model:value="appConfigForm.name" placeholder="请输入应用名称" />
          </n-form-item>
          <n-form-item label="应用负责人">
            <n-input v-model:value="appConfigForm.owner" placeholder="请输入负责人" />
          </n-form-item>
          <n-form-item label="开发人员">
            <n-input v-model:value="appConfigForm.developers" placeholder="多个用逗号分隔" />
          </n-form-item>
          <n-form-item label="测试人员">
            <n-input v-model:value="appConfigForm.testers" placeholder="多个用逗号分隔" />
          </n-form-item>
          <n-form-item label="Git 地址">
            <n-input v-model:value="appConfigForm.gitAddress" placeholder="https://github.com/..." />
          </n-form-item>
          <n-form-item label="应用状态">
            <n-select
              v-model:value="appConfigForm.appState"
              :options="appStateOptions"
              placeholder="选择应用状态"
            />
          </n-form-item>
          <n-form-item label="开发语言">
            <n-select
              v-model:value="appConfigForm.language"
              :options="languageOptions"
              placeholder="选择开发语言"
            />
          </n-form-item>
          <n-form-item label="应用介绍">
            <n-input
              v-model:value="appConfigForm.description"
              type="textarea"
              :rows="3"
              placeholder="请输入应用介绍"
            />
          </n-form-item>
          <n-form-item label="域名">
            <n-input v-model:value="appConfigForm.domain" placeholder="app.example.com" />
          </n-form-item>
        </n-form>
        <n-space justify="end">
          <n-button type="primary" :loading="configLoading" @click="handleSaveAppConfig">
            保存
          </n-button>
        </n-space>
      </n-tab-pane>

      <n-tab-pane name="build-config" tab="构建配置">
        <n-form label-placement="left" label-width="100">
          <n-form-item label="构建环境">
            <n-select
              v-model:value="buildConfigForm.buildEnv"
              :options="buildEnvOptions"
              placeholder="选择构建环境"
            />
          </n-form-item>
          <n-form-item label="编译工具">
            <n-input v-model:value="buildConfigForm.buildTool" placeholder="maven / gradle / go / npm" />
          </n-form-item>
          <n-form-item label="编译配置">
            <n-input
              v-model:value="buildConfigForm.buildConfig"
              type="textarea"
              :rows="2"
              placeholder="-DskipTests"
            />
          </n-form-item>
          <n-form-item label="自定义配置">
            <n-input
              v-model:value="buildConfigForm.customConfig"
              type="textarea"
              :rows="3"
              placeholder="自定义构建配置"
            />
          </n-form-item>
          <n-form-item label="Dockerfile">
            <n-input
              v-model:value="buildConfigForm.dockerfile"
              type="textarea"
              :rows="8"
              placeholder="FROM openjdk:17..."
              :input-props="{ style: 'font-family: monospace;' }"
            />
          </n-form-item>
        </n-form>
        <n-space justify="end">
          <n-button type="primary" :loading="configLoading" @click="handleSaveBuildConfig">
            保存
          </n-button>
        </n-space>
      </n-tab-pane>

      <n-tab-pane name="deploy-config" tab="部署配置">
        <n-form label-placement="left" label-width="120">
          <n-form-item label="服务端口">
            <n-input-number
              v-model:value="deployConfigForm.servicePort"
              :min="1"
              :max="65535"
              placeholder="8080"
            />
          </n-form-item>
          <n-form-item label="CPU Request">
            <n-select
              v-model:value="deployConfigForm.cpuRequest"
              :options="cpuOptions"
              placeholder="选择 CPU 请求"
            />
          </n-form-item>
          <n-form-item label="CPU Limit">
            <n-select
              v-model:value="deployConfigForm.cpuLimit"
              :options="cpuOptions"
              placeholder="选择 CPU 限制"
            />
          </n-form-item>
          <n-form-item label="内存 Request">
            <n-select
              v-model:value="deployConfigForm.memoryRequest"
              :options="memoryOptions"
              placeholder="选择内存请求"
            />
          </n-form-item>
          <n-form-item label="内存 Limit">
            <n-select
              v-model:value="deployConfigForm.memoryLimit"
              :options="memoryOptions"
              placeholder="选择内存限制"
            />
          </n-form-item>
          <n-form-item label="环境">
            <n-select
              v-model:value="deployConfigForm.environment"
              :options="environmentOptions"
              placeholder="选择部署环境"
            />
          </n-form-item>
          <n-form-item label="环境变量">
            <n-input
              v-model:value="deployConfigForm.envVars"
              type="textarea"
              :rows="4"
              placeholder="KEY=value&#10;DB_HOST=mysql.dev.svc"
            />
          </n-form-item>
        </n-form>
        <n-space justify="end">
          <n-button type="primary" :loading="configLoading" @click="handleSaveDeployConfig">
            保存
          </n-button>
        </n-space>
      </n-tab-pane>

      <n-tab-pane name="tech-stack" tab="技术栈配置">
        <n-form label-placement="left" label-width="100">
          <n-form-item label="技术栈名称">
            <n-input v-model:value="techStackConfigForm.name" placeholder="Java 17 / Go 1.21" />
          </n-form-item>
          <n-form-item label="开发语言">
            <n-select
              v-model:value="techStackConfigForm.language"
              :options="languageOptions"
              placeholder="选择开发语言"
            />
          </n-form-item>
          <n-form-item label="版本">
            <n-input v-model:value="techStackConfigForm.version" placeholder="17 / 1.21 / 3.11" />
          </n-form-item>
          <n-form-item label="基础镜像">
            <n-input v-model:value="techStackConfigForm.baseImage" placeholder="openjdk:17" />
          </n-form-item>
          <n-form-item label="构建镜像">
            <n-input v-model:value="techStackConfigForm.buildImage" placeholder="maven:3.9" />
          </n-form-item>
          <n-form-item label="运行时镜像">
            <n-input v-model:value="techStackConfigForm.runtimeImage" placeholder="openjdk:17-jre" />
          </n-form-item>
        </n-form>
        <n-space justify="end">
          <n-button type="primary" :loading="configLoading" @click="handleSaveTechStackConfig">
            保存
          </n-button>
        </n-space>
      </n-tab-pane>
    </n-tabs>
  </n-modal>
</template>
