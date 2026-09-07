import { ref } from 'vue'

export function useEditDrawer(fetchDetail: () => Promise<void>) {
  const editDialogVisible = ref(false)
  const editFullscreen = ref(false)

  const handleEdit = () => { editDialogVisible.value = true }
  const handleEditSuccess = () => { editDialogVisible.value = false; fetchDetail() }
  const handleEditCancel = () => { editDialogVisible.value = false }

  return { editDialogVisible, editFullscreen, handleEdit, handleEditSuccess, handleEditCancel }
}
