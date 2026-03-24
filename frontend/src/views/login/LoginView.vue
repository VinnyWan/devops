<script setup lang="ts">
import { ref } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useAuth } from '@/composables/useAuth'
import type { AuthType } from '@/types/auth'

const router = useRouter()
const route = useRoute()
const { login } = useAuth()

const loading = ref(false)
const errorMessage = ref('')
const form = ref({ username: '', password: '', authType: 'local' as AuthType })

async function handleLogin() {
  errorMessage.value = ''
  if (!form.value.username || !form.value.password) {
    errorMessage.value = '请输入用户名和密码'
    return
  }
  loading.value = true
  try {
    await login(form.value.username, form.value.password, form.value.authType)
    const redirect = (route.query.redirect as string) || '/dashboard'
    router.push(redirect)
  } catch (e: unknown) {
    errorMessage.value = e instanceof Error ? e.message : '登录失败'
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="login-container">
    <form class="login-card" @submit.prevent="handleLogin">
      <h1 class="title">DevOps 管理平台</h1>
      <label class="field">
        <span>认证方式</span>
        <select v-model="form.authType">
          <option value="local">本地</option>
          <option value="ldap">LDAP</option>
          <option value="oauth2">OAuth2</option>
        </select>
      </label>
      <label class="field">
        <span>用户名</span>
        <input v-model.trim="form.username" placeholder="请输入用户名" autocomplete="username" />
      </label>
      <label class="field">
        <span>密码</span>
        <input
          v-model.trim="form.password"
          type="password"
          placeholder="请输入密码"
          autocomplete="current-password"
        />
      </label>
      <p v-if="errorMessage" class="error-text">{{ errorMessage }}</p>
      <button type="submit" :disabled="loading">
        {{ loading ? '登录中...' : '登录' }}
      </button>
    </form>
  </div>
</template>

<style scoped>
.login-container {
  height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #f0f2f5;
}

.login-card {
  width: min(380px, calc(100vw - 32px));
  box-sizing: border-box;
  background: #fff;
  border: 1px solid #e5e7eb;
  border-radius: 12px;
  padding: 24px;
  box-shadow: 0 8px 24px rgb(15 23 42 / 8%);
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.title {
  margin: 0 0 8px;
  font-size: 22px;
  line-height: 1.2;
  color: #111827;
}

.field {
  display: flex;
  flex-direction: column;
  gap: 6px;
  color: #334155;
  font-size: 14px;
}

.field select,
.field input {
  height: 38px;
  border: 1px solid #cbd5e1;
  border-radius: 8px;
  padding: 0 12px;
  font-size: 14px;
  outline: none;
  transition: border-color 0.2s;
}

.field select:focus,
.field input:focus {
  border-color: #18a058;
}

button {
  margin-top: 4px;
  height: 40px;
  border: none;
  border-radius: 8px;
  background: #18a058;
  color: #fff;
  font-size: 14px;
  font-weight: 600;
  cursor: pointer;
}

button:disabled {
  opacity: 0.7;
  cursor: not-allowed;
}

.error-text {
  margin: 0;
  color: #dc2626;
  font-size: 13px;
}
</style>
