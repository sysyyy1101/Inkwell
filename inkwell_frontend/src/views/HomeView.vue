<template>
  <div class="container">
    <section class="hero">
      <h1 class="hero__title font-display">把想法写成帖子, 让讨论沉淀下来。</h1>
      <p class="hero__desc">
        榜单由 Reddit 热度算法实时计算, 投票与评论数存在 Redis, 正文落在 MySQL。选择你感兴趣的版块,
        开始一次认真的讨论。
      </p>
      <div class="hero__actions">
        <RouterLink v-if="auth.isLogin" class="btn btn--primary" :to="{ name: 'publish' }">
          <AppIcon name="plus" :size="16" />
          发布帖子
        </RouterLink>
        <RouterLink v-else class="btn btn--primary" :to="{ name: 'signup' }">免费注册</RouterLink>
        <RouterLink class="btn btn--link" :to="{ name: 'community', params: { id: 1 } }">
          逛逛版块
          <AppIcon name="chevronRight" :size="15" />
        </RouterLink>
      </div>
    </section>

    <div class="layout">
      <aside class="sidebar">
        <SideNav />

        <section class="card cta">
          <div class="card__body">
            <h2 class="cta__title">{{ auth.isLogin ? '有想法就说出来' : '加入 Inkwell' }}</h2>
            <p class="cta__desc">
              {{
                auth.isLogin
                  ? '发帖后你会自动获得一票赞成, 帖子会立刻进入 Redis 榜单。'
                  : '注册后即可发帖、投票、评论, 参与版块里的讨论。'
              }}
            </p>
            <RouterLink
              class="btn btn--primary btn--block"
              :to="{ name: auth.isLogin ? 'publish' : 'signup' }"
            >
              <AppIcon :name="auth.isLogin ? 'plus' : 'user'" :size="16" />
              {{ auth.isLogin ? '发布帖子' : '立即注册' }}
            </RouterLink>
          </div>
        </section>
      </aside>

      <div class="feed">
        <div class="feed__toolbar">
          <FeedTabs :model-value="tab" :options="tabOptions" @update:model-value="handleTabChange" />
          <button class="btn btn--ghost btn--sm" type="button" :disabled="loading" @click="refresh">
            <AppIcon name="refresh" :size="15" />
            刷新
          </button>
        </div>

        <p v-if="notice" class="feed__notice">
          <AppIcon name="info" :size="16" />
          {{ notice }}
        </p>

        <SkeletonList v-if="loading" :count="4" />

        <EmptyState
          v-else-if="error"
          icon="alert"
          title="列表加载失败"
          :description="error"
        >
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
              :rank="showRank ? index + 1 : undefined"
            />
          </div>
          <LoadMoreButton :has-more="hasMore" :loading="loadingMore" @load-more="loadMore" />
        </template>

        <EmptyState
          v-else
          icon="compass"
          title="这个榜单还没有帖子"
          description="榜单来自 Redis, 发帖或投票之后才会出现在这里。也可以直接看看全部帖子。"
        >
          <button class="btn btn--primary btn--sm" type="button" @click="handleTabChange('new')">
            查看全部帖子
          </button>
        </EmptyState>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'

import AppIcon from '@/components/AppIcon.vue'
import EmptyState from '@/components/EmptyState.vue'
import FeedTabs from '@/components/FeedTabs.vue'
import LoadMoreButton from '@/components/LoadMoreButton.vue'
import PostCard from '@/components/PostCard.vue'
import SideNav from '@/components/SideNav.vue'
import SkeletonList from '@/components/SkeletonList.vue'
import type { FeedTabOption } from '@/components/feedTabs'
import type { OrderType } from '@/api'
import { usePostFeed } from '@/composables/usePostFeed'
import { useAuthStore } from '@/stores/auth'

type TabKey = 'hot' | 'new' | 'time'

const auth = useAuthStore()

