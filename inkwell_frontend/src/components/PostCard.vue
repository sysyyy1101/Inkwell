<template>
  <article class="post-card card card--interactive">
    <VoteControl
      class="post-card__vote"
      :post-id="post.post_id"
      :count="localCount"
      :direction="direction"
      :disabled="!voteOpen"
      :disabled-reason="voteDisabledReason"
      size="sm"
      layout="vertical"
      @change="handleVoteChange"
    />

    <div class="post-card__main">
      <div class="post-card__meta">
        <span v-if="typeof rank === 'number'" class="post-card__rank">{{ rank }}</span>
        <RouterLink class="chip" :to="{ name: 'community', params: { id: post.community_id } }">
          {{ post.community_name || '未分类' }}
        </RouterLink>
        <span v-if="showTime" class="post-card__time">{{ formatRelativeTime(post.create_time) }}</span>
        <span v-if="!voteOpen" class="chip chip--warning">已停止投票</span>
      </div>

      <h3 class="post-card__title">
        <RouterLink :to="{ name: 'post', params: { id: post.post_id } }">{{ post.title }}</RouterLink>
      </h3>

      <p v-if="summary" class="post-card__summary">{{ summary }}</p>

      <div class="post-card__footer">
        <RouterLink
          class="post-card__author"
          :to="{ name: 'user', params: { id: post.author_id } }"
          :title="`查看 ${post.author_name || '匿名用户'} 的主页`"
        >
          <span class="avatar avatar--sm" :style="avatarStyle">{{ initial }}</span>
          {{ post.author_name || '匿名用户' }}
        </RouterLink>

        <span class="post-card__stats">
          <span class="post-card__stat">
            <AppIcon name="trend" :size="15" />
            {{ formatCount(localCount) }}
          </span>
          <RouterLink
            class="post-card__stat post-card__stat--link"
            :to="{ name: 'post', params: { id: post.post_id }, hash: '#comments' }"
          >
            <AppIcon name="message" :size="15" />
            {{ formatCount(post.comment_num) }}
          </RouterLink>
        </span>
      </div>
    </div>
  </article>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'

import AppIcon from '@/components/AppIcon.vue'
import VoteControl from '@/components/VoteControl.vue'
import type { PostListItem, VoteDirection } from '@/api'
import { useAuthStore } from '@/stores/auth'
import { useVoteStore } from '@/stores/votes'
import {
  avatarText,
  avatarTone,
  formatCount,
  formatRelativeTime,
  isVoteWindowOpen,
  truncateText,
} from '@/utils/format'

const props = withDefaults(
  defineProps<{
    post: PostListItem
    /** 榜单排名, 传了才显示序号 */
    rank?: number
    /** 卡片上是否显示相对时间; 用户主页的日期轴已经按天展示过, 那里传 false 避免重复 */
    showTime?: boolean
  }>(),
  {
    rank: undefined,
    showTime: true,
  },
)

const auth = useAuthStore()
const votes = useVoteStore()

/** 票数: 后端返回的值 + 本地投票后的增量, 列表刷新时同步回后端值 */
const localCount = ref(props.post.vote_num)
watch(
  () => props.post.vote_num,
  (value) => {
    localCount.value = value
  },
)

const direction = computed<VoteDirection>(() => {
  const stored = votes.get(auth.userID, props.post.post_id)
  if (stored !== null) {
    return stored
  }
  // 后端发帖时会把作者的赞成票写进 Redis, 所以作者本人的初始状态是"已赞成"
  return auth.userID && auth.userID === props.post.author_id ? 1 : 0
})

const voteOpen = computed(() => isVoteWindowOpen(props.post.create_time))
const voteDisabledReason = computed(() => (voteOpen.value ? '' : '发帖已超过 7 天, 不能再投票'))

const summary = computed(() => truncateText(props.post.summary || '', 140))
const initial = computed(() => avatarText(props.post.author_name))
const avatarStyle = computed(() => avatarTone(props.post.author_name))

function handleVoteChange(payload: { direction: VoteDirection; count: number }): void {
  localCount.value = payload.count
  if (auth.userID) {
    votes.record(auth.userID, props.post.post_id, payload.direction)
  }
}
</script>

<style scoped lang="scss">
.post-card {
  display: flex;
  gap: var(--space-4);
  padding: var(--space-5);
}

.post-card__vote {
  align-self: flex-start;
}

.post-card__main {
  display: flex;
  min-width: 0;
  flex: 1;
  flex-direction: column;
  gap: 10px;
}

.post-card__meta {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 8px;
}

.post-card__rank {
  display: inline-flex;
  width: 22px;
  height: 22px;
  align-items: center;
  justify-content: center;
  border-radius: var(--radius-xs);
  background: var(--surface-hover);
  color: var(--text-tertiary);
  font-size: var(--text-xs);
  font-weight: var(--weight-bold);
}

.post-card__time {
  color: var(--text-tertiary);
  font-size: var(--text-xs);
}

.post-card__title {
  font-size: var(--text-md);
  font-weight: var(--weight-bold);
  letter-spacing: -0.02em;
  line-height: 1.45;
  overflow-wrap: anywhere;

  a {
    background-image: linear-gradient(var(--accent), var(--accent));
    background-position: 0 100%;
    background-repeat: no-repeat;
    background-size: 0 1px;
    transition: background-size var(--transition), color var(--transition);

    &:hover {
      color: var(--accent);
      background-size: 100% 1px;
    }
  }
}

.post-card__summary {
  color: var(--text-secondary);
  font-size: var(--text-base);
  line-height: var(--leading-normal);
  overflow-wrap: anywhere;
}

.post-card__footer {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-3);
  margin-top: 2px;
}

.post-card__author {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  color: var(--text-secondary);
  font-size: var(--text-sm);
  transition: color var(--transition);

  &:hover {
    color: var(--accent);
  }
}

.post-card__stats {
  display: inline-flex;
  align-items: center;
  gap: var(--space-4);
  color: var(--text-tertiary);
  font-size: var(--text-sm);
  font-variant-numeric: tabular-nums;
}

.post-card__stat {
  display: inline-flex;
  align-items: center;
  gap: 6px;
}

.post-card__stat--link {
  transition: color var(--transition);

  &:hover {
    color: var(--accent);
  }
}
</style>
