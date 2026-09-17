<template>
  <div class="composer" :class="{ 'composer--compact': compact }">
    <template v-if="auth.isLogin">
      <div class="composer__top">
        <span class="avatar avatar--sm" :style="avatarStyle">{{ initial }}</span>
        <span class="composer__who">
          {{ auth.username }}
          <span v-if="replyToName" class="composer__reply-to">回复 {{ replyToName }}</span>
        </span>
        <button v-if="replyToName" class="btn btn--ghost btn--sm" type="button" @click="emit('cancel')">
          取消回复
        </button>
      </div>

      <textarea
        ref="textareaRef"
        v-model="content"
        class="textarea composer__input"
        :placeholder="placeholder"
        :maxlength="maxLength"
        :disabled="submitting"
        @keydown.ctrl.enter="submit"
        @keydown.meta.enter="submit"
      />

      <div class="composer__footer">
        <span class="composer__counter" :class="{ 'is-warning': content.length > maxLength * 0.9 }">
          {{ content.length }} / {{ maxLength }}
        </span>
        <button
          class="btn btn--primary"
          type="button"
          :disabled="!canSubmit"
          @click="submit"
        >
          <AppIcon name="send" :size="16" />
          {{ submitting ? '发布中…' : submitLabel }}
        </button>
      </div>
    </template>

    <div v-else class="composer__guest">
      <span class="composer__guest-text">登录后即可参与讨论</span>
      <button class="btn btn--primary btn--sm" type="button" @click="goLogin">
        <AppIcon name="user" :size="16" />
        去登录
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import AppIcon from '@/components/AppIcon.vue'
import { commentApi, toErrorMessage } from '@/api'
import { useAuthStore } from '@/stores/auth'
import { useToastStore } from '@/stores/toast'
import { avatarText, avatarTone } from '@/utils/format'

const props = withDefaults(
  defineProps<{
    postId: string
    /** 回复哪条评论, 直接回复帖子时传 '0' */
    parentId?: string
    /** 被回复人的名字, 用于提示"回复 xxx" */
    replyToName?: string
    placeholder?: string
    compact?: boolean
    maxLength?: number
    autoFocus?: boolean
  }>(),
  {
    parentId: '0',
    replyToName: '',
    placeholder: '说点什么… (Ctrl + Enter 快速发布)',
    compact: false,
    maxLength: 2000,
    autoFocus: false,
  },
)

const emit = defineEmits<{
  /** 评论成功后返回新评论 ID, 由父组件决定怎么插入列表 */
  (event: 'created', commentId: string): void
  (event: 'cancel'): void
}>()

const auth = useAuthStore()
const toast = useToastStore()
const router = useRouter()
const route = useRoute()

const content = ref('')
const submitting = ref(false)
const textareaRef = ref<HTMLTextAreaElement | null>(null)

const canSubmit = computed(() => content.value.trim().length > 0 && !submitting.value)
const submitLabel = computed(() => (props.replyToName ? '回复' : '发表评论'))
const initial = computed(() => avatarText(auth.username ?? ''))
const avatarStyle = computed(() => avatarTone(auth.username ?? ''))

async function submit(): Promise<void> {
  if (!canSubmit.value) {
    return
  }
  if (!auth.isLogin) {
    goLogin()
    return
  }
  submitting.value = true
  try {
    const result = await commentApi.createComment({
      post_id: props.postId,
      parent_id: props.parentId,
      content: content.value.trim(),
    })
    content.value = ''
    toast.success(props.replyToName ? '回复成功' : '评论成功')
    emit('created', result.comment_id)
  } catch (error) {
    toast.error(toErrorMessage(error))
  } finally {
    submitting.value = false
  }
}

function goLogin(): void {
  toast.notifyThenRedirect('info', '请先登录', () => {
    void router.push({ name: 'login', query: { redirect: route.fullPath } })
  })
}

watch(
  () => props.autoFocus,
  (value) => {
    if (value) {
      void nextTick(() => textareaRef.value?.focus())
    }
  },
)

onMounted(() => {
  if (props.autoFocus) {
    void nextTick(() => textareaRef.value?.focus())
  }
})

defineExpose({ focus: () => textareaRef.value?.focus() })
</script>

<style scoped lang="scss">
.composer {
  display: flex;
  flex-direction: column;
  gap: var(--space-3);
  padding: var(--space-5);
  border: 1px solid var(--border);
  border-radius: var(--radius-lg);
  background: var(--surface);
}

.composer--compact {
  padding: var(--space-4);
  border-radius: var(--radius-md);
}

.composer__top {
  display: flex;
  align-items: center;
  gap: 10px;
}

.composer__who {
  display: inline-flex;
  flex: 1;
  align-items: center;
  gap: 8px;
  color: var(--text);
  font-size: var(--text-sm);
  font-weight: var(--weight-bold);
}

.composer__reply-to {
  padding: 2px 10px;
  border-radius: var(--radius-pill);
  background: var(--accent-soft);
  color: var(--accent);
  font-size: var(--text-xs);
  font-weight: var(--weight-medium);
}

.composer__input {
  min-height: 96px;
}

.composer--compact .composer__input {
  min-height: 76px;
}

.composer__footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-3);
}

.composer__counter {
  color: var(--text-tertiary);
  font-family: var(--font-mono);
  font-size: var(--text-xs);
}

.composer__counter.is-warning {
  color: var(--warning);
}

.composer__guest {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-3);
}

.composer__guest-text {
  color: var(--text-secondary);
  font-size: var(--text-sm);
}
</style>
