<template>
  <Teleport to="body">
    <div class="toast-host" aria-live="polite" aria-atomic="false">
      <TransitionGroup name="toast">
        <div v-for="item in toast.toasts" :key="item.id" class="toast" :class="`toast--${item.type}`">
          <AppIcon :name="iconOf(item.type)" :size="18" />
          <span class="toast__message">{{ item.message }}</span>
          <button class="toast__close" type="button" aria-label="关闭提示" @click="toast.dismiss(item.id)">
            <AppIcon name="close" :size="14" />
          </button>
        </div>
      </TransitionGroup>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
import AppIcon from '@/components/AppIcon.vue'
import type { IconName } from '@/components/icons'
import { useToastStore, type ToastType } from '@/stores/toast'

const toast = useToastStore()

function iconOf(type: ToastType): IconName {
  if (type === 'success') {
    return 'check'
  }
  if (type === 'error') {
    return 'alert'
  }
  return 'info'
}
</script>

<style scoped lang="scss">
.toast-host {
  position: fixed;
  left: 50%;
  top: calc(var(--header-height) + 16px);
  z-index: 60;
  display: flex;
  width: min(92vw, 420px);
  flex-direction: column;
  gap: 10px;
  transform: translateX(-50%);
  pointer-events: none;
}

.toast {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 12px 14px;
  border: 1px solid var(--border);
  border-radius: var(--radius-md);
  background: var(--surface);
  box-shadow: var(--shadow-md);
  pointer-events: auto;
}

.toast__message {
  flex: 1;
  font-size: var(--text-base);
}

.toast__close {
  display: inline-flex;
  width: 24px;
  height: 24px;
  align-items: center;
  justify-content: center;
  border-radius: var(--radius-pill);
  color: var(--text-tertiary);
  transition: background-color var(--transition), color var(--transition);

  &:hover {
    background: var(--surface-hover);
    color: var(--text);
  }
}

.toast--success {
  color: var(--success);

  .toast__message {
    color: var(--text);
  }
}

.toast--error {
  color: var(--danger);

  .toast__message {
    color: var(--text);
  }
}

.toast--info {
  color: var(--accent);

  .toast__message {
    color: var(--text);
  }
}

.toast-enter-active,
.toast-leave-active {
  transition: opacity 240ms ease, transform 240ms cubic-bezier(0.32, 0.72, 0, 1);
}

.toast-enter-from {
  opacity: 0;
  transform: translateY(-12px) scale(0.97);
}

.toast-leave-to {
  opacity: 0;
  transform: translateY(-8px) scale(0.98);
}

.toast-move {
  transition: transform 240ms ease;
}
</style>
