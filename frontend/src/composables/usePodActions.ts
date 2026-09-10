import { type Ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessageBox, ElMessage } from 'element-plus'
import { useI18n } from 'vue-i18n'
import { deletePod } from '@/api/resource'

export function usePodActions(clusterName: Ref<string>) {
  const router = useRouter()
  const { t } = useI18n()

  const handlePodLogs = (pod: { namespace: string; name: string }) => {
    const route = router.resolve({
      name: 'fullscreen-logs',
      query: { cluster: clusterName.value, namespace: pod.namespace, pod: pod.name }
    })
    window.open(route.href, '_blank')
  }

  const handlePodExec = (pod: { namespace: string; name: string }) => {
    const route = router.resolve({
      name: 'fullscreen-terminal',
      query: { cluster: clusterName.value, namespace: pod.namespace, pod: pod.name }
    })
    window.open(route.href, '_blank')
  }

  const handlePodDelete = async (
    pod: { namespace: string; name: string },
    onSuccess?: () => void,
    force = false
  ) => {
    try {
      await ElMessageBox.confirm(
        force
          ? t('workload.forceDeletePodConfirm', { name: pod.name })
          : t('workload.deletePodConfirm', { name: pod.name }),
        force ? t('workload.forceDeletePod') : t('workload.deletePod'),
        { confirmButtonText: t('common.confirm'), cancelButtonText: t('common.cancel'), type: 'warning' }
      )
      await deletePod({ namespace: pod.namespace, name: pod.name, force })
      ElMessage.success(t('common.deleteSuccess', { type: 'Pod' }))
      onSuccess?.()
    } catch (e: any) {
      if (e !== 'cancel') {
        ElMessage.error(e?.response?.data?.message || t('common.deleteFailed', { type: 'Pod' }))
      }
    }
  }

  return { handlePodLogs, handlePodExec, handlePodDelete }
}
