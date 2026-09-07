<script setup lang="ts">
import { Plus, Delete } from '@element-plus/icons-vue'
import type { Container } from '../form-types'
import { createEmptyContainer, createEmptyEnv } from '../form-types'

const props = withDefaults(defineProps<{
  containers: Container[]
  title?: string
  showPullPolicy?: boolean
  minContainers?: number
  indexColor?: string
}>(), {
  title: '容器配置',
  showPullPolicy: true,
  minContainers: 1,
  indexColor: 'var(--el-color-primary)',
})

function addContainer() { props.containers.push(createEmptyContainer()) }
function removeContainer(i: number) { if (props.containers.length > props.minContainers) props.containers.splice(i, 1) }
function addPort(ci: number) { props.containers[ci].ports.push({ name: '', containerPort: null, protocol: 'TCP' }) }
function removePort(ci: number, pi: number) { props.containers[ci].ports.splice(pi, 1) }
function addEnv(ci: number) { props.containers[ci].env.push(createEmptyEnv()) }
function removeEnv(ci: number, ei: number) { props.containers[ci].env.splice(ei, 1) }
</script>

<template>
  <div v-for="(container, ci) in containers" :key="ci" class="container-card">
    <div class="container-card-header">
      <div class="container-title">
        <span class="container-index" :style="{ background: indexColor }">{{ ci + 1 }}</span>
        <span>{{ container.name || '未命名容器' }}</span>
      </div>
      <el-button v-if="containers.length > minContainers" type="danger" text size="small" @click="removeContainer(ci)">
        <el-icon><Delete /></el-icon> 移除
      </el-button>
    </div>
    <div class="fields-grid">
      <el-form-item label="容器名称" required>
        <el-input v-model="container.name" placeholder="nginx" />
      </el-form-item>
      <el-form-item label="镜像" required>
        <el-input v-model="container.image" placeholder="nginx:1.25" />
      </el-form-item>
      <el-form-item v-if="showPullPolicy" label="拉取策略">
        <el-select v-model="container.imagePullPolicy" style="width: 100%;">
          <el-option label="Always" value="Always" />
          <el-option label="IfNotPresent" value="IfNotPresent" />
          <el-option label="Never" value="Never" />
        </el-select>
      </el-form-item>
      <el-form-item label="启动命令 (command)">
        <el-input type="textarea" :autosize="{ minRows: 1, maxRows: 4 }" v-model="container.command" placeholder="每行一个参数，如: /bin/sh" />
      </el-form-item>
      <el-form-item label="启动参数 (args)" class="full-width">
        <el-input type="textarea" :autosize="{ minRows: 1, maxRows: 4 }" v-model="container.args" placeholder="每行一个参数，如: -c echo hello" />
      </el-form-item>
    </div>

    <!-- Ports -->
    <el-divider content-position="left">端口</el-divider>
    <div v-for="(port, pi) in container.ports" :key="pi" class="kv-row">
      <el-input v-model="port.name" placeholder="名称" style="width: 120px;" />
      <el-input-number v-model="port.containerPort" :min="1" :max="65535" placeholder="端口" style="flex: 1;" />
      <el-select v-model="port.protocol" style="width: 100px;">
        <el-option label="TCP" value="TCP" /><el-option label="UDP" value="UDP" />
      </el-select>
      <el-button type="danger" text circle @click="removePort(ci, pi)">
        <el-icon><Delete /></el-icon>
      </el-button>
    </div>
    <el-button text type="primary" size="small" @click="addPort(ci)">
      <el-icon><Plus /></el-icon> 添加端口
    </el-button>

    <!-- Env -->
    <el-divider content-position="left">环境变量</el-divider>
    <div v-for="(env, ei) in container.env" :key="ei" class="env-row">
      <el-input v-model="env.name" placeholder="名称" style="width: 140px;" />
      <el-select v-model="env.type" style="width: 140px;" @change="env.value = ''; env.configMapName = ''; env.configMapKey = ''; env.secretName = ''; env.secretKey = ''; env.fieldPath = ''">
        <el-option label="直接值" value="plain" />
        <el-option label="ConfigMap" value="configMapKeyRef" />
        <el-option label="Secret" value="secretKeyRef" />
        <el-option label="字段引用" value="fieldRef" />
      </el-select>
      <el-input v-if="env.type === 'plain'" v-model="env.value" placeholder="值" style="flex: 1;" />
      <template v-if="env.type === 'configMapKeyRef'">
        <el-input v-model="env.configMapName" placeholder="ConfigMap 名称" style="flex: 1;" />
        <el-input v-model="env.configMapKey" placeholder="Key" style="width: 140px;" />
      </template>
      <template v-if="env.type === 'secretKeyRef'">
        <el-input v-model="env.secretName" placeholder="Secret 名称" style="flex: 1;" />
        <el-input v-model="env.secretKey" placeholder="Key" style="width: 140px;" />
      </template>
      <el-input v-if="env.type === 'fieldRef'" v-model="env.fieldPath" placeholder="如: metadata.name" style="flex: 1;" />
      <el-button type="danger" text circle @click="removeEnv(ci, ei)">
        <el-icon><Delete /></el-icon>
      </el-button>
    </div>
    <el-button text type="primary" size="small" @click="addEnv(ci)">
      <el-icon><Plus /></el-icon> 添加环境变量
    </el-button>

    <!-- Resources -->
    <el-divider content-position="left">资源配额</el-divider>
    <div class="resources-grid">
      <div class="resource-group">
        <div class="resource-group-title">Requests</div>
        <div class="resource-fields">
          <el-form-item label="CPU"><el-input v-model="container.resources.requests.cpu" placeholder="100m" /></el-form-item>
          <el-form-item label="Memory"><el-input v-model="container.resources.requests.memory" placeholder="128Mi" /></el-form-item>
        </div>
      </div>
      <div class="resource-group">
        <div class="resource-group-title">Limits</div>
        <div class="resource-fields">
          <el-form-item label="CPU"><el-input v-model="container.resources.limits.cpu" placeholder="500m" /></el-form-item>
          <el-form-item label="Memory"><el-input v-model="container.resources.limits.memory" placeholder="512Mi" /></el-form-item>
        </div>
      </div>
    </div>
  </div>
  <el-button text type="primary" @click="addContainer" class="add-container-btn">
    <el-icon><Plus /></el-icon> 添加{{ title }}
  </el-button>
