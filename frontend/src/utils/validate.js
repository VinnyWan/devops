export const required = (message = '此字段为必填项') => ({
  required: true,
  message,
  trigger: 'blur'
})

export const email = (message = '请输入有效的邮箱地址') => ({
  type: 'email',
  message,
  trigger: 'blur'
})

export const ipAddress = (message = '请输入有效的IP地址') => ({
  pattern: /^((25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$/,
  message,
  trigger: 'blur'
})
