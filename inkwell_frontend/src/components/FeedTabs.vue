<template>
  <div ref="root" class="segmented" :class="{ 'is-ready': ready }" role="tablist">
    <!-- 选中的底色是一个会滑动的滑块, 而不是瞬间切换的背景 -->
    <span class="segmented__thumb" :style="thumbStyle" aria-hidden="true" />

    <button
      v-for="(option, index) in options"
      :key="option.value"
      :ref="(el) => setItemRef(el, index)"
      class="segmented__item"
      :class="{ 'is-active': option.value === modelValue }"
      type="button"
      role="tab"
      :aria-selected="option.value === modelValue"
      @click="emit('update:modelValue', option.value)"
    >
      <AppIcon v-if="option.icon" :name="option.icon" :size="15" />
      {{ option.label }}
    </button>
  </div>
</template>

<script setup lang="ts">
import {
  computed,
  nextTick,
  onBeforeUnmount,
  onMounted,
  reactive,
  ref,
  watch,
  type ComponentPublicInstance,
} from 'vue'

import AppIcon from '@/components/AppIcon.vue'
import type { FeedTabOption } from '@/components/feedTabs'

const props = defineProps<{
  modelValue: string
  options: FeedTabOption[]
}>()

const emit = defineEmits<{
  (event: 'update:modelValue', value: string): void
}>()

const root = ref<HTMLElement | null>(null)
const itemRefs = ref<HTMLElement[]>([])
/** 首次测量前先不上过渡, 否则滑块会从左边滑过来 */
const ready = ref(false)
const thumb = reactive({ left: 0, width: 0 })

const thumbStyle = computed(() => ({
  transform: `translateX(${thumb.left}px)`,
  width: `${thumb.width}px`,
}))

function setItemRef(el: Element | ComponentPublicInstance | null, index: number): void {
  if (el instanceof HTMLElement) {
    itemRefs.value[index] = el
  }
}

/** 把滑块移到当前选中项的位置(只改 transform/width, 不触发重排) */
function measure(): void {
  const activeIndex = props.options.findIndex((option) => option.value === props.modelValue)
  const el = itemRefs.value[activeIndex]
  if (!el) {
    return
  }
  thumb.left = el.offsetLeft
  thumb.width = el.offsetWidth
  ready.value = true
}

let resizeObserver: ResizeObserver | null = null

onMounted(async () => {
  await nextTick()
  measure()
  resizeObserver = new ResizeObserver(() => measure())
  if (root.value) {
    resizeObserver.observe(root.value)
  }
})

onBeforeUnmount(() => {
  resizeObserver?.disconnect()
  resizeObserver = null
})

watch(
  () => props.modelValue,
  async () => {
    await nextTick()
    measure()
  },
)

watch(
  () => props.options,
  async () => {
    await nextTick()
    measure()
  },
)
</script>

<style scoped lang="scss">
.segmented {
  position: relative;
  display: inline-flex;
  gap: 2px;
  padding: 3px;
  border: 1px solid var(--border);
  border-radius: var(--radius-md);
  background: var(--segmented-bg);
}

.segmented__thumb {
  position: absolute;
  left: 0;
  top: 3px;
  bottom: 3px;
  z-index: 0;
  border: 1px solid var(--segmented-thumb-border);
  border-radius: var(--radius-sm);
  background: var(--segmented-thumb);
  box-shadow: var(--segmented-thumb-shadow);
  /* 只动 transform, 保证滑动是合成层动画 */
  transition: transform 320ms cubic-bezier(0.34, 1.2, 0.64, 1), width 320ms cubic-bezier(0.34, 1.2, 0.64, 1),
    background-color var(--transition), border-color var(--transition);

  /* 顶部一道极浅的高光, 让滑块看起来是"凸起"的 */
  &::after {
    content: '';
    position: absolute;
    inset: 0 6px auto;
    height: 1px;
    border-radius: var(--radius-pill);
    background: linear-gradient(90deg, transparent, rgba(255, 255, 255, 0.9), transparent);
    opacity: 0.6;
  }
}

[data-theme='dark'] .segmented__thumb::after {
  background: linear-gradient(90deg, transparent, rgba(255, 255, 255, 0.22), transparent);
  opacity: 0.9;
}

.segmented:not(.is-ready) .segmented__thumb {
  transition: none;
}

.segmented__item {
  position: relative;
  z-index: 1;
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 7px 15px;
  border: 0;
  border-radius: var(--radius-sm);
  background: transparent;
  color: var(--text-tertiary);
  font-size: var(--text-sm);
  font-weight: var(--weight-medium);
  transition: color var(--transition), transform 140ms ease;

  &:hover {
    color: var(--text-secondary);
  }

  &:active {
    transform: scale(0.97);
  }

  &.is-active {
    color: var(--text);
    font-weight: var(--weight-semibold);
  }

  &.is-active .icon {
    color: var(--accent);
  }
}
</style>
