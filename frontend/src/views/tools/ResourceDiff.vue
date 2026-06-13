<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage } from 'element-plus'
import { Refresh, CopyDocument, Download } from '@element-plus/icons-vue'
import request from '@/api/request'
import * as monaco from 'monaco-editor'

const { t } = useI18n()
const loading = ref(false)
const namespaces = ref<string[]>([])
const resources = ref([
  { type: 'deployment', label: 'Deployments' },
  { type: 'statefulset', label: 'StatefulSets' },
  { type: 'daemonset', label: 'DaemonSets' },
  { type: 'service', label: 'Services' },
  { type: 'configmap', label: 'ConfigMaps' },
  { type: 'secret', label: 'Secrets' },
])

const selectedType = ref('deployment')
const selectedNamespace = ref('')
const selectedResource1 = ref('')
const selectedResource2 = ref('')
const resourceList = ref<any[]>([])

const yaml1 = ref('')
const yaml2 = ref('')
const diffResult = ref('')

const editorRef1 = ref<HTMLElement | null>(null)
const editorRef2 = ref<HTMLElement | null>(null)
const diffEditorRef = ref<HTMLElement | null>(null)

let editor1: monaco.editor.IStandaloneCodeEditor | null = null
let editor2: monaco.editor.IStandaloneCodeEditor | null = null
let diffEditor: monaco.editor.IStandaloneDiffEditor | null = null

const viewMode = ref('side-by-side') // 'side-by-side' or 'unified'

async function fetchNamespaces() {
  try {
    const res: any = await request.get('/k8s/namespace/list')
    namespaces.value = res.data?.map((ns: any) => ns.name) || []
    if (namespaces.value.length > 0) {
      selectedNamespace.value = namespaces.value[0]
    }
  } catch (e: any) {
    ElMessage.error('Failed to load namespaces')
  }
}

async function fetchResources() {
  if (!selectedNamespace.value) return

  try {
    const res: any = await request.get(`/k8s/${selectedType.value}/list`, {
      params: { namespace: selectedNamespace.value }
    })
    resourceList.value = res.data || []
  } catch (e: any) {
    ElMessage.error('Failed to load resources')
  }
}

async function fetchYaml(resourceName: string, side: 'left' | 'right') {
  if (!resourceName) return

  try {
    const res: any = await request.get(`/k8s/${selectedType.value}/get-yaml`, {
      params: {
        namespace: selectedNamespace.value,
        name: resourceName
      }
    })

    if (side === 'left') {
      yaml1.value = res.data || ''
    } else {
      yaml2.value = res.data || ''
    }
  } catch (e: any) {
    ElMessage.error(`Failed to load YAML for ${resourceName}`)
  }
}

function initEditors() {
  if (editorRef1.value && !editor1) {
    editor1 = monaco.editor.create(editorRef1.value, {
      value: yaml1.value,
      language: 'yaml',
      readOnly: false,
      minimap: { enabled: false },
      lineNumbers: 'on',
      scrollBeyondLastLine: false,
      wordWrap: 'on',
      theme: 'vs-dark',
    })
  }

  if (editorRef2.value && !editor2) {
    editor2 = monaco.editor.create(editorRef2.value, {
      value: yaml2.value,
      language: 'yaml',
      readOnly: false,
      minimap: { enabled: false },
      lineNumbers: 'on',
      scrollBeyondLastLine: false,
      wordWrap: 'on',
      theme: 'vs-dark',
    })
  }

  if (diffEditorRef.value && !diffEditor) {
    diffEditor = monaco.editor.createDiffEditor(diffEditorRef.value, {
      readOnly: true,
      minimap: { enabled: false },
      lineNumbers: 'on',
      scrollBeyondLastLine: false,
      wordWrap: 'on',
      theme: 'vs-dark',
      renderSideBySide: viewMode.value === 'side-by-side',
    })
  }
}

function updateDiff() {
  if (!diffEditor) return

  const original = monaco.editor.createModel(yaml1.value, 'yaml')
  const modified = monaco.editor.createModel(yaml2.value, 'yaml')

  diffEditor.setModel({ original, modified })
}

function copyToClipboard(text: string) {
  navigator.clipboard.writeText(text).then(() => {
    ElMessage.success('Copied to clipboard')
  }).catch(() => {
    ElMessage.error('Failed to copy')
  })
}

