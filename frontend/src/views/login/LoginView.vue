<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useAuthStore } from '@/stores/auth'
import { ElMessage, type FormInstance, type FormRules } from 'element-plus'
import { User, Lock, Connection } from '@element-plus/icons-vue'
import { getOidcLoginUrl } from '@/api/auth'

const router = useRouter()
const route = useRoute()
const authStore = useAuthStore()
const { t } = useI18n()

const formRef = ref<FormInstance>()
const loading = ref(false)
const oidcLoading = ref(false)
const oidcEnabled = ref(false)

const form = reactive({
  username: '',
  password: '',
})

const rules: FormRules = {
  username: [{ required: true, message: t('login.usernameRequired'), trigger: 'blur' }],
  password: [{ required: true, message: t('login.passwordRequired'), trigger: 'blur' }],
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
    const redirect = route.query.redirect
    router.push(typeof redirect === 'string' ? redirect : '/dashboard')
  } catch (e: any) {
    ElMessage.error(e?.message || t('login.loginFailed'))
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
      ElMessage.error(t('login.oidcUrlFailed'))
    }
  } catch {
    ElMessage.error(t('login.oidcFailed'))
  } finally {
    oidcLoading.value = false
  }
}
</script>

<template>
  <div class="login-page">
    <!-- Animated background -->
    <div class="bg-layer">
      <div class="mesh-gradient"></div>
      <div class="grid-pattern"></div>
      <!-- Floating geometric shapes -->
      <div class="floating-shapes">
        <div class="shape hexagon shape-1"></div>
        <div class="shape hexagon shape-2"></div>
        <div class="shape hexagon shape-3"></div>
        <div class="shape circle shape-4"></div>
        <div class="shape circle shape-5"></div>
        <div class="shape ring shape-6"></div>
        <div class="shape ring shape-7"></div>
        <div class="shape dot shape-8"></div>
        <div class="shape dot shape-9"></div>
        <div class="shape dot shape-10"></div>
        <!-- Connection lines -->
        <svg class="connections" viewBox="0 0 1440 900" preserveAspectRatio="none">
          <line x1="200" y1="150" x2="400" y2="350" class="conn-line" />
          <line x1="1100" y1="200" x2="900" y2="450" class="conn-line" />
          <line x1="300" y1="700" x2="550" y2="550" class="conn-line" />
          <line x1="1200" y1="600" x2="950" y2="700" class="conn-line" />
          <line x1="700" y1="100" x2="850" y2="300" class="conn-line delay" />
        </svg>
      </div>
    </div>

    <!-- Login card -->
    <div class="login-container">
      <div class="login-card">
        <!-- Logo -->
        <div class="login-brand">
          <div class="brand-icon">
            <svg viewBox="0 0 48 48" width="48" height="48">
              <defs>
                <linearGradient id="login-hex" x1="0" y1="0" x2="1" y2="1">
                  <stop offset="0%" stop-color="#6366f1" />
                  <stop offset="50%" stop-color="#818cf8" />
                  <stop offset="100%" stop-color="#3b82f6" />
                </linearGradient>
                <linearGradient id="login-inner" x1="0" y1="0" x2="0" y2="1">
                  <stop offset="0%" stop-color="#c7d2fe" />
                  <stop offset="100%" stop-color="#e0e7ff" />
                </linearGradient>
              </defs>
              <path d="M24 2 L44 14 L44 34 L24 46 L4 34 L4 14 Z" fill="url(#login-hex)" />
              <path d="M24 8 L38 16 L38 32 L24 40 L10 32 L10 16 Z" fill="url(#login-inner)" opacity="0.15" />
              <path
                d="M30 16 C26.5 13.5 21.5 13.5 18 16 C14.5 18.5 14 23 14 24 C14 28 16 31 20 33 C23 34.5 27 34.5 30 33 L30 26 L24 26 L24 23 L30 23 Z"
                fill="white"
                opacity="0.95"
              />
              <circle cx="36" cy="12" r="3" fill="#a78bfa" opacity="0.8" />
            </svg>
          </div>
          <h1 class="brand-name">GKube</h1>
          <p class="brand-tagline">{{ t('login.subtitle') }}</p>
        </div>

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
              :placeholder="t('login.usernamePlaceholder')"
              size="large"
              :prefix-icon="User"
              autocomplete="username"
              @keyup.enter="handleLogin"
            />
          </el-form-item>
          <el-form-item prop="password">
            <el-input
              v-model="form.password"
              type="password"
              :placeholder="t('login.passwordPlaceholder')"
              size="large"
              show-password
              :prefix-icon="Lock"
              autocomplete="current-password"
              @keyup.enter="handleLogin"
            />
          </el-form-item>

          <el-form-item>
            <el-button
              type="primary"
              native-type="submit"
              :loading="loading"
              size="large"
              class="login-btn"
            >
              {{ t('login.loginButton') }}
            </el-button>
          </el-form-item>

          <template v-if="oidcEnabled">
            <div class="oidc-divider">
              <span class="divider-line"></span>
              <span class="divider-text">{{ t('login.or') }}</span>
              <span class="divider-line"></span>
            </div>

            <el-form-item>
              <el-button
                :loading="oidcLoading"
                size="large"
                class="oidc-btn"
                @click="handleOIDCLogin"
              >
                <el-icon class="oidc-icon"><Connection /></el-icon>
                {{ t('login.oidcLogin') }}
              </el-button>
            </el-form-item>
          </template>
        </el-form>

        <div class="login-footer">
          <span>Powered by <strong>GKube</strong></span>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.login-page {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  position: relative;
  overflow: hidden;
  background: #0a0a1a;
}

