import { ElMessage, ElMessageBox } from 'element-plus'
import { useI18n } from 'vue-i18n'
import { cordonNode, deleteNode } from '@/api/resource'

/**
 * 节点通用操作：封锁/解除封锁、删除。
 * 在 NodeList 与 NodeDetail 间共享，避免重复的确认弹窗 + API 调用样板。
 *
 * @param onChanged 操作成功后的刷新回调（删除可由 after 覆盖，如跳转列表）
 */
export function useNodeActions(onChanged: () => void) {
  const { t } = useI18n()

  async function handleCordon(name: string, unschedulable: boolean) {
    const actionLabel = unschedulable ? t('node.uncordon') : t('node.cordon')
    try {
      await ElMessageBox.confirm(t('node.cordonConfirm', { action: actionLabel, name }), t('common.confirm'), { type: 'warning' })
      await cordonNode({ name, cordon: !unschedulable })
      ElMessage.success(t('node.cordonSuccess', { action: actionLabel }))
      onChanged()
    } catch (e: any) {
      if (e !== 'cancel') ElMessage.error(e?.message || t('node.cordonFailed', { action: actionLabel }))
    }
  }

  /**
   * 删除节点（清理 etcd 里的 Node 对象残留）。
   * 仅适用于已永久下线的节点；若节点仍在线（kubelet 运行中），删除后会自动重新注册。
   *
   * @param name 节点名
   * @param ready 节点是否就绪。true=在线，确认框会额外警告"删除后会重新注册"
   * @param after 删除成功后的回调（如跳转列表），未传则调 onChanged
   */
  async function handleDelete(name: string, ready: boolean, after?: () => void) {
    const baseMsg = t('node.deleteOfflineConfirm', { name })
    const onlineMsg = t('node.deleteOnlineWarning', { name })
    try {
      await ElMessageBox.confirm(
        ready ? onlineMsg : baseMsg,
        t('common.confirmDelete'),
        { type: 'error', confirmButtonText: t('common.delete'), cancelButtonText: t('common.cancel') },
      )
      await deleteNode({ name })
      ElMessage.success(t('common.deleteSuccess', { type: 'Node' }))
      after ? after() : onChanged()
    } catch (e: any) {
      if (e !== 'cancel') ElMessage.error(e?.message || t('common.deleteFailed', { type: 'Node' }))
    }
  }

  return { handleCordon, handleDelete }
}
