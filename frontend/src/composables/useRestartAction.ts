import { ElMessageBox, ElMessage } from 'element-plus'
import { useI18n } from 'vue-i18n'

export function useRestartAction(
  kind: 'deployment' | 'statefulset' | 'daemonset',
  restartApi: (params: { namespace: string; name: string }) => Promise<any>
) {
  const { t } = useI18n()
  const kindLabel = { deployment: 'Deployment', statefulset: 'StatefulSet', daemonset: 'DaemonSet' }[kind]

  const handleRestart = async (
    namespace: string,
    name: string,
    onSuccess?: () => void
  ) => {
    try {
      await ElMessageBox.confirm(
        t('workload.restartConfirm', { kind: kindLabel, name }),
        t('workload.restartConfirmTitle'),
        { confirmButtonText: t('common.confirm'), cancelButtonText: t('common.cancel'), type: 'warning' }
      )
      await restartApi({ namespace, name })
      ElMessage.success(t('common.restartSuccess'))
      onSuccess?.()
    } catch (e: any) {
      if (e !== 'cancel') {
        ElMessage.error(e?.response?.data?.message || t('common.restartFailed'))
      }
    }
  }

  return { handleRestart }
}
