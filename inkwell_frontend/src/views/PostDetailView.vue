<template>
  <div class="container layout">
    <aside class="sidebar">
      <SideNav />
    </aside>

    <div class="detail">
      <RouterLink class="back-link" :to="{ name: 'home' }">
        <AppIcon name="arrowLeft" :size="16" />
        返回首页
      </RouterLink>

      <div v-if="loading" class="card detail__skeleton">
        <span class="skeleton detail__skeleton-title" />
        <span class="skeleton detail__skeleton-line" />
        <span class="skeleton detail__skeleton-line" />
        <span class="skeleton detail__skeleton-line detail__skeleton-line--short" />
      </div>

      <EmptyState v-else-if="error" icon="alert" title="帖子加载失败" :description="error">
        <RouterLink class="btn btn--primary btn--sm" :to="{ name: 'home' }">回到首页</RouterLink>
      </EmptyState>

      <template v-else-if="post">
        <article class="card post">
          <header class="post__header">
            <div class="post__meta">
              <RouterLink class="chip chip--accent" :to="{ name: 'community', params: { id: post.community_id } }">
                {{ post.community_name || '未分类' }}
              </RouterLink>
              <span class="post__meta-time" :title="formatDateTime(post.create_time)">
                {{ formatRelativeTime(post.create_time) }}发布
              </span>
              <span v-if="!voteOpen" class="chip chip--warning">已停止投票</span>
            </div>

            <h1 class="post__title font-display">{{ post.title }}</h1>

            <RouterLink
              class="post__author"
              :to="{ name: 'user', params: { id: post.author_id } }"
              :title="`查看 ${post.author_name || '匿名用户'} 的主页`"
            >
              <span class="avatar avatar--lg" :style="avatarStyle">{{ initial }}</span>
              <span class="post__author-info">
                <span class="post__author-name">{{ post.author_name || '匿名用户' }}</span>
                <span class="post__author-time">{{ formatDateTime(post.create_time) }}</span>
              </span>
            </RouterLink>
          </header>

          <div class="post__content">{{ post.content }}</div>

          <footer class="post__footer">
            <VoteControl
              :post-id="post.post_id"
              :count="voteCount"
              :direction="direction"
              :disabled="!voteOpen"
              :disabled-reason="voteDisabledReason"
              size="lg"
              layout="horizontal"
              @change="handleVoteChange"
            />

            <div class="post__actions">
              <a class="btn btn--ghost btn--sm" href="#comments">
                <AppIcon name="message" :size="16" />
                {{ formatCount(commentCount) }} 条评论
              </a>
              <button class="btn btn--ghost btn--sm" type="button" @click="copyLink">
                <AppIcon name="link" :size="16" />
                复制链接
              </button>
            </div>
          </footer>
        </article>

        <!-- 评论列表加载完/新评论插入后会回报总数, 保证这里的"N 条评论"立刻同步 -->
        <CommentList :post-id="post.post_id" @update:total="commentCount = $event" />
      </template>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useRoute } from 'vue-router'

import AppIcon from '@/components/AppIcon.vue'
import CommentList from '@/components/CommentList.vue'
import EmptyState from '@/components/EmptyState.vue'
import SideNav from '@/components/SideNav.vue'
import VoteControl from '@/components/VoteControl.vue'
import { ApiError, postApi, toErrorMessage, type PostDetail, type VoteDirection } from '@/api'
import { useAuthStore } from '@/stores/auth'
import { useToastStore } from '@/stores/toast'
import { useVoteStore } from '@/stores/votes'
import {
  avatarText,
  avatarTone,
  formatCount,
  formatDateTime,
  formatRelativeTime,
  isVoteWindowOpen,
} from '@/utils/format'

const route = useRoute()
const auth = useAuthStore()
const toast = useToastStore()
const votes = useVoteStore()

const postId = computed(() => String(route.params.id ?? ''))
const post = ref<PostDetail | null>(null)
const loading = ref(true)
const error = ref<string | null>(null)

const voteCount = ref(0)
const direction = ref<VoteDirection>(0)
/** 评论数: 先用详情接口返回的值, 之后由评论列表回报的权威数量覆盖 */
const commentCount = ref(0)

const voteOpen = computed(() => (post.value ? isVoteWindowOpen(post.value.create_time) : true))
const voteDisabledReason = computed(() => (voteOpen.value ? '' : '发帖已超过 7 天, 不能再投票'))
const initial = computed(() => avatarText(post.value?.author_name ?? ''))
const avatarStyle = computed(() => avatarTone(post.value?.author_name ?? ''))

