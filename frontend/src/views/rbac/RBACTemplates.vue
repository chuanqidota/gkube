<script setup lang="ts">
import { ref, reactive } from 'vue'
import { DocumentCopy, View, Edit, Setting, User, Connection, Box } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import type { Component } from 'vue'
import request from '@/api/request'

interface RBACTemplate {
  id: string
  name: string
  description: string
  icon: Component
  iconColor: string
  scope: 'Cluster' | 'Namespace'
  roleKind: 'ClusterRole' | 'Role'
  rules: { apiGroups: string[]; resources: string[]; verbs: string[] }[]
}

const templates: RBACTemplate[] = [
  {
    id: 'readonly',
    name: '只读用户',
    description: '可以查看所有资源，但不能修改任何内容。适合审计人员或新员工。',
    icon: View,
    iconColor: '#67c23a',
    scope: 'Cluster',
    roleKind: 'ClusterRole',
    rules: [
      { apiGroups: [''], resources: ['*'], verbs: ['get', 'list', 'watch'] },
      { apiGroups: ['apps'], resources: ['*'], verbs: ['get', 'list', 'watch'] },
      { apiGroups: ['batch'], resources: ['*'], verbs: ['get', 'list', 'watch'] },
      { apiGroups: ['networking.k8s.io'], resources: ['*'], verbs: ['get', 'list', 'watch'] },
      { apiGroups: ['storage.k8s.io'], resources: ['*'], verbs: ['get', 'list', 'watch'] },
    ],
  },
  {
    id: 'developer',
    name: '开发人员',
    description: '可以管理工作负载（Deployment、Pod 等）和配置（ConfigMap、Secret），但不能管理集群资源。',
    icon: Edit,
    iconColor: '#409eff',
    scope: 'Namespace',
    roleKind: 'Role',
    rules: [
      { apiGroups: [''], resources: ['pods', 'pods/log', 'pods/exec', 'services', 'endpoints', 'configmaps', 'secrets', 'persistentvolumeclaims'], verbs: ['get', 'list', 'watch', 'create', 'update', 'patch', 'delete'] },
      { apiGroups: ['apps'], resources: ['deployments', 'statefulsets', 'daemonsets', 'replicasets'], verbs: ['get', 'list', 'watch', 'create', 'update', 'patch', 'delete'] },
      { apiGroups: ['batch'], resources: ['jobs', 'cronjobs'], verbs: ['get', 'list', 'watch', 'create', 'update', 'patch', 'delete'] },
      { apiGroups: ['autoscaling'], resources: ['horizontalpodautoscalers'], verbs: ['get', 'list', 'watch', 'create', 'update', 'patch', 'delete'] },
      { apiGroups: [''], resources: ['events'], verbs: ['get', 'list', 'watch'] },
    ],
  },
  {
    id: 'namespace-admin',
    name: '命名空间管理员',
    description: '在指定命名空间内拥有完全控制权限，包括 RBAC 管理。适合团队负责人。',
    icon: User,
    iconColor: '#e6a23c',
    scope: 'Namespace',
    roleKind: 'Role',
    rules: [
      { apiGroups: [''], resources: ['*'], verbs: ['*'] },
      { apiGroups: ['apps'], resources: ['*'], verbs: ['*'] },
      { apiGroups: ['batch'], resources: ['*'], verbs: ['*'] },
      { apiGroups: ['networking.k8s.io'], resources: ['*'], verbs: ['*'] },
      { apiGroups: ['autoscaling'], resources: ['*'], verbs: ['*'] },
      { apiGroups: ['policy'], resources: ['*'], verbs: ['*'] },
      { apiGroups: ['rbac.authorization.k8s.io'], resources: ['*'], verbs: ['*'] },
    ],
  },
  {
    id: 'ops-admin',
    name: '运维管理员',
    description: '可以管理节点、存储、网络等基础设施资源，但不管理 RBAC。适合运维团队。',
    icon: Setting,
    iconColor: '#909399',
    scope: 'Cluster',
    roleKind: 'ClusterRole',
    rules: [
      { apiGroups: [''], resources: ['nodes', 'namespaces', 'persistentvolumes', 'persistentvolumeclaims', 'services', 'endpoints', 'configmaps', 'secrets', 'events', 'serviceaccounts'], verbs: ['*'] },
      { apiGroups: ['apps'], resources: ['*'], verbs: ['*'] },
      { apiGroups: ['batch'], resources: ['*'], verbs: ['*'] },
      { apiGroups: ['networking.k8s.io'], resources: ['*'], verbs: ['*'] },
      { apiGroups: ['storage.k8s.io'], resources: ['*'], verbs: ['*'] },
      { apiGroups: ['autoscaling'], resources: ['*'], verbs: ['*'] },
      { apiGroups: ['policy'], resources: ['*'], verbs: ['*'] },
      { apiGroups: ['snapshot.storage.k8s.io'], resources: ['*'], verbs: ['*'] },
    ],
  },
  {
    id: 'network-admin',
    name: '网络管理员',
    description: '专注于网络资源管理：Service、Ingress、NetworkPolicy。适合网络团队。',
    icon: Connection,
    iconColor: '#f56c6c',
    scope: 'Cluster',
    roleKind: 'ClusterRole',
    rules: [
      { apiGroups: [''], resources: ['services', 'endpoints'], verbs: ['*'] },
      { apiGroups: ['networking.k8s.io'], resources: ['ingresses', 'ingressclasses', 'networkpolicies'], verbs: ['*'] },
      { apiGroups: [''], resources: ['pods', 'nodes', 'namespaces'], verbs: ['get', 'list', 'watch'] },
    ],
  },
  {
    id: 'storage-admin',
    name: '存储管理员',
    description: '专注于存储资源管理：PV、PVC、StorageClass、VolumeSnapshot。适合存储团队。',
    icon: Box,
    iconColor: '#b37feb',
    scope: 'Cluster',
    roleKind: 'ClusterRole',
    rules: [
      { apiGroups: [''], resources: ['persistentvolumes', 'persistentvolumeclaims'], verbs: ['*'] },
      { apiGroups: ['storage.k8s.io'], resources: ['storageclasses', 'csinodes', 'csidrivers'], verbs: ['*'] },
      { apiGroups: ['snapshot.storage.k8s.io'], resources: ['volumesnapshots', 'volumesnapshotclasses', 'volumesnapshotcontents'], verbs: ['*'] },
      { apiGroups: [''], resources: ['namespaces', 'nodes'], verbs: ['get', 'list', 'watch'] },
    ],
  },
]

