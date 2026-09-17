<template>
  <div class="container">
    <div class="not-found">
      <p class="not-found__code font-display">404</p>
      <h1 class="not-found__title font-display">这个页面不存在</h1>
      <p class="not-found__desc">
        你访问的地址没有对应的页面。可以回到首页看看最新的帖子, 或者直接去某个版块逛逛。
      </p>
      <div class="not-found__actions">
        <RouterLink class="btn btn--primary" :to="{ name: 'home' }">
          <AppIcon name="home" :size="16" />
          回到首页
        </RouterLink>
        <RouterLink class="btn btn--secondary" :to="{ name: 'community', params: { id: 1 } }">
          <AppIcon name="compass" :size="16" />
          逛逛版块
        </RouterLink>
      </div>
      <ServiceStatus />
    </div>
  </div>
</template>

<script setup lang="ts">
import AppIcon from '@/components/AppIcon.vue'
import ServiceStatus from '@/components/ServiceStatus.vue'
</script>

<style scoped lang="scss">
.not-found {
  display: flex;
  max-width: 520px;
  flex-direction: column;
  align-items: center;
  gap: var(--space-3);
  margin: var(--space-8) auto 0;
  padding: var(--space-8) var(--space-6);
  border: 1px solid var(--border);
  border-radius: var(--radius-xl);
  background: var(--surface);
  box-shadow: var(--shadow-sm);
  text-align: center;
}

.not-found__code {
  color: var(--text);
  font-size: 60px;
  font-weight: var(--weight-black);
  letter-spacing: var(--tracking-tighter);
  line-height: 1;
}

.not-found__title {
  font-size: var(--text-xl);
  font-weight: var(--weight-heavy);
  letter-spacing: -0.03em;
}

.not-found__desc {
  color: var(--text-secondary);
  font-size: var(--text-base);
}

.not-found__actions {
  display: flex;
  flex-wrap: wrap;
  justify-content: center;
  gap: var(--space-3);
  margin: var(--space-3) 0;
}
</style>
