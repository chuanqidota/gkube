<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getIngressDetail, deleteIngress, getIngressEvents, getIngressTLSCertStatus, ingressApi } from '@/api/resource'
import { FullScreen, Aim } from '@element-plus/icons-vue'
import YamlDrawer from '@/components/YamlDrawer.vue'
import DetailPageLayout from '@/components/DetailPageLayout.vue'
import DetailPageHeader from '@/components/DetailPageHeader.vue'
import EventsTable from '@/components/EventsTable.vue'
import LabelsBlock from '@/components/LabelsBlock.vue'
import IngressForm from './components/IngressForm.vue'
import { useAutoRefresh } from '@/composables/useAutoRefresh'
import { useEditDrawer } from '@/composables/useEditDrawer'

const { t } = useI18n()
const route = useRoute()
const router = useRouter()
const loading = ref(false)
const ingressRaw = ref<any>(null)
const yamlDialogVisible = ref(false)

// Events
const events = ref<any[]>([])
const eventsLoading = ref(false)

// TLS Cert Status
const tlsCerts = ref<any[]>([])
const tlsCertsLoading = ref(false)

// Edit dialog
const { editDialogVisible, editFullscreen, handleEdit, handleEditSuccess, handleEditCancel } = useEditDrawer(async () => {
  fetchDetail()
})

const namespace = route.params.namespace as string
const name = route.params.name as string

const ingress = computed(() => {
  const raw = ingressRaw.value
  if (!raw) return null
  const meta = raw.metadata || {}
  const spec = raw.spec || {}

  const rules = (spec.rules || []).map((rule: any) => ({
    host: rule.host || '*',
    paths: (rule.http?.paths || []).map((p: any) => ({
      path: p.path || '/',
      pathType: p.pathType || 'ImplementationSpecific',
      backend: {
        serviceName: p.backend?.service?.name || '',
        servicePort: p.backend?.service?.port?.number || p.backend?.service?.port?.name || '',
      },
    })),
  }))

  const tls = (spec.tls || []).map((t: any) => ({
    hosts: t.hosts || [],
    secretName: t.secretName || '',
  }))

  const defaultBackend = spec.defaultBackend
    ? {
        serviceName: spec.defaultBackend.service?.name || '',
        servicePort: spec.defaultBackend.service?.port?.number || spec.defaultBackend.service?.port?.name || '',
      }
    : null

  const address = (raw.status?.loadBalancer?.ingress || [])
    .map((i: any) => i.ip || i.hostname)
    .filter(Boolean)
    .join(', ')

  return {
    name: meta.name || '',
    namespace: meta.namespace || '',
    ingressClassName: spec.ingressClassName || '',
    labels: meta.labels || {},
    address,
    defaultBackend,
    rules,
    tls,
  }
})

const statusTag = computed(() => {
  if (ingress.value?.ingressClassName) {
    return { text: ingress.value.ingressClassName, type: 'info' as const }
  }
  return undefined
})

async function fetchDetail() {
  loading.value = true
  try {
    const res: any = await getIngressDetail({ namespace, name })
    ingressRaw.value = res.data
    if (res.data?.spec?.tls?.length > 0) {
      fetchTLSCerts()
    } else {
      tlsCerts.value = []
    }
  } catch (e: any) {
    ElMessage.error(e?.message || t('network.loadDetailFailed'))
  } finally {
    loading.value = false
  }
}

async function fetchEvents() {
  eventsLoading.value = true
  try {
    const res: any = await getIngressEvents({ namespace, name })
    events.value = res.data || []
  } catch (e) {
    events.value = []
  } finally {
    eventsLoading.value = false
  }
}

async function fetchTLSCerts() {
  tlsCertsLoading.value = true
  try {
    const res: any = await getIngressTLSCertStatus({ namespace, name })
    tlsCerts.value = res.data || []
  } catch (e) {
    tlsCerts.value = []
  } finally {
    tlsCertsLoading.value = false
  }
}

function handleOpenYaml() {
  yamlDialogVisible.value = true
}

function handleYamlSaved() {
  fetchDetail()
}

async function handleDelete() {
  try {
    await ElMessageBox.confirm(
      `确定要删除 Ingress "${name}" 吗？此操作不可恢复。`,
      '确认删除',
      { type: 'error', confirmButtonText: '删除', cancelButtonText: '取消' }
    )
    await deleteIngress({ namespace, name })
    ElMessage.success(t('network.ingressDeleted'))
    router.push('/network/ingresses')
  } catch (e: any) {
    if (e !== 'cancel') {
      ElMessage.error(e?.message || t('common.deleteFailed'))
    }
  }
}

function certStatusType(status: string) {
  switch (status) {
    case 'valid': return 'success'
    case 'expiring': return 'warning'
    case 'expired': return 'danger'
    case 'error': return 'danger'
    default: return 'info'
  }
}

function certStatusText(status: string) {
  switch (status) {
    case 'valid': return '有效'
    case 'expiring': return '即将过期'
    case 'expired': return '已过期'
    case 'error': return '异常'
    default: return '未知'
  }
}

