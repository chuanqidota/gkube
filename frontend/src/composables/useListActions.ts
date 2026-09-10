import { ref } from 'vue'
import { ElMessageBox, ElMessage } from 'element-plus'
import { useI18n } from 'vue-i18n'

interface ListActionsOptions {
  kind: 'deployment' | 'statefulset' | 'daemonset'
  scaleApi?: (params: { namespace: string; name: string; replicas: number }) => Promise<any>
  restartApi: (params: { namespace: string; name: string }) => Promise<any>
  updateImageApi: (params: { namespace: string; name: string; containerName: string; image: string }) => Promise<any>
  detailApi: (params: { namespace: string; name: string }) => Promise<any>
  onActionSuccess?: () => void
}

export function useListActions(options: ListActionsOptions) {
  const { t } = useI18n()
  const { kind, scaleApi, restartApi, updateImageApi, detailApi, onActionSuccess } = options

  // Scale dialog (only for deployment/statefulset)
  const scaleDialogVisible = ref(false)
  const scaleTarget = ref<{ namespace: string; name: string } | null>(null)
  const scaleReplicas = ref<number>(1)
  const scaleLoading = ref(false)

  const canScale = kind !== 'daemonset'

  function openScaleDialog(row: any) {
    if (!canScale) return
    scaleTarget.value = { namespace: row.namespace, name: row.name }
    scaleReplicas.value = row.replicas ?? row.ready_replicas ?? 1
    scaleDialogVisible.value = true
  }

  async function handleScaleConfirm() {
    if (!scaleTarget.value || !scaleApi) return
    scaleLoading.value = true
    try {
      await scaleApi({ ...scaleTarget.value, replicas: scaleReplicas.value })
      ElMessage.success(t('workload.scaledToReplicas', { name: scaleTarget.value.name, n: scaleReplicas.value }))
      scaleDialogVisible.value = false
      onActionSuccess?.()
    } catch (e: any) {
      ElMessage.error(e?.message || t('common.scaleFailed'))
    } finally {
      scaleLoading.value = false
    }
  }

  // Restart
  async function handleRestart(row: any) {
    const kindLabel = { deployment: 'Deployment', statefulset: 'StatefulSet', daemonset: 'DaemonSet' }[kind]
    try {
      await ElMessageBox.confirm(
        t('workload.restartConfirm', { kind: kindLabel, name: row.name }),
        t('workload.restartConfirmTitle'),
        { type: 'warning', confirmButtonText: t('common.confirm'), cancelButtonText: t('common.cancel') }
      )
      await restartApi({ namespace: row.namespace, name: row.name })
      ElMessage.success(t('workload.restartSuccessMsg', { name: row.name }))
      onActionSuccess?.()
    } catch (e: any) {
      if (e !== 'cancel') ElMessage.error(e?.message || t('workload.restartFailedMsg'))
    }
  }

  // Update image dialog
  const imageDialogVisible = ref(false)
  const imageTarget = ref<{ namespace: string; name: string } | null>(null)
  const imageForm = ref({ containerName: '', image: '' })
  const imageContainers = ref<{ name: string; image: string }[]>([])
  const imageLoading = ref(false)

  async function openImageDialog(row: any) {
    imageTarget.value = { namespace: row.namespace, name: row.name }
    imageForm.value = { containerName: '', image: '' }
    imageContainers.value = []
    imageDialogVisible.value = true
    try {
      const res: any = await detailApi({ namespace: row.namespace, name: row.name })
      const containers = res?.data?.spec?.template?.spec?.containers || res?.spec?.template?.spec?.containers || []
      imageContainers.value = containers.map((c: any) => ({ name: c.name, image: c.image || '' }))
      if (imageContainers.value.length > 0) {
        imageForm.value.containerName = imageContainers.value[0].name
        imageForm.value.image = imageContainers.value[0].image
      }
    } catch (e: any) {
      ElMessage.error(e?.message || t('workload.fetchContainerFailed'))
    }
  }

  async function handleImageConfirm() {
    if (!imageTarget.value || !imageForm.value.containerName || !imageForm.value.image) {
      ElMessage.warning(t('workload.selectContainerAndImage'))
      return
    }
    imageLoading.value = true
    try {
      await updateImageApi({ ...imageTarget.value, ...imageForm.value })
      ElMessage.success(t('workload.imageUpdateSuccess'))
      imageDialogVisible.value = false
      onActionSuccess?.()
    } catch (e: any) {
      ElMessage.error(e?.message || t('workload.imageUpdateFailed'))
    } finally {
      imageLoading.value = false
    }
  }

  return {
    // Scale
    canScale,
    scaleDialogVisible,
    scaleTarget,
    scaleReplicas,
    scaleLoading,
    openScaleDialog,
    handleScaleConfirm,
    // Restart
    handleRestart,
    // Image
    imageDialogVisible,
    imageTarget,
    imageForm,
    imageContainers,
    imageLoading,
    openImageDialog,
    handleImageConfirm,
  }
}
