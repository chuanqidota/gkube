<template>
  <div class="oidc-callback">
    <el-card class="callback-card">
      <template #header>
        <div class="card-header">
          <span>正在完成登录...</span>
        </div>
      </template>
      <div class="callback-content">
        <el-icon v-if="loading" class="is-loading" :size="48">
          <Loading />
        </el-icon>
        <el-result
          v-else-if="error"
          icon="error"
          :title="error"
          sub-title="请返回登录页面重试"
        >
          <template #extra>
            <el-button type="primary" @click="goToLogin">返回登录</el-button>
          </template>
        </el-result>
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { Loading } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import { handleOidcCallback } from '@/api/auth'

const router = useRouter()
const route = useRoute()
const authStore = useAuthStore()

const loading = ref(true)
const error = ref('')

onMounted(async () => {
  const code = route.query.code as string
  const state = route.query.state as string

  if (!code || !state) {
    error.value = '无效的回调参数'
    loading.value = false
    return
  }

  try {
    const res = await handleOidcCallback(code, state)
    const data = res.data

    authStore.setToken(data.accessToken)
    authStore.setRefreshToken(data.refreshToken)
    authStore.setUserInfo(data.user)
    ElMessage.success('OIDC 登录成功')
    router.push('/')
  } catch (err: any) {
    error.value = err?.message || '网络错误，请重试'
  } finally {
    loading.value = false
  }
})

const goToLogin = () => {
  router.push('/login')
}
</script>

<style scoped>
.oidc-callback {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 100vh;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
}

.callback-card {
  width: 400px;
}

.card-header {
  text-align: center;
  font-size: 18px;
  font-weight: bold;
}

.callback-content {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 200px;
}
</style>
