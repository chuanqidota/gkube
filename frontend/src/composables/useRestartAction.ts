import { ElMessageBox, ElMessage } from 'element-plus'

export function useRestartAction(
  kind: 'deployment' | 'statefulset' | 'daemonset',
  restartApi: (params: { namespace: string; name: string }) => Promise<any>
) {
  const kindLabel = { deployment: 'Deployment', statefulset: 'StatefulSet', daemonset: 'DaemonSet' }[kind]

  const handleRestart = async (
    namespace: string,
    name: string,
    onSuccess?: () => void
  ) => {
    try {
      await ElMessageBox.confirm(
        `确定重启 ${kindLabel} "${name}"？这将触发所有 Pod 滚动更新。`,
        '重启确认',
        { confirmButtonText: '确定', cancelButtonText: '取消', type: 'warning' }
      )
      await restartApi({ namespace, name })
      ElMessage.success('重启指令已发送')
      onSuccess?.()
    } catch (e: any) {
      if (e !== 'cancel') {
        ElMessage.error(e?.response?.data?.message || '重启失败')
      }
    }
  }

  return { handleRestart }
}
