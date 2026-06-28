import { ref, watchEffect } from 'vue'

export type Theme = 'light' | 'dark'

const THEME_KEY = 'gk-theme'

/**
 * Get the stored theme or default to 'light'
 */
export function getTheme(): Theme {
  const stored = localStorage.getItem(THEME_KEY)
  if (stored === 'light' || stored === 'dark') return stored
  // Respect system preference
  if (window.matchMedia('(prefers-color-scheme: dark)').matches) return 'dark'
  return 'light'
}

/**
 * Apply theme to document and persist to localStorage
 */
export function setTheme(theme: Theme): void {
  document.documentElement.setAttribute('data-theme', theme)
  localStorage.setItem(THEME_KEY, theme)
}

/**
 * Initialize theme on app startup (call once in main.ts)
 */
export function initTheme(): void {
  setTheme(getTheme())
}

/**
 * Composable for reactive theme state
 */
export function useTheme() {
  const currentTheme = ref<Theme>(getTheme())

  function toggle() {
    currentTheme.value = currentTheme.value === 'light' ? 'dark' : 'light'
  }

  function set(theme: Theme) {
    currentTheme.value = theme
  }

  // Sync to DOM and localStorage whenever it changes
  watchEffect(() => {
    setTheme(currentTheme.value)
  })

  const isDark = ref(currentTheme.value === 'dark')

  watchEffect(() => {
    isDark.value = currentTheme.value === 'dark'
  })

  return {
    currentTheme,
    isDark,
    toggle,
    set,
  }
}
