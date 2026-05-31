<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { ElMessage } from 'element-plus'

const router = useRouter()
const authStore = useAuthStore()

const form = ref({ username: '', password: '' })
const loading = ref(false)

async function handleLogin() {
  loading.value = true
  try {
    await authStore.login(form.value)
    router.push('/dashboard')
  } catch (e: any) {
    ElMessage.error(e?.message || 'Login failed')
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="login-container">
    <h2>gkube</h2>
    <el-form @submit.prevent="handleLogin">
      <el-form-item>
        <el-input v-model="form.username" placeholder="Username" />
      </el-form-item>
      <el-form-item>
        <el-input v-model="form.password" type="password" placeholder="Password" />
      </el-form-item>
      <el-button type="primary" :loading="loading" @click="handleLogin" style="width:100%">
        Login
      </el-button>
    </el-form>
  </div>
</template>

<style scoped>
.login-container {
  max-width: 360px;
  margin: 160px auto;
  padding: 24px;
}
</style>
