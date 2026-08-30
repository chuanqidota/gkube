<script setup lang="ts">
import { ref, computed, watch, nextTick } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage } from 'element-plus'
import type { FormInstance, FormRules } from 'element-plus'
import { createBinding, updateBinding, getNamespaceList, searchUsers } from '@/api/rbac'

const { t } = useI18n()

const props = defineProps<{
  visible: boolean
  clusterId: number
  clusterName: string  // 用于加载命名空间列表
  binding?: any | null
  roles: any[]
}>()

const emit = defineEmits<{
  (e: 'update:visible', val: boolean): void
  (e: 'success'): void
}>()

const dialogVisible = computed({
  get: () => props.visible,
  set: (val) => emit('update:visible', val),
})

const isEditMode = computed(() => !!props.binding)
const formRef = ref<FormInstance>()
const saving = ref(false)

const form = ref({
  userId: '' as string | number,
  roleId: '' as string | number,
  namespaces: [] as string[],
})

const userOptions = ref<any[]>([])
const userLoading = ref(false)

// 默认角色：空间观察者
function setDefaultRole() {
  if (!props.roles || props.roles.length === 0) return
  const viewerRole = props.roles.find((r: any) => r.name === 'ns-viewer')
    || props.roles.find((r: any) => r.name?.includes('viewer'))
  if (viewerRole) form.value.roleId = viewerRole.id
}

const nsList = ref<string[]>([])
const nsLoading = ref(false)
const nsLoadError = ref(false)

const rules = computed<FormRules>(() => ({
  userId: [{ required: true, message: t('rbac.selectUser'), trigger: 'change' }],
  roleId: [{ required: true, message: t('rbac.selectRole'), trigger: 'change' }],
  namespaces: [{ required: true, type: 'array', min: 1, message: t('rbac.selectNamespace'), trigger: 'change' }],
}))

// 只显示命名空间级角色
const filteredRoles = computed(() => {
  return props.roles.filter(r => r.scopeType === 'namespace')
})

const selectedRole = computed(() => {
  return props.roles.find(r => r.id === form.value.roleId)
})

const roleDescriptionMap: Record<string, string> = {
  'ns-admin':  t('rbac.roleDesc.nsAdmin'),
  'ns-editor': t('rbac.roleDesc.nsEditor'),
  'ns-viewer': t('rbac.roleDesc.nsViewer'),
}

async function loadAllUsers() {
  if (userOptions.value.length > 0) return
  userLoading.value = true
  try {
    const res: any = await searchUsers({ page: 1, size: 100 })
    userOptions.value = res?.data?.items || []
  } catch {
    userOptions.value = []
  } finally {
    userLoading.value = false
  }
}

// 新增模式重置
watch(() => props.visible, (val) => {
  if (val && !isEditMode.value) {
    form.value.userId = ''
    form.value.roleId = ''
    form.value.namespaces = []
    nextTick(() => formRef.value?.clearValidate())
    loadAllUsers()
    setDefaultRole()
  }
})

// 编辑模式
watch(() => props.binding, (b) => {
  if (b) {
    form.value.userId = b.userId
    form.value.roleId = b.roleId
    form.value.namespaces = b.namespace ? [b.namespace] : []
    userOptions.value = [{ id: b.userId, username: b.username, display_name: b.displayName }]
  }
}, { immediate: true })

// 加载命名空间列表
watch(() => props.visible, async (val) => {
  if (!val || isEditMode.value || !props.clusterName) return
  nsLoading.value = true
  nsLoadError.value = false
  try {
    const res: any = await getNamespaceList(props.clusterName)
    const nsData = res?.data ?? res
    const nsArr: any[] = Array.isArray(nsData) ? nsData : (nsData?.items || [])
    nsList.value = nsArr.map((ns: any) => ns.metadata?.name || ns.name || ns)
  } catch {
    nsLoadError.value = true
  } finally {
    nsLoading.value = false
  }
})

// 选择"全部"时自动选中所有命名空间
function handleSelectAllNs() {
  form.value.namespaces = [...nsList.value]
}