function formatDate(iso: string) {
  if (!iso) return '-'
  return iso.replace('T', ' ').replace(/Z$/, '').replace(/\+.*/, '')
}

const tlsCertMap = computed(() => {
  const map: Record<string, any> = {}
  for (const cert of tlsCerts.value) {
    map[cert.secretName] = cert
  }
  return map
})

const { isRunning, countdown, currentInterval, availableIntervals, toggle, refresh: manualRefresh, setIntervalOption } = useAutoRefresh(async () => {
  fetchDetail()
  fetchEvents()
}, { autoStart: false })

onMounted(() => {
  fetchDetail()
  fetchEvents()
})
</script>

<template>
  <DetailPageLayout :resizable="true">
    <!-- Header -->
    <DetailPageHeader
      :title="name"
      :status-tag="statusTag"
      :namespace="namespace"
      :loading="loading"
      :is-running="isRunning"
      :countdown="countdown"
      :current-interval="currentInterval"
      :available-intervals="availableIntervals"
      @refresh="manualRefresh()"
      @toggle="toggle()"
      @set-interval-option="setIntervalOption"
      @back="router.push('/network/ingresses')"
    >
      <template #meta>
        <span class="replicas-info" v-if="ingress?.rules?.length">
          {{ ingress.rules.length }} 条规则
        </span>
      </template>
      <template #actions>
        <el-button type="info" @click="handleEdit">编辑</el-button>
        <el-button @click="handleOpenYaml">YAML</el-button>
        <el-button type="danger" @click="handleDelete">删除</el-button>
      </template>
    </DetailPageHeader>

    <!-- Left panel: basic info -->
    <template v-if="ingress" #left>
      <div class="panel-title">基本信息</div>
      <div class="info-body">
        <el-descriptions :column="1" border size="small">
          <el-descriptions-item label="名称">{{ ingress.name }}</el-descriptions-item>
          <el-descriptions-item label="命名空间">{{ ingress.namespace }}</el-descriptions-item>
          <el-descriptions-item label="Ingress Class">{{ ingress.ingressClassName || '-' }}</el-descriptions-item>
          <el-descriptions-item label="地址">
            <span v-if="ingress.address">{{ ingress.address }}</span>
            <span v-else class="text-muted">-</span>
          </el-descriptions-item>
        </el-descriptions>

        <!-- Labels -->
        <div v-if="ingress.labels && Object.keys(ingress.labels).length > 0" style="margin-top: var(--gk-space-4);">
          <h4 style="margin: 0 0 8px; font-size: var(--gk-font-size-sm);">Labels</h4>
          <LabelsBlock :labels="ingress.labels" />
        </div>
      </div>
    </template>

    <!-- Right-top: Rules + TLS -->
    <template v-if="ingress" #right-top>
      <div class="right-panel-inner">
        <!-- Default Backend -->
        <div v-if="ingress.defaultBackend" class="section-block">
          <div class="panel-title">默认后端</div>
          <div class="rules-body">
            <el-alert type="info" :closable="false" show-icon>
              <template #title>
                所有未匹配规则的流量将转发至
                <el-button
                  v-if="ingress.defaultBackend.serviceName"
                  link
                  type="primary"
                  size="small"
                  @click="router.push(`/network/services/${namespace}/${ingress.defaultBackend.serviceName}`)"
                >{{ ingress.defaultBackend.serviceName }}</el-button>
                <span v-else>-</span>
                :{{ ingress.defaultBackend.servicePort || '-' }}
              </template>
            </el-alert>
          </div>
        </div>

        <!-- Rules -->
        <div v-if="ingress.rules && ingress.rules.length > 0" class="section-block">
          <div class="panel-title">
            路由规则
            <span class="count-badge">{{ ingress.rules.length }} 条</span>
          </div>
          <div class="rules-body">
            <el-table :data="ingress.rules" border stripe size="small">
              <el-table-column prop="host" label="Host" min-width="200" show-overflow-tooltip />
              <el-table-column label="Paths" min-width="300">
                <template #default="{ row }">
                  <div v-if="row.paths && row.paths.length > 0">
                    <div v-for="(p, idx) in row.paths" :key="idx" style="margin-bottom: 4px;">
                      <el-tag size="small" type="info">{{ p.pathType || 'ImplementationSpecific' }}</el-tag>
                      {{ p.path || '/' }} ->
                      <el-button
                        v-if="p.backend?.serviceName"
                        link
                        type="primary"
                        size="small"
                        @click="router.push(`/network/services/${namespace}/${p.backend.serviceName}`)"
                      >{{ p.backend.serviceName }}</el-button>
                      <span v-else>-</span>:{{ p.backend?.servicePort || '-' }}
                    </div>
                  </div>
                  <span v-else>-</span>
                </template>
              </el-table-column>
            </el-table>
          </div>
        </div>

        <!-- TLS -->
        <div v-if="ingress.tls && ingress.tls.length > 0" class="section-block">
          <div class="panel-title">
            TLS
            <span class="count-badge">{{ ingress.tls.length }} 条</span>
          </div>
          <div class="rules-body" v-loading="tlsCertsLoading">
            <el-table :data="ingress.tls" border stripe size="small">
              <el-table-column label="Hosts" min-width="180">
                <template #default="{ row }">
                  <el-tag v-for="h in (row.hosts || [])" :key="h" size="small" style="margin-right: 4px;">{{ h }}</el-tag>
                </template>
              </el-table-column>
              <el-table-column prop="secretName" label="Secret Name" min-width="160" />
              <el-table-column label="证书状态" min-width="160">
                <template #default="{ row }">
                  <template v-if="tlsCertMap[row.secretName]">
                    <el-tag
                      :type="certStatusType(tlsCertMap[row.secretName].status)"
                      size="small"
                      effect="dark"
                    >
                      {{ certStatusText(tlsCertMap[row.secretName].status) }}
                    </el-tag>
                    <div class="cert-detail">{{ tlsCertMap[row.secretName].message }}</div>
                  </template>
                  <span v-else class="text-muted">-</span>
                </template>
              </el-table-column>
              <el-table-column label="过期时间" min-width="160">
                <template #default="{ row }">
                  <template v-if="tlsCertMap[row.secretName]?.notAfter">
                    <div>{{ formatDate(tlsCertMap[row.secretName].notAfter) }}</div>
                    <div class="cert-issuer">签发: {{ tlsCertMap[row.secretName].issuer || '-' }}</div>
                  </template>
                  <span v-else class="text-muted">-</span>
                </template>
              </el-table-column>
            </el-table>
          </div>
        </div>
      </div>
    </template>

    <!-- Right-bottom: Events -->
    <template v-if="ingress" #right-bottom>
      <div class="panel-title">
        事件
        <span class="count-badge">{{ events.length }} 条</span>
      </div>
      <EventsTable :events="events" :loading="eventsLoading" time-field="last_seen" />
    </template>

    <!-- YAML Drawer -->
    <YamlDrawer
      v-model="yamlDialogVisible"
      :get-yaml="ingressApi.getYaml"
      :update-yaml="ingressApi.updateYaml"
      :namespace="namespace"
      :name="name"
      title="Ingress YAML"
      @saved="handleYamlSaved"
    />

    <!-- Edit Drawer -->
    <el-drawer
      v-model="editDialogVisible"
      title="编辑 Ingress"
      :size="editFullscreen ? '100%' : '85%'"
      direction="rtl"
      :destroy-on-close="true"
      :body-style="{ padding: '0', height: '100%' }"
    >
      <template #header>
        <div class="drawer-header">
          <span class="drawer-title">编辑 Ingress</span>
          <el-tooltip :content="editFullscreen ? '退出全屏' : '全屏'" placement="top">
            <el-icon class="fullscreen-btn" @click="editFullscreen = !editFullscreen">
              <FullScreen v-if="!editFullscreen" />
              <Aim v-else />
            </el-icon>
          </el-tooltip>
        </div>
      </template>
      <div style="height: calc(100dvh - 52px); overflow-y: auto;">
        <IngressForm
          v-if="editDialogVisible && ingressRaw"
          :is-edit="true"
          :initial-data="ingressRaw"
          @success="handleEditSuccess"
          @cancel="handleEditCancel"
        />
      </div>
    </el-drawer>
  </DetailPageLayout>
