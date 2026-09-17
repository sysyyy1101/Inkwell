<template>
  <div class="vote" :class="[`vote--${layout}`, `vote--${size}`]">
    <button
      class="vote__button vote__button--up"
      :class="{ 'is-active': direction === 1 }"
      type="button"
      :disabled="pending || disabled"
      :aria-pressed="direction === 1"
      :title="disabled ? disabledReason : direction === 1 ? '取消赞成' : '赞成'"
      @click="cast(1)"
    >
      <AppIcon name="arrowUp" :size="iconSize" />
    </button>

    <span
      class="vote__count"
      :class="{ 'is-active': direction !== 0, 'is-bumped': bumped }"
    >
      {{ formatCount(count) }}
    </span>

    <button
      class="vote__button vote__button--down"
      :class="{ 'is-active': direction === -1 }"
      type="button"
      :disabled="pending || disabled"
      :aria-pressed="direction === -1"
      :title="disabled ? disabledReason : direction === -1 ? '取消反对' : '反对'"
      @click="cast(-1)"
    >
      <AppIcon name="arrowDown" :size="iconSize" />
    </button>
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import AppIcon from '@/components/AppIcon.vue'
import { ApiError, postApi, toErrorMessage, type VoteDirection } from '@/api'
import { useAuthStore } from '@/stores/auth'
import { useToastStore } from '@/stores/toast'
import { formatCount } from '@/utils/format'

const props = withDefaults(
  defineProps<{
    postId: string
    count: number
    direction: VoteDirection
    size?: 'sm' | 'md' | 'lg'
    layout?: 'horizontal' | 'vertical'
    disabled?: boolean
    disabledReason?: string
  }>(),
  {
    size: 'md',
    layout: 'horizontal',
    disabled: false,
    disabledReason: '',
  },
)

const emit = defineEmits<{
  /** 投票成功后把新的方向和票数交给父组件(父组件负责更新展示与本地记录) */
  (event: 'change', payload: { direction: VoteDirection; count: number }): void
}>()

const auth = useAuthStore()
const toast = useToastStore()
const router = useRouter()
const route = useRoute()

const pending = ref(false)
/** 票数变化时给数字加一次轻微弹动, 让"投票成功"有即时的视觉反馈 */
const bumped = ref(false)
let bumpTimer: ReturnType<typeof setTimeout> | null = null

watch(
  () => props.count,
  async () => {
    bumped.value = false
    await nextTick()
    bumped.value = true
    if (bumpTimer) {
      clearTimeout(bumpTimer)
    }
    bumpTimer = setTimeout(() => {
      bumped.value = false
    }, 300)
  },
)

onBeforeUnmount(() => {
  if (bumpTimer) {
    clearTimeout(bumpTimer)
  }
})

const iconSize = computed(() => (props.size === 'lg' ? 18 : props.size === 'sm' ? 12 : 14))

/**
 * 点击赞成/反对:
 * - 没投过 -> 投这个方向;
 * - 已经投过同一方向 -> 取消投票(0);
 * - 投的是相反方向 -> 直接反转, 后端支持。
 * 显示的是净票数(赞成 - 反对), 具体数值以后端返回的 vote_num 为准。
 */
async function cast(target: 1 | -1): Promise<void> {
  if (pending.value || props.disabled) {
    return
  }
  if (!auth.isLogin) {
    toast.notifyThenRedirect('info', '登录后即可参与投票', () => {
      void router.push({ name: 'login', query: { redirect: route.fullPath } })
    })
    return
  }

  const current = props.direction
  const next: VoteDirection = current === target ? 0 : target
  pending.value = true
  try {
    const result = await postApi.votePost(props.postId, next)
    // 票数以服务端返回的权威值为准。
    // 本机记录可能缺失或过期(换设备/清缓存), 自己推算会和服务端对不上, 所以优先用服务端的数字。
    // 兜底按净分推算: 新净分 = 旧净分 + (新方向 - 旧方向), 即赞成 +1 / 反对 -1。
    const count =
      typeof result?.vote_num === 'number'
        ? result.vote_num
        : props.count + (next - current)
    emit('change', { direction: next, count })
    toast.success(next === 0 ? '已取消投票' : next === 1 ? '已赞成' : '已反对')
  } catch (error) {
    // 后端拒绝"重复投同一方向的票", 说明服务端的记录和本地不一致(例如换设备投过)。
    // 这种情况把本地状态同步成服务端的状态即可, 票数不变。
    if (error instanceof ApiError && error.code === 1001 && error.message.includes('已经投过票了')) {
      emit('change', { direction: next, count: props.count })
      toast.info('投票状态已同步')
      return
    }
    toast.error(toErrorMessage(error))
  } finally {
    pending.value = false
  }
}
</script>

<style scoped lang="scss">
.vote {
  display: inline-flex;
  align-items: center;
  gap: 2px;
  padding: 3px;
  border: 1px solid var(--border);
  border-radius: var(--radius-sm);
  background: var(--surface-2);
}

.vote--vertical {
  flex-direction: column;
  gap: 1px;
  padding: 4px 3px;
}

.vote__button {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  border-radius: var(--radius-xs);
  color: var(--text-tertiary);
  transition: background-color var(--transition), color var(--transition);

  &:active:not(:disabled) {
    transform: scale(0.88);
  }

  &:hover:not(:disabled) {
    background: var(--surface-hover);
    color: var(--text);
  }

  &:disabled {
    cursor: not-allowed;
    opacity: 0.4;
  }
}

.vote--sm .vote__button {
  width: 26px;
  height: 26px;
}

.vote--md .vote__button {
  width: 32px;
  height: 32px;
}

.vote--lg .vote__button {
  width: 40px;
  height: 40px;
}

.vote__button--up.is-active {
  background: var(--vote-up-soft);
  color: var(--vote-up);
}

.vote__button--down.is-active {
  background: var(--vote-down-soft);
  color: var(--vote-down);
}

.vote__count {
  min-width: 34px;
  color: var(--text-secondary);
  font-size: var(--text-sm);
  font-variant-numeric: tabular-nums;
  font-weight: var(--weight-semibold);
  text-align: center;
}

.vote__count.is-active {
  color: var(--text);
}

.vote__count.is-bumped {
  display: inline-block;
  animation: count-bump 300ms cubic-bezier(0.34, 1.4, 0.64, 1);
}

.vote--lg .vote__count {
  font-size: var(--text-md);
}
</style>
