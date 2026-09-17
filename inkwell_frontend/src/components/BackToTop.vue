<template>
  <Transition name="fade">
    <button
      v-if="visible"
      class="to-top"
      type="button"
      aria-label="回到顶部"
      title="回到顶部"
      @click="scrollToTop"
    >
      <AppIcon name="arrowUp" :size="16" />
    </button>
  </Transition>
</template>

<script setup lang="ts">
import { onBeforeUnmount, onMounted, ref } from 'vue'

import AppIcon from '@/components/AppIcon.vue'

const visible = ref(false)

function handleScroll(): void {
  visible.value = window.scrollY > 480
}

function scrollToTop(): void {
  window.scrollTo({ top: 0, behavior: 'smooth' })
}

onMounted(() => {
  window.addEventListener('scroll', handleScroll, { passive: true })
  handleScroll()
})

onBeforeUnmount(() => {
  window.removeEventListener('scroll', handleScroll)
})
</script>

<style scoped lang="scss">
.to-top {
  position: fixed;
  right: 24px;
  bottom: 28px;
  z-index: 25;
  display: inline-flex;
  width: 44px;
  height: 44px;
  align-items: center;
  justify-content: center;
  border: 1px solid var(--border);
  border-radius: var(--radius-pill);
  background: var(--surface);
  color: var(--text-secondary);
  box-shadow: var(--shadow-sm);
  transition: color var(--transition), transform var(--transition);

  &:hover {
    color: var(--text);
    transform: translateY(-2px);
  }

  &:active {
    transform: translateY(0) scale(0.94);
  }
}
</style>
