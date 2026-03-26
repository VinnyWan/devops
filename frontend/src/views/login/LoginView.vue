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
    <!-- 背景装饰 -->
    <div class="login-bg-decoration">
      <div class="decoration-circle decoration-circle--1"></div>
      <div class="decoration-circle decoration-circle--2"></div>
      <div class="decoration-circle decoration-circle--3"></div>
    </div>

    <form class="login-card animate-slide-in-up" @submit.prevent="handleLogin">
      <!-- Logo区域 -->
      <div class="login-header">
        <div class="logo-wrapper">
          <svg viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg" class="logo-icon">
            <path d="M12 2L2 7L12 12L22 7L12 2Z" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
            <path d="M2 17L12 22L22 17" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
            <path d="M2 12L12 17L22 12" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
          </svg>
        </div>
        <h1 class="title">DevOps 管理平台</h1>
        <p class="subtitle">统一运维管理解决方案</p>
      </div>

      <!-- 表单区域 -->
      <div class="login-form">
        <label class="field">
          <span class="field-label">认证方式</span>
          <div class="select-wrapper">
            <select v-model="form.authType" class="field-select">
              <option value="local">本地认证</option>
              <option value="ldap">LDAP</option>
              <option value="oauth2">OAuth2</option>
            </select>
            <svg class="select-arrow" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
              <path d="M6 9L12 15L18 9" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
            </svg>
          </div>
        </label>

        <label class="field">
          <span class="field-label">用户名</span>
          <div class="input-wrapper">
            <svg class="input-icon" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
              <path d="M20 21V19C20 17.9391 19.5786 16.9217 18.8284 16.1716C18.0783 15.4214 17.0609 15 16 15H8C6.93913 15 5.92172 15.4214 5.17157 16.1716C4.42143 16.9217 4 17.9391 4 19V21" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
              <path d="M12 11C14.2091 11 16 9.20914 16 7C16 4.79086 14.2091 3 12 3C9.79086 3 8 4.79086 8 7C8 9.20914 9.79086 11 12 11Z" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
            </svg>
            <input
              v-model.trim="form.username"
              placeholder="请输入用户名"
              autocomplete="username"
              class="field-input"
            />
          </div>
        </label>

        <label class="field">
          <span class="field-label">密码</span>
          <div class="input-wrapper">
            <svg class="input-icon" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
              <rect x="3" y="11" width="18" height="11" rx="2" ry="2" stroke="currentColor" stroke-width="2"/>
              <path d="M7 11V7C7 5.67392 7.52678 4.40215 8.46447 3.46447C9.40215 2.52678 10.6739 2 12 2C13.3261 2 14.5979 2.52678 15.5355 3.46447C16.4732 4.40215 17 5.67392 17 7V11" stroke="currentColor" stroke-width="2"/>
            </svg>
            <input
              v-model.trim="form.password"
              type="password"
              placeholder="请输入密码"
              autocomplete="current-password"
              class="field-input"
            />
          </div>
        </label>

        <!-- 错误提示 -->
        <div v-if="errorMessage" class="error-message">
          <svg class="error-icon" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
            <path d="M12 22C6.477 22 2 17.523 2 12C2 6.477 6.477 2 12 2C17.523 2 22 6.477 22 12C22 17.523 17.523 22 12 22Z" fill="currentColor" fill-opacity="0.1"/>
            <path d="M12 8V12M12 16H12.01" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
          </svg>
          <span>{{ errorMessage }}</span>
        </div>

        <!-- 登录按钮 -->
        <button type="submit" class="login-button" :class="{ 'login-button--loading': loading }" :disabled="loading">
          <span v-if="!loading" class="button-text">登录</span>
          <span v-else class="button-loading">
            <svg class="loading-spinner" viewBox="0 0 24 24">
              <circle cx="12" cy="12" r="10" stroke="currentColor" stroke-width="3" fill="none" stroke-dasharray="31.4 31.4" stroke-linecap="round"/>
            </svg>
            登录中...
          </span>
        </button>
      </div>

      <!-- 底部信息 -->
      <div class="login-footer">
        <p>© 2024 DevOps Platform. All rights reserved.</p>
      </div>
    </form>
  </div>
</template>

<style scoped>
.login-container {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, var(--gray-50) 0%, var(--primary-50) 100%);
  position: relative;
  overflow: hidden;
}

/* 背景装饰 */
.login-bg-decoration {
  position: absolute;
  inset: 0;
  pointer-events: none;
  overflow: hidden;
}

.decoration-circle {
  position: absolute;
  border-radius: 50%;
  opacity: 0.4;
}

.decoration-circle--1 {
  width: 400px;
  height: 400px;
  background: linear-gradient(135deg, var(--primary-200), var(--primary-100));
  top: -100px;
  right: -100px;
}

.decoration-circle--2 {
  width: 300px;
  height: 300px;
  background: linear-gradient(135deg, var(--primary-100), var(--primary-50));
  bottom: -50px;
  left: -50px;
}

.decoration-circle--3 {
  width: 200px;
  height: 200px;
  background: linear-gradient(135deg, var(--primary-200), transparent);
  top: 50%;
  left: 20%;
  transform: translateY(-50%);
}

/* 登录卡片 */
.login-card {
  width: min(420px, calc(100vw - 32px));
  background: var(--card-bg);
  border-radius: var(--radius-xl);
  padding: var(--spacing-8);
  box-shadow: var(--shadow-xl);
  border: 1px solid var(--border-light);
  position: relative;
  z-index: 1;
}

