<template>
  <button
    class="icon-button theme-toggle"
    type="button"
    :aria-label="isDark ? '切换到浅色模式' : '切换到深色模式'"
    :title="isDark ? '切换到浅色模式' : '切换到深色模式'"
    @click="theme.toggle()"
  >
    <Transition name="icon-swap" mode="out-in">
      <AppIcon :key="isDark ? 'sun' : 'moon'" :name="isDark ? 'sun' : 'moon'" :size="18" />
    </Transition>
  </button>
</template>

<script setup lang="ts">
import { computed } from 'vue'

import AppIcon from '@/components/AppIcon.vue'
import { useThemeStore } from '@/stores/theme'

const theme = useThemeStore()
const isDark = computed(() => theme.mode === 'dark')
</script>

<style scoped lang="scss">
.theme-toggle .icon {
  transition: transform 420ms cubic-bezier(0.34, 1.4, 0.64, 1);
}

.theme-toggle:hover .icon {
  transform: rotate(-18deg);
}

/* 日/月图标切换时做一个旋转交叉淡入, 而不是硬切 */
.icon-swap-enter-active,
.icon-swap-leave-active {
  transition: opacity 180ms ease, transform 240ms cubic-bezier(0.34, 1.4, 0.64, 1);
}

.icon-swap-enter-from {
  opacity: 0;
  transform: rotate(-70deg) scale(0.6);
}

.icon-swap-leave-to {
  opacity: 0;
  transform: rotate(70deg) scale(0.6);
}
</style>