// Apply dialog
const applyDialogVisible = ref(false)
const selectedTemplate = ref<RBACTemplate | null>(null)
const applyForm = reactive({
  subjectType: 'User' as 'User' | 'Group' | 'ServiceAccount',
  subjectName: '',
  namespace: 'default',
})
const applying = ref(false)

function openApplyDialog(template: RBACTemplate) {
  selectedTemplate.value = template
  applyForm.subjectType = 'User'
  applyForm.subjectName = ''
  applyForm.namespace = 'default'
  applyDialogVisible.value = true
}

async function handleApply() {
  if (!selectedTemplate.value) return
  if (!applyForm.subjectName.trim()) {
    ElMessage.warning('请输入主体名称')
    return
  }

  applying.value = true
  const tpl = selectedTemplate.value

  try {
    const roleName = `${tpl.id}-${applyForm.subjectName}-${Date.now()}`
    const isNamespaceScope = tpl.scope === 'Namespace'
    const ns = isNamespaceScope ? applyForm.namespace : ''

    // 1. Create Role/ClusterRole
    const roleYaml = buildRoleYaml(roleName, ns, tpl.roleKind, tpl.rules)
    const createRoleEndpoint = isNamespaceScope ? '/k8s/role/create' : '/k8s/clusterrole/create'
    await request.post(createRoleEndpoint, {
      yaml: roleYaml,
      ...(isNamespaceScope ? { namespace: ns } : {}),
    })

    // 2. Create RoleBinding/ClusterRoleBinding
    const bindingName = `${roleName}-binding`
    const bindingKind = isNamespaceScope ? 'RoleBinding' : 'ClusterRoleBinding'
    const bindingYaml = buildBindingYaml(bindingName, ns, bindingKind, tpl.roleKind, roleName, applyForm.subjectType, applyForm.subjectName)
    const createBindingEndpoint = isNamespaceScope ? '/k8s/rolebinding/create' : '/k8s/clusterrolebinding/create'
    await request.post(createBindingEndpoint, {
      yaml: bindingYaml,
      ...(isNamespaceScope ? { namespace: ns } : {}),
    })

    ElMessage.success(`模板 "${tpl.name}" 已成功应用到 ${applyForm.subjectType}/${applyForm.subjectName}`)
    applyDialogVisible.value = false
  } catch (e: any) {
    ElMessage.error(e?.message || '应用模板失败')
  } finally {
    applying.value = false
  }
}

