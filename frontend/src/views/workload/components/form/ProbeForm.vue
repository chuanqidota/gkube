<script setup lang="ts">
interface Probe {
  type: string
  httpGetPath: string
  httpGetPort: number | null
  tcpSocketPort: number | null
  execCommand: string
  initialDelaySeconds: number
  periodSeconds: number
  timeoutSeconds: number
  failureThreshold: number
}

const props = defineProps<{
  modelValue: Probe | null
  label: string
  description: string
}>()

const emit = defineEmits<{
  'update:modelValue': [value: Probe | null]
}>()

function enableProbe() {
  emit('update:modelValue', {
    type: 'httpGet',
    httpGetPath: '/',
    httpGetPort: 80,
    tcpSocketPort: null,
    execCommand: '',
    initialDelaySeconds: 15,
    periodSeconds: 10,
    timeoutSeconds: 5,
    failureThreshold: 3,
  })
}

function disableProbe() {
  emit('update:modelValue', null)
}

function updateField<K extends keyof Probe>(key: K, value: Probe[K]) {
  if (!props.modelValue) return
  emit('update:modelValue', { ...props.modelValue, [key]: value })
}
</script>

<template>
  <div class="probe-card">
    <div class="probe-header">
      <div>
        <span class="probe-label">{{ label }}</span>
        <span class="probe-desc">{{ description }}</span>
      </div>
      <el-switch
        :model-value="!!modelValue"
        @update:model-value="(v: boolean) => (v ? enableProbe() : disableProbe())"
      />
    </div>
    <template v-if="modelValue">
      <div class="fields-grid" style="margin-top: var(--gk-space-4)">
        <el-form-item label="检测类型">
          <el-select
            :model-value="modelValue.type"
            style="width: 100%"
            @update:model-value="(v: string) => updateField('type', v)"
          >
            <el-option label="HTTP GET" value="httpGet" />
            <el-option label="TCP Socket" value="tcpSocket" />
            <el-option label="Exec" value="exec" />
          </el-select>
        </el-form-item>
        <el-form-item v-if="modelValue.type === 'httpGet'" label="路径">
          <el-input
            :model-value="modelValue.httpGetPath"
            placeholder="/"
            @update:model-value="(v: string) => updateField('httpGetPath', v)"
          />
        </el-form-item>
        <el-form-item v-if="modelValue.type === 'httpGet'" label="端口">
          <el-input-number
            :model-value="modelValue.httpGetPort"
            :min="1"
            :max="65535"
            style="width: 100%"
            @update:model-value="(v: number) => updateField('httpGetPort', v)"
          />
        </el-form-item>
        <el-form-item v-if="modelValue.type === 'tcpSocket'" label="端口">
          <el-input-number
            :model-value="modelValue.tcpSocketPort"
            :min="1"
            :max="65535"
            style="width: 100%"
            @update:model-value="(v: number) => updateField('tcpSocketPort', v)"
          />
        </el-form-item>
        <el-form-item v-if="modelValue.type === 'exec'" label="命令">
          <el-input
            type="textarea"
            :autosize="{ minRows: 1, maxRows: 4 }"
            :model-value="modelValue.execCommand"
            placeholder="每行一个参数，如: cat /tmp/healthy"
            @update:model-value="(v: string) => updateField('execCommand', v)"
          />
        </el-form-item>
        <el-form-item label="初始延迟(秒)">
          <el-input-number
            :model-value="modelValue.initialDelaySeconds"
            :min="0"
            style="width: 100%"
            @update:model-value="(v: number) => updateField('initialDelaySeconds', v)"
          />
        </el-form-item>
        <el-form-item label="检测周期(秒)">
          <el-input-number
            :model-value="modelValue.periodSeconds"
            :min="1"
            style="width: 100%"
            @update:model-value="(v: number) => updateField('periodSeconds', v)"
          />
        </el-form-item>
        <el-form-item label="超时时间(秒)">
          <el-input-number
            :model-value="modelValue.timeoutSeconds"
            :min="1"
            style="width: 100%"
            @update:model-value="(v: number) => updateField('timeoutSeconds', v)"
          />
        </el-form-item>
        <el-form-item label="失败阈值">
          <el-input-number
            :model-value="modelValue.failureThreshold"
            :min="1"
            style="width: 100%"
            @update:model-value="(v: number) => updateField('failureThreshold', v)"
          />
        </el-form-item>
      </div>
    </template>
  </div>
</template>

<style scoped>
.probe-card {
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 6px;
  padding: 12px 16px;
  margin-bottom: 12px;
}

.probe-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.probe-label {
  font-size: 14px;
  font-weight: 500;
  color: var(--el-text-color-primary);
}

.probe-desc {
  font-size: 12px;
  color: var(--el-text-color-secondary);
  margin-left: 8px;
}

.fields-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
  gap: 0 16px;
}
</style>
