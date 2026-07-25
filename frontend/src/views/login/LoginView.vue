<script setup lang="ts">
import { ref, reactive, watch, computed, nextTick, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useAuthStore } from '@/stores/auth'
import { ElMessage, type FormInstance, type FormRules } from 'element-plus'
import { User, Lock } from '@element-plus/icons-vue'
import Logo from '@/components/Logo.vue'
import { useClusterGraph, graphState } from '@/composables/useClusterGraph'

const router = useRouter()
const route = useRoute()
const authStore = useAuthStore()
const { t } = useI18n()

const formRef = ref<FormInstance>()
const loading = ref(false)
const usernameRef = ref()
const canvasRef = ref<HTMLCanvasElement>()

// Ambient control-plane canvas. The composable reads graphState (below) to
// react to form focus / typing / submit / error without this view re-rendering.
useClusterGraph(canvasRef)

const form = reactive({
  username: '',
  password: '',
})

const rules: FormRules = {
  username: [{ required: true, message: t('login.usernameRequired'), trigger: 'blur' }],
  password: [{ required: true, message: t('login.passwordRequired'), trigger: 'blur' }],
}

// ── Left-right linkage ──
// Typing a character bumps typing.at so the canvas spawns a particle burst at
// the cluster hub matching the active field's zone.
watch(
  () => form.username,
  () => {
    if (graphState.activeField === 'username') {
      graphState.typing = { field: 'username', at: Date.now() }
    }
  },
)
watch(
  () => form.password,
  () => {
    if (graphState.activeField === 'password') {
      graphState.typing = { field: 'password', at: Date.now() }
    }
  },
)

function onUsernameFocus() {
  graphState.activeField = 'username'
}
function onPasswordFocus() {
  graphState.activeField = 'password'
}
function onBlur() {
  // only clear if neither field holds focus
  void nextTick(() => {
    const ae = document.activeElement
    if (!ae || !ae.closest('.login-form')) graphState.activeField = null
  })
}

const canSubmit = computed(() => form.username.trim().length > 0 && form.password.length > 0)

async function handleLogin() {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return

  loading.value = true
  graphState.status = 'loading'
  graphState.burstTick++
  try {
    await authStore.login({ username: form.username, password: form.password })
    const redirect = route.query.redirect
    router.push(typeof redirect === 'string' ? redirect : '/dashboard')
  } catch (e: any) {
    graphState.status = 'error'
    ElMessage.error(e?.message || t('login.loginFailed'))
  } finally {
    loading.value = false
  }
}

// Autofocus the username field so the canvas shows an active zone on load.
onMounted(() => {
  usernameRef.value?.focus?.()
})
</script>

<template>
  <div class="login-page">
    <!-- Left: ambient control-plane canvas -->
    <section class="canvas-panel">
      <div class="canvas-atmosphere" aria-hidden="true"></div>
      <canvas ref="canvasRef" class="graph-canvas" aria-hidden="true" />
      <div class="canvas-head">
        <Logo :size="36" show-text :text-size="20" tone="light" />
      </div>
      <div class="canvas-foot">
        <span class="mono">{{ t('login.clusterCaption') }}</span>
      </div>
    </section>

    <!-- Right: form -->
    <section class="form-panel">
      <div class="form-card">
        <header class="form-head">
          <h1 class="form-title">{{ t('login.welcome') }}</h1>
          <p class="form-subtitle">{{ t('login.subtitle') }}</p>
        </header>

        <el-form
          ref="formRef"
          :model="form"
          :rules="rules"
          label-position="top"
          @submit.prevent="handleLogin"
          class="login-form"
        >
          <el-form-item :label="t('login.usernameLabel')" prop="username">
            <el-input
              ref="usernameRef"
              v-model="form.username"
              :placeholder="t('login.usernamePlaceholder')"
              size="large"
              :prefix-icon="User"
              autocomplete="username"
              @focus="onUsernameFocus"
              @blur="onBlur"
              @keyup.enter="handleLogin"
            />
          </el-form-item>
          <el-form-item :label="t('login.passwordLabel')" prop="password">
            <el-input
              v-model="form.password"
              type="password"
              :placeholder="t('login.passwordPlaceholder')"
              size="large"
              show-password
              :prefix-icon="Lock"
              autocomplete="current-password"
              @focus="onPasswordFocus"
              @blur="onBlur"
              @keyup.enter="handleLogin"
            />
          </el-form-item>

          <el-form-item>
            <el-button
              type="primary"
              native-type="submit"
              :loading="loading"
              :disabled="!canSubmit"
              size="large"
              class="login-btn"
            >
              {{ t('login.loginButton') }}
            </el-button>
          </el-form-item>
        </el-form>

        <footer class="form-foot">
          <span class="mono">Powered by GKube</span>
        </footer>
      </div>
    </section>
  </div>