/* ─── Background layers ─── */
.bg-layer {
  position: absolute;
  inset: 0;
  z-index: 0;
}

.mesh-gradient {
  position: absolute;
  inset: 0;
  background:
    radial-gradient(ellipse 80% 60% at 20% 30%, rgba(99, 102, 241, 0.25) 0%, transparent 60%),
    radial-gradient(ellipse 60% 80% at 80% 70%, rgba(59, 130, 246, 0.2) 0%, transparent 60%),
    radial-gradient(ellipse 50% 50% at 50% 50%, rgba(139, 92, 246, 0.1) 0%, transparent 70%);
  animation: meshShift 20s ease-in-out infinite alternate;
}

@keyframes meshShift {
  0% {
    background:
      radial-gradient(ellipse 80% 60% at 20% 30%, rgba(99, 102, 241, 0.25) 0%, transparent 60%),
      radial-gradient(ellipse 60% 80% at 80% 70%, rgba(59, 130, 246, 0.2) 0%, transparent 60%),
      radial-gradient(ellipse 50% 50% at 50% 50%, rgba(139, 92, 246, 0.1) 0%, transparent 70%);
  }
  100% {
    background:
      radial-gradient(ellipse 60% 80% at 70% 60%, rgba(99, 102, 241, 0.3) 0%, transparent 60%),
      radial-gradient(ellipse 80% 60% at 30% 20%, rgba(59, 130, 246, 0.25) 0%, transparent 60%),
      radial-gradient(ellipse 50% 50% at 60% 40%, rgba(139, 92, 246, 0.15) 0%, transparent 70%);
  }
}

.grid-pattern {
  position: absolute;
  inset: 0;
  background-image:
    linear-gradient(rgba(99, 102, 241, 0.04) 1px, transparent 1px),
    linear-gradient(90deg, rgba(99, 102, 241, 0.04) 1px, transparent 1px);
  background-size: 60px 60px;
  mask-image: radial-gradient(ellipse 70% 70% at 50% 50%, black 30%, transparent 70%);
}

/* ─── Floating shapes ─── */
.floating-shapes {
  position: absolute;
  inset: 0;
  pointer-events: none;
}

.shape {
  position: absolute;
  opacity: 0;
  animation: floatIn 1.2s ease-out forwards;
}

.hexagon {
  width: 40px;
  height: 40px;
  background: linear-gradient(135deg, rgba(99, 102, 241, 0.15), rgba(59, 130, 246, 0.1));
  clip-path: polygon(50% 0%, 100% 25%, 100% 75%, 50% 100%, 0% 75%, 0% 25%);
}

.circle {
  width: 24px;
  height: 24px;
  border-radius: 50%;
  border: 2px solid rgba(99, 102, 241, 0.2);
}

