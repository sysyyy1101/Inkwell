import { ref } from 'vue'
import { defineStore } from 'pinia'

export type ToastType = 'success' | 'error' | 'info'

export interface Toast {
  id: number
  type: ToastType
  message: string
}

let seed = 0

/** 轻量的全局提示, 用于接口成功/失败反馈 */
export const useToastStore = defineStore('toast', () => {
  const toasts = ref<Toast[]>([])
  const timers = new Map<number, ReturnType<typeof setTimeout>>()

  function dismiss(id: number): void {
    toasts.value = toasts.value.filter((item) => item.id !== id)
    const timer = timers.get(id)
    if (timer) {
      clearTimeout(timer)
      timers.delete(id)
    }
  }

  function push(type: ToastType, message: string, duration = type === 'error' ? 4200 : 2600): number {
    seed += 1
    const id = seed
    toasts.value = [...toasts.value, { id, type, message }]
    timers.set(
      id,
      setTimeout(() => dismiss(id), duration),
    )
    return id
  }

  function success(message: string): number {
    return push('success', message)
  }

  function error(message: string): number {
    return push('error', message)
  }

  function info(message: string): number {
    return push('info', message)
  }

  /**
   * 先弹提示, 等它渲染出来再执行跳转。
   *
   * 实测: 如果在同一次点击事件里先 toast 再 router.push, 路由切换会把这次
   * ToastHost 的 DOM 更新吞掉, 用户完全看不到提示, 所以这里把跳转放到下一帧之后。
   */
  function notifyThenRedirect(
    type: ToastType,
    message: string,
    redirect: () => void,
    delay = 600,
  ): void {
    push(type, message)
    window.setTimeout(redirect, delay)
  }

  return {
    toasts,
    push,
    success,
    error,
    info,
    notifyThenRedirect,
    dismiss,
  }
})
