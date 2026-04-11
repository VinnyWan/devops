// 表单验证规则

// 必填验证
export const required = (message = '此字段为必填项') => ({
  required: true,
  message,
  trigger: 'blur'
})

// 邮箱验证
export const email = (message = '请输入有效的邮箱地址') => ({
  type: 'email',
  message,
  trigger: 'blur'
})

// 密码强度验证（至少8位，包含字母和数字）
export const password = (message = '密码至少8位，需包含字母和数字') => ({
  pattern: /^(?=.*[A-Za-z])(?=.*\d)[A-Za-z\d@$!%*#?&]{8,}$/,
  message,
  trigger: 'blur'
})

// IP地址验证
export const ipAddress = (message = '请输入有效的IP地址') => ({
  pattern: /^((25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$/,
  message,
  trigger: 'blur'
})

// URL验证
export const url = (message = '请输入有效的URL') => ({
  type: 'url',
  message,
  trigger: 'blur'
})

// 手机号验证
export const phone = (message = '请输入有效的手机号') => ({
  pattern: /^1[3-9]\d{9}$/,
  message,
  trigger: 'blur'
})
