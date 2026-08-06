import { ElMessage, ElMessageBox } from 'element-plus'
import { cordonNode, deleteNode } from '@/api/resource'

/**
 * 节点通用操作：封锁/解除封锁、删除。
 * 在 NodeList 与 NodeDetail 间共享，避免重复的确认弹窗 + API 调用样板。
 *
 * @param onChanged 操作成功后的刷新回调（删除可由 after 覆盖，如跳转列表）
 */
export function useNodeActions(onChanged: () => void) {
  async function handleCordon(name: string, unschedulable: boolean) {
    const actionLabel = unschedulable ? '解除封锁' : '封锁'
    try {
      await ElMessageBox.confirm(`确定要${actionLabel}节点 "${name}" 吗？`, '确认操作', { type: 'warning' })
      await cordonNode({ name, cordon: !unschedulable })
      ElMessage.success(`节点已${actionLabel}`)
      onChanged()
    } catch (e: any) {
      if (e !== 'cancel') ElMessage.error(e?.message || `${actionLabel}失败`)
    }
  }

  async function handleDelete(name: string, after?: () => void) {
    try {
      await ElMessageBox.confirm(
        `确定要删除节点 "${name}" 吗？此操作不可恢复，节点将从集群中移除。`,
        '确认删除',
        { type: 'error', confirmButtonText: '删除', cancelButtonText: '取消' },
      )
      await deleteNode({ name })
      ElMessage.success('节点已删除')
      after ? after() : onChanged()
    } catch (e: any) {
      if (e !== 'cancel') ElMessage.error(e?.message || '删除失败')
    }
  }

  return { handleCordon, handleDelete }
}
