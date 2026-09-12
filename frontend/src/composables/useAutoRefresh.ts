import { ref, onUnmounted } from 'vue'

export interface AutoRefreshOptions {
  /** Default interval in ms (default: 15000) */
  interval?: number
  /** Auto-start on creation (default: true) */
  autoStart?: boolean
  /** Available interval options in seconds */
  intervalOptions?: number[]
  /**
   * 手动刷新时调用的 fetch 函数。默认回退到 fetchFn。
   * 自动刷新通常走 silent（不显示遮罩），手动刷新需显示遮罩——
   * 传入此参数可解耦两者，避免手动刷新也静默。
   */
  manualFetch?: () => Promise<void>
}

const DEFAULT_INTERVAL_OPTIONS = [5, 10, 15, 30, 60]
const MAX_BACKOFF = 60_000

export function useAutoRefresh(fetchFn: () => Promise<void>, options: AutoRefreshOptions = {}) {
  const {
    interval = 15000,
    autoStart = false,
    intervalOptions = DEFAULT_INTERVAL_OPTIONS,
    manualFetch,
  } = typeof options === 'number' ? { interval: options } : options

  const isRunning = ref(autoStart)
  const currentInterval = ref(interval)
  const countdown = ref(Math.floor(interval / 1000))
  const availableIntervals = intervalOptions

  let pollTimer: ReturnType<typeof setTimeout> | null = null
  let countdownTimer: ReturnType<typeof setInterval> | null = null
  let consecutiveFailures = 0

  function getEffectiveDelay() {
    if (consecutiveFailures >= 3) {
      return Math.min(currentInterval.value * Math.pow(2, consecutiveFailures - 3), MAX_BACKOFF)
    }
    return currentInterval.value
  }

  function startCountdown(delayMs?: number) {
    const effective = delayMs ?? getEffectiveDelay()
    countdown.value = Math.floor(effective / 1000)
    if (countdownTimer) clearInterval(countdownTimer)
    countdownTimer = setInterval(() => {
      countdown.value--
      if (countdown.value <= 0) {
        countdown.value = Math.floor(getEffectiveDelay() / 1000)
      }
    }, 1000)
  }

  async function pollingTick() {
    try {
      await fetchFn()
      consecutiveFailures = 0
    } catch (e) {
      consecutiveFailures++
      console.warn(`[useAutoRefresh] Poll failed (${consecutiveFailures}x):`, e)
    } finally {
      if (isRunning.value) {
        const delay = getEffectiveDelay()
        pollTimer = setTimeout(pollingTick, delay)
        startCountdown(delay)
      }
    }
  }

  function start() {
    isRunning.value = true
    consecutiveFailures = 0
    pollingTick()
  }

  function stop() {
    isRunning.value = false
    if (pollTimer) {
      clearTimeout(pollTimer)
      pollTimer = null
    }
    if (countdownTimer) {
      clearInterval(countdownTimer)
      countdownTimer = null
    }
  }

  function toggle() {
    if (isRunning.value) {
      stop()
    } else {
      start()
    }
  }

  async function refresh() {
    // 手动刷新优先用 manualFetch（可显示遮罩），否则回退到 fetchFn
    const fn = manualFetch || fetchFn
    // 停止当前轮询，避免与手动刷新并发
    if (isRunning.value && pollTimer) {
      clearTimeout(pollTimer)
      pollTimer = null
      if (countdownTimer) {
        clearInterval(countdownTimer)
        countdownTimer = null
      }
    }
    try {
      await fn()
      consecutiveFailures = 0
    } catch (e) {
      consecutiveFailures++
      console.warn('[useAutoRefresh] Manual refresh failed:', e)
    } finally {
      // 手动刷新完成后，恢复轮询计时器
      if (isRunning.value) {
        const delay = getEffectiveDelay()
        pollTimer = setTimeout(pollingTick, delay)
        startCountdown(delay)
      }
    }
  }

  function setIntervalOption(seconds: number) {
    currentInterval.value = seconds * 1000
    if (isRunning.value) {
      stop()
      start()
    }
  }

  // Auto-start if enabled
  if (autoStart) {
    start()
  }

  // Cleanup on unmount
  onUnmounted(() => {
    stop()
  })

  return {
    isRunning,
    countdown,
    currentInterval,
    availableIntervals,
    toggle,
    refresh,
    start,
    stop,
    setIntervalOption,
  }
}
