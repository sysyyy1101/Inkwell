<template>
  <div class="skeleton-list">
    <div v-for="index in count" :key="index" class="skeleton-item card">
      <span class="skeleton skeleton-item__vote" />
      <div class="skeleton-item__main">
        <span class="skeleton skeleton-item__meta" />
        <span class="skeleton skeleton-item__title" />
        <span class="skeleton skeleton-item__line" />
        <span class="skeleton skeleton-item__line skeleton-item__line--short" />
        <div class="skeleton-item__footer">
          <span class="skeleton skeleton-item__author" />
          <span class="skeleton skeleton-item__stats" />
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
withDefaults(defineProps<{ count?: number }>(), { count: 4 })
</script>

<style scoped lang="scss">
.skeleton-list {
  display: flex;
  flex-direction: column;
  gap: var(--space-4);
}

.skeleton-item {
  display: flex;
  gap: var(--space-4);
  padding: var(--space-5);
}

.skeleton-item__vote {
  width: 34px;
  height: 92px;
  border-radius: var(--radius-md);
}

.skeleton-item__main {
  display: flex;
  flex: 1;
  flex-direction: column;
  gap: 10px;
}

.skeleton-item__meta {
  width: 140px;
  height: 20px;
}

.skeleton-item__title {
  width: 62%;
  height: 22px;
}

.skeleton-item__line {
  width: 100%;
  height: 14px;
}

.skeleton-item__line--short {
  width: 72%;
}

.skeleton-item__footer {
  display: flex;
  justify-content: space-between;
  margin-top: 6px;
}

.skeleton-item__author {
  width: 120px;
  height: 20px;
}

.skeleton-item__stats {
  width: 96px;
  height: 20px;
}
</style>
