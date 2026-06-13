<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { Refresh, Download, Upload, Check, Close } from '@element-plus/icons-vue'
import request from '@/api/request'
import * as monaco from 'monaco-editor'

const loading = ref(false)
const selectedResource = ref('deployment')
const selectedNamespace = ref('default')
const resourceName = ref('')
const namespaces = ref<string[]>([])
const yamlContent = ref('')
const validationResult = ref<{ valid: boolean; message: string } | null>(null)
const editorRef = ref<HTMLElement | null>(null)
let editor: monaco.editor.IStandaloneCodeEditor | null = null

const resourceTypes = [
  { value: 'deployment', label: 'Deployment' },
  { value: 'statefulset', label: 'StatefulSet' },
  { value: 'daemonset', label: 'DaemonSet' },
  { value: 'service', label: 'Service' },
  { value: 'configmap', label: 'ConfigMap' },
  { value: 'secret', label: 'Secret' },
  { value: 'ingress', label: 'Ingress' },
  { value: 'persistentvolumeclaim', label: 'PVC' },
  { value: 'resourcequota', label: 'ResourceQuota' },
  { value: 'limitrange', label: 'LimitRange' },
  { value: 'networkpolicy', label: 'NetworkPolicy' },
  { value: 'horizontalpodautoscaler', label: 'HPA' },
]

const sampleYAML: Record<string, string> = {
  deployment: `apiVersion: apps/v1
kind: Deployment
metadata:
  name: my-app
  namespace: default
  labels:
    app: my-app
spec:
  replicas: 3
  selector:
    matchLabels:
      app: my-app
  template:
    metadata:
      labels:
        app: my-app
    spec:
      containers:
      - name: my-app
        image: nginx:latest
        ports:
        - containerPort: 80
        resources:
          requests:
            cpu: "100m"
            memory: "128Mi"
          limits:
            cpu: "500m"
            memory: "512Mi"`,
  service: `apiVersion: v1
kind: Service
metadata:
  name: my-service
  namespace: default
spec:
  selector:
    app: my-app
  ports:
  - port: 80
    targetPort: 80
  type: ClusterIP`,
  configmap: `apiVersion: v1
kind: ConfigMap
metadata:
  name: my-config
  namespace: default
data:
  config.yaml: |
    database:
      host: localhost
      port: 5432`,
  secret: `apiVersion: v1
kind: Secret
metadata:
  name: my-secret
  namespace: default
type: Opaque
data:
  username: YWRtaW4=
  password: cGFzc3dvcmQ=`,
}

async function fetchNamespaces() {
  try {
    const res: any = await request.get('/k8s/namespace/list')
    namespaces.value = res.data?.map((ns: any) => ns.name) || []
  } catch {
    namespaces.value = ['default']
  }
}

function initEditor() {
  if (editorRef.value && !editor) {
    editor = monaco.editor.create(editorRef.value, {
      value: yamlContent.value,
      language: 'yaml',
      minimap: { enabled: false },
      lineNumbers: 'on',
      scrollBeyondLastLine: false,
      wordWrap: 'on',
      theme: 'vs-dark',
      automaticLayout: true,
    })

    editor.onDidChangeModelContent(() => {
      yamlContent.value = editor?.getValue() || ''
    })
  }
}

function loadSample() {
  yamlContent.value = sampleYAML[selectedResource.value] || ''
  if (editor) {
    editor.setValue(yamlContent.value)
  }
}

async function loadFromCluster() {
  if (!resourceName.value) {
    ElMessage.warning('Please enter a resource name')
    return
  }

  loading.value = true
  try {
    const res: any = await request.get(`/k8s/${selectedResource.value}/get-yaml`, {
      params: {
        namespace: selectedNamespace.value,
        name: resourceName.value,
      }
    })
    yamlContent.value = res.data || ''
    if (editor) {
      editor.setValue(yamlContent.value)
    }
    ElMessage.success('YAML loaded')
  } catch (e: any) {
    ElMessage.error(e?.message || 'Failed to load YAML')
  } finally {
    loading.value = false
  }
}

