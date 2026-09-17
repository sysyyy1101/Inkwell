import { onBeforeUnmount, onMounted, type Ref } from 'vue'

/**
 * 点击元素外部时关闭浮层。
 * 监听的是 document 上的捕获阶段, 因此浮层内部按钮的点击不会误触发关闭。
 */
export function useClickOutside(target: Ref<HTMLElement | null>, onOutside: () => void): void {
  function handlePointerDown(event: MouseEvent | TouchEvent): void {
    const element = target.value
    if (!element) {
      return
    }
    if (!element.contains(event.target as Node)) {
      onOutside()
    }
  }

  function handleKeydown(event: KeyboardEvent): void {
    if (event.key === 'Escape') {
      onOutside()
    }
  }

  onMounted(() => {
    document.addEventListener('mousedown', handlePointerDown)
    document.addEventListener('touchstart', handlePointerDown)
    document.addEventListener('keydown', handleKeydown)
  })

  onBeforeUnmount(() => {
    document.removeEventListener('mousedown', handlePointerDown)
    document.removeEventListener('touchstart', handlePointerDown)
    document.removeEventListener('keydown', handleKeydown)
  })
}