const tabOptions: FeedTabOption[] = [
  { value: 'hot', label: '热门', icon: 'flame' },
  { value: 'new', label: '最新', icon: 'clock' },
  { value: 'time', label: '时间榜', icon: 'trend' },
]

const tab = ref<TabKey>('hot')
const notice = ref('')

/** 榜单(Redis)与全部帖子(MySQL)各用一个数据源, 切来切去不会丢已经加载的内容 */
const rankFeed = usePostFeed({ source: 'rank' })
const mysqlFeed = usePostFeed({ source: 'mysql' })

const activeFeed = computed(() => (tab.value === 'new' ? mysqlFeed : rankFeed))
const items = computed(() => activeFeed.value.items.value)
const loading = computed(() => activeFeed.value.loading.value)
const loadingMore = computed(() => activeFeed.value.loadingMore.value)
const hasMore = computed(() => activeFeed.value.hasMore.value)
const error = computed(() => activeFeed.value.error.value)
const showRank = computed(() => tab.value !== 'new')

/** 按当前 Tab 决定要不要请求接口, 已经加载过的直接复用 */
async function ensureLoaded(): Promise<void> {
  if (tab.value === 'new') {
    if (!mysqlFeed.items.value.length) {
      await mysqlFeed.reload()
    }
    return
  }
  const order: OrderType = tab.value === 'hot' ? 'score' : 'time'
  if (rankFeed.order.value !== order || !rankFeed.items.value.length) {
    rankFeed.order.value = order
    await rankFeed.reload()
  }
}

async function handleTabChange(value: string): Promise<void> {
  tab.value = value as TabKey
  notice.value = ''
  await ensureLoaded()
}

async function refresh(): Promise<void> {
  notice.value = ''
  await activeFeed.value.reload()
}

async function loadMore(): Promise<void> {
  await activeFeed.value.loadMore()
}

onMounted(async () => {
  await ensureLoaded()
  // 榜单是 Redis 数据, 清空过 Redis 或者刚初始化数据库时会为空,
  // 这时自动切到"最新"(直接读 MySQL), 保证首页第一眼有内容。
  if (tab.value === 'hot' && !rankFeed.loading.value && rankFeed.items.value.length === 0) {
    tab.value = 'new'
    notice.value = '榜单暂时没有数据(帖子需要发帖/投票后才会进入 Redis 榜单), 已为你切换到最新帖子。'
    await ensureLoaded()
  }
})
</script>

<style scoped lang="scss">
.hero {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: var(--space-4);
  max-width: 720px;
  margin-bottom: var(--space-8);
}

.hero__title {
  font-size: var(--text-2xl);
  font-weight: var(--weight-heavy);
  letter-spacing: var(--tracking-tighter);
  line-height: var(--leading-tight);

  @media (min-width: 768px) {
    font-size: var(--text-3xl);
  }
}

.hero__desc {
  max-width: 620px;
  color: var(--text-secondary);
  font-size: var(--text-md);
  line-height: var(--leading-relaxed);
}

.hero__actions {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: var(--space-5);
  margin-top: var(--space-2);
}

.feed {
  display: flex;
  flex-direction: column;
  gap: var(--space-4);
  min-width: 0;
}

.feed__toolbar {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-3);
}

.feed__notice {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 12px 16px;
  border: 1px solid var(--border);
  border-radius: var(--radius-md);
  background: var(--accent-soft);
  color: var(--accent);
  font-size: var(--text-sm);
}

.feed__list {
  display: flex;
  flex-direction: column;
  gap: var(--space-4);
}

.cta__title {
  font-size: var(--text-md);
  font-weight: var(--weight-bold);
}

.cta__desc {
  margin: 8px 0 var(--space-4);
  color: var(--text-secondary);
  font-size: var(--text-sm);
  line-height: 1.6;
}

/* 窄屏下版块导航已经是一行可横滑的胶囊, 发帖入口交给顶栏按钮, 这里不再占高度 */
.cta {
  display: none;

  @media (min-width: 1024px) {
    display: block;
  }
}
</style>