</template>

<style scoped>
.login-page {
  min-height: 100dvh;
  display: grid;
  grid-template-columns: 58fr 42fr;
  background: #07080f;
  /* One theme for the whole page (always-dark gateway). No mid-page flips. */
  color: #e2e8f0;
}

/* ─── Canvas panel ─── */
.canvas-panel {
  position: relative;
  display: flex;
  flex-direction: column;
  background: #07080f;
  border-right: 1px solid rgba(148, 163, 184, 0.08);
  overflow: hidden;
}

.canvas-atmosphere {
  position: absolute;
  inset: 0;
  z-index: 0;
  pointer-events: none;
  background:
    radial-gradient(ellipse 60% 50% at 28% 38%, rgba(59, 130, 246, 0.14) 0%, transparent 62%),
    radial-gradient(ellipse 50% 60% at 78% 72%, rgba(99, 102, 241, 0.10) 0%, transparent 60%),
    radial-gradient(ellipse 90% 70% at 50% 120%, rgba(59, 130, 246, 0.06) 0%, transparent 70%),
    linear-gradient(180deg, #090b14 0%, #07080f 55%, #05060c 100%);
}

.canvas-atmosphere::before {
  content: '';
  position: absolute;
  inset: 0;
  background-image: radial-gradient(rgba(148, 163, 184, 0.10) 0.5px, transparent 0.5px);
  background-size: 26px 26px;
  mask-image: radial-gradient(ellipse 75% 75% at 50% 50%, black 25%, transparent 80%);
  -webkit-mask-image: radial-gradient(ellipse 75% 75% at 50% 50%, black 25%, transparent 80%);
}

.canvas-atmosphere::after {
  content: '';
  position: absolute;
  inset: 0;
  background:
    radial-gradient(ellipse 100% 100% at 50% 50%, transparent 55%, rgba(0, 0, 0, 0.55) 100%);
}

.graph-canvas {
  position: absolute;
  inset: 0;
  z-index: 1;
  width: 100%;
  height: 100%;
  display: block;
}

.canvas-head {
  position: relative;
  z-index: 2;
  padding: var(--gk-space-8) var(--gk-space-10);
}

.canvas-foot {
  position: relative;
  z-index: 2;
  margin-top: auto;
  padding: var(--gk-space-6) var(--gk-space-10);
}

.canvas-foot::before {
  content: '';
  position: absolute;
  left: var(--gk-space-10);
  top: 0;
  width: 28px;
  height: 1px;
  background: linear-gradient(90deg, var(--gk-color-primary), transparent);
}

.mono {
  font-family: var(--gk-font-mono);
  font-size: 12px;
  letter-spacing: 0.06em;
  color: rgba(148, 163, 184, 0.6);
}

/* ─── Form panel ─── */
.form-panel {
  display: flex;
  align-items: center;
  justify-content: center;
  padding: var(--gk-space-8) var(--gk-space-6);
  background: #0f1117;
}

.form-card {
  width: 100%;
  max-width: 380px;
}

.form-head {
  margin-bottom: var(--gk-space-10);
}

.form-title {
  margin: 0 0 var(--gk-space-3);
  font-family: var(--gk-font-sans);
  font-size: 30px;
  font-weight: 700;
  letter-spacing: -0.025em;
  line-height: 1.1;
  color: #f8fafc;
}

.form-subtitle {
  margin: 0;
  font-size: 14px;
  line-height: 1.5;
  color: rgba(148, 163, 184, 0.75);
  max-width: 38ch;
}

/* ─── Form controls (token-driven, WCAG AA on dark) ─── */
.login-form :deep(.el-form-item) {
  margin-bottom: var(--gk-space-6);
}

.login-form :deep(.el-form-item__label) {
  color: rgba(203, 213, 225, 0.92);
  font-size: 13px;
  font-weight: 500;
  padding-bottom: var(--gk-space-2);
  line-height: 1.4;
}

.login-form :deep(.el-input__wrapper) {
  background: rgba(148, 163, 184, 0.05);
  border: 1px solid rgba(148, 163, 184, 0.16);
  border-radius: var(--gk-radius-md);
  box-shadow: none;
  padding: 8px 14px;
  transition: border-color var(--gk-transition-base), background var(--gk-transition-base),
    box-shadow var(--gk-transition-base), transform var(--gk-transition-fast);
}

.login-form :deep(.el-input__wrapper:hover) {
  background: rgba(148, 163, 184, 0.09);
  border-color: rgba(96, 165, 250, 0.4);
}

.login-form :deep(.el-input__wrapper.is-focus) {
  background: rgba(59, 130, 246, 0.08);
  border-color: var(--gk-color-primary);
  box-shadow: 0 0 0 4px rgba(59, 130, 246, 0.16);
  transform: translateY(-1px);
}

.login-form :deep(.el-input__inner) {
  color: #e2e8f0;
  font-size: 14px;
}

.login-form :deep(.el-input__inner::placeholder) {
  color: rgba(148, 163, 184, 0.45);
}

.login-form :deep(.el-input__prefix .el-icon),
.login-form :deep(.el-input__suffix .el-icon) {
  color: rgba(148, 163, 184, 0.7);
}

.login-form :deep(.el-input__wrapper.is-focus .el-input__prefix .el-icon) {
  color: var(--gk-color-primary-light);
}

.login-btn {
  width: 100%;
  height: 50px;
  font-size: 15px;
  font-weight: 600;
  border-radius: var(--gk-radius-md);
  letter-spacing: 0.04em;
  background: var(--gk-color-primary);
  border: none;
  margin-top: var(--gk-space-2);
  transition: transform var(--gk-transition-fast), box-shadow var(--gk-transition-base),
    background var(--gk-transition-base), opacity var(--gk-transition-fast);
}

.login-btn:hover:not(:disabled):not(.is-loading) {
  background: var(--gk-color-primary-light);
  box-shadow: 0 10px 28px rgba(59, 130, 246, 0.36);
  transform: translateY(-1px);
}

.login-btn:active:not(:disabled):not(.is-loading) {
  transform: translateY(1px);
  box-shadow: 0 4px 12px rgba(59, 130, 246, 0.24);
}

.login-btn:disabled {
  opacity: 0.45;
  cursor: not-allowed;
}

.form-foot {
  margin-top: var(--gk-space-8);
  text-align: center;
}

/* ─── Responsive: collapse to single column under 768px ─── */
@media (max-width: 768px) {
  .login-page {
    grid-template-columns: 1fr;
    grid-template-rows: 32vh 1fr;
  }

  .canvas-panel {
    border-right: none;
    border-bottom: 1px solid rgba(148, 163, 184, 0.08);
  }

  .canvas-head {
    padding: var(--gk-space-4) var(--gk-space-5);
  }

  .canvas-foot {
    display: none;
  }

  .form-panel {
    padding: var(--gk-space-6) var(--gk-space-5);
  }

  .form-head {
    margin-bottom: var(--gk-space-6);
  }

  .form-title {
    font-size: 24px;
  }
}
</style>
