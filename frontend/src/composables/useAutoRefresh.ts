import { ref, onUnmounted } from 'vue'

export interface AutoRefreshOptions {
  /** Default interval in ms (default: 15000) */
  interval?: number
  /** Auto-start on creation (default: true) */
  autoStart?: boolean
  /** Available interval options in seconds */
  intervalOptions?: number[]
}

const DEFAULT_INTERVAL_OPTIONS = [5, 10, 15, 30, 60]

export function useAutoRefresh(fetchFn: () => Promise<void>, options: AutoRefreshOptions = {}) {
  const {
    interval = 15000,
    autoStart = false,
    intervalOptions = DEFAULT_INTERVAL_OPTIONS,
  } = typeof options === 'number' ? { interval: options } : options

  const isRunning = ref(autoStart)
  const currentInterval = ref(interval)
  const countdown = ref(Math.floor(interval / 1000))
  const availableIntervals = intervalOptions

  let pollTimer: ReturnType<typeof setInterval> | null = null
  let countdownTimer: ReturnType<typeof setInterval> | null = null

  function startCountdown() {
    countdown.value = Math.floor(currentInterval.value / 1000)
    if (countdownTimer) clearInterval(countdownTimer)
    countdownTimer = setInterval(() => {
      countdown.value--
      if (countdown.value <= 0) {
        countdown.value = Math.floor(currentInterval.value / 1000)
      }
    }, 1000)
  }

  function startPolling() {
    if (pollTimer) clearInterval(pollTimer)
    pollTimer = setInterval(() => {
      fetchFn()
    }, currentInterval.value)
  }

  function start() {
    isRunning.value = true
    startCountdown()
    startPolling()
  }

  function stop() {
    isRunning.value = false
    if (pollTimer) {
      clearInterval(pollTimer)
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

  function refresh() {
    fetchFn()
    if (isRunning.value) {
      stop()
      start()
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
