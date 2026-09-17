<template>
  <section id="comments" class="comments card">
    <header class="comments__header">
      <h2 class="comments__title">
        <AppIcon name="message" :size="18" />
        评论
        <span class="comments__count">{{ total }}</span>
      </h2>
    </header>

    <div class="comments__body">
      <CommentComposer :post-id="postId" placeholder="写下你的看法…" @created="handleCreated" />

      <div v-if="loading" class="comments__loading">
        <span v-for="index in 3" :key="index" class="skeleton comments__skeleton" />
      </div>

      <p v-else-if="error" class="comments__error">
        {{ error }}
        <button class="btn btn--ghost btn--sm" type="button" @click="load">重新加载</button>
      </p>

      <EmptyState
        v-else-if="!roots.length"
        icon="message"
        title="还没有评论"
        description="成为第一个参与讨论的人吧。"
      />

      <div v-else class="comments__list">
        <CommentItem
          v-for="(item, index) in roots"
          :key="item.comment_id"
          class="enter-rise"
          :style="{ '--enter-index': Math.min(index, 6) }"
          :comment="item"
          :children-map="childrenMap"
          :comment-index="commentIndex"
          :depth="0"
          :post-id="postId"
          :reply-to-id="replyToId"
          :current-user-id="auth.userID"
          @reply="handleReply"
          @cancel-reply="replyToId = null"
          @created="handleCreated"
        />
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import AppIcon from '@/components/AppIcon.vue'
import CommentComposer from '@/components/CommentComposer.vue'
import CommentItem from '@/components/CommentItem.vue'
import EmptyState from '@/components/EmptyState.vue'
import { commentApi, toErrorMessage, type PostComment } from '@/api'
import { useAuthStore } from '@/stores/auth'
import { useToastStore } from '@/stores/toast'

const props = defineProps<{ postId: string }>()

const emit = defineEmits<{
  /** 评论总数变化(加载完成/新评论插入): 详情页用它同步"N 条评论" */
  (event: 'update:total', total: number): void
}>()

const auth = useAuthStore()
const toast = useToastStore()
const router = useRouter()
const route = useRoute()

const comments = ref<PostComment[]>([])
const loading = ref(true)
const error = ref<string | null>(null)
const replyToId = ref<string | null>(null)

const total = computed(() => comments.value.length)

/** 顶层评论: parent_id 为 '0' */
const roots = computed(() => comments.value.filter((item) => item.parent_id === '0'))

/** 楼中楼: parent_id -> 直接子评论 */
const childrenMap = computed(() => {
  const map = new Map<string, PostComment[]>()
  comments.value.forEach((item) => {
    if (item.parent_id === '0') {
      return
    }
    const list = map.get(item.parent_id)
    if (list) {
      list.push(item)
    } else {
      map.set(item.parent_id, [item])
    }
  })
  return map
})

/** comment_id -> 评论, 供子组件查被回复人 */
const commentIndex = computed(() => {
  const index = new Map<string, PostComment>()
  comments.value.forEach((item) => index.set(item.comment_id, item))
  return index
})

async function load(): Promise<void> {
  loading.value = true
  error.value = null
  try {
    comments.value = await commentApi.fetchCommentsByPost(props.postId)
    // 列表就是这个帖子的全部评论, 用它同步详情页的评论数
    emit('update:total', comments.value.length)
  } catch (e) {
    error.value = toErrorMessage(e)
    comments.value = []
  } finally {
    loading.value = false
  }
}

function handleReply(target: PostComment): void {
  if (!auth.isLogin) {
    toast.notifyThenRedirect('info', '登录后才能回复', () => {
      void router.push({ name: 'login', query: { redirect: route.fullPath } })
    })
    return
  }
  replyToId.value = replyToId.value === target.comment_id ? null : target.comment_id
}

/**
 * 评论发布成功后, 后端只返回了评论 ID。
 * 这里用 GET /comment?ids=xxx 把这刚刚创建的评论取回来再插进列表,
 * 这样作者名/时间/ID 都以服务端数据为准, 不在前端拼假数据。
 */
async function handleCreated(commentId: string): Promise<void> {
  replyToId.value = null
  try {
    const [created] = await commentApi.fetchCommentsByIds([commentId])
    if (created) {
      comments.value = [...comments.value, created]
      emit('update:total', comments.value.length)
      return
    }
  } catch {
    // 拉取失败时退化成整列表刷新
  }
  await load()
}

onMounted(() => {
  void load()
})

defineExpose({ reload: load })
</script>

<style scoped lang="scss">
.comments__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: var(--space-5) var(--space-5) var(--space-4);
  border-bottom: 1px solid var(--divider);
}

.comments__title {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  font-size: var(--text-md);
  font-weight: var(--weight-bold);
}

.comments__count {
  padding: 2px 10px;
  border-radius: var(--radius-pill);
  background: var(--surface-hover);
  color: var(--text-secondary);
  font-size: var(--text-xs);
  font-weight: var(--weight-medium);
}

.comments__body {
  display: flex;
  flex-direction: column;
  gap: var(--space-6);
  padding: var(--space-5);
}

.comments__list {
  display: flex;
  flex-direction: column;
  gap: var(--space-6);
}

.comments__loading {
  display: grid;
  gap: var(--space-3);
}

.comments__skeleton {
  height: 68px;
}

.comments__error {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-3);
  color: var(--text-secondary);
  font-size: var(--text-sm);
}
</style>
