<template>
  <div class="container">
    <div class="profile-page">
      <RouterLink class="back-link" :to="{ name: 'home' }">
        <AppIcon name="arrowLeft" :size="16" />
        返回首页
      </RouterLink>

      <section v-if="loadingProfile" class="card profile">
        <span class="skeleton avatar avatar--hero" />
        <div class="profile__body">
          <span class="skeleton profile__skeleton-name" />
          <span class="skeleton profile__skeleton-line" />
        </div>
      </section>

      <EmptyState
        v-else-if="profileError"
        icon="alert"
        title="用户加载失败"
        :description="profileError"
      >
        <RouterLink class="btn btn--primary btn--sm" :to="{ name: 'home' }">回到首页</RouterLink>
      </EmptyState>

      <template v-else-if="profile">
        <section class="card profile">
          <span class="avatar avatar--hero" :style="avatarStyle">{{ initial }}</span>

          <div class="profile__body">
            <div class="profile__name-row">
              <h1 class="profile__name font-display">{{ profile.username }}</h1>
              <span v-if="isMe" class="chip chip--accent">我</span>
            </div>
            <div class="profile__meta">
              <span class="profile__meta-item">
                <AppIcon name="calendar" :size="15" />
                加入于 {{ formatDate(profile.create_time) }}
              </span>
              <span class="profile__meta-item profile__meta-item--id" :title="profile.user_id">
                <AppIcon name="link" :size="15" />
                ID {{ profile.user_id }}
              </span>
            </div>
          </div>

          <RouterLink v-if="isMe" class="btn btn--primary" :to="{ name: 'publish' }">
            <AppIcon name="compose" :size="16" />
            发布帖子
          </RouterLink>
        </section>

        <div class="timeline-head">
          <h2 class="timeline-head__title">
            {{ isMe ? '我发布的帖子' : `${profile.username} 发布的帖子` }}
          </h2>
          <button class="btn btn--ghost btn--sm" type="button" :disabled="loading" @click="refresh">
            <AppIcon name="refresh" :size="15" />
            刷新
          </button>
        </div>

        <SkeletonList v-if="loading" :count="3" />

        <EmptyState v-else-if="error" icon="alert" title="列表加载失败" :description="error">
          <button class="btn btn--primary btn--sm" type="button" @click="refresh">重新加载</button>
        </EmptyState>

        <template v-else-if="dayGroups.length">
          <!-- 朋友圈式时间轴: 左边是按天聚合的日期(同一天只出现一次), 右边是当天的帖子 -->
          <section class="timeline">
            <div v-for="group in dayGroups" :key="group.key" class="timeline__group">
              <div class="timeline__date">
                <span class="timeline__day font-display">{{ group.day }}</span>
                <span class="timeline__weekday">{{ group.weekday }}</span>
              </div>
              <div class="timeline__posts">
                <PostCard
                  v-for="(post, index) in group.posts"
                  :key="post.post_id"
                  class="enter-rise"
                  :style="{ '--enter-index': Math.min(index, 6) }"
                  :post="post"
                  :show-time="false"
                />
              </div>
            </div>
          </section>
          <LoadMoreButton :has-more="hasMore" :loading="loadingMore" @load-more="loadMore" />
        </template>

        <EmptyState
          v-else
          icon="compose"
          :title="isMe ? '你还没有发过帖子' : '这个人还没有发过帖子'"
          :description="isMe ? '写下第一篇帖子, 让大家看到你的想法。' : '可以先去首页看看大家都在聊什么。'"
        >
          <RouterLink v-if="isMe" class="btn btn--primary btn--sm" :to="{ name: 'publish' }">
            去发帖
          </RouterLink>
          <RouterLink v-else class="btn btn--secondary btn--sm" :to="{ name: 'home' }">
            回到首页
          </RouterLink>
        </EmptyState>
      </template>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useRoute } from 'vue-router'

