<script setup lang="ts">
import { ref, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage, ElMessageBox } from 'element-plus'
import { updateBindingBatch, deleteBinding } from '@/api/rbac'

interface MemberBinding {
  bindingId: number
  roleId: number
  roleName: string
  roleDisplayName: string
  scopeType: 'cluster' | 'namespace'
  namespace: string
}
interface Member {
  userId: number
  username: string
  displayName: string
  isSuperAdmin: boolean
  bindings: MemberBinding[]
}

const props = defineProps<{
  visible: boolean
  clusterId: number
  member: Member
  roles: any[]
}>()

const emit = defineEmits<{
  'update:visible': [val: boolean]
  success: []
}>()

const { t } = useI18n()
const saving = ref(false)

const dialogVisible = computed({
  get: () => props.visible,
  set: (val: boolean) => emit('update:visible', val),
})

// 每条授权可选的角色：与其 scopeType 一致的角色
function availableRoles(scopeType: string) {
  return props.roles.filter((r) => r.scopeType === scopeType)
}

async function handleChangeRole(b: MemberBinding, newRoleId: number) {
  if (newRoleId === b.roleId) return
  saving.value = true
  try {
    await updateBindingBatch({
      userId: props.member.userId,
      clusterId: props.clusterId,
      fromRoleId: b.roleId,
      toRoleId: newRoleId,
    })
    ElMessage.success(t('rbac.updateSuccess'))
    emit('success')
    dialogVisible.value = false
  } catch (e: any) {
    ElMessage.error(e?.message || t('common.failed'))
  } finally {
    saving.value = false
  }
}

async function handleRemoveBinding(b: MemberBinding) {
  try {
    await ElMessageBox.confirm(
      t('rbac.removeRoleConfirm', {
        name: props.member.username,
        role:
          b.scopeType === 'cluster' ? b.roleDisplayName : `${b.roleDisplayName} (${b.namespace})`,
      }),
      t('common.confirm'),
      { type: 'warning' },
    )
  } catch {
    return
  }
  saving.value = true
  try {
    await deleteBinding(b.bindingId)
    ElMessage.success(t('rbac.removeRoleSuccess'))
    emit('success')
    dialogVisible.value = false
  } catch (e: any) {
    ElMessage.error(e?.message || t('common.deleteFailed'))
  } finally {
    saving.value = false
  }
}
</script>

<template>
  <el-dialog
    v-model="dialogVisible"
    :title="t('rbac.editMember', { name: member.username })"
    width="640px"
  >
    <div v-loading="saving" class="binding-list">
      <div v-for="b in member.bindings" :key="b.bindingId" class="binding-row">
        <div class="scope-cell">
          <el-tag v-if="b.scopeType === 'cluster'" size="small" type="info">{{
            t('rbac.clusterScope')
          }}</el-tag>
          <el-tag v-else size="small" type="warning">{{ b.namespace }}</el-tag>
        </div>
        <el-select
          :model-value="b.roleId"
          size="default"
          style="width: 220px"
          :disabled="saving"
          @change="(v: any) => handleChangeRole(b, Number(v))"
        >
          <el-option
            v-for="r in availableRoles(b.scopeType)"
            :key="r.id"
            :label="r.displayName"
            :value="r.id"
          />
        </el-select>
        <el-button size="small" type="danger" @click="handleRemoveBinding(b)">
          {{ t('rbac.removeBinding') }}
        </el-button>
      </div>
      <el-empty v-if="member.bindings.length === 0" :description="t('rbac.noRole')" />
    </div>
    <el-alert type="info" :closable="false" style="margin-top: 12px">
      {{ t('rbac.scopeChangeWarning') }}
    </el-alert>
    <template #footer>
      <el-button @click="dialogVisible = false">{{ t('common.close') }}</el-button>
    </template>
  </el-dialog>
</template>

<style scoped>
.binding-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}
.binding-row {
  display: flex;
  align-items: center;
  gap: 12px;
  flex-wrap: wrap;
}
.scope-cell {
  width: 140px;
}
</style>
