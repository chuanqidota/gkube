import { ref, onUnmounted } from 'vue'

export function useAutoRefresh(fetchFn: () => Promise<void>, interval = 15000, autoStart = true) {
  const isRunning = ref(autoStart)
  const countdown = ref(Math.floor(interval / 1000))
  let pollTimer: ReturnType<typeof setInterval> | null = null
  let countdownTimer: ReturnType<typeof setInterval> | null = null

  function startCountdown() {
    countdown.value = Math.floor(interval / 1000)
    countdownTimer = setInterval(() => {
      countdown.value--
      if (countdown.value <= 0) {
        countdown.value = Math.floor(interval / 1000)
      }
    }, 1000)
  }

  function startPolling() {
    pollTimer = setInterval(() => {
      fetchFn()
    }, interval)
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
    toggle,
    refresh,
    start,
    stop,
  }
}
