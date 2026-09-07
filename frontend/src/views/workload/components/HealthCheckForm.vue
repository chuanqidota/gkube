<script setup lang="ts">
import type { Container, LifecycleHandler } from './form-types'
import ProbeForm from './form/ProbeForm.vue'

defineProps<{
  containers: Container[]
}>()

function newLifecycleHandler(): LifecycleHandler {
  return { type: 'exec', execCommand: '', httpGetPath: '/', httpGetPort: 80, tcpSocketPort: null }
}
</script>

<template>
  <div class="health-check-form">
    <div v-for="(container, ci) in containers" :key="ci" style="margin-bottom: 24px;">
      <div class="mount-container-name">{{ container.name || `容器 ${ci + 1}` }}</div>

      <!-- Liveness -->
      <ProbeForm
        v-model="container.livenessProbe"
        label="存活探针"
        description="容器是否正在运行"
      />

      <!-- Readiness -->
      <ProbeForm
        v-model="container.readinessProbe"
        label="就绪探针"
        description="容器是否准备好接收流量"
      />

      <!-- Startup Probe -->
      <ProbeForm
        v-model="container.startupProbe"
        label="启动探针"
        description="慢启动应用专用，成功后切换到存活探针"
      />

      <!-- Lifecycle Hooks -->
      <div class="probe-card">
        <div class="probe-header">
          <div>
            <span class="probe-label">生命周期钩子</span>
            <span class="probe-desc">容器启动后/停止前执行的操作</span>
          </div>
        </div>
        <div style="margin-top: 12px;">
          <div style="margin-bottom: 12px;">
            <div style="display: flex; align-items: center; gap: 12px; margin-bottom: 8px;">
              <span style="font-size: 13px; font-weight: 600; color: var(--el-text-color-regular);">postStart（启动后）</span>
              <el-switch :model-value="!!container.lifecycle.postStart" @update:model-value="(v: boolean) => v ? (container.lifecycle.postStart = newLifecycleHandler()) : (container.lifecycle.postStart = null)" />
            </div>
            <template v-if="container.lifecycle.postStart">
              <div class="fields-grid">
                <el-form-item label="类型">
                  <el-select v-model="container.lifecycle.postStart.type" style="width: 100%;">
                    <el-option label="Exec" value="exec" />
                    <el-option label="HTTP GET" value="httpGet" />
                    <el-option label="TCP Socket" value="tcpSocket" />
                  </el-select>
                </el-form-item>
                <el-form-item v-if="container.lifecycle.postStart.type === 'exec'" label="命令">
                  <el-input type="textarea" :autosize="{ minRows: 1, maxRows: 4 }" v-model="container.lifecycle.postStart.execCommand" placeholder="每行一个参数" />
                </el-form-item>
                <el-form-item v-if="container.lifecycle.postStart.type === 'httpGet'" label="路径">
                  <el-input v-model="container.lifecycle.postStart.httpGetPath" placeholder="/" />
                </el-form-item>
                <el-form-item v-if="container.lifecycle.postStart.type === 'httpGet'" label="端口">
                  <el-input-number v-model="container.lifecycle.postStart.httpGetPort" :min="1" :max="65535" style="width: 100%;" />
                </el-form-item>
                <el-form-item v-if="container.lifecycle.postStart.type === 'tcpSocket'" label="端口">
                  <el-input-number v-model="container.lifecycle.postStart.tcpSocketPort" :min="1" :max="65535" style="width: 100%;" />
                </el-form-item>
              </div>
            </template>
          </div>
          <div>
            <div style="display: flex; align-items: center; gap: 12px; margin-bottom: 8px;">
              <span style="font-size: 13px; font-weight: 600; color: var(--el-text-color-regular);">preStop（停止前）</span>
              <el-switch :model-value="!!container.lifecycle.preStop" @update:model-value="(v: boolean) => v ? (container.lifecycle.preStop = newLifecycleHandler()) : (container.lifecycle.preStop = null)" />
            </div>
            <template v-if="container.lifecycle.preStop">
              <div class="fields-grid">
                <el-form-item label="类型">
                  <el-select v-model="container.lifecycle.preStop.type" style="width: 100%;">
                    <el-option label="Exec" value="exec" />
                    <el-option label="HTTP GET" value="httpGet" />
                    <el-option label="TCP Socket" value="tcpSocket" />
                  </el-select>
                </el-form-item>
                <el-form-item v-if="container.lifecycle.preStop.type === 'exec'" label="命令">
                  <el-input type="textarea" :autosize="{ minRows: 1, maxRows: 4 }" v-model="container.lifecycle.preStop.execCommand" placeholder="每行一个参数" />
                </el-form-item>
                <el-form-item v-if="container.lifecycle.preStop.type === 'httpGet'" label="路径">
                  <el-input v-model="container.lifecycle.preStop.httpGetPath" placeholder="/" />
                </el-form-item>
                <el-form-item v-if="container.lifecycle.preStop.type === 'httpGet'" label="端口">
                  <el-input-number v-model="container.lifecycle.preStop.httpGetPort" :min="1" :max="65535" style="width: 100%;" />
                </el-form-item>
                <el-form-item v-if="container.lifecycle.preStop.type === 'tcpSocket'" label="端口">
                  <el-input-number v-model="container.lifecycle.preStop.tcpSocketPort" :min="1" :max="65535" style="width: 100%;" />
                </el-form-item>
              </div>
            </template>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.health-check-form {
  /* Container for the health check section */
}

.mount-container-name {
  font-size: 14px;
  font-weight: 600;
  color: var(--el-text-color-primary);
  margin-bottom: 12px;
  padding-bottom: 8px;
  border-bottom: 1px solid var(--el-border-color-lighter);
}

.probe-card {
  background: var(--el-fill-color-blank);
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 8px;
  padding: 16px;
  margin-bottom: 12px;
}

.probe-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.probe-label {
  font-size: 13px;
  font-weight: 600;
  color: var(--el-text-color-regular);
}

.probe-desc {
  font-size: 12px;
  color: var(--el-text-color-placeholder);
  margin-left: 8px;
}

.fields-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
  gap: 12px;
}
</style>
