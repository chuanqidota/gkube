<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { ElMessage } from 'element-plus'
import { FullScreen, CopyDocument } from '@element-plus/icons-vue'
import yaml from 'js-yaml'
import WorkloadForm from './components/WorkloadForm.vue'
import YamlEditor from '@/components/YamlEditor.vue'
import CloneDialog from '@/components/CloneDialog.vue'
import { useCloneCreate, dumpCloneYaml } from '@/composables/useCloneCreate'
import {
  getDeploymentYaml, getStatefulSetYaml, getDaemonSetYaml,
  getDeploymentList, getStatefulSetList, getDaemonSetList,
} from '@/api/resource'

const props = defineProps<{
  kind: 'Deployment' | 'StatefulSet' | 'DaemonSet'
  createApi: (data: { namespace: string; yaml: string }) => Promise<any>
  listRoute: string
}>()

const router = useRouter()
const { t } = useI18n()
const mode = ref<'form' | 'yaml'>('form')
const yamlEditorRef = ref()
const submitting = ref(false)

// Clone target parsed data, fed into WorkloadForm via :initial-data
const parsedData = ref<any>(null)

// Build kind -> { listFn, yamlFn } map once, reused everywhere
const RESOURCE_API_MAP: Record<string, { list: (params: any) => Promise<any>; yaml: (params: any) => Promise<any> }> = {
  Deployment: { list: getDeploymentList, yaml: getDeploymentYaml },
  StatefulSet: { list: getStatefulSetList, yaml: getStatefulSetYaml },
  DaemonSet: { list: getDaemonSetList, yaml: getDaemonSetYaml },
}

const {
  cloneMode, cloneNamespace, cloneName, cloneNsOptions, cloneNameOptions,
  cloneNsLoading, cloneNameLoading, cloneLoading, cloneTarget,
  startClone, cancelClone, handleLoadClone,
} = useCloneCreate({
  api: RESOURCE_API_MAP[props.kind],
  onCloneToForm: (parsed) => { parsedData.value = parsed; yamlContent.value = defaultYaml; mode.value = 'form' },
  onCloneToYaml: (parsed) => { yamlContent.value = dumpCloneYaml(parsed); parsedData.value = null; mode.value = 'yaml' },
})

const defaultYaml = {
  Deployment: `apiVersion: apps/v1
kind: Deployment
metadata:
  name: my-deployment
  namespace: default
  labels:
    app: my-deployment
spec:
  replicas: 1
  selector:
    matchLabels:
      app: my-deployment
  template:
    metadata:
      labels:
        app: my-deployment
    spec:
      containers:
        - name: my-deployment
          image: nginx:latest
          ports:
            - containerPort: 80
`,
  StatefulSet: `apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: my-statefulset
  namespace: default
  labels:
    app: my-statefulset
spec:
  replicas: 1
  serviceName: my-statefulset
  selector:
    matchLabels:
      app: my-statefulset
  template:
    metadata:
      labels:
        app: my-statefulset
    spec:
      containers:
        - name: my-statefulset
          image: nginx:latest
          ports:
            - containerPort: 80
          volumeMounts:
            - name: data
              mountPath: /data
  volumeClaimTemplates:
    - metadata:
        name: data
      spec:
        accessModes: ["ReadWriteOnce"]
        resources:
          requests:
            storage: 1Gi
`,
  DaemonSet: `apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: my-daemonset
  namespace: default
  labels:
    app: my-daemonset
spec:
  selector:
    matchLabels:
      app: my-daemonset
  template:
    metadata:
      labels:
        app: my-daemonset
    spec:
      containers:
        - name: my-daemonset
          image: nginx:latest
          ports:
            - containerPort: 80
`,
}[props.kind] || ''

const yamlContent = ref(defaultYaml)

async function handleYamlSubmit() {
  if (!yamlContent.value.trim()) {
    ElMessage.error('YAML content is required')
    return
  }
  submitting.value = true
  try {
    const parsed = yaml.load(yamlContent.value) as any
    const ns = parsed?.metadata?.namespace || 'default'
    await props.createApi({ namespace: ns, yaml: yamlContent.value })
    ElMessage.success(`${props.kind} created successfully`)
    router.push(props.listRoute)
  } catch (e: any) {
    ElMessage.error(e?.message || 'Create failed')
  } finally {
    submitting.value = false
  }
}