function buildRoleYaml(name: string, namespace: string, kind: string, rules: RBACTemplate['rules']): string {
  const lines: string[] = []
  lines.push(`apiVersion: rbac.authorization.k8s.io/v1`)
  lines.push(`kind: ${kind}`)
  lines.push(`metadata:`)
  lines.push(`  name: ${name}`)
  if (namespace) lines.push(`  namespace: ${namespace}`)
  lines.push(`rules:`)
  for (const rule of rules) {
    lines.push(`  - apiGroups:`)
    for (const ag of rule.apiGroups) lines.push(`      - "${ag}"`)
    lines.push(`    resources:`)
    for (const r of rule.resources) lines.push(`      - "${r}"`)
    lines.push(`    verbs:`)
    for (const v of rule.verbs) lines.push(`      - "${v}"`)
  }
  return lines.join('\n')
}

function buildBindingYaml(
  name: string,
  namespace: string,
  bindingKind: string,
  roleKind: string,
  roleName: string,
  subjectKind: string,
  subjectName: string,
): string {
  const lines: string[] = []
  lines.push(`apiVersion: rbac.authorization.k8s.io/v1`)
  lines.push(`kind: ${bindingKind}`)
  lines.push(`metadata:`)
  lines.push(`  name: ${name}`)
  if (namespace) lines.push(`  namespace: ${namespace}`)
  lines.push(`roleRef:`)
  lines.push(`  apiGroup: rbac.authorization.k8s.io`)
  lines.push(`  kind: ${roleKind}`)
  lines.push(`  name: ${roleName}`)
  lines.push(`subjects:`)
  lines.push(`  - kind: ${subjectKind}`)
  lines.push(`    name: ${subjectName}`)
  if (subjectKind === 'ServiceAccount' && namespace) {
    lines.push(`    namespace: ${namespace}`)
  }
  return lines.join('\n')
}

// Preview dialog
const previewDialogVisible = ref(false)
const previewTitle = ref('')
const previewYaml = ref('')

function openPreview(template: RBACTemplate) {
  const roleName = `<${template.roleKind.toLowerCase()}-name>`
  const roleYaml = buildRoleYaml(roleName, template.scope === 'Namespace' ? '<namespace>' : '', template.roleKind, template.rules)
  const bindingYaml = buildBindingYaml(
    `<binding-name>`,
    template.scope === 'Namespace' ? '<namespace>' : '',
    template.scope === 'Namespace' ? 'RoleBinding' : 'ClusterRoleBinding',
    template.roleKind,
    roleName,
    '<subject-kind>',
    '<subject-name>',
  )
  previewTitle.value = template.name
  previewYaml.value = `# ${template.roleKind}\n${roleYaml}\n\n---\n\n# ${template.scope === 'Namespace' ? 'RoleBinding' : 'ClusterRoleBinding'}\n${bindingYaml}`
  previewDialogVisible.value = true
}
</script>

