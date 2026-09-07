<script setup lang="ts">
import { Plus, Delete } from '@element-plus/icons-vue'
import { computed } from 'vue'
import type { Volume, VolumeClaimTemplate, VolumeMount } from '../form-types'

const props = defineProps<{
  volumes: Volume[]
  volumeClaimTemplates?: VolumeClaimTemplate[]
  /** Direct containers with volumeMounts (used by WorkloadForm) */
  containers?: { name: string; volumeMounts: { name: string; mountPath: string; subPath: string; readOnly: boolean }[] }[]
  /** Pre-computed volume mount data (used by CronJobForm, JobForm) */
  volumeMounts?: { containerName: string; mounts: VolumeMount[] }[]
  kind?: string
}>()

// Normalize to unified format
const mountData = computed(() => {
  if (props.volumeMounts) return props.volumeMounts
  if (props.containers) return props.containers.map((c, i) => ({
    containerName: c.name || `容器 ${i + 1}`,
    mounts: c.volumeMounts,
  }))
  return []
})

function addVolume() { props.volumes.push({ name: '', type: 'emptyDir', hostPath: '', hostPathType: 'DirectoryOrCreate', configMapName: '', secretName: '', pvcName: '' }) }
function removeVolume(i: number) { props.volumes.splice(i, 1) }
function addVolumeMount(ci: number) { mountData.value[ci].mounts.push({ name: '', mountPath: '', subPath: '', readOnly: false }) }
function removeVolumeMount(ci: number, mi: number) { mountData.value[ci].mounts.splice(mi, 1) }
function addVolumeClaimTemplate() { props.volumeClaimTemplates?.push({ name: '', storageSize: '1Gi', storageClassName: '', accessModes: ['ReadWriteOnce'] }) }
function removeVolumeClaimTemplate(i: number) { props.volumeClaimTemplates?.splice(i, 1) }
</script>

<template>
  <!-- Volume Claim Templates (StatefulSet only) -->
  <template v-if="kind === 'StatefulSet' && volumeClaimTemplates">
    <el-form-item label="持久卷声明模板 (VolumeClaimTemplates)">
      <div style="width: 100%;">
        <div v-for="(vct, vi) in volumeClaimTemplates" :key="vi" class="volume-card">
          <div class="volume-row">
            <el-input v-model="vct.name" placeholder="模板名称 (如: data)" />
            <el-button type="danger" text circle @click="removeVolumeClaimTemplate(vi)">
              <el-icon><Delete /></el-icon>
            </el-button>
          </div>
          <div class="fields-grid" style="margin-top: 8px;">
            <el-form-item label="存储大小">
              <el-input v-model="vct.storageSize" placeholder="1Gi" />
            </el-form-item>
            <el-form-item label="存储类名">
              <el-input v-model="vct.storageClassName" placeholder="留空使用默认 StorageClass" />
            </el-form-item>
            <el-form-item label="访问模式" class="full-width">
              <el-checkbox-group v-model="vct.accessModes">
                <el-checkbox label="ReadWriteOnce" />
                <el-checkbox label="ReadOnlyMany" />
                <el-checkbox label="ReadWriteMany" />
              </el-checkbox-group>
            </el-form-item>
          </div>
        </div>
        <el-button text type="primary" @click="addVolumeClaimTemplate" size="small">
          <el-icon><Plus /></el-icon> 添加持久卷声明模板
        </el-button>
      </div>
    </el-form-item>
    <el-divider />
  </template>

  <el-form-item label="数据卷">
    <div style="width: 100%;">
      <div v-for="(vol, vi) in volumes" :key="vi" class="volume-card">
        <div class="volume-row">
          <el-input v-model="vol.name" placeholder="卷名称" />
          <el-select v-model="vol.type" style="width: 160px;">
            <el-option label="emptyDir" value="emptyDir" />
            <el-option label="hostPath" value="hostPath" />
            <el-option label="ConfigMap" value="configMap" />
            <el-option label="Secret" value="secret" />
            <el-option label="PVC" value="pvc" />
          </el-select>
          <el-button type="danger" text circle @click="removeVolume(vi)">
            <el-icon><Delete /></el-icon>
          </el-button>
        </div>
        <template v-if="vol.type === 'hostPath'">
          <el-input v-model="vol.hostPath" placeholder="主机路径 (e.g. /data)" style="margin-top: 8px;" />
          <el-select v-model="vol.hostPathType" style="margin-top: 8px; width: 100%;">
            <el-option label="DirectoryOrCreate" value="DirectoryOrCreate" />
            <el-option label="Directory" value="Directory" />
            <el-option label="FileOrCreate" value="FileOrCreate" />
            <el-option label="File" value="File" />
            <el-option label="Socket" value="Socket" />
            <el-option label="CharDevice" value="CharDevice" />
            <el-option label="BlockDevice" value="BlockDevice" />
          </el-select>
        </template>
        <el-input v-if="vol.type === 'configMap'" v-model="vol.configMapName" placeholder="ConfigMap 名称" style="margin-top: 8px;" />
        <el-input v-if="vol.type === 'secret'" v-model="vol.secretName" placeholder="Secret 名称" style="margin-top: 8px;" />
        <el-input v-if="vol.type === 'pvc'" v-model="vol.pvcName" placeholder="PVC 名称" style="margin-top: 8px;" />
      </div>
      <el-button text type="primary" @click="addVolume" size="small">
        <el-icon><Plus /></el-icon> 添加数据卷
      </el-button>
    </div>
  </el-form-item>

  <el-divider v-if="volumes.length > 0" />

  <el-form-item v-if="volumes.length > 0" label="卷挂载">
    <div style="width: 100%;">
      <div v-for="(containerData, ci) in mountData" :key="ci" style="margin-bottom: var(--gk-space-4);">
        <div class="mount-container-name">{{ containerData.containerName || `容器 ${ci + 1}` }}</div>
        <div v-for="(mount, mi) in containerData.mounts" :key="mi" class="kv-row">
          <el-select v-model="mount.name" placeholder="选择卷" style="width: 160px;">
            <el-option v-for="v in volumes.filter(v => v.name)" :key="v.name" :label="v.name" :value="v.name" />
          </el-select>
          <el-input v-model="mount.mountPath" placeholder="挂载路径" />
          <el-input v-model="mount.subPath" placeholder="子路径" style="width: 120px;" />
          <el-checkbox v-model="mount.readOnly">只读</el-checkbox>
          <el-button type="danger" text circle @click="removeVolumeMount(ci, mi)">
            <el-icon><Delete /></el-icon>
          </el-button>
        </div>
        <el-button text type="primary" size="small" @click="addVolumeMount(ci)">
          <el-icon><Plus /></el-icon> 添加挂载
        </el-button>
      </div>
    </div>
  </el-form-item>
</template>

<style scoped>
.volume-card {
  border: 1px solid var(--el-border-color-extra-light);
  border-radius: var(--gk-radius-md);
  padding: 14px;
  margin-bottom: 8px;
  background: var(--el-fill-color-lighter);
}

.volume-row {
  display: flex;
  gap: 8px;
  align-items: center;
}

.volume-row :deep(.el-input) {
  flex: 1;
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
