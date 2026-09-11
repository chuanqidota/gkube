<script setup lang="ts">
import { ref, onErrorCaptured } from 'vue'

const error = ref<Error | null>(null)
const retryCount = ref(0)

onErrorCaptured((err) => {
  error.value = err
  console.error('[ErrorBoundary]', err)
  return false
})

function retry() {
  error.value = null
  retryCount.value++
}
</script>

<template>
  <div v-if="error" class="error-boundary">
    <el-result icon="error" title="页面渲染出错" :sub-title="error.message">
      <template #extra>
        <el-button type="primary" @click="retry">重试</el-button>
        <el-button @click="$router.push('/')">返回首页</el-button>
      </template>
    </el-result>
  </div>
  <div v-else :key="retryCount">
    <slot />
  </div>
</template>

<style scoped>
.error-boundary {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 400px;
}
</style>
