<script setup lang="ts">
import { Plus, Delete } from '@element-plus/icons-vue'

interface Label {
  key: string
  value: string
}
interface Annotation {
  key: string
  value: string
}

defineProps<{
  labels: Label[]
  annotations: Annotation[]
}>()

function addLabel(labels: Label[]) {
  labels.push({ key: '', value: '' })
}
function removeLabel(labels: Label[], i: number) {
  labels.splice(i, 1)
}
function addAnnotation(annotations: Annotation[]) {
  annotations.push({ key: '', value: '' })
}
function removeAnnotation(annotations: Annotation[], i: number) {
  annotations.splice(i, 1)
}
</script>

<template>
  <el-form-item label="标签">
    <div style="width: 100%">
      <div v-for="(label, i) in labels" :key="i" class="kv-row">
        <el-input v-model="label.key" placeholder="Key" />
        <el-input v-model="label.value" placeholder="Value" />
        <el-button
          type="danger"
          text
          circle
          :disabled="labels.length <= 1"
          @click="removeLabel(labels, i)"
        >
          <el-icon><Delete /></el-icon>
        </el-button>
      </div>
      <el-button text type="primary" size="small" @click="addLabel(labels)">
        <el-icon><Plus /></el-icon> 添加标签
      </el-button>
    </div>
  </el-form-item>
  <el-form-item label="注解">
    <div style="width: 100%">
      <div v-for="(ann, i) in annotations" :key="i" class="kv-row">
        <el-input v-model="ann.key" placeholder="Key" />
        <el-input v-model="ann.value" placeholder="Value" />
        <el-button type="danger" text circle @click="removeAnnotation(annotations, i)">
          <el-icon><Delete /></el-icon>
        </el-button>
      </div>
      <el-button text type="primary" size="small" @click="addAnnotation(annotations)">
        <el-icon><Plus /></el-icon> 添加注解
      </el-button>
    </div>
  </el-form-item>
</template>

<style scoped>
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
