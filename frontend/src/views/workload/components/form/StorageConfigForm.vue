<script setup lang="ts">
import { Plus, Delete } from '@element-plus/icons-vue'
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import type { Volume, VolumeClaimTemplate, VolumeMount } from '../form-types'

const props = defineProps<{
  /** Direct containers with volumeMounts (used by WorkloadForm) */
  containers?: {
    name: string
    volumeMounts: { name: string; mountPath: string; subPath: string; readOnly: boolean }[]
  }[]
  /** Pre-computed volume mount data (used by CronJobForm, JobForm) */
  volumeMounts?: { containerName: string; mounts: VolumeMount[] }[]
  kind?: string
}>()

const volumes = defineModel<Volume[]>('volumes', { required: true })
const volumeClaimTemplates = defineModel<VolumeClaimTemplate[]>('volumeClaimTemplates')

const { t } = useI18n()

// Normalize to unified format
const mountData = computed(() => {
  if (props.volumeMounts) return props.volumeMounts
  if (props.containers)
    return props.containers.map((c, i) => ({
      containerName: c.name || t('storageConfig.containerLabel', { n: i + 1 }),
      mounts: c.volumeMounts,
    }))
  return []
})

function addVolume() {
  volumes.value.push({
    name: '',
    type: 'emptyDir',
    hostPath: '',
    hostPathType: 'DirectoryOrCreate',
    configMapName: '',
    secretName: '',
    pvcName: '',
  })
}
function removeVolume(i: number) {
  volumes.value.splice(i, 1)
}
function addVolumeMount(ci: number) {
  mountData.value[ci].mounts.push({ name: '', mountPath: '', subPath: '', readOnly: false })
}
function removeVolumeMount(ci: number, mi: number) {
  mountData.value[ci].mounts.splice(mi, 1)
}
function addVolumeClaimTemplate() {
  volumeClaimTemplates.value?.push({
    name: '',
    storageSize: '1Gi',
    storageClassName: '',
    accessModes: ['ReadWriteOnce'],
  })
}
function removeVolumeClaimTemplate(i: number) {
  volumeClaimTemplates.value?.splice(i, 1)
}
</script>

<template>
  <!-- Volume Claim Templates (StatefulSet only) -->
  <template v-if="kind === 'StatefulSet' && volumeClaimTemplates">
    <el-form-item :label="t('storageConfig.volumeClaimTemplates')">
      <div style="width: 100%">
        <div v-for="(vct, vi) in volumeClaimTemplates" :key="vi" class="volume-card">
          <div class="volume-row">
            <el-input v-model="vct.name" :placeholder="t('storageConfig.templateName')" />
            <el-button type="danger" text circle @click="removeVolumeClaimTemplate(vi)">
              <el-icon><Delete /></el-icon>
            </el-button>
          </div>
          <div class="fields-grid" style="margin-top: 8px">
            <el-form-item :label="t('storageConfig.storageSize')">
              <el-input v-model="vct.storageSize" placeholder="1Gi" />
            </el-form-item>
            <el-form-item :label="t('storageConfig.storageClassName')">
              <el-input
                v-model="vct.storageClassName"
                :placeholder="t('storageConfig.storageClassPlaceholder')"
              />
            </el-form-item>
            <el-form-item :label="t('storageConfig.accessModes')" class="full-width">
              <el-checkbox-group v-model="vct.accessModes">
                <el-checkbox label="ReadWriteOnce" />
                <el-checkbox label="ReadOnlyMany" />
                <el-checkbox label="ReadWriteMany" />
              </el-checkbox-group>
            </el-form-item>
          </div>
        </div>
        <el-button text type="primary" size="small" @click="addVolumeClaimTemplate">
          <el-icon><Plus /></el-icon> {{ t('storageConfig.addVolumeClaimTemplate') }}
        </el-button>
      </div>
    </el-form-item>
    <el-divider />
  </template>

  <el-form-item :label="t('storageConfig.volumes')">
    <div style="width: 100%">
      <div v-for="(vol, vi) in volumes" :key="vi" class="volume-card">
        <div class="volume-row">
          <el-input v-model="vol.name" :placeholder="t('storageConfig.volumeName')" />
          <el-select v-model="vol.type" style="width: 160px">
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
          <el-input
            v-model="vol.hostPath"
            :placeholder="t('storageConfig.hostPathPlaceholder')"
            style="margin-top: 8px"
          />
          <el-select v-model="vol.hostPathType" style="margin-top: 8px; width: 100%">
            <el-option label="DirectoryOrCreate" value="DirectoryOrCreate" />
            <el-option label="Directory" value="Directory" />
            <el-option label="FileOrCreate" value="FileOrCreate" />
            <el-option label="File" value="File" />
            <el-option label="Socket" value="Socket" />
            <el-option label="CharDevice" value="CharDevice" />
            <el-option label="BlockDevice" value="BlockDevice" />
          </el-select>
        </template>
        <el-input
          v-if="vol.type === 'configMap'"
          v-model="vol.configMapName"
          :placeholder="t('storageConfig.configMapNamePlaceholder')"
          style="margin-top: 8px"
        />
        <el-input
          v-if="vol.type === 'secret'"
          v-model="vol.secretName"
          :placeholder="t('storageConfig.secretNamePlaceholder')"
          style="margin-top: 8px"
        />
        <el-input
          v-if="vol.type === 'pvc'"
          v-model="vol.pvcName"
          :placeholder="t('storageConfig.pvcNamePlaceholder')"
          style="margin-top: 8px"
        />
      </div>
      <el-button text type="primary" size="small" @click="addVolume">
        <el-icon><Plus /></el-icon> {{ t('storageConfig.addVolume') }}
      </el-button>
    </div>
  </el-form-item>

  <el-divider v-if="volumes.length > 0" />

  <el-form-item v-if="volumes.length > 0" :label="t('storageConfig.volumeMounts')">
    <div style="width: 100%">
      <div
        v-for="(containerData, ci) in mountData"
        :key="ci"
        style="margin-bottom: var(--gk-space-4)"
      >
        <div class="mount-container-name">
          {{ containerData.containerName || t('storageConfig.containerLabel', { n: ci + 1 }) }}
        </div>
        <div v-for="(mount, mi) in containerData.mounts" :key="mi" class="kv-row">
          <el-select
            v-model="mount.name"
            :placeholder="t('storageConfig.selectVolume')"
            style="width: 160px"
          >
            <el-option
              v-for="v in volumes.filter((v) => v.name)"
              :key="v.name"
              :label="v.name"
              :value="v.name"
            />
          </el-select>
          <el-input v-model="mount.mountPath" :placeholder="t('storageConfig.mountPath')" />
          <el-input
            v-model="mount.subPath"
            :placeholder="t('storageConfig.subPath')"
            style="width: 120px"
          />
          <el-checkbox v-model="mount.readOnly">{{ t('storageConfig.readOnly') }}</el-checkbox>
          <el-button type="danger" text circle @click="removeVolumeMount(ci, mi)">
            <el-icon><Delete /></el-icon>
          </el-button>
        </div>
        <el-button text type="primary" size="small" @click="addVolumeMount(ci)">
          <el-icon><Plus /></el-icon> {{ t('storageConfig.addMount') }}
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
