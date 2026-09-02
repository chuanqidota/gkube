<script setup lang="ts">
import { ref, computed, watch, nextTick } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage } from 'element-plus'
import type { FormInstance, FormRules } from 'element-plus'
import { createBinding, getNamespaceList, searchUsers } from '@/api/rbac'

const { t } = useI18n()

const props = defineProps<{
  visible: boolean
  clusterId: number
  clusterName: string
  binding?: any | null   // 兼容旧调用（不再使用编辑模式，编辑走 MemberEditDialog）
  roles: any[]
}>()

const emit = defineEmits<{
  (e: 'update:visible', val: boolean): void
  (e: 'success'): void
}>()

const dialogVisible = computed({
  get: () => props.visible,
  set: (val: boolean) => emit('update:visible', val),
})

const formRef = ref<FormInstance>()
const saving = ref(false)

// scope: 'cluster' | 'namespace'
const form = ref({
  userId: '' as string | number,
  scope: 'namespace' as 'cluster' | 'namespace',
  roleId: '' as string | number,
  namespaces: [] as string[],
})

const userOptions = ref<any[]>([])
const userLoading = ref(false)

const nsList = ref<string[]>([])
const nsLoading = ref(false)
const nsLoadError = ref(false)

const rules = computed<FormRules>(() => ({
  userId: [{ required: true, message: t('rbac.selectUser'), trigger: 'change' }],
  roleId: [{ required: true, message: t('rbac.selectRole'), trigger: 'change' }],
  namespaces: [{
    validator: (_rule: any, _value: any, callback: (err?: Error) => void) => {
      if (form.value.scope === 'namespace' && form.value.namespaces.length === 0) {
        callback(new Error(t('rbac.selectNamespace')))
      } else {
        callback()
      }
    },
    trigger: 'change',
  }],
}))

// 角色按作用域过滤：集群级 -> scopeType=cluster；空间级 -> scopeType=namespace
const filteredRoles = computed(() => {
  return props.roles.filter(r => r.scopeType === form.value.scope)
})

const selectedRole = computed(() => {
  return filteredRoles.value.find(r => r.id === form.value.roleId)
})

// 作用域切换时重置角色选择
watch(() => form.value.scope, () => {
  form.value.roleId = ''
  nextTick(() => formRef.value?.clearValidate('roleId'))
})

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

watch(() => props.visible, (val) => {
  if (val) {
    form.value.userId = ''
    form.value.scope = 'namespace'
    form.value.roleId = ''
    form.value.namespaces = []
    nextTick(() => formRef.value?.clearValidate())
    loadAllUsers()
  }
})

watch(() => props.visible, async (val) => {
  if (!val || !props.clusterName) return
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

async function handleSubmit() {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return

  saving.value = true
  try {
    await createBinding({
      userId: Number(form.value.userId),
      roleId: Number(form.value.roleId),
      clusterId: props.clusterId,
      // 集群级不传 namespaces（后端归一化为 [""]）；空间级传数组走批量插入
      ...(form.value.scope === 'namespace' ? { namespaces: form.value.namespaces } : {}),
    })
    ElMessage.success(t('rbac.createSuccess'))
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
  <el-dialog v-model="dialogVisible" :title="t('rbac.addUser')" width="560px" :close-on-click-modal="false">
    <el-form ref="formRef" :model="form" :rules="rules" label-width="80px" label-position="right">
      <el-form-item :label="t('rbac.username')" prop="userId">
        <el-select
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
      </el-form-item>

      <el-form-item :label="t('rbac.scope')">
        <el-radio-group v-model="form.scope">
          <el-radio value="namespace">{{ t('rbac.namespaceScope') }}</el-radio>
          <el-radio value="cluster">{{ t('rbac.clusterScope') }}</el-radio>
        </el-radio-group>
      </el-form-item>

      <el-form-item v-if="form.scope === 'namespace'" :label="t('rbac.namespace')" prop="namespaces">
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
          <el-option v-for="ns in nsList" :key="ns" :label="ns" :value="ns" />
        </el-select>
        <template v-else>
          <el-alert type="warning" :closable="false">
            {{ t('rbac.clusterOfflineNsHint') }}
          </el-alert>
        </template>
      </el-form-item>

      <el-form-item :label="t('rbac.role')" prop="roleId">
        <el-select v-model="form.roleId" :placeholder="t('rbac.selectRole')" style="width: 100%;">
          <el-option-group
            :label="form.scope === 'cluster' ? t('rbac.clusterScope') : t('rbac.namespaceScope')"
          >
            <el-option
              v-for="r in filteredRoles"
              :key="r.id"
              :label="`${r.displayName} (${r.name})`"
              :value="r.id"
            />
          </el-option-group>
        </el-select>
      </el-form-item>

      <el-alert v-if="selectedRole" type="info" :closable="false" style="margin-top: 4px;">
        {{ selectedRole.displayName }}
      </el-alert>
    </el-form>

    <template #footer>
      <el-button @click="dialogVisible = false">{{ t('common.cancel') }}</el-button>
      <el-button type="primary" :loading="saving" @click="handleSubmit">
        {{ t('common.confirm') }}
      </el-button>
    </template>
  </el-dialog>
</template>
