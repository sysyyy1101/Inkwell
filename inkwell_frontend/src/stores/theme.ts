import { ref, watch } from 'vue'
import { defineStore } from 'pinia'

export type ThemeMode = 'light' | 'dark'

const STORAGE_KEY = 'inkwell.theme'

function readTheme(): ThemeMode | null {
  try {
    const saved = localStorage.getItem(STORAGE_KEY)
    return saved === 'light' || saved === 'dark' ? saved : null
  } catch {
    return null
  }
}

function systemTheme(): ThemeMode {
  return window.matchMedia?.('(prefers-color-scheme: dark)').matches ? 'dark' : 'light'
}

function applyTheme(mode: ThemeMode): void {
  document.documentElement.dataset.theme = mode
  const meta = document.querySelector('meta[name="theme-color"]')
  meta?.setAttribute('content', mode === 'dark' ? '#0b0b0f' : '#f5f5f7')
}

/** 主题: 首次进入跟随系统, 用户手动切换后记住选择 */
export const useThemeStore = defineStore('theme', () => {
  const stored = readTheme()
  const mode = ref<ThemeMode>(stored ?? systemTheme())
  const followedSystem = ref(stored === null)

  applyTheme(mode.value)

  watch(mode, (next) => {
    applyTheme(next)
    try {
      localStorage.setItem(STORAGE_KEY, next)
    } catch {
      // 忽略存储失败
    }
    followedSystem.value = false
  })

  function toggle(): void {
    mode.value = mode.value === 'dark' ? 'light' : 'dark'
  }

  return {
    mode,
    followedSystem,
    toggle,
  }
})
