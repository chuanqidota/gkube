<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'

import { ElMessage } from 'element-plus'
import { createCrd } from '@/api/resource'
import YamlEditor from '@/components/YamlEditor.vue'

const router = useRouter()
const loading = ref(false)
const yamlContent = ref(`apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: myresources.example.com
spec:
  group: example.com
  versions:
    - name: v1
      served: true
      storage: true
      schema:
        openAPIV3Schema:
          type: object
          properties:
            spec:
              type: object
              properties:
                name:
                  type: string
  scope: Namespaced
  names:
    plural: myresources
    singular: myresource
    kind: MyResource
    shortNames:
      - mr
`)

async function handleCreate() {
  if (!yamlContent.value.trim()) {
    ElMessage.warning('Please enter CRD YAML')
    return
  }
  loading.value = true
  try {
    await createCrd({ yaml: yamlContent.value })
    ElMessage.success('CRD created successfully')
    router.push('/crd')
  } catch (e: any) {
    ElMessage.error(e?.message || 'Failed to create CRD')
  } finally { loading.value = false }
}
</script>

<template>
  <div class="page-container">
    <div class="page-header">
      <h2 style="margin: 0;">创建 CustomResourceDefinition</h2>
      <el-button @click="router.push('/crd')">Back to List</el-button>
    </div>
    <el-card shadow="never">
      <el-alert title="Enter the CRD YAML below. The CRD defines a new custom resource type in your cluster." type="info" :closable="false" show-icon style="margin-bottom: 16px;" />
      <YamlEditor v-model="yamlContent" height="500px" />
      <div style="margin-top: 16px;">
        <el-button type="primary" :loading="loading" @click="handleCreate">创建 CRD</el-button>
      </div>
    </el-card>
  </div>
</template>

<style scoped>
.page-container { padding: 20px; }
.page-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px; }
</style>