import AppIcon from '@/components/AppIcon.vue'
import EmptyState from '@/components/EmptyState.vue'
import LoadMoreButton from '@/components/LoadMoreButton.vue'
import PostCard from '@/components/PostCard.vue'
import SkeletonList from '@/components/SkeletonList.vue'
import { ApiError, toErrorMessage, userApi, type PostListItem, type UserProfile } from '@/api'
import { usePostFeed } from '@/composables/usePostFeed'
import { useAuthStore } from '@/stores/auth'
import { avatarText, avatarTone, formatDate, toTimestamp } from '@/utils/format'

/** 时间轴上的一天: 同一天的帖子共用这一个日期 */
interface DayGroup {
  key: string
  day: string
  weekday: string
  posts: PostListItem[]
}

const WEEKDAYS = ['星期日', '星期一', '星期二', '星期三', '星期四', '星期五', '星期六']

/** 一天的 key: 按浏览器本地时区归到"哪一天" */
function dayKey(date: Date): string {
  return `${date.getFullYear()}-${date.getMonth() + 1}-${date.getDate()}`
}

/** 日期主标题: 今天 / 昨天 / 9月15日 / 2025年9月15日 */
function dayTitle(date: Date, now: Date): string {
  const sameDay = (a: Date, b: Date) => dayKey(a) === dayKey(b)
  if (sameDay(date, now)) {
    return '今天'
  }
  const yesterday = new Date(now.getFullYear(), now.getMonth(), now.getDate() - 1)
  if (sameDay(date, yesterday)) {
    return '昨天'
  }
  const monthDay = `${date.getMonth() + 1}月${date.getDate()}日`
  return date.getFullYear() === now.getFullYear() ? monthDay : `${date.getFullYear()}年${monthDay}`
}

const route = useRoute()
const auth = useAuthStore()

const userId = computed(() => String(route.params.id ?? ''))
const profile = ref<UserProfile | null>(null)
const loadingProfile = ref(true)
const profileError = ref<string | null>(null)

const feed = usePostFeed({ source: 'user', userId })
const items = computed(() => feed.items.value)
const loading = computed(() => feed.loading.value)
const loadingMore = computed(() => feed.loadingMore.value)
const hasMore = computed(() => feed.hasMore.value)
const error = computed(() => feed.error.value)

const isMe = computed(() => Boolean(auth.userID) && auth.userID === userId.value)
const initial = computed(() => avatarText(profile.value?.username ?? ''))
const avatarStyle = computed(() => avatarTone(profile.value?.username ?? ''))

/**
 * 把帖子按"天"聚合给时间轴用。
 * 列表本身是按发帖时间倒序的, 所以这里直接用 Map 保持插入顺序就是"新的一天在前"。
 * 分页加载更多时新数据会并进已有的分组, 同一天不会重复出现日期。
 */
const dayGroups = computed<DayGroup[]>(() => {
  const now = new Date()
  const groups = new Map<string, DayGroup>()
  items.value.forEach((post) => {
    const timestamp = toTimestamp(post.create_time)
    const date = new Date(timestamp)
    const key = dayKey(date)
    let group = groups.get(key)
    if (!group) {
      group = { key, day: dayTitle(date, now), weekday: WEEKDAYS[date.getDay()], posts: [] }
      groups.set(key, group)
    }
    group.posts.push(post)
  })
  return [...groups.values()]
})

async function loadProfile(): Promise<void> {
  loadingProfile.value = true
  profileError.value = null
  profile.value = null
  try {
    profile.value = await userApi.fetchUserProfile(userId.value)
  } catch (e) {
    // 后端对不存在的用户返回 1009, 这里直接展示成"用户不存在"
    profileError.value =
      e instanceof ApiError && e.code === 1009 ? '这个用户不存在或已被删除。' : toErrorMessage(e)
  } finally {
    loadingProfile.value = false
  }
}

async function refresh(): Promise<void> {
  await feed.reload()
}

async function loadMore(): Promise<void> {
  await feed.loadMore()
}

