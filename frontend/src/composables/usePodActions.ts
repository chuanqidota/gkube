import { type Ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessageBox, ElMessage } from 'element-plus'
import { deletePod } from '@/api/resource'

export function usePodActions(clusterName: Ref<string>) {
  const router = useRouter()

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
          ? `确定强制删除 Pod "${pod.name}"？这将跳过优雅终止。`
          : `确定删除 Pod "${pod.name}"？`,
        force ? '强制删除 Pod' : '删除 Pod',
        { confirmButtonText: '确定', cancelButtonText: '取消', type: 'warning' }
      )
      await deletePod({ namespace: pod.namespace, name: pod.name, force })
      ElMessage.success('删除成功')
      onSuccess?.()
    } catch (e: any) {
      if (e !== 'cancel') {
        ElMessage.error(e?.response?.data?.message || '删除失败')
      }
    }
  }

  return { handlePodLogs, handlePodExec, handlePodDelete }
}