/* Logo区域 */
.login-header {
  text-align: center;
  margin-bottom: var(--spacing-8);
}

.logo-wrapper {
  width: 56px;
  height: 56px;
  margin: 0 auto var(--spacing-4);
  background: linear-gradient(135deg, var(--primary-500), var(--primary-600));
  border-radius: var(--radius-lg);
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow: 0 8px 24px rgba(59, 130, 246, 0.3);
}

.logo-icon {
  width: 32px;
  height: 32px;
  color: white;
}

.title {
  margin: 0 0 var(--spacing-2);
  font-size: var(--font-size-2xl);
  font-weight: var(--font-weight-bold);
  color: var(--text-primary);
  letter-spacing: -0.02em;
}

.subtitle {
  margin: 0;
  font-size: var(--font-size-sm);
  color: var(--text-secondary);
}

/* 表单区域 */
.login-form {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-5);
}

.field {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-2);
}

.field-label {
  font-size: var(--font-size-sm);
  font-weight: var(--font-weight-medium);
  color: var(--text-secondary);
}

/* Select 样式 */
.select-wrapper {
  position: relative;
}

.field-select {
  width: 100%;
  height: var(--touch-target-min);
  padding: 0 var(--spacing-5);
  padding-right: var(--spacing-10);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-lg);
  font-size: var(--font-size-base);
  color: var(--text-primary);
  background: var(--card-bg);
  cursor: pointer;
  appearance: none;
  transition: all var(--transition-fast);
}

.field-select:hover {
  border-color: var(--primary-300);
}

.field-select:focus {
  outline: none;
  border-color: var(--primary-500);
  box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
}

.select-arrow {
  position: absolute;
  right: var(--spacing-4);
  top: 50%;
  transform: translateY(-50%);
  width: 16px;
  height: 16px;
  color: var(--text-muted);
  pointer-events: none;
}

/* Input 样式 */
.input-wrapper {
  position: relative;
  display: flex;
  align-items: center;
}

.input-icon {
  position: absolute;
  left: var(--spacing-4);
  width: 18px;
  height: 18px;
  color: var(--text-muted);
  pointer-events: none;
  transition: color var(--transition-fast);
}

.field-input {
  width: 100%;
  height: var(--touch-target-min);
  padding: 0 var(--spacing-4);
  padding-left: var(--spacing-11);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-lg);
  font-size: var(--font-size-base);
  color: var(--text-primary);
  background: var(--card-bg);
  transition: all var(--transition-fast);
}

.field-input::placeholder {
  color: var(--text-muted);
}

.field-input:hover {
  border-color: var(--primary-300);
}

.field-input:focus {
  outline: none;
  border-color: var(--primary-500);
  box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
}

.field-input:focus + .input-icon,
.input-wrapper:focus-within .input-icon {
  color: var(--primary-500);
}

/* 错误提示 */
.error-message {
  display: flex;
  align-items: center;
  gap: var(--spacing-2);
  padding: var(--spacing-3) var(--spacing-4);
  background: var(--error-bg);
  border-radius: var(--radius-md);
  color: var(--error-600);
  font-size: var(--font-size-sm);
  animation: shake 0.3s ease-in-out;
}

.error-icon {
  width: 16px;
  height: 16px;
  flex-shrink: 0;
  color: var(--error-500);
}

@keyframes shake {
  0%, 100% { transform: translateX(0); }
  25% { transform: translateX(-4px); }
  75% { transform: translateX(4px); }
}

/* 登录按钮 */
.login-button {
  width: 100%;
  height: var(--touch-target-min);
  margin-top: var(--spacing-2);
  border: none;
  border-radius: var(--radius-lg);
  background: linear-gradient(135deg, var(--primary-500), var(--primary-600));
  color: white;
  font-size: var(--font-size-base);
  font-weight: var(--font-weight-semibold);
  cursor: pointer;
  transition: all var(--transition-fast);
  position: relative;
  overflow: hidden;
}

.login-button::before {
  content: '';
  position: absolute;
  inset: 0;
  background: linear-gradient(135deg, var(--primary-600), var(--primary-700));
  opacity: 0;
  transition: opacity var(--transition-fast);
}

.login-button:hover:not(:disabled)::before {
  opacity: 1;
}

.login-button:hover:not(:disabled) {
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(59, 130, 246, 0.4);
}

.login-button:active:not(:disabled) {
  transform: translateY(0);
}

.login-button:disabled {
  opacity: 0.7;
  cursor: not-allowed;
}

.button-text,
.button-loading {
  position: relative;
  z-index: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: var(--spacing-2);
}

.loading-spinner {
  width: 18px;
  height: 18px;
  animation: spin 1s linear infinite;
}

@keyframes spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}

/* 底部信息 */
.login-footer {
  margin-top: var(--spacing-6);
  text-align: center;
}

.login-footer p {
  margin: 0;
  font-size: var(--font-size-xs);
  color: var(--text-muted);
}

/* 响应式 */
@media (max-width: 640px) {
  .login-card {
    padding: var(--spacing-6);
  }

  .login-header {
    margin-bottom: var(--spacing-6);
  }

  .title {
    font-size: var(--font-size-xl);
  }
}
</style>
