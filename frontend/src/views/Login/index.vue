<template>
  <div class="login-container">
    <el-card class="login-card">
      <h2>运维开发平台</h2>
      <el-form ref="formRef" :model="form" :rules="rules" @submit.prevent="handleLogin" label-width="80px">
        <el-form-item label="租户" prop="tenantCode">
          <el-input v-model="form.tenantCode" placeholder="请输入租户代码" />
        </el-form-item>
        <el-form-item label="认证" prop="authType">
          <el-select v-model="form.authType" placeholder="请选择认证方式" style="width: 100%">
            <el-option label="本地" value="local" />
            <el-option label="LDAP" value="ldap" />
            <el-option label="OIDC" value="oidc" />
          </el-select>
        </el-form-item>
        <el-form-item label="用户" prop="username">
          <el-input v-model="form.username" placeholder="请输入用户名" />
        </el-form-item>
        <el-form-item label="密码" prop="password">
          <el-input v-model="form.password" type="password" placeholder="请输入密码" show-password />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" native-type="submit" :loading="loading" style="width: 100%">
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
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
}
.login-card {
  width: 400px;
  padding: 20px;
  border-radius: 16px;
}
h2 {
  text-align: center;
  margin-bottom: 30px;
}
:deep(.el-input__wrapper) {
  border-radius: 12px;
}
:deep(.el-select .el-input__wrapper) {
  border-radius: 12px;
}
:deep(.el-button) {
  border-radius: 12px;
}
</style>
