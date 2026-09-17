<template>
  <div class="app-shell">
    <div class="app-bg" aria-hidden="true" />
    <AppHeader />

    <main class="app-main">
      <RouterView v-slot="{ Component }">
        <Transition name="fade" mode="out-in">
          <component :is="Component" />
        </Transition>
      </RouterView>
    </main>

    <AppFooter />
    <BackToTop />
    <ToastHost />
  </div>
</template>

<script setup lang="ts">
import { RouterView } from 'vue-router'

import AppFooter from '@/components/AppFooter.vue'
import AppHeader from '@/components/AppHeader.vue'
import BackToTop from '@/components/BackToTop.vue'
import ToastHost from '@/components/ToastHost.vue'
</script>

<style lang="scss" scoped>
.app-shell {
  display: flex;
  min-height: 100vh;
  flex-direction: column;
}

.app-main {
  flex: 1;
  padding: var(--space-7) 0 var(--space-9);

  @media (min-width: 768px) {
    padding: var(--space-8) 0 var(--space-10);
  }
}
</style>