/** 主页之间互相跳转时(例如从评论里点另一个人的头像), 路由参数变了要重新加载 */
watch(userId, async () => {
  await loadProfile()
  if (!profileError.value) {
    await feed.reload()
  }
})

onMounted(async () => {
  await loadProfile()
  if (!profileError.value) {
    await feed.reload()
  }
})
</script>

<style scoped lang="scss">
.profile-page {
  max-width: 1040px;
  margin: 0 auto;
}

.back-link {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  margin-bottom: var(--space-5);
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

.profile {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: var(--space-5);
  margin-bottom: var(--space-7);
  padding: var(--space-6);
}

.avatar--hero {
  width: 64px;
  height: 64px;
  border-radius: var(--radius-pill);
  font-size: var(--text-xl);
}

.profile__body {
  display: flex;
  min-width: 0;
  flex: 1;
  flex-direction: column;
  gap: var(--space-2);
}

.profile__name-row {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: var(--space-2);
}

.profile__name {
  font-size: var(--text-xl);
  font-weight: var(--weight-heavy);
  letter-spacing: var(--tracking-tighter);
  overflow-wrap: anywhere;
}

.profile__meta {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-4);
}

.profile__meta-item {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  color: var(--text-tertiary);
  font-size: var(--text-sm);
}

.profile__meta-item--id {
  font-family: var(--font-mono);
  font-size: var(--text-xs);
}

.profile__skeleton-name {
  width: 180px;
  height: 26px;
}

.profile__skeleton-line {
  width: 260px;
  height: 16px;
}

.timeline-head {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-3);
  margin-bottom: var(--space-5);
}

.timeline-head__title {
  font-size: var(--text-md);
  font-weight: var(--weight-bold);
}

/**
 * 朋友圈式时间轴: 一"天"一行。
 * 左列是日期(同一天只出现一次, 整列跟着滚动吸顶), 右列是当天的帖子,
 * 中间那条竖线 + 圆点提供时间轴的视觉引导。
 */
.timeline {
  display: flex;
  flex-direction: column;
  gap: var(--space-7);
}

.timeline__group {
  display: grid;
  grid-template-columns: 132px minmax(0, 1fr);
  gap: var(--space-4);
  align-items: start;
}

.timeline__date {
  position: sticky;
  top: calc(var(--header-height) + 20px);
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  gap: 3px;
  padding-right: var(--space-5);
  text-align: right;
  /* 时间轴竖线 */
  border-right: 1px solid var(--divider);

  /* 线上的圆点, 对齐日期第一行 */
  &::after {
    content: '';
    position: absolute;
    top: 12px;
    right: -0.5px;
    width: 9px;
    height: 9px;
    transform: translateX(50%);
    border-radius: var(--radius-pill);
    background: var(--accent);
    box-shadow: 0 0 0 3px var(--accent-soft);
  }
}

.timeline__day {
  font-size: var(--text-xl);
  font-weight: var(--weight-heavy);
  letter-spacing: var(--tracking-tighter);
  line-height: 1.1;
  color: var(--text);
}

.timeline__weekday {
  color: var(--text-tertiary);
  font-size: var(--text-xs);
  letter-spacing: var(--tracking-wide);
}

.timeline__posts {
  display: flex;
  min-width: 0;
  flex-direction: column;
  gap: var(--space-4);
}

/* 窄屏: 日期变成一行小标题, 时间轴竖线收起来 */
@media (max-width: 767px) {
  .timeline {
    gap: var(--space-6);
  }

  .timeline__group {
    grid-template-columns: minmax(0, 1fr);
    gap: var(--space-3);
  }

  .timeline__date {
    position: static;
    flex-direction: row;
    align-items: baseline;
    gap: var(--space-2);
    padding: 0 0 var(--space-2);
    text-align: left;
    border-right: 0;
    border-bottom: 1px solid var(--divider);

    &::after {
      display: none;
    }
  }

  .timeline__day {
    font-size: var(--text-lg);
  }
}
</style>