async function load(): Promise<void> {
  loading.value = true
  error.value = null
  try {
    const detail = await postApi.fetchPostDetail(postId.value)
    post.value = detail
    voteCount.value = detail.vote_num
    commentCount.value = detail.comment_num
    direction.value = resolveDirection(detail.author_id)
  } catch (e) {
    post.value = null
    error.value =
      e instanceof ApiError && e.code === 1009 ? '这个帖子不存在或已被删除。' : toErrorMessage(e)
  } finally {
    loading.value = false
  }
}

/** 本地记录的投票方向优先; 没有记录时, 作者本人默认已赞成(后端发帖时自动投票) */
function resolveDirection(authorId: string): VoteDirection {
  const stored = votes.get(auth.userID, postId.value)
  if (stored !== null) {
    return stored
  }
  return auth.userID && auth.userID === authorId ? 1 : 0
}

function handleVoteChange(payload: { direction: VoteDirection; count: number }): void {
  direction.value = payload.direction
  voteCount.value = payload.count
  if (auth.userID) {
    votes.record(auth.userID, postId.value, payload.direction)
  }
}

async function copyLink(): Promise<void> {
  try {
    await navigator.clipboard.writeText(window.location.href)
    toast.success('链接已复制')
  } catch {
    toast.error('复制失败, 请手动复制地址栏链接')
  }
}

// 同一组件在帖子之间跳转时(例如从评论里的链接), 需要重新拉数据
watch(postId, async () => {
  await load()
})

onMounted(async () => {
  await load()
  // 从列表页点"查看讨论"进来时带着 #comments, 等评论渲染完再滚动过去
  if (route.hash === '#comments') {
    setTimeout(() => {
      document.getElementById('comments')?.scrollIntoView({ behavior: 'smooth', block: 'start' })
    }, 300)
  }
  // 登录状态可能在页面打开后才变化(登录后回到本页), 这时要重新解析投票方向
  if (!auth.isLogin) {
    direction.value = 0
  }
})
</script>

<style scoped lang="scss">
.detail {
  display: flex;
  max-width: 860px;
  flex-direction: column;
  gap: var(--space-5);
}

.back-link {
  display: inline-flex;
  align-items: center;
  align-self: flex-start;
  gap: 6px;
  padding: 6px 12px 6px 8px;
  border-radius: var(--radius-pill);
  color: var(--text-secondary);
  font-size: var(--text-sm);
  transition: background-color var(--transition), color var(--transition);

  &:hover {
    background: var(--surface-hover);
    color: var(--text);
  }
}

.post {
  padding: var(--space-6);

  @media (min-width: 768px) {
    padding: var(--space-7);
  }
}

.post__header {
  display: flex;
  flex-direction: column;
  gap: var(--space-4);
}

.post__meta {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 10px;
}

.post__meta-time {
  color: var(--text-tertiary);
  font-size: var(--text-sm);
}

.post__title {
  font-size: var(--text-xl);
  font-weight: var(--weight-heavy);
  letter-spacing: -0.03em;
  line-height: 1.3;
  overflow-wrap: anywhere;

  @media (min-width: 768px) {
    font-size: var(--text-2xl);
  }
}

.post__author {
  display: inline-flex;
  align-self: flex-start;
  align-items: center;
  gap: 10px;
  border-radius: var(--radius-pill);

  &:hover .post__author-name {
    color: var(--accent);
  }
}

.post__author-info {
  display: flex;
  flex-direction: column;
}

.post__author-name {
  font-size: var(--text-sm);
  font-weight: var(--weight-bold);
  transition: color var(--transition);
}

.post__author-time {
  color: var(--text-tertiary);
  font-size: var(--text-xs);
}

.post__content {
  margin: var(--space-6) 0;
  padding-top: var(--space-6);
  border-top: 1px solid var(--divider);
  color: var(--text);
  font-size: var(--text-md);
  line-height: var(--leading-relaxed);
  overflow-wrap: anywhere;
  white-space: pre-wrap;
}

.post__footer {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-4);
  padding-top: var(--space-5);
  border-top: 1px solid var(--divider);
}

.post__actions {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: var(--space-2);
}

.detail__skeleton {
  display: flex;
  flex-direction: column;
  gap: var(--space-4);
  padding: var(--space-7);
}

.detail__skeleton-title {
  width: 70%;
  height: 30px;
}

.detail__skeleton-line {
  width: 100%;
  height: 16px;
}

.detail__skeleton-line--short {
  width: 60%;
}
</style>
