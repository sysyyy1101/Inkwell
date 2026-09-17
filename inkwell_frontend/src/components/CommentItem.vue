<template>
  <div class="comment" :class="[`comment--depth-${Math.min(depth, 2)}`]">
    <RouterLink
      class="comment__avatar-link"
      :to="userLink"
      :title="`查看 ${comment.author_name || '匿名用户'} 的主页`"
    >
      <span class="avatar avatar--lg" :style="avatarStyle">{{ initial }}</span>
    </RouterLink>

    <div class="comment__body">
      <header class="comment__header">
        <span class="comment__author">
          <RouterLink class="comment__author-link" :to="userLink">
            {{ comment.author_name || '匿名用户' }}
          </RouterLink>
          <span v-if="isMine" class="chip chip--accent">我</span>
          <span v-if="replyToName" class="comment__reply-to">回复 {{ replyToName }}</span>
        </span>
        <time class="comment__time" :datetime="comment.create_time" :title="formatDateTime(comment.create_time)">
          {{ formatRelativeTime(comment.create_time) }}
        </time>
        <button class="comment__action" type="button" @click="emit('reply', comment)">
          <AppIcon name="reply" :size="14" />
          回复
        </button>
      </header>

      <p class="comment__content">{{ comment.content }}</p>

      <CommentComposer
        v-if="replyToId === comment.comment_id"
        class="comment__composer"
        :post-id="postId"
        :parent-id="comment.comment_id"
        :reply-to-name="comment.author_name"
        placeholder="回复这条评论…"
        compact
        auto-focus
        @created="(id) => emit('created', id)"
        @cancel="emit('cancel-reply')"
      />

      <div v-if="children.length" class="comment__children">
        <CommentItem
          v-for="(child, index) in children"
          :key="child.comment_id"
          class="enter-rise"
          :style="{ '--enter-index': Math.min(index, 4) }"
          :comment="child"
          :children-map="childrenMap"
          :comment-index="commentIndex"
          :depth="depth + 1"
          :post-id="postId"
          :reply-to-id="replyToId"
          :current-user-id="currentUserId"
          @reply="(target) => emit('reply', target)"
          @cancel-reply="emit('cancel-reply')"
          @created="(id) => emit('created', id)"
        />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'

import AppIcon from '@/components/AppIcon.vue'
import CommentComposer from '@/components/CommentComposer.vue'
import type { PostComment } from '@/api'
import { avatarText, avatarTone, formatDateTime, formatRelativeTime } from '@/utils/format'

const props = defineProps<{
  comment: PostComment
  /** 评论树: parent_id -> 子评论列表 */
  childrenMap: Map<string, PostComment[]>
  /** comment_id -> 评论, 用来查"被回复的是谁" */
  commentIndex: Map<string, PostComment>
  depth: number
  postId: string
  replyToId: string | null
  currentUserId: string | null
}>()

const emit = defineEmits<{
  (event: 'reply', target: PostComment): void
  (event: 'cancel-reply'): void
  /** 子评论发布成功, 冒泡给列表去拉取权威数据 */
  (event: 'created', commentId: string): void
}>()

const children = computed(() => props.childrenMap.get(props.comment.comment_id) ?? [])
const isMine = computed(() => Boolean(props.currentUserId) && props.currentUserId === props.comment.author_id)
const initial = computed(() => avatarText(props.comment.author_name))
const avatarStyle = computed(() => avatarTone(props.comment.author_name))
/** 作者主页链接: 点自己的头像回自己的主页, 点别人的看别人的帖子 */
const userLink = computed(() => ({ name: 'user', params: { id: props.comment.author_id } }))

/** 被回复人的名字: 从全文索引里找父评论的作者 */
const replyToName = computed(() =>
  props.comment.parent_id === '0'
    ? ''
    : (props.commentIndex.get(props.comment.parent_id)?.author_name ?? ''),
)
</script>

<style scoped lang="scss">
.comment {
  display: flex;
  gap: var(--space-3);
}

.comment__avatar-link {
  display: inline-flex;
  align-self: flex-start;
  border-radius: var(--radius-pill);
}

.comment__body {
  display: flex;
  min-width: 0;
  flex: 1;
  flex-direction: column;
  gap: 6px;
}

.comment__header {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 8px;
}

.comment__author {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  color: var(--text);
  font-size: var(--text-sm);
  font-weight: var(--weight-bold);
}

.comment__author-link {
  transition: color var(--transition);

  &:hover {
    color: var(--accent);
  }
}

.comment__reply-to {
  color: var(--text-tertiary);
  font-size: var(--text-xs);
  font-weight: var(--weight-regular);
}

.comment__time {
  color: var(--text-tertiary);
  font-size: var(--text-xs);
}

.comment__action {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 3px 9px;
  border-radius: var(--radius-pill);
  color: var(--text-tertiary);
  font-size: var(--text-xs);
  transition: background-color var(--transition), color var(--transition);

  &:hover {
    background: var(--surface-hover);
    color: var(--accent);
  }
}

.comment__content {
  color: var(--text-secondary);
  font-size: var(--text-base);
  line-height: var(--leading-relaxed);
  overflow-wrap: anywhere;
  white-space: pre-wrap;
}

.comment__composer {
  margin-top: var(--space-2);
}

.comment__children {
  display: flex;
  flex-direction: column;
  gap: var(--space-5);
  margin-top: var(--space-4);
  padding-left: var(--space-4);
  border-left: 2px solid var(--divider);
}

/* 更深的层级不再继续缩进, 只保留左边的引导线, 避免窄屏被压得太扁 */
.comment--depth-2 .comment__children {
  padding-left: var(--space-2);
}
</style>