async function handleSubmit() {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return

  saving.value = true
  try {
    if (isEditMode.value) {
      await updateBinding(props.binding.id, { roleId: Number(form.value.roleId) })
      ElMessage.success(t('rbac.updateSuccess'))
    } else {
      // 过滤掉 __all__ 标记
      const namespaces = form.value.namespaces.filter(ns => ns !== '__all__')
      let succeeded = 0
      let failed = 0
      for (const ns of namespaces) {
        try {
          await createBinding({
            userId: Number(form.value.userId),
            roleId: Number(form.value.roleId),
            clusterId: props.clusterId,
            namespace: ns,
          })
          succeeded++
        } catch {
          failed++
        }
      }
      if (succeeded > 0) {
        ElMessage.success(t('rbac.createSuccess') + (failed > 0 ? ` (${succeeded}/${succeeded + failed})` : ''))
      }
      if (failed > 0 && succeeded === 0) {
        ElMessage.error(t('common.failed'))
      }
    }
    emit('success')
    dialogVisible.value = false
  } catch (e: any) {
    ElMessage.error(e?.message || t('common.failed'))
  } finally {
    saving.value = false
  }
}
</script>

<template>
  <el-dialog
    v-model="dialogVisible"
    :title="isEditMode ? t('rbac.editBinding') : t('rbac.addUser')"
    :width="isEditMode ? '480px' : '560px'"
    :close-on-click-modal="false"
  >
    <el-form
      ref="formRef"
      :model="form"
      :rules="rules"
      label-width="80px"
      label-position="right"
    >
      <!-- 命名空间 -->
      <el-form-item v-if="!isEditMode" :label="t('rbac.namespace')" prop="namespaces">
        <el-select
          v-if="!nsLoadError"
          v-model="form.namespaces"
          filterable
          multiple
          collapse-tags
          collapse-tags-tooltip
          :loading="nsLoading"
          :placeholder="t('rbac.selectNamespace')"
          style="width: 100%;"
        >
          <el-option :label="t('rbac.all')" value="__all__" @click="handleSelectAllNs" />
          <el-option v-for="ns in nsList" :key="ns" :label="ns" :value="ns" />
        </el-select>
        <template v-else>
          <el-alert type="warning" :closable="false" style="margin-bottom: 8px;">
            {{ t('rbac.clusterOfflineNsHint') }}
          </el-alert>
        </template>
      </el-form-item>
      <el-form-item v-if="isEditMode" :label="t('rbac.namespace')">
        <el-tag type="info" size="small">{{ binding?.namespace || '—' }}</el-tag>
      </el-form-item>

      <!-- 用户 -->
      <el-form-item :label="t('rbac.username')" prop="userId">
        <el-select
          v-if="!isEditMode"
          v-model="form.userId"
          filterable
          :loading="userLoading"
          :placeholder="t('rbac.selectUser')"
          style="width: 100%;"
        >
          <el-option
            v-for="u in userOptions"
            :key="u.id"
            :label="`${u.username} (${u.display_name || '-'})`"
            :value="Number(u.id)"
          />
        </el-select>
        <el-input v-else :model-value="`${binding?.username} (${binding?.displayName || '-'})`" disabled />
      </el-form-item>

      <!-- 角色 -->
      <el-form-item :label="t('rbac.role')" prop="roleId">
        <el-select v-model="form.roleId" :placeholder="t('rbac.selectRole')" style="width: 100%;">
          <el-option
            v-for="r in filteredRoles"
            :key="r.id"
            :label="`${r.displayName} (${r.name})`"
            :value="r.id"
          />
        </el-select>
      </el-form-item>

      <!-- 角色说明 -->
      <el-alert
        v-if="selectedRole"
        type="info"
        :closable="false"
        style="margin-top: 4px;"
      >
        {{ selectedRole.displayName }}: {{ roleDescriptionMap[selectedRole.name] || '' }}
      </el-alert>

      <el-alert
        v-if="isEditMode"
        type="warning"
        :closable="false"
        style="margin-top: 12px;"
      >
        {{ t('rbac.scopeChangeWarning') }}
      </el-alert>
    </el-form>

    <template #footer>
      <el-button @click="dialogVisible = false">{{ t('common.cancel') }}</el-button>
      <el-button type="primary" :loading="saving" @click="handleSubmit">
        {{ isEditMode ? t('common.save') : t('common.confirm') }}
      </el-button>
    </template>
  </el-dialog>
</template>