.ring {
  width: 56px;
  height: 56px;
  border-radius: 50%;
  border: 1.5px solid rgba(139, 92, 246, 0.12);
}

.dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: rgba(167, 139, 250, 0.4);
}

.shape-1 { top: 10%; left: 8%; animation-delay: 0.2s; }
.shape-2 { top: 25%; right: 12%; animation-delay: 0.5s; width: 28px; height: 28px; }
.shape-3 { bottom: 20%; left: 15%; animation-delay: 0.8s; width: 32px; height: 32px; }
.shape-4 { top: 60%; right: 8%; animation-delay: 0.3s; }
.shape-5 { bottom: 35%; left: 5%; animation-delay: 0.7s; width: 16px; height: 16px; }
.shape-6 { top: 15%; left: 35%; animation-delay: 1s; }
.shape-7 { bottom: 15%; right: 25%; animation-delay: 0.6s; width: 40px; height: 40px; }
.shape-8 { top: 40%; left: 20%; animation-delay: 0.4s; }
.shape-9 { top: 70%; right: 15%; animation-delay: 0.9s; }
.shape-10 { bottom: 40%; left: 40%; animation-delay: 1.1s; width: 8px; height: 8px; }

@keyframes floatIn {
  from { opacity: 0; transform: translateY(20px) scale(0.8); }
  to { opacity: 1; transform: translateY(0) scale(1); }
}

.shape-1 { animation: floatIn 1.2s ease-out 0.2s forwards, floatDrift 12s ease-in-out 1.4s infinite; }
.shape-2 { animation: floatIn 1.2s ease-out 0.5s forwards, floatDrift 15s ease-in-out 1.7s infinite; }
.shape-3 { animation: floatIn 1.2s ease-out 0.8s forwards, floatDrift 18s ease-in-out 2s infinite; }
.shape-6 { animation: floatIn 1.2s ease-out 1s forwards, floatDrift 20s ease-in-out 2.2s infinite; }

@keyframes floatDrift {
  0%, 100% { transform: translateY(0) rotate(0deg); }
  25% { transform: translateY(-12px) rotate(3deg); }
  50% { transform: translateY(-6px) rotate(-2deg); }
  75% { transform: translateY(-18px) rotate(1deg); }
}

/* Connection lines */
.connections {
  position: absolute;
  inset: 0;
  width: 100%;
  height: 100%;
}

.conn-line {
  stroke: rgba(99, 102, 241, 0.08);
  stroke-width: 1;
  stroke-dasharray: 8 6;
  animation: dashFlow 8s linear infinite;
}

.conn-line.delay {
  animation-delay: 3s;
}

@keyframes dashFlow {
  to { stroke-dashoffset: -56; }
}

/* ─── Login card ─── */
.login-container {
  position: relative;
  z-index: 1;
  width: 100%;
  max-width: 440px;
  padding: 0 16px;
}

.login-card {
  padding: 48px 40px 40px;
  background: rgba(15, 15, 35, 0.7);
  border: 1px solid rgba(99, 102, 241, 0.15);
  border-radius: 20px;
  backdrop-filter: blur(24px) saturate(1.4);
  box-shadow:
    0 0 0 1px rgba(255, 255, 255, 0.03),
    0 8px 40px rgba(0, 0, 0, 0.4),
    0 0 80px rgba(99, 102, 241, 0.06);
  animation: cardAppear 0.8s cubic-bezier(0.16, 1, 0.3, 1) forwards;
  opacity: 0;
}

@keyframes cardAppear {
  from {
    opacity: 0;
    transform: translateY(30px) scale(0.97);
  }
  to {
    opacity: 1;
    transform: translateY(0) scale(1);
  }
}

/* ─── Brand section ─── */
.login-brand {
  text-align: center;
  margin-bottom: 40px;
}

.brand-icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 20px;
  animation: iconPulse 3s ease-in-out infinite;
}

@keyframes iconPulse {
  0%, 100% { filter: drop-shadow(0 0 8px rgba(99, 102, 241, 0.3)); }
  50% { filter: drop-shadow(0 0 20px rgba(99, 102, 241, 0.5)); }
}

.brand-name {
  margin: 0 0 8px;
  font-size: 32px;
  font-weight: 700;
  color: #f0f0ff;
  letter-spacing: 3px;
  text-transform: uppercase;
}

