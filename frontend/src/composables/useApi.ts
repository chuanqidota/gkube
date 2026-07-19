import { ref, type Ref } from 'vue'
import { ElMessage } from 'element-plus'

export interface UseApiOptions {
  /** Show an ElMessage.error toast on failure (default: true) */
  showError?: boolean
  /** Fallback error message when the error has no message */
  errorMessage?: string
  /** Called on success with the resolved value */
  onSuccess?: (data: any) => void
  /** Called on failure with the error */
  onError?: (err: any) => void
}

/**
 * Shared async wrapper that manages loading + error state and centralizes
 * error-toast display, replacing the repeated
 *   loading.value = true; try {...} catch (e) { ElMessage.error(...) } finally {...}
 * pattern scattered across views.
 */
export function useApi<T = any>(
  fn: (...args: any[]) => Promise<T>,
  options: UseApiOptions = {}
) {
  const { showError = true, errorMessage, onSuccess, onError } = options
  const loading = ref(false)
  const error: Ref<Error | null> = ref(null)
  const data = ref<T | null>(null) as Ref<T | null>

  async function run(...args: any[]): Promise<T | undefined> {
    loading.value = true
    error.value = null
    try {
      const result = await fn(...args)
      data.value = result
      onSuccess?.(result)
      return result
    } catch (e: any) {
      error.value = e instanceof Error ? e : new Error(String(e))
      if (showError) {
        ElMessage.error(e?.message || errorMessage || '操作失败')
      }
      onError?.(e)
      return undefined
    } finally {
      loading.value = false
    }
  }

  return { loading, error, data, run }
}
