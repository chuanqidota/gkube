<template>
  <el-drawer
    :model-value="modelValue"
    @update:model-value="$emit('update:modelValue', $event)"
    :title="drawerTitle"
    size="85%"
    direction="rtl"
    class="yaml-drawer"
    :body-style="{ padding: '0', height: '100%' }"
    :destroy-on-close="true"
  >
    <div v-loading="loading" style="height: calc(100vh - 52px);">
      <YamlEditor
        v-if="!loading"
        v-model="yamlContent"
        height="100%"
        auto-format
        show-save-buttons
        :saving="saving"
        @save="handleSave"
        @cancel="handleCancel"
      />
    </div>
  </el-drawer>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { ElMessage } from 'element-plus'
import YamlEditor from './YamlEditor.vue'
import {
  // Workload
  getDeploymentYaml, updateDeploymentYaml,
  getStatefulSetYaml, updateStatefulSetYaml,
  getDaemonSetYaml, updateDaemonSetYaml,
  getPodYaml, updatePodYaml,
  getJobYaml, updateJobYaml,
  getCronJobYaml, updateCronJobYaml,
  getReplicaSetYaml,
  // Network
  getServiceYaml, updateServiceYaml,
  getIngressYaml, updateIngressYaml,
  getNetworkPolicyYaml, updateNetworkPolicyYaml,
  // Storage
  getPvYaml, updatePvYaml,
  getPvcYaml,
  getStorageClassYaml, updateStorageClassYaml,
  getVolumeSnapshotYaml, updateVolumeSnapshot,
  getVolumeSnapshotClassYaml, updateVolumeSnapshotClass,
  // Config
  getConfigMapYaml, updateConfigMap,
  getSecretYaml, updateSecret,
  getResourceQuotaYaml, updateResourceQuota,
  getLimitRangeYaml, updateLimitRange,
  // Node
  getNodeYaml, updateNodeYaml,
  // Namespace
  getNamespaceYaml, updateNamespace,
  // RBAC
  getClusterRoleYaml,
  getClusterRoleBindingYaml,
  getRoleYaml,
  getRoleBindingYaml,
  getServiceAccountYaml,
  // CRD
  getCrdYaml, updateCrd,
  // HPA
  getHpaYaml, updateHpa,
  // PDB
  getPdbYaml, updatePdb,
} from '@/api/resource'

// Resource type definition
export type ResourceType =
  | 'deployment' | 'statefulset' | 'daemonset' | 'pod' | 'job' | 'cronjob' | 'replicaset'
  | 'service' | 'ingress' | 'networkpolicy'
  | 'pv' | 'pvc' | 'storageclass' | 'volumesnapshot' | 'volumesnapshotclass'
  | 'configmap' | 'secret' | 'resourcequota' | 'limitrange'
  | 'node' | 'namespace'
  | 'clusterrole' | 'clusterrolebinding' | 'role' | 'rolebinding' | 'serviceaccount'
  | 'crd' | 'hpa' | 'pdb'

// Resource display names
const resourceDisplayNames: Record<ResourceType, string> = {
  deployment: 'Deployment',
  statefulset: 'StatefulSet',
  daemonset: 'DaemonSet',
  pod: 'Pod',
  job: 'Job',
  cronjob: 'CronJob',
  replicaset: 'ReplicaSet',
  service: 'Service',
  ingress: 'Ingress',
  networkpolicy: 'NetworkPolicy',
  pv: 'PersistentVolume',
  pvc: 'PersistentVolumeClaim',
  storageclass: 'StorageClass',
  volumesnapshot: 'VolumeSnapshot',
  volumesnapshotclass: 'VolumeSnapshotClass',
  configmap: 'ConfigMap',
  secret: 'Secret',
  resourcequota: 'ResourceQuota',
  limitrange: 'LimitRange',
  node: 'Node',
  namespace: 'Namespace',
  clusterrole: 'ClusterRole',
  clusterrolebinding: 'ClusterRoleBinding',
  role: 'Role',
  rolebinding: 'RoleBinding',
  serviceaccount: 'ServiceAccount',
  crd: 'CRD',
  hpa: 'HPA',
  pdb: 'PDB',
}

// Check if resource is cluster-scoped (no namespace)
const clusterScopedResources: ResourceType[] = [
  'pv', 'storageclass', 'volumesnapshotclass',
  'node', 'namespace',
  'clusterrole', 'clusterrolebinding',
  'crd',
]

// API function signatures
interface ResourceApi {
  getYaml: (params: any) => Promise<any>
  updateYaml: ((data: any) => Promise<any>) | null
}

