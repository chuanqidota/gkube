<script setup lang="ts">
import { Plus, Delete } from '@element-plus/icons-vue'
import type { Container } from '../form-types'

defineProps<{
  containers: Container[]
}>()

function addCapability(sc: Container['securityContext'], type: 'add' | 'drop') { (type === 'add' ? sc.capabilitiesAdd : sc.capabilitiesDrop).push('') }
function removeCapability(sc: Container['securityContext'], type: 'add' | 'drop', i: number) { (type === 'add' ? sc.capabilitiesAdd : sc.capabilitiesDrop).splice(i, 1) }
</script>

<template>
  <div v-for="(container, ci) in containers" :key="ci" style="margin-bottom: 24px;">
    <div class="mount-container-name">{{ container.name || `容器 ${ci + 1}` }}</div>
    <div class="security-grid">
      <div class="security-item">
        <div class="security-item-label">运行用户 ID</div>
        <el-input-number v-model="container.securityContext.runAsUser" :min="0" placeholder="UID" style="width: 100%;" />
      </div>
      <div class="security-item">
        <div class="security-item-label">非 Root 运行</div>
        <el-switch v-model="container.securityContext.runAsNonRoot" />
      </div>
      <div class="security-item">
        <div class="security-item-label">只读根文件系统</div>
        <el-switch v-model="container.securityContext.readOnlyRootFilesystem" />
      </div>
      <div class="security-item">
        <div class="security-item-label">特权模式</div>
        <el-switch v-model="container.securityContext.privileged" />
      </div>
    </div>
    <!-- Capabilities -->
    <div style="margin-top: var(--gk-space-4);">
      <el-divider content-position="left">Linux Capabilities</el-divider>
      <div class="fields-grid">
        <el-form-item label="添加 (Add)">
          <div style="width: 100%;">
            <div v-for="(_cap, i) in container.securityContext.capabilitiesAdd" :key="i" class="kv-row">
              <el-select v-model="container.securityContext.capabilitiesAdd[i]" filterable allow-create placeholder="如: NET_ADMIN" style="flex: 1;">
                <el-option label="NET_ADMIN" value="NET_ADMIN" />
                <el-option label="NET_RAW" value="NET_RAW" />
                <el-option label="SYS_ADMIN" value="SYS_ADMIN" />
                <el-option label="SYS_PTRACE" value="SYS_PTRACE" />
                <el-option label="SYS_TIME" value="SYS_TIME" />
                <el-option label="SYS_RESOURCE" value="SYS_RESOURCE" />
                <el-option label="DAC_OVERRIDE" value="DAC_OVERRIDE" />
                <el-option label="DAC_READ_SEARCH" value="DAC_READ_SEARCH" />
                <el-option label="SETUID" value="SETUID" />
                <el-option label="SETGID" value="SETGID" />
                <el-option label="CHOWN" value="CHOWN" />
                <el-option label="FOWNER" value="FOWNER" />
                <el-option label="KILL" value="KILL" />
                <el-option label="MKNOD" value="MKNOD" />
              </el-select>
              <el-button type="danger" text circle @click="removeCapability(container.securityContext, 'add', i)">
                <el-icon><Delete /></el-icon>
              </el-button>
            </div>
            <el-button text type="primary" size="small" @click="addCapability(container.securityContext, 'add')">
              <el-icon><Plus /></el-icon> 添加
            </el-button>
          </div>
        </el-form-item>
        <el-form-item label="移除 (Drop)">
          <div style="width: 100%;">
            <div v-for="(_cap, i) in container.securityContext.capabilitiesDrop" :key="i" class="kv-row">
              <el-select v-model="container.securityContext.capabilitiesDrop[i]" filterable allow-create placeholder="如: ALL" style="flex: 1;">
                <el-option label="ALL" value="ALL" />
                <el-option label="NET_ADMIN" value="NET_ADMIN" />
                <el-option label="NET_RAW" value="NET_RAW" />
                <el-option label="SYS_ADMIN" value="SYS_ADMIN" />
                <el-option label="SYS_PTRACE" value="SYS_PTRACE" />
                <el-option label="SYS_TIME" value="SYS_TIME" />
                <el-option label="SYS_RESOURCE" value="SYS_RESOURCE" />
                <el-option label="DAC_OVERRIDE" value="DAC_OVERRIDE" />
                <el-option label="DAC_READ_SEARCH" value="DAC_READ_SEARCH" />
                <el-option label="SETUID" value="SETUID" />
                <el-option label="SETGID" value="SETGID" />
                <el-option label="CHOWN" value="CHOWN" />
                <el-option label="FOWNER" value="FOWNER" />
                <el-option label="KILL" value="KILL" />
                <el-option label="MKNOD" value="MKNOD" />
              </el-select>
              <el-button type="danger" text circle @click="removeCapability(container.securityContext, 'drop', i)">
                <el-icon><Delete /></el-icon>
              </el-button>
            </div>
            <el-button text type="primary" size="small" @click="addCapability(container.securityContext, 'drop')">
              <el-icon><Plus /></el-icon> 添加
            </el-button>
          </div>
        </el-form-item>
      </div>
    </div>
  </div>
</template>

<style scoped>
.mount-container-name {
  font-size: 13px;
  font-weight: 600;
  color: var(--el-text-color-regular);
  margin-bottom: 10px;
  padding: 4px 10px;
  background: var(--el-fill-color-light);
  border-radius: 4px;
  display: inline-block;
}

.security-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 20px;
}

.security-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 16px;
  border: 1px solid var(--el-border-color-extra-light);
  border-radius: var(--gk-radius-md);
  background: var(--el-fill-color-lighter);
}

.security-item-label {
  font-size: 14px;
  color: var(--el-text-color-regular);
}

.fields-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 0 32px;
}

.fields-grid :deep(.el-form-item) {
  margin-bottom: 16px;
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
</style>