function handleCancel() {
  router.push(props.listRoute)
}

function handleFormat() {
  yamlEditorRef.value?.handleFormat()
}

function handleCopy() {
  yamlEditorRef.value?.handleCopy()
}

function handleMaximize() {
  yamlEditorRef.value?.toggleFullscreen()
}
</script>

<template>
  <div class="workload-create">
    <div class="mode-switcher">
      <el-segmented v-model="mode" :options="[{ label: t('common.formCreate'), value: 'form' }, { label: t('common.yamlCreate'), value: 'yaml' }]" size="small" />
      <el-button size="small" style="margin-left: 12px;" @click="startClone">
        <el-icon><CopyDocument /></el-icon> 从现有资源克隆
      </el-button>
    </div>

    <CloneDialog
      :kind-label="kind"
      v-model="cloneMode"
      v-model:ns-value="cloneNamespace"
      v-model:name-value="cloneName"
      v-model:target="cloneTarget"
      :ns-options="cloneNsOptions"
      :name-options="cloneNameOptions"
      :ns-loading="cloneNsLoading"
      :name-loading="cloneNameLoading"
      :loading="cloneLoading"
      @load="handleLoadClone"
      @cancel="cancelClone"
    />

    <!-- WorkloadForm 始终挂载，用 v-show 控制显隐，避免克隆后切换 mode 时组件被销毁重建导致数据丢失 -->
    <WorkloadForm v-show="mode === 'form'" :kind="kind" :initial-data="parsedData" />

    <div v-if="mode !== 'form'" class="yaml-mode">
      <div class="yaml-card">
        <div class="yaml-card-header">
          <div class="yaml-card-left">
            <span class="yaml-card-title">YAML 配置</span>
            <el-button-group>
              <el-button size="small" @click="handleFormat">Format</el-button>
              <el-button size="small" @click="handleCopy">复制</el-button>
            </el-button-group>
            <el-tooltip content="最大化" placement="top">
              <el-icon class="maximize-btn" @click="handleMaximize"><FullScreen /></el-icon>
            </el-tooltip>
          </div>
          <div class="yaml-card-actions">
            <el-button size="small" @click="handleCancel">取消</el-button>
            <el-button size="small" type="primary" :loading="submitting" @click="handleYamlSubmit">创建</el-button>
          </div>
        </div>
        <div class="yaml-card-body">
          <YamlEditor ref="yamlEditorRef" v-model="yamlContent" height="calc(100vh - 180px)" :read-only="false" editable auto-format :show-toolbar="false" title="YAML 配置">
            <template #fullscreen-actions>
              <el-button size="small" @click="handleCancel">取消</el-button>
              <el-button size="small" type="primary" :loading="submitting" @click="handleYamlSubmit">创建</el-button>
            </template>
          </YamlEditor>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.workload-create { max-width: 1100px; margin: 0 auto; padding: 20px 0; }
.mode-switcher { display: flex; justify-content: center; margin-bottom: 12px; }
.yaml-mode { padding: 0 16px; }

.yaml-card {
  border: 1px solid var(--el-border-color-light);
  border-radius: 8px;
  overflow: hidden;
  background: var(--el-bg-color);
  box-shadow: 0 2px 12px 0 rgba(0, 0, 0, 0.05);
}

.yaml-card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 16px;
  background: var(--el-fill-color-lighter);
  border-bottom: 1px solid var(--el-border-color-light);
}

.yaml-card-left {
  display: flex;
  align-items: center;
  gap: 12px;
}

.yaml-card-title {
  font-size: 14px;
  font-weight: 600;
  color: var(--el-text-color-primary);
}

.yaml-card-actions {
  display: flex;
  gap: 8px;
}

.yaml-card-body {
  padding: 0;
}

.maximize-btn {
  cursor: pointer;
  font-size: 16px;
  color: var(--el-text-color-secondary);
  margin-left: 4px;
  transition: color 0.2s;
}
.maximize-btn:hover {
  color: var(--el-color-primary);
}
</style>
