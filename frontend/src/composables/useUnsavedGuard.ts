import { ref, onUnmounted } from 'vue'
import { onBeforeRouteLeave } from 'vue-router'
import { ElMessageBox } from 'element-plus'

export function useUnsavedGuard() {
  const isDirty = ref(false)

  function handleBeforeUnload(e: BeforeUnloadEvent) {
    e.preventDefault()
    e.returnValue = ''
  }

  function markDirty() {
    if (!isDirty.value) {
      isDirty.value = true
      window.addEventListener('beforeunload', handleBeforeUnload)
    }
  }

  function markClean() {
    isDirty.value = false
    window.removeEventListener('beforeunload', handleBeforeUnload)
  }

  onBeforeRouteLeave((_to, _from) => {
    if (!isDirty.value) {
      return true
    }
    return ElMessageBox.confirm('有未保存的更改，确定离开吗？', '提示', {
      type: 'warning',
      confirmButtonText: '离开',
      cancelButtonText: '取消',
    })
      .then(() => true)
      .catch(() => false)
  })

  onUnmounted(() => {
    window.removeEventListener('beforeunload', handleBeforeUnload)
  })

  return { isDirty, markDirty, markClean }
}
