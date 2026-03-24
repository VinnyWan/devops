<script setup lang="ts">
import { ref } from 'vue'
import { NSpace, NInput, NSelect, NButton } from 'naive-ui'

interface FilterConfig {
  name: string
  label: string
  type: 'select' | 'date'
  options?: { label: string; value: any }[]
}

interface Props {
  filters?: FilterConfig[]
  onSearch: (params: Record<string, any>) => void
}

const props = withDefaults(defineProps<Props>(), {
  filters: () => []
})

const keyword = ref('')
const filterValues = ref<Record<string, any>>({})

const handleSearch = () => {
  props.onSearch({
    keyword: keyword.value,
    ...filterValues.value
  })
}

const handleReset = () => {
  keyword.value = ''
  filterValues.value = {}
  props.onSearch({})
}
</script>

<template>
  <NSpace>
    <NInput
      v-model:value="keyword"
      placeholder="关键词搜索"
      clearable
      style="width: 200px"
      @keyup.enter="handleSearch"
    />
    <NSelect
      v-for="filter in filters"
      :key="filter.name"
      v-model:value="filterValues[filter.name]"
      :options="filter.options"
      :placeholder="filter.label"
      clearable
      style="width: 150px"
    />
    <NButton type="primary" @click="handleSearch">搜索</NButton>
    <NButton @click="handleReset">重置</NButton>
  </NSpace>
</template>
