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

  /**
   * 删除节点（清理 etcd 里的 Node 对象残留）。
   * 仅适用于已永久下线的节点；若节点仍在线（kubelet 运行中），删除后会自动重新注册。
   *
   * @param name 节点名
   * @param ready 节点是否就绪。true=在线，确认框会额外警告"删除后会重新注册"
   * @param after 删除成功后的回调（如跳转列表），未传则调 onChanged
   */
  async function handleDelete(name: string, ready: boolean, after?: () => void) {
    const baseMsg = `此操作用于清理已下线节点的残留记录。若节点仍在线（kubelet 运行中），删除后会自动重新注册。确定要删除节点 "${name}" 吗？`
    const onlineMsg = `节点 "${name}" 当前在线（Ready）。删除后 kubelet 会重新注册该节点，删除将无效。如确需删除，请先停止节点上的 kubelet。仍要继续吗？`
    try {
      await ElMessageBox.confirm(
        ready ? onlineMsg : baseMsg,
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
