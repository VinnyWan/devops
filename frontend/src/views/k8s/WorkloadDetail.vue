<template>
  <div class="workload-detail">
    <el-page-header @back="$router.back()" :content="`${kind} / ${name}`" />
    <el-divider />
    <p>工作负载详情页 ({{ kind }}: {{ namespace }}/{{ name }}) - 待实现</p>
  </div>
</template>

<script setup>
import { useRoute } from 'vue-router'

const route = useRoute()
const kind = route.params.kind
const clusterName = route.params.clusterName
const namespace = route.params.namespace
const name = route.params.name
</script>

<style scoped>
.workload-detail {
  padding: 20px;
}
</style>
