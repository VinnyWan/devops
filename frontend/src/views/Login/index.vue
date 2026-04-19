<template>
  <div class="login-container">
    <div class="login-bg-shape shape-1"></div>
    <div class="login-bg-shape shape-2"></div>
    <div class="login-bg-shape shape-3"></div>
    <el-card class="login-card">
      <div class="login-header">
        <div class="login-logo">S</div>
        <h2>运维开发平台</h2>
        <p class="login-subtitle">SRE Platform</p>
      </div>
      <el-form ref="formRef" :model="form" :rules="rules" @submit.prevent="handleLogin" label-position="top">
        <el-form-item label="租户代码" prop="tenantCode">
          <el-input v-model="form.tenantCode" placeholder="请输入租户代码" />
        </el-form-item>
        <el-form-item label="认证方式" prop="authType">
          <el-select v-model="form.authType" placeholder="请选择认证方式" style="width: 100%">
            <el-option label="本地认证" value="local" />
            <el-option label="LDAP" value="ldap" />
            <el-option label="OIDC" value="oidc" />
          </el-select>
        </el-form-item>
        <el-form-item label="用户名" prop="username">
          <el-input v-model="form.username" placeholder="请输入用户名" />
        </el-form-item>
        <el-form-item label="密码" prop="password">
          <el-input v-model="form.password" type="password" placeholder="请输入密码" show-password />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" native-type="submit" :loading="loading" class="login-btn">
            登录
          </el-button>
        </el-form-item>
      </el-form>
    </el-card>
  </div>
</template>

<script setup>
import { reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { useUserStore } from '../../stores/user'
import { login } from '../../api/user'
import { required } from '../../utils/validate'

const router = useRouter()
const userStore = useUserStore()
const formRef = ref()
const loading = ref(false)

const form = reactive({
  tenantCode: 'default',
  authType: 'local',
  username: '',
  password: ''
})

const rules = {
  tenantCode: [required('请输入租户代码')],
  authType: [required('请选择认证方式')],
  username: [required('请输入用户名')],
  password: [required('请输入密码')]
}

const handleLogin = async () => {
  loading.value = true

  try {
    const valid = await formRef.value.validate()
    if (!valid) {
      loading.value = false
      return
    }

    const res = await login(form)

    userStore.setToken(res.data.token)
    userStore.setUserInfo(res.data.user)
    ElMessage.success('登录成功')
    await router.push('/')
  } catch (error) {
    loading.value = false
    ElMessage.error(error.response?.data?.message || '登录失败')
  }
}
</script>

<style scoped>
.login-container {
  display: flex;
  justify-content: center;
  align-items: center;
  height: 100vh;
  background: var(--color-bg);
  position: relative;
  overflow: hidden;
}

/* Decorative background shapes */
.login-bg-shape {
  position: absolute;
  border-radius: 50%;
  opacity: 0.5;
}
.shape-1 {
  width: 600px;
  height: 600px;
  background: radial-gradient(circle, var(--color-primary-lighter) 0%, transparent 70%);
  top: -200px;
  right: -100px;
}
.shape-2 {
  width: 400px;
  height: 400px;
  background: radial-gradient(circle, #E0E7FF 0%, transparent 70%);
  bottom: -100px;
  left: -50px;
}
.shape-3 {
  width: 200px;
  height: 200px;
  background: radial-gradient(circle, var(--color-success-light) 0%, transparent 70%);
  top: 50%;
  left: 15%;
}

.login-card {
  width: 420px;
  border-radius: var(--radius-lg);
  border: 1px solid var(--color-border-light);
  box-shadow: var(--shadow-lg);
  position: relative;
  z-index: 1;
  background: var(--color-bg-white);
}

.login-header {
  text-align: center;
  margin-bottom: var(--spacing-lg);
}

.login-logo {
  width: 48px;
  height: 48px;
  border-radius: var(--radius-md);
  background: var(--color-primary);
  color: white;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  font-weight: 700;
  font-size: 20px;
  margin-bottom: var(--spacing-md);
}

.login-header h2 {
  margin: 0 0 4px 0;
  font-size: var(--font-size-xl);
  font-weight: 600;
  color: var(--color-text);
}

.login-subtitle {
  margin: 0;
  font-size: var(--font-size-sm);
  color: var(--color-text-tertiary);
  letter-spacing: 0.05em;
  text-transform: uppercase;
}

.login-btn {
  width: 100%;
  height: 42px;
  font-size: var(--font-size-base);
  border-radius: var(--radius-sm);
  font-weight: 500;
}

:deep(.el-form-item__label) {
  font-weight: 500;
  color: var(--color-text);
}
</style>