.brand-tagline {
  margin: 0;
  font-size: 14px;
  color: rgba(200, 200, 230, 0.6);
  letter-spacing: 0.5px;
}

/* ─── Form styles ─── */
.login-form {
  margin-top: 0;
}

.login-form :deep(.el-form-item) {
  margin-bottom: 22px;
}

.login-form :deep(.el-input__wrapper) {
  background: rgba(255, 255, 255, 0.04);
  border: 1px solid rgba(99, 102, 241, 0.12);
  border-radius: 12px;
  box-shadow: none;
  padding: 4px 14px;
  transition: all 0.3s ease;
}

.login-form :deep(.el-input__wrapper:hover) {
  background: rgba(255, 255, 255, 0.06);
  border-color: rgba(99, 102, 241, 0.25);
}

.login-form :deep(.el-input__wrapper.is-focus) {
  background: rgba(255, 255, 255, 0.06);
  border-color: rgba(99, 102, 241, 0.5);
  box-shadow: 0 0 0 3px rgba(99, 102, 241, 0.1);
}

.login-form :deep(.el-input__inner) {
  color: #e0e0f0;
  font-size: 14px;
}

.login-form :deep(.el-input__inner::placeholder) {
  color: rgba(180, 180, 210, 0.4);
}

.login-form :deep(.el-input__prefix .el-icon) {
  color: rgba(167, 139, 250, 0.6);
}

.login-form :deep(.el-input__suffix .el-icon) {
  color: rgba(167, 139, 250, 0.5);
}

.login-btn {
  width: 100%;
  height: 48px;
  font-size: 15px;
  font-weight: 600;
  border-radius: 12px;
  letter-spacing: 3px;
  background: linear-gradient(135deg, #6366f1 0%, #4f46e5 50%, #3b82f6 100%);
  border: none;
  margin-top: 4px;
  transition: all 0.3s ease;
  position: relative;
  overflow: hidden;
}

.login-btn::before {
  content: '';
  position: absolute;
  inset: 0;
  background: linear-gradient(135deg, #818cf8 0%, #6366f1 50%, #60a5fa 100%);
  opacity: 0;
  transition: opacity 0.3s ease;
}

.login-btn:hover {
  transform: translateY(-1px);
  box-shadow: 0 8px 24px rgba(99, 102, 241, 0.35);
}

.login-btn:hover::before {
  opacity: 1;
}

.login-btn:active {
  transform: translateY(0);
}

/* ─── OIDC divider ─── */
.oidc-divider {
  display: flex;
  align-items: center;
  gap: 16px;
  margin: 8px 0 22px;
}

.divider-line {
  flex: 1;
  height: 1px;
  background: linear-gradient(90deg, transparent, rgba(99, 102, 241, 0.2), transparent);
}

.divider-text {
  color: rgba(180, 180, 210, 0.4);
  font-size: 12px;
  text-transform: uppercase;
  letter-spacing: 2px;
}

.oidc-btn {
  width: 100%;
  height: 48px;
  font-size: 14px;
  border-radius: 12px;
  background: rgba(255, 255, 255, 0.04);
  border: 1px solid rgba(99, 102, 241, 0.15);
  color: #c8c8e6;
  transition: all 0.3s ease;
}

.oidc-btn:hover {
  background: rgba(99, 102, 241, 0.08);
  border-color: rgba(99, 102, 241, 0.3);
  color: #e0e0ff;
}

.oidc-icon {
  margin-right: 8px;
}

/* ─── Footer ─── */
.login-footer {
  text-align: center;
  margin-top: 32px;
  font-size: 12px;
  color: rgba(150, 150, 180, 0.35);
}

.login-footer strong {
  color: rgba(167, 139, 250, 0.5);
  font-weight: 600;
}

/* ─── Responsive ─── */
@media (max-width: 480px) {
  .login-card {
    padding: 36px 24px 32px;
    border-radius: 16px;
  }

  .brand-name {
    font-size: 26px;
    letter-spacing: 2px;
  }

  .login-btn,
  .oidc-btn {
    height: 44px;
  }

  .hexagon,
  .ring,
  .connections {
    display: none;
  }
}
</style>