function validateYAML() {
  try {
    JSON.parse(yamlContent.value)
    validationResult.value = { valid: true, message: 'YAML is valid' }
  } catch (e: any) {
    // Simple YAML validation
    const lines = yamlContent.value.split('\n')
    let valid = true
    let message = 'YAML appears valid'

    for (let i = 0; i < lines.length; i++) {
      const line = lines[i]
      if (line.includes(':') && !line.trim().startsWith('#')) {
        const parts = line.split(':')
        if (parts.length < 2) {
          valid = false
          message = `Line ${i + 1}: Invalid YAML syntax`
          break
        }
      }
    }

    validationResult.value = { valid, message }
  }
}

async function applyYAML() {
  if (!yamlContent.value) {
    ElMessage.warning('YAML content is empty')
    return
  }

  loading.value = true
  try {
    await request.post(`/k8s/${selectedResource.value}/create`, {
      yaml: yamlContent.value,
      namespace: selectedNamespace.value,
    })
    ElMessage.success('Resource applied successfully')
  } catch (e: any) {
    ElMessage.error(e?.message || 'Failed to apply resource')
  } finally {
    loading.value = false
  }
}

function downloadYAML() {
  const blob = new Blob([yamlContent.value], { type: 'text/yaml' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = `${selectedResource.value}.yaml`
  a.click()
  URL.revokeObjectURL(url)
}

function uploadYAML() {
  const input = document.createElement('input')
  input.type = 'file'
  input.accept = '.yaml,.yml'
  input.onchange = (e: any) => {
    const file = e.target.files[0]
    if (file) {
      const reader = new FileReader()
      reader.onload = (e: any) => {
        yamlContent.value = e.target.result
        if (editor) {
          editor.setValue(yamlContent.value)
        }
      }
      reader.readAsText(file)
    }
  }
  input.click()
}

onMounted(() => {
  fetchNamespaces()
  loadSample()
  initEditor()
})
</script>

<template>
  <div class="page-container">
    <el-card shadow="never" class="filter-card">
      <div class="filter-bar">
        <h3 style="margin: 0;">YAML 编辑器</h3>
        <div class="filter-right">
          <el-select v-model="selectedResource" style="width: 180px;" @change="loadSample">
            <el-option v-for="r in resourceTypes" :key="r.value" :label="r.label" :value="r.value" />
          </el-select>
          <el-select v-model="selectedNamespace" style="width: 150px;">
            <el-option v-for="ns in namespaces" :key="ns" :label="ns" :value="ns" />
          </el-select>
          <el-input v-model="resourceName" placeholder="资源名称" style="width: 150px;" />
          <el-button @click="loadFromCluster"><el-icon><Download /></el-icon> 加载</el-button>
          <el-button @click="loadSample">示例</el-button>
          <el-button @click="uploadYAML"><el-icon><Upload /></el-icon> 上传</el-button>
        </div>
      </div>
    </el-card>

    <el-row :gutter="16">
      <el-col :span="18">
        <el-card shadow="never">
          <div ref="editorRef" style="height: 600px;"></div>
        </el-card>
      </el-col>

      <el-col :span="6">
        <el-card shadow="never">
          <template #header>
            <h4 style="margin: 0;">操作</h4>
          </template>
          <div class="actions">
            <el-button type="primary" style="width: 100%;" @click="applyYAML" :loading="loading">
              <el-icon><Check /></el-icon> 应用
            </el-button>
            <el-button style="width: 100%;" @click="validateYAML">
              <el-icon><Check /></el-icon> 验证
            </el-button>
            <el-button style="width: 100%;" @click="downloadYAML">
              <el-icon><Download /></el-icon> 下载
            </el-button>
          </div>

          <el-divider />

          <div v-if="validationResult" class="validation-result">
            <el-alert
              :title="validationResult.message"
              :type="validationResult.valid ? 'success' : 'error'"
              show-icon
            />
          </div>

          <el-divider />

          <h4>帮助</h4>
          <div class="help-text">
            <p>• 选择资源类型加载示例 YAML</p>
            <p>• 从集群加载现有资源</p>
            <p>• 编辑后点击"应用"创建资源</p>
            <p>• 支持上传/下载 YAML 文件</p>
          </div>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<style scoped>
.page-container { padding: 20px; }
.filter-card { margin-bottom: 16px; }
.filter-bar { display: flex; justify-content: space-between; align-items: center; }
.filter-right { display: flex; align-items: center; gap: 8px; }
.actions { display: flex; flex-direction: column; gap: 8px; }
.validation-result { margin-top: 16px; }
.help-text { font-size: 12px; color: #909399; line-height: 1.8; }
</style>
