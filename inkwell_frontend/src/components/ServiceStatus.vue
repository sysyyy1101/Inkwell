<template>
  <span class="status" :class="`status--${state}`">
    <span class="status__dot" aria-hidden="true" />
    <span class="status__text">{{ text }}</span>
    <button
      class="status__refresh"
      type="button"
      :disabled="state === 'checking'"
      :title="`重新检测 (GET /api/v1/ping)`"
      aria-label="重新检测后端服务状态"
      @click="check()"
    >
      <AppIcon name="refresh" :size="13" />
    </button>
  </span>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'

import AppIcon from '@/components/AppIcon.vue'
import { systemApi } from '@/api'

type StatusState = 'checking' | 'online' | 'offline'

const state = ref<StatusState>('checking')

const text = computed(() => {
  if (state.value === 'checking') {
    return '检测中'
  }
  return state.value === 'online' ? '后端服务正常' : '后端服务不可用'
})

/** 调用 GET /ping 探活, 用来确认前端确实连得上后端 */
async function check(): Promise<void> {
  state.value = 'checking'
  try {
    const result = await systemApi.ping()
    state.value = result === 'pong' ? 'online' : 'offline'
  } catch {
    state.value = 'offline'
  }
}

onMounted(() => {
  void check()
})
</script>

<style scoped lang="scss">
.status {
  display: inline-flex;
  align-items: center;
  gap: 7px;
  color: var(--text-tertiary);
  font-size: var(--text-sm);
}

.status__dot {
  width: 8px;
  height: 8px;
  border-radius: var(--radius-pill);
  background: var(--text-tertiary);
  transition: background-color var(--transition), box-shadow var(--transition);
}

.status--online .status__dot {
  background: var(--success);
  box-shadow: 0 0 0 3px var(--success-soft);
}

.status--offline .status__dot {
  background: var(--danger);
  box-shadow: 0 0 0 3px var(--danger-soft);
}

.status--checking .status__dot {
  animation: pulse 1.2s ease-in-out infinite;
}

.status__refresh {
  display: inline-flex;
  width: 22px;
  height: 22px;
  align-items: center;
  justify-content: center;
  border-radius: var(--radius-pill);
  color: var(--text-tertiary);
  transition: background-color var(--transition), color var(--transition);

  &:hover:not(:disabled) {
    background: var(--surface-hover);
    color: var(--text);
  }

  &:disabled {
    opacity: 0.5;
  }
}

@keyframes pulse {
  0%,
  100% {
    opacity: 1;
  }
  50% {
    opacity: 0.35;
  }
}
</style>