function downloadYaml(content: string, filename: string) {
  const blob = new Blob([content], { type: 'text/yaml' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = filename
  a.click()
  URL.revokeObjectURL(url)
}

async function loadComparison() {
  if (!selectedResource1.value || !selectedResource2.value) {
    ElMessage.warning('Please select two resources to compare')
    return
  }

  loading.value = true
  try {
    await Promise.all([
      fetchYaml(selectedResource1.value, 'left'),
      fetchYaml(selectedResource2.value, 'right')
    ])

    if (editor1) editor1.setValue(yaml1.value)
    if (editor2) editor2.setValue(yaml2.value)

    updateDiff()
  } finally {
    loading.value = false
  }
}

function switchViewMode() {
  if (diffEditor) {
    diffEditor.updateOptions({
      renderSideBySide: viewMode.value === 'side-by-side'
    })
  }
}

watch(selectedType, () => {
  selectedResource1.value = ''
  selectedResource2.value = ''
  fetchResources()
})

watch(selectedNamespace, () => {
  selectedResource1.value = ''
  selectedResource2.value = ''
  fetchResources()
})

onMounted(() => {
  fetchNamespaces()
  initEditors()
})
</script>

<template>
  <div class="page-container">
    <el-card shadow="never" class="filter-card">
      <div class="filter-bar">
        <h3 style="margin: 0;">Resource Diff / Compare</h3>
        <div class="filter-right">
          <el-select v-model="selectedType" style="width: 150px;">
            <el-option v-for="r in resources" :key="r.type" :label="r.label" :value="r.type" />
          </el-select>
          <el-select v-model="selectedNamespace" placeholder="Namespace" style="width: 150px;">
            <el-option v-for="ns in namespaces" :key="ns" :label="ns" :value="ns" />
          </el-select>
          <el-button type="primary" @click="loadComparison"><el-icon><Refresh /></el-icon> Compare</el-button>
        </div>
      </div>
    </el-card>

    <el-row :gutter="16">
      <el-col :span="12">
        <el-card shadow="never">
          <template #header>
            <div class="card-header">
              <span>Resource 1</span>
              <el-select v-model="selectedResource1" placeholder="Select resource" style="width: 250px;">
                <el-option v-for="r in resourceList" :key="r.name" :label="r.name" :value="r.name" />
              </el-select>
            </div>
          </template>
          <div ref="editorRef1" class="yaml-editor"></div>
          <div class="editor-actions">
            <el-button size="small" @click="copyToClipboard(yaml1)"><el-icon><CopyDocument /></el-icon> Copy</el-button>
            <el-button size="small" @click="downloadYaml(yaml1, selectedResource1 + '.yaml')"><el-icon><Download /></el-icon> Download</el-button>
          </div>
        </el-card>
      </el-col>
      <el-col :span="12">
        <el-card shadow="never">
          <template #header>
            <div class="card-header">
              <span>Resource 2</span>
              <el-select v-model="selectedResource2" placeholder="Select resource" style="width: 250px;">
                <el-option v-for="r in resourceList" :key="r.name" :label="r.name" :value="r.name" />
              </el-select>
            </div>
          </template>
          <div ref="editorRef2" class="yaml-editor"></div>
          <div class="editor-actions">
            <el-button size="small" @click="copyToClipboard(yaml2)"><el-icon><CopyDocument /></el-icon> Copy</el-button>
            <el-button size="small" @click="downloadYaml(yaml2, selectedResource2 + '.yaml')"><el-icon><Download /></el-icon> Download</el-button>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <el-card shadow="never" style="margin-top: 16px;">
      <template #header>
        <div class="card-header">
          <span>Diff Result</span>
          <el-radio-group v-model="viewMode" size="small" @change="switchViewMode">
            <el-radio-button value="side-by-side">Side by Side</el-radio-button>
            <el-radio-button value="unified">Unified</el-radio-button>
          </el-radio-group>
        </div>
      </template>
      <div ref="diffEditorRef" class="diff-editor"></div>
    </el-card>
  </div>
</template>

<style scoped>
.page-container { padding: 20px; }
.filter-card { margin-bottom: 16px; }
.filter-bar { display: flex; justify-content: space-between; align-items: center; }
.filter-right { display: flex; align-items: center; gap: 8px; }
.card-header { display: flex; justify-content: space-between; align-items: center; }
.yaml-editor { height: 400px; border: 1px solid #dcdfe6; border-radius: 4px; }
.diff-editor { height: 500px; border: 1px solid #dcdfe6; border-radius: 4px; }
.editor-actions { margin-top: 8px; display: flex; gap: 8px; }
</style>
