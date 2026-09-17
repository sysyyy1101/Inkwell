<template>
  <div class="load-more">
    <button
      v-if="hasMore"
      class="btn btn--secondary"
      type="button"
      :disabled="loading"
      @click="emit('load-more')"
    >
      <span v-if="loading" class="load-more__spinner" aria-hidden="true" />
      {{ loading ? '加载中…' : '加载更多' }}
    </button>
    <p v-else class="load-more__end">已经到底啦</p>
  </div>
</template>

<script setup lang="ts">
defineProps<{
  hasMore: boolean
  loading: boolean
}>()

const emit = defineEmits<{
  (event: 'load-more'): void
}>()
</script>

<style scoped lang="scss">
.load-more {
  display: flex;
  justify-content: center;
  padding: var(--space-2) 0;
}

.load-more__end {
  color: var(--text-tertiary);
  font-size: var(--text-sm);
}

.load-more__spinner {
  width: 14px;
  height: 14px;
  border: 2px solid var(--border-strong);
  border-top-color: var(--accent);
  border-radius: var(--radius-pill);
  animation: spin 0.7s linear infinite;
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}
</style>
