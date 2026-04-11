<template>
  <div class="page-container">
    <div class="page-header">
      <h3>权限管理</h3>
    </div>

    <div style="margin-bottom: 16px; display: flex; gap: 12px;">
      <el-input v-model="keyword" placeholder="搜索权限名称" style="width: 300px;" clearable @clear="fetchData" @keyup.enter="fetchData">
        <template #append>
          <el-button @click="fetchData"><el-icon><Search /></el-icon></el-button>
        </template>
      </el-input>
      <el-select v-model="resourceFilter" placeholder="按资源过滤" clearable @change="fetchData" style="width: 200px;">
        <el-option v-for="r in resources" :key="r" :label="r" :value="r" />
      </el-select>
    </div>

    <el-table :data="tableData" stripe v-loading="loading" style="width: 100%">
      <el-table-column prop="name" label="权限名称" width="200" />
      <el-table-column prop="resource" label="资源" width="150" />
      <el-table-column prop="action" label="操作" width="120" />
      <el-table-column prop="description" label="描述" />
    </el-table>

    <div style="margin-top: 16px; display: flex; justify-content: flex-end;">
      <el-pagination v-model:current-page="page" v-model:page-size="pageSize" :total="total" :page-sizes="[10, 20, 50]" layout="total, sizes, prev, pager, next" @current-change="fetchData" @size-change="fetchData" />
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { Search } from '@element-plus/icons-vue'
import { getPermissionList } from '@/api/permission'

const loading = ref(false)
const tableData = ref([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(10)
const keyword = ref('')
const resourceFilter = ref('')
const resources = ref([])

const fetchData = async () => {
  loading.value = true
  try {
    const params = { page: page.value, pageSize: pageSize.value }
    if (keyword.value) params.keyword = keyword.value
    if (resourceFilter.value) params.resource = resourceFilter.value
    const res = await getPermissionList(params)
    tableData.value = res.data?.list || res.data || []
    total.value = res.data?.total || 0
    const resSet = new Set((res.data?.list || res.data || []).map(p => p.resource))
    resources.value = [...resSet]
  } finally { loading.value = false }
}

onMounted(fetchData)
</script>

<style scoped>
.page-container { background: #fff; border-radius: 4px; padding: 24px; }
.page-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 24px; }
.page-header h3 { margin: 0; font-size: 18px; font-weight: 500; }
</style>