</template>

<style scoped>
.container-card {
  border: 1px solid var(--el-border-color-extra-light);
  border-radius: var(--gk-radius-md);
  padding: 16px;
  margin-bottom: 12px;
  background: var(--el-fill-color-blank);
  transition: box-shadow 0.2s;
}

.container-card:hover {
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.04);
}

.container-card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
  padding-bottom: 10px;
  border-bottom: 1px solid var(--el-border-color-extra-light);
}

.container-title {
  display: flex;
  align-items: center;
  gap: 10px;
  font-weight: 600;
  font-size: 15px;
  color: var(--el-text-color-primary);
}

.container-index {
  width: 26px;
  height: 26px;
  border-radius: 50%;
  color: #fff;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 13px;
  font-weight: 600;
}

.add-container-btn {
  margin-top: 8px;
}

.fields-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 0 32px;
}

.fields-grid :deep(.el-form-item) {
  margin-bottom: 16px;
}

.fields-grid :deep(.el-form-item:last-child) {
  margin-bottom: 0;
}

.fields-grid :deep(.el-form-item.full-width) {
  grid-column: 1 / -1;
}

.kv-row {
  display: flex;
  gap: 8px;
  margin-bottom: 8px;
  align-items: center;
}

.kv-row :deep(.el-input) {
  flex: 1;
}

.env-row {
  display: flex;
  gap: 8px;
  margin-bottom: 8px;
  align-items: center;
}

.env-row :deep(.el-input) {
  flex: 1;
}

.resources-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 20px;
}

.resource-group {
  border: 1px solid var(--el-border-color-extra-light);
  border-radius: var(--gk-radius-md);
  padding: 14px;
  background: var(--el-fill-color-lighter);
}

.resource-group-title {
  font-size: 13px;
  font-weight: 600;
  color: var(--el-text-color-regular);
  margin-bottom: 12px;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.resource-fields {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 16px;
}

.resource-fields :deep(.el-form-item) {
  margin-bottom: 0;
}
</style>
