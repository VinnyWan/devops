<template>
  <el-select v-model="selectedNamespace" placeholder="选择命名空间" @change="handleChange" style="width: 200px">
    <el-option label="全部命名空间" value="" />
    <el-option v-for="ns in namespaces" :key="ns.name" :label="ns.name" :value="ns.name" />
  </el-select>
</template>

<script setup>
import { ref, watch } from 'vue'
import { getNamespaceList } from '@/api/namespace'

const props = defineProps({
  clusterName: { type: [String, Number], required: true }
})

const selectedNamespace = defineModel()
const namespaces = ref([])

const fetchNamespaces = async () => {
  if (!props.clusterName) return
  const res = await getNamespaceList({ clusterName: props.clusterName })
  namespaces.value = res.data || []
}

const handleChange = (val) => {
  selectedNamespace.value = val
}

watch(() => props.clusterName, fetchNamespaces, { immediate: true })
</script>
