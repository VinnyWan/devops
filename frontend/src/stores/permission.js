import { defineStore } from 'pinia'
import { ref } from 'vue'
import { getPermissions } from '@/api/auth'

export const usePermissionStore = defineStore('permission', () => {
  const menus = ref([])
  const buttons = ref({})
  const fieldRules = ref({})
  const apis = ref([])

  async function fetchPermissions() {
    const res = await getPermissions()
    menus.value = res.menus || []
    buttons.value = res.buttons || {}
    fieldRules.value = res.fieldRules || {}
    apis.value = res.apis || []
  }

  function hasButton(resource, action) {
    const key = `${resource}:list`
    return buttons.value[key]?.includes(action) ?? false
  }

  function getFieldAction(resource, field) {
    return fieldRules.value[resource]?.[field] || 'visible'
  }

  function hasApi(resource, action) {
    return apis.value.includes(`${resource}:${action}`)
  }

  return {
    menus, buttons, fieldRules, apis,
    fetchPermissions, hasButton, getFieldAction, hasApi
  }
})