// Resource API registry
const resourceApis: Record<ResourceType, ResourceApi> = {
  // Workload
  deployment: { getYaml: getDeploymentYaml, updateYaml: updateDeploymentYaml },
  statefulset: { getYaml: getStatefulSetYaml, updateYaml: updateStatefulSetYaml },
  daemonset: { getYaml: getDaemonSetYaml, updateYaml: updateDaemonSetYaml },
  pod: { getYaml: getPodYaml, updateYaml: updatePodYaml },
  job: { getYaml: getJobYaml, updateYaml: updateJobYaml },
  cronjob: { getYaml: getCronJobYaml, updateYaml: updateCronJobYaml },
  replicaset: { getYaml: getReplicaSetYaml, updateYaml: null },
  // Network
  service: { getYaml: getServiceYaml, updateYaml: updateServiceYaml },
  ingress: { getYaml: getIngressYaml, updateYaml: updateIngressYaml },
  networkpolicy: { getYaml: getNetworkPolicyYaml, updateYaml: updateNetworkPolicyYaml },
  // Storage
  pv: { getYaml: getPvYaml, updateYaml: updatePvYaml },
  pvc: { getYaml: getPvcYaml, updateYaml: null },
  storageclass: { getYaml: getStorageClassYaml, updateYaml: updateStorageClassYaml },
  volumesnapshot: { getYaml: getVolumeSnapshotYaml, updateYaml: updateVolumeSnapshot },
  volumesnapshotclass: { getYaml: getVolumeSnapshotClassYaml, updateYaml: updateVolumeSnapshotClass },
  // Config
  configmap: { getYaml: getConfigMapYaml, updateYaml: updateConfigMap },
  secret: { getYaml: getSecretYaml, updateYaml: updateSecret },
  resourcequota: { getYaml: getResourceQuotaYaml, updateYaml: updateResourceQuota },
  limitrange: { getYaml: getLimitRangeYaml, updateYaml: updateLimitRange },
  // Node & Namespace
  node: { getYaml: getNodeYaml, updateYaml: updateNodeYaml },
  namespace: { getYaml: getNamespaceYaml, updateYaml: updateNamespace },
  // RBAC (read-only for now)
  clusterrole: { getYaml: getClusterRoleYaml, updateYaml: null },
  clusterrolebinding: { getYaml: getClusterRoleBindingYaml, updateYaml: null },
  role: { getYaml: getRoleYaml, updateYaml: null },
  rolebinding: { getYaml: getRoleBindingYaml, updateYaml: null },
  serviceaccount: { getYaml: getServiceAccountYaml, updateYaml: null },
  // Others
  crd: { getYaml: getCrdYaml, updateYaml: updateCrd },
  hpa: { getYaml: getHpaYaml, updateYaml: updateHpa },
  pdb: { getYaml: getPdbYaml, updateYaml: updatePdb },
}

const props = withDefaults(defineProps<{
  modelValue: boolean
  resourceType: ResourceType
  namespace?: string
  name: string
  title?: string
}>(), {
  namespace: '',
  title: '',
})

const emit = defineEmits<{
  'update:modelValue': [value: boolean]
  'saved': []
}>()

const loading = ref(false)
const saving = ref(false)
const yamlContent = ref('')

// Computed title
const drawerTitle = computed(() => {
  if (props.title) return props.title
  const typeName = resourceDisplayNames[props.resourceType] || props.resourceType
  return `${typeName} YAML: ${props.name}`
})

// Resources that don't need name in update data (name is extracted from YAML)
const noNameInUpdateData: ResourceType[] = [
  'volumesnapshot', 'volumesnapshotclass', 'hpa', 'pdb',
]

// Build params for getYaml based on resource type
function buildGetYamlParams() {
  const isCluster = clusterScopedResources.includes(props.resourceType)
  if (isCluster) {
    return { name: props.name }
  }
  return { namespace: props.namespace, name: props.name }
}

// Build data for updateYaml based on resource type
function buildUpdateData() {
  const isCluster = clusterScopedResources.includes(props.resourceType)
  const needsName = !noNameInUpdateData.includes(props.resourceType)

  if (isCluster) {
    return needsName
      ? { name: props.name, yaml: yamlContent.value }
      : { yaml: yamlContent.value }
  }
  return needsName
    ? { namespace: props.namespace, name: props.name, yaml: yamlContent.value }
    : { namespace: props.namespace, yaml: yamlContent.value }
}

// Load YAML
async function fetchYaml() {
  if (!props.name) return

  loading.value = true
  yamlContent.value = ''

  try {
    const api = resourceApis[props.resourceType]
    const params = buildGetYamlParams()
    const res = await api.getYaml(params)
    // 兼容两种后端返回格式：直接返回字符串 或 包装在 { yaml: "..." } 中
    const raw = res.data ?? res
    yamlContent.value = typeof raw === 'object' && raw?.yaml ? raw.yaml : raw
  } catch (error: any) {
    ElMessage.error('获取 YAML 失败: ' + (error.message || '未知错误'))
  } finally {
    loading.value = false
  }
}

// Save YAML
async function handleSave() {
  const api = resourceApis[props.resourceType]
  if (!api.updateYaml) {
    ElMessage.warning('该资源类型不支持编辑')
    return
  }

  saving.value = true
  try {
    const data = buildUpdateData()
    await api.updateYaml(data)
    ElMessage.success('保存成功')
    emit('saved')
    emit('update:modelValue', false)
  } catch (error: any) {
    ElMessage.error('保存失败: ' + (error.message || '未知错误'))
  } finally {
    saving.value = false
  }
}

// Cancel and reload
function handleCancel() {
  fetchYaml()
}

// Watch for drawer open
watch(() => props.modelValue, (visible) => {
  if (visible && props.name) {
    fetchYaml()
  }
})

// Expose for parent to manually refresh
defineExpose({ fetchYaml })
</script>

<style>
.yaml-drawer .el-drawer__header {
  padding: 6px 16px;
  margin-bottom: 0;
  min-height: auto;
}
</style>