<template>
  <div class="page-container">
    <el-card shadow="never" class="intro-card">
      <div class="intro-content">
        <el-icon :size="24" color="#409eff"><DocumentCopy /></el-icon>
        <div>
          <h3 style="margin: 0 0 4px 0;">RBAC 模板</h3>
          <p style="margin: 0; color: #909399; font-size: 14px;">
            选择常用角色模板，一键创建 Role/ClusterRole 并绑定到指定用户。点击「应用」开始配置。
          </p>
        </div>
      </div>
    </el-card>

    <div class="template-grid">
      <el-card
        v-for="tpl in templates"
        :key="tpl.id"
        shadow="hover"
        class="template-card"
      >
        <div class="template-header">
          <div class="template-icon" :style="{ backgroundColor: tpl.iconColor + '15' }">
            <el-icon :size="28" :color="tpl.iconColor"><component :is="tpl.icon" /></el-icon>
          </div>
          <div>
            <h3 style="margin: 0 0 4px 0;">{{ tpl.name }}</h3>
            <el-tag :type="tpl.scope === 'Cluster' ? 'danger' : ''" size="small">
              {{ tpl.scope === 'Cluster' ? '集群级' : '命名空间级' }}
            </el-tag>
            <el-tag type="info" size="small" style="margin-left: 4px;">
              {{ tpl.roleKind }}
            </el-tag>
          </div>
        </div>
        <p class="template-desc">{{ tpl.description }}</p>
        <div class="template-rules-summary">
          <el-tag v-for="rule in tpl.rules.slice(0, 3)" :key="rule.resources.join(',')" size="small" type="info" style="margin: 2px;">
            {{ rule.resources.join(', ') }}
          </el-tag>
          <el-tag v-if="tpl.rules.length > 3" size="small" type="info" style="margin: 2px;">
            +{{ tpl.rules.length - 3 }} more
          </el-tag>
        </div>
        <div class="template-actions">
          <el-button size="small" @click="openPreview(tpl)">
            <el-icon><View /></el-icon> 预览
          </el-button>
          <el-button type="primary" size="small" @click="openApplyDialog(tpl)">
            <el-icon><DocumentCopy /></el-icon> 应用
          </el-button>
        </div>
      </el-card>
    </div>

    <!-- Apply Dialog -->
    <el-dialog
      v-model="applyDialogVisible"
      :title="`应用模板: ${selectedTemplate?.name || ''}`"
      width="520px"
      destroy-on-close
    >
      <el-form label-width="100px">
        <el-form-item label="模板">
          <el-tag>{{ selectedTemplate?.name }}</el-tag>
          <el-tag :type="selectedTemplate?.scope === 'Cluster' ? 'danger' : ''" style="margin-left: 4px;" size="small">
            {{ selectedTemplate?.scope === 'Cluster' ? '集群级' : '命名空间级' }}
          </el-tag>
        </el-form-item>
        <el-form-item label="主体类型">
          <el-select v-model="applyForm.subjectType" style="width: 100%;">
            <el-option label="User" value="User" />
            <el-option label="Group" value="Group" />
            <el-option label="ServiceAccount" value="ServiceAccount" />
          </el-select>
        </el-form-item>
        <el-form-item label="主体名称">
          <el-input v-model="applyForm.subjectName" placeholder="输入用户名、组名或SA名称" />
        </el-form-item>
        <el-form-item v-if="selectedTemplate?.scope === 'Namespace'" label="命名空间">
          <el-input v-model="applyForm.namespace" placeholder="输入命名空间名称" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="applyDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="applying" @click="handleApply">
          确认应用
        </el-button>
      </template>
    </el-dialog>

    <!-- Preview Dialog -->
    <el-dialog
      v-model="previewDialogVisible"
      :title="`模板预览: ${previewTitle}`"
      width="700px"
      destroy-on-close
    >
      <el-input
        v-model="previewYaml"
        type="textarea"
        :rows="30"
        readonly
        style="font-family: monospace;"
      />
    </el-dialog>
  </div>
</template>

<style scoped>
.page-container {
  padding: 20px;
}
.intro-card {
  margin-bottom: 20px;
  border-radius: 8px;
}
.intro-content {
  display: flex;
  align-items: center;
  gap: 16px;
}
.template-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(380px, 1fr));
  gap: 16px;
}
.template-card {
  border-radius: 8px;
  transition: transform 0.2s;
}
.template-card:hover {
  transform: translateY(-2px);
}
.template-header {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 12px;
}
.template-icon {
  width: 48px;
  height: 48px;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}
.template-desc {
  color: #606266;
  font-size: 14px;
  line-height: 1.6;
  margin: 0 0 12px 0;
}
.template-rules-summary {
  margin-bottom: 16px;
}
.template-actions {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
}
</style>