</template>

<style scoped>
.replicas-info {
  font-size: 12px;
  color: var(--gk-color-text-primary);
}

.panel-title {
  font-size: var(--gk-font-size-sm);
  font-weight: 600;
  padding: var(--gk-space-2) var(--gk-space-4);
  background: var(--gk-neutral-100);
  border-bottom: 1px solid var(--gk-color-border-light);
  display: flex;
  align-items: center;
  gap: 6px;
  flex-shrink: 0;
}

.count-badge {
  font-weight: 400;
  font-size: var(--gk-font-size-xs);
  color: var(--gk-color-text-secondary);
}

.info-body {
  flex: 1;
  overflow-y: auto;
  padding: 14px;
}

.right-panel-inner {
  flex: 1;
  overflow-y: auto;
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.section-block {
  border: 1px solid var(--gk-color-border-light);
  border-radius: var(--gk-radius-md);
  display: flex;
  flex-direction: column;
  overflow: hidden;
  background: var(--el-bg-color);
  flex-shrink: 0;
}

.rules-body {
  padding: 14px;
  overflow-y: auto;
}

.text-muted {
  color: var(--gk-color-text-secondary);
  font-size: 12px;
}

.cert-detail {
  font-size: 12px;
  color: var(--gk-color-text-secondary);
  margin-top: 2px;
  line-height: 1.4;
}

.cert-issuer {
  font-size: 11px;
  color: var(--gk-color-text-secondary);
  margin-top: 2px;
}

.drawer-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  width: 100%;
}

.drawer-title {
  font-size: var(--gk-font-size-lg);
  font-weight: 600;
}

.fullscreen-btn {
  cursor: pointer;
  font-size: 18px;
  color: var(--gk-color-text-primary);
  transition: color 0.2s;
}

.fullscreen-btn:hover {
  color: var(--el-color-primary);
}
</style>
