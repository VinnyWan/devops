<script setup lang="ts">
import { computed, defineAsyncComponent } from 'vue'
import { useRoute } from 'vue-router'

const route = useRoute()
const MainLayout = defineAsyncComponent(() => import('@/layouts/MainLayout.vue'))
const BlankLayout = defineAsyncComponent(() => import('@/layouts/BlankLayout.vue'))
const layoutComponent = computed(() => (route.meta.layout === 'blank' ? BlankLayout : MainLayout))
</script>

<template>
  <component :is="layoutComponent" />
</template>
