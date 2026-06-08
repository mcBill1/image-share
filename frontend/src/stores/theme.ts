import { defineStore } from 'pinia'
import { ref, watch } from 'vue'

export type ThemeMode = 'system' | 'light' | 'dark'

export const useThemeStore = defineStore('theme', () => {
  const mode = ref<ThemeMode>((localStorage.getItem('theme') as ThemeMode) || 'system')

  function getSystemTheme(): 'light' | 'dark' {
    return window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light'
  }

  function getEffectiveTheme(): 'light' | 'dark' {
    if (mode.value === 'system') return getSystemTheme()
    return mode.value
  }

  function applyTheme() {
    const effective = getEffectiveTheme()
    const html = document.documentElement
    html.classList.remove('light', 'dark')
    html.classList.add(effective)
    html.setAttribute('data-theme', effective)
  }

  function setTheme(newMode: ThemeMode) {
    mode.value = newMode
    localStorage.setItem('theme', newMode)
    applyTheme()
  }

  function initTheme() {
    applyTheme()
    window.matchMedia('(prefers-color-scheme: dark)').addEventListener('change', () => {
      if (mode.value === 'system') applyTheme()
    })
  }

  watch(mode, () => applyTheme())

  return { mode, setTheme, initTheme, getEffectiveTheme }
})
