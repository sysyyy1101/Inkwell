<template>
  <svg
    class="icon"
    :width="sizeValue"
    :height="sizeValue"
    viewBox="0 0 24 24"
    role="img"
    :aria-label="label || undefined"
    :aria-hidden="label ? undefined : 'true'"
    :fill="icon.fill ? 'currentColor' : 'none'"
    stroke="currentColor"
    :stroke-width="icon.fill ? 0 : strokeWidth"
    stroke-linecap="round"
    stroke-linejoin="round"
  >
    <path v-for="(d, index) in icon.paths" :key="index" :d="d" />
  </svg>
</template>

<script setup lang="ts">
import { computed } from 'vue'

import { ICON_PATHS, type IconDefinition, type IconName } from '@/components/icons'

const ICONS: Record<IconName, IconDefinition> = ICON_PATHS

const props = withDefaults(
  defineProps<{
    name: IconName
    size?: number | string
    strokeWidth?: number
    label?: string
  }>(),
  {
    size: 20,
    strokeWidth: 1.7,
    label: '',
  },
)

const icon = computed(() => ICONS[props.name])
const sizeValue = computed(() => (typeof props.size === 'number' ? String(props.size) : props.size))
</script>

<style scoped>
.icon {
  flex: none;
  display: block;
}
</style>
