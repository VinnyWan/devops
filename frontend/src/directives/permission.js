import { usePermissionStore } from '@/stores/permission'

// 按钮权限指令 v-permission="'user:create'"
export const vPermission = {
  mounted(el, binding) {
    const store = usePermissionStore()
    const [resource, action] = binding.value.split(':')
    if (!store.hasApi(resource, action)) {
      el.parentNode?.removeChild(el)
    }
  }
}

// 字段权限指令 v-permission-field="'user.salary'"
export const vPermissionField = {
  mounted(el, binding) {
    const store = usePermissionStore()
    const [resource, field] = binding.value.split('.')
    const action = store.getFieldAction(resource, field)
    if (action === 'hidden') {
      el.style.display = 'none'
    } else if (action === 'readonly') {
      const input = el.querySelector('input, textarea, select')
      if (input) {
        input.disabled = true
      }
    }
  }
}
