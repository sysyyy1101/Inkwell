<template>
  <nav class="side-nav card" aria-label="版块导航">
    <!-- 和版块同款样式的"全部"入口, 选中的时候表示当前在首页看所有帖子 -->
    <RouterLink class="side-nav__item" :class="{ 'is-active': isHome }" :to="{ name: 'home' }">
      <AppIcon name="compass" :size="15" />
      <span class="side-nav__name">全部</span>
    </RouterLink>

    <div class="side-nav__section">
      <header class="side-nav__header">
        <span class="side-nav__title">版块</span>
        <button
          class="side-nav__refresh"
          type="button"
          :disabled="community.loading"
          aria-label="刷新版块列表"
          title="刷新版块列表"
          @click="community.reload()"
        >
          <AppIcon name="refresh" :size="14" />
        </button>
      </header>

      <div v-if="community.loading && !community.list.length" class="side-nav__list">
        <span v-for="index in 6" :key="index" class="skeleton side-nav__skeleton" />
      </div>

      <p v-else-if="community.error" class="side-nav__error">
        {{ community.error }}
        <button class="side-nav__retry" type="button" @click="community.reload()">重试</button>
      </p>

      <div v-else-if="!community.list.length" class="side-nav__error">还没有版块</div>

      <div v-else class="side-nav__list">
        <RouterLink
          v-for="item in community.list"
          :key="item.community_id"
          class="side-nav__item"
          :class="{ 'is-active': activeId === String(item.community_id) }"
          :to="{ name: 'community', params: { id: item.community_id } }"
        >
          <span class="side-nav__dot" :style="dotTone(item.community_name)" />
          <span class="side-nav__name">{{ item.community_name }}</span>
        </RouterLink>
      </div>
    </div>
  </nav>
</template>

<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { useRoute } from 'vue-router'

import AppIcon from '@/components/AppIcon.vue'
import { useCommunityStore } from '@/stores/community'
import { avatarTone } from '@/utils/format'

const route = useRoute()
const community = useCommunityStore()

const isHome = computed(() => route.name === 'home')
const activeId = computed(() => (route.name === 'community' ? String(route.params.id ?? '') : ''))

/** 每个版块一个稳定的低饱和色相, 具体颜色在样式里按主题取 */
function dotTone(name: string): Record<string, string> {
  return avatarTone(name)
}

onMounted(() => {
  void community.ensureLoaded()
})
</script>

<style scoped lang="scss">
.side-nav {
  display: flex;
  flex-direction: column;
  gap: 2px;
  padding: var(--space-3);
}

.side-nav__section {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.side-nav__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-2);
  margin: var(--space-3) 0 var(--space-1);
  padding: 0 12px;
}

.side-nav__title {
  color: var(--text-tertiary);
  font-size: var(--text-xs);
  font-weight: var(--weight-semibold);
  letter-spacing: var(--tracking-wide);
  text-transform: uppercase;
}

.side-nav__refresh {
  display: inline-flex;
  width: 24px;
  height: 24px;
  align-items: center;
  justify-content: center;
  border-radius: var(--radius-pill);
  color: var(--text-tertiary);
  transition: background-color var(--transition), color var(--transition);

  &:hover:not(:disabled) {
    background: var(--surface-hover);
    color: var(--text);
  }

  &:disabled {
    opacity: 0.5;
  }
}

.side-nav__list {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.side-nav__item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 9px 12px;
  border-radius: var(--radius-sm);
  color: var(--text-secondary);
  font-size: var(--text-base);
  transition: background-color var(--transition), color var(--transition), transform var(--transition);

  &:hover {
    background: var(--surface-hover);
    color: var(--text);
    transform: translateX(2px);
  }

  &.is-active {
    background: var(--accent-soft);
    color: var(--accent);
    font-weight: var(--weight-semibold);

    &:hover {
      transform: none;
    }
  }
}

.side-nav__dot {
  width: 7px;
  height: 7px;
  flex: none;
  border-radius: var(--radius-pill);
  background: hsl(var(--tone-hue, 214) calc(var(--tone-sat, 26%) + 16%) 48% / 0.85);
}

[data-theme='dark'] .side-nav__dot {
  background: hsl(var(--tone-hue, 214) calc(var(--tone-sat, 26%) + 12%) 70% / 0.8);
}

.side-nav__name {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.side-nav__skeleton {
  height: 38px;
}

.side-nav__error {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-2);
  padding: 8px 12px;
  color: var(--text-tertiary);
  font-size: var(--text-sm);
}

.side-nav__retry {
  color: var(--accent);
  font-size: var(--text-xs);
}

/* 窄屏: 版块导航变成一行可横向滚动的胶囊, 不再占满竖向空间 */
@media (max-width: 1023px) {
  .side-nav {
    flex-direction: row;
    align-items: center;
    gap: 8px;
    padding: var(--space-2);
    overflow-x: auto;
    scrollbar-width: none;

    &::-webkit-scrollbar {
      display: none;
    }
  }

  .side-nav__section {
    flex-direction: row;
    align-items: center;
    gap: 8px;
  }

  .side-nav__header {
    margin: 0;
    padding: 0;
  }

  .side-nav__title {
    display: none;
  }

  .side-nav__list {
    flex-direction: row;
    gap: 6px;
  }

  .side-nav__item {
    padding: 7px 14px;
    border-radius: var(--radius-pill);
    background: var(--surface-hover);
    font-size: var(--text-sm);
    white-space: nowrap;
  }

  .side-nav__dot {
    display: none;
  }
}
</style>
