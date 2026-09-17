<template>
  <div class="container">
    <RouterLink class="back-link" :to="{ name: 'home' }">
      <AppIcon name="arrowLeft" :size="16" />
      返回首页
    </RouterLink>

    <section v-if="loadingDetail" class="card community-hero">
      <div class="community-hero__body">
        <span class="skeleton community-hero__skeleton-title" />
        <span class="skeleton community-hero__skeleton-line" />
      </div>
    </section>

    <EmptyState
      v-else-if="detailError"
      icon="alert"
      title="版块加载失败"
      :description="detailError"
    >
      <RouterLink class="btn btn--primary btn--sm" :to="{ name: 'home' }">回到首页</RouterLink>
    </EmptyState>

    <template v-else-if="detail">
      <section class="card community-hero">
        <div class="community-hero__body">
          <span class="chip chip--accent">
            <AppIcon name="layers" :size="14" />
            版块
          </span>
          <h1 class="community-hero__title font-display">{{ detail.community_name }}</h1>
          <p class="community-hero__desc">{{ detail.introduction || '这个版块还没有简介。' }}</p>
          <div class="community-hero__meta">
            <span class="community-hero__meta-item">
              <AppIcon name="calendar" :size="15" />
              创建于 {{ formatDate(detail.create_time) }}
            </span>
            <span class="community-hero__meta-item">
              <AppIcon name="link" :size="15" />
              版块 ID {{ detail.community_id }}
            </span>
          </div>
        </div>
        <RouterLink
          class="btn btn--primary"
          :to="{ name: 'publish', query: { community: String(detail.community_id) } }"
        >
          <AppIcon name="plus" :size="16" />
          在此版块发帖
        </RouterLink>
      </section>

      <div class="layout">
        <aside class="sidebar">
          <SideNav />
        </aside>

        <div class="feed">
          <div class="feed__toolbar">
            <FeedTabs :model-value="order" :options="orderOptions" @update:model-value="handleOrderChange" />
            <button class="btn btn--ghost btn--sm" type="button" :disabled="loading" @click="refresh">
              <AppIcon name="refresh" :size="15" />
              刷新
            </button>
          </div>

          <SkeletonList v-if="loading" :count="3" />

          <EmptyState v-else-if="error" icon="alert" title="列表加载失败" :description="error">
            <button class="btn btn--primary btn--sm" type="button" @click="refresh">重新加载</button>
          </EmptyState>

          <template v-else-if="items.length">
            <div class="feed__list">
              <PostCard
                v-for="(post, index) in items"
                :key="post.post_id"
                class="enter-rise"
                :style="{ '--enter-index': Math.min(index, 8) }"
                :post="post"
                :rank="index + 1"
              />
            </div>
            <LoadMoreButton :has-more="hasMore" :loading="loadingMore" @load-more="loadMore" />
          </template>

          <EmptyState
            v-else
            icon="compass"
            title="这个版块还没有帖子"
            description="版块榜单来自 Redis, 发帖之后就会出现在这里。"
          >
            <RouterLink
              class="btn btn--primary btn--sm"
              :to="{ name: 'publish', query: { community: String(detail.community_id) } }"
            >
              来发第一帖
            </RouterLink>
          </EmptyState>
        </div>
      </div>
    </template>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useRoute } from 'vue-router'

import AppIcon from '@/components/AppIcon.vue'
import EmptyState from '@/components/EmptyState.vue'
import FeedTabs from '@/components/FeedTabs.vue'
import LoadMoreButton from '@/components/LoadMoreButton.vue'
import PostCard from '@/components/PostCard.vue'
import SideNav from '@/components/SideNav.vue'
import SkeletonList from '@/components/SkeletonList.vue'
import type { FeedTabOption } from '@/components/feedTabs'
import { ApiError, communityApi, toErrorMessage, type CommunityDetail, type OrderType } from '@/api'
import { usePostFeed } from '@/composables/usePostFeed'
import { formatDate } from '@/utils/format'

const route = useRoute()

const orderOptions: FeedTabOption[] = [
  { value: 'score', label: '热门', icon: 'flame' },
  { value: 'time', label: '最新', icon: 'clock' },
]

const communityId = computed(() => String(route.params.id ?? ''))
const detail = ref<CommunityDetail | null>(null)
const loadingDetail = ref(true)
const detailError = ref<string | null>(null)

const feed = usePostFeed({ source: 'rank', communityId })
const items = computed(() => feed.items.value)
const loading = computed(() => feed.loading.value)
const loadingMore = computed(() => feed.loadingMore.value)
const hasMore = computed(() => feed.hasMore.value)
const error = computed(() => feed.error.value)
const order = computed(() => feed.order.value)

async function loadDetail(): Promise<void> {
  loadingDetail.value = true
  detailError.value = null
  detail.value = null
  try {
    // 后端对非法的版块 ID 会返回 1009, 这里直接展示成"版块不存在"
    detail.value = await communityApi.fetchCommunityDetail(communityId.value)
  } catch (e) {
    detailError.value =
      e instanceof ApiError && e.code === 1009 ? '这个版块不存在或已被删除。' : toErrorMessage(e)
  } finally {
    loadingDetail.value = false
  }
}

async function handleOrderChange(value: string): Promise<void> {
  await feed.setOrder(value as OrderType)
}

async function refresh(): Promise<void> {
  await feed.reload()
}

async function loadMore(): Promise<void> {
  await feed.loadMore()
}

/** 同一个组件在不同版块之间跳转时, 路由参数变了要重新加载 */
watch(communityId, async () => {
  await loadDetail()
  if (!detailError.value) {
    await feed.reload()
  }
})

onMounted(async () => {
  await loadDetail()
  if (!detailError.value) {
    await feed.reload()
  }
})
</script>

<style scoped lang="scss">
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

.community-hero {
  display: flex;
  flex-direction: column;
  gap: var(--space-5);
  margin-bottom: var(--space-7);
  padding: var(--space-6);
  overflow: hidden;
  background: var(--surface);

  @media (min-width: 860px) {
    flex-direction: row;
    align-items: flex-end;
    justify-content: space-between;
  }
}

.community-hero__body {
  display: flex;
  flex-direction: column;
  gap: var(--space-3);
}

.community-hero__title {
  font-size: var(--text-2xl);
  font-weight: var(--weight-heavy);
  letter-spacing: var(--tracking-tighter);
}

.community-hero__desc {
  max-width: 620px;
  color: var(--text-secondary);
  font-size: var(--text-md);
}

.community-hero__meta {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-4);
  margin-top: var(--space-1);
}

.community-hero__meta-item {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  color: var(--text-tertiary);
  font-size: var(--text-sm);
}

.community-hero__skeleton-title {
  width: 180px;
  height: 32px;
}

.community-hero__skeleton-line {
  width: 320px;
  height: 16px;
}

.feed {
  display: flex;
  min-width: 0;
  flex-direction: column;
  gap: var(--space-4);
}

.feed__toolbar {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-3);
}

.feed__list {
  display: flex;
  flex-direction: column;
  gap: var(--space-4);
}
</style>
