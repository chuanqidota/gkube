<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { ElMessage, type FormInstance, type FormRules } from 'element-plus'
import { Ship, User, Lock, Connection } from '@element-plus/icons-vue'
import { getOidcLoginUrl } from '@/api/auth'

const router = useRouter()
const authStore = useAuthStore()

const formRef = ref<FormInstance>()
const loading = ref(false)
const oidcLoading = ref(false)
const oidcEnabled = ref(false)
const rememberMe = ref(false)

const form = reactive({
  username: '',
  password: '',
})

const rules: FormRules = {
  username: [{ required: true, message: '请输入用户名', trigger: 'blur' }],
  password: [{ required: true, message: '请输入密码', trigger: 'blur' }],
}

onMounted(async () => {
  try {
    await getOidcLoginUrl()
    oidcEnabled.value = true
  } catch {
    oidcEnabled.value = false
  }
})

async function handleLogin() {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return

  loading.value = true
  try {
    await authStore.login({ username: form.username, password: form.password })
    router.push('/dashboard')
  } catch (e: any) {
    ElMessage.error(e?.message || '登录失败')
  } finally {
    loading.value = false
  }
}

async function handleOIDCLogin() {
  oidcLoading.value = true
  try {
    const res = await getOidcLoginUrl()
    if (res.data?.url) {
      window.location.href = res.data.url
    } else {
      ElMessage.error('获取 OIDC 登录地址失败')
    }
  } catch {
    ElMessage.error('OIDC 登录失败')
  } finally {
    oidcLoading.value = false
  }
}
</script>

<template>
  <div class="login-wrapper">
    <div class="login-card">
      <!-- Logo -->
      <div class="login-logo">
        <el-icon :size="48" color="#409EFF"><Ship /></el-icon>
      </div>
      <h1 class="login-title">gkube</h1>
      <p class="login-subtitle">Kubernetes 集群管理平台</p>

      <!-- Form -->
      <el-form
        ref="formRef"
        :model="form"
        :rules="rules"
        @submit.prevent="handleLogin"
        class="login-form"
      >
        <el-form-item prop="username">
          <el-input
            v-model="form.username"
            placeholder="请输入用户名"
            size="large"
            :prefix-icon="User"
          />
        </el-form-item>
        <el-form-item prop="password">
          <el-input
            v-model="form.password"
            type="password"
            placeholder="请输入密码"
            size="large"
            show-password
            :prefix-icon="Lock"
            @keyup.enter="handleLogin"
          />
        </el-form-item>

        <div class="login-options">
          <el-checkbox v-model="rememberMe">记住我</el-checkbox>
        </div>

        <el-form-item>
          <el-button
            type="primary"
            :loading="loading"
            size="large"
            class="login-btn"
            @click="handleLogin"
          >
            登 录
          </el-button>
        </el-form-item>

        <div v-if="oidcEnabled" class="oidc-divider">
          <el-divider>
            <span class="divider-text">或</span>
          </el-divider>
        </div>

        <el-form-item v-if="oidcEnabled">
          <el-button
            :loading="oidcLoading"
            size="large"
            class="oidc-btn"
            @click="handleOIDCLogin"
          >
            <el-icon class="oidc-icon"><Connection /></el-icon>
            使用 OIDC 登录
          </el-button>
        </el-form-item>
      </el-form>

      <div class="login-footer">
        <span>Powered by gkube</span>
      </div>
    </div>
  </div>
</template>

<style scoped>
.login-wrapper {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #1a1a2e 0%, #16213e 50%, #0f3460 100%);
  position: relative;
  overflow: hidden;
}

.login-wrapper::before {
  content: '';
  position: absolute;
  top: -50%;
  left: -50%;
  width: 200%;
  height: 200%;
  background: radial-gradient(circle, rgba(64, 158, 255, 0.08) 0%, transparent 60%);
  animation: pulse 8s ease-in-out infinite;
}

@keyframes pulse {
  0%, 100% { transform: scale(1); opacity: 0.5; }
  50% { transform: scale(1.1); opacity: 1; }
}

.login-card {
  width: 420px;
  padding: 48px 40px 36px;
  background: rgba(255, 255, 255, 0.95);
  border-radius: 16px;
  box-shadow: 0 25px 80px rgba(0, 0, 0, 0.3);
  backdrop-filter: blur(10px);
  position: relative;
  z-index: 1;
}

.login-logo {
  text-align: center;
  margin-bottom: 16px;
}

.login-title {
  text-align: center;
  margin: 0 0 8px;
  font-size: 32px;
  font-weight: 700;
  color: #1a1a2e;
  letter-spacing: 2px;
}

.login-subtitle {
  text-align: center;
  margin: 0 0 36px;
  font-size: 14px;
  color: #909399;
}

.login-form {
  margin-top: 8px;
}

.login-form :deep(.el-input__wrapper) {
  border-radius: 8px;
  box-shadow: 0 0 0 1px #dcdfe6 inset;
  padding: 4px 12px;
}

.login-form :deep(.el-input__wrapper:hover) {
  box-shadow: 0 0 0 1px #c0c4cc inset;
}

.login-form :deep(.el-input__wrapper.is-focus) {
  box-shadow: 0 0 0 1px #409eff inset;
}

.login-options {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.login-btn {
  width: 100%;
  height: 44px;
  font-size: 16px;
  font-weight: 600;
  border-radius: 8px;
  letter-spacing: 4px;
}

.oidc-divider {
  margin: 0 0 16px;
}

.divider-text {
  color: #909399;
  font-size: 12px;
}

.oidc-btn {
  width: 100%;
  height: 44px;
  font-size: 14px;
  border-radius: 8px;
  background: #f5f7fa;
  border-color: #dcdfe6;
  color: #606266;
}

.oidc-btn:hover {
  background: #ecf5ff;
  border-color: #b3d8ff;
  color: #409eff;
}

.oidc-icon {
  margin-right: 8px;
}

.login-footer {
  text-align: center;
  margin-top: 24px;
  font-size: 12px;
  color: #c0c4cc;
}
</style>
