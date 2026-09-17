<template>
  <div ref="root" class="menu">
    <div class="menu__trigger" :class="{ 'is-open': open }">
      <!-- 头像直接进自己的主页, 右边那半块仍然负责展开菜单 -->
      <RouterLink
        class="menu__avatar-link"
        :to="profileLink"
        :aria-label="`进入 ${auth.username} 的主页`"
        :title="`进入 ${auth.username} 的主页`"
      >
        <span class="avatar" :style="avatarStyle">{{ initial }}</span>
      </RouterLink>
      <button
        class="menu__toggle"
        type="button"
        :aria-expanded="open"
        aria-haspopup="menu"
        @click="open = !open"
      >
        <span class="menu__name">{{ auth.username }}</span>
        <AppIcon name="chevronDown" :size="14" />
      </button>
    </div>

    <Transition name="fade">
      <div v-if="open" class="menu__panel" role="menu">
        <div class="menu__header">
          <span class="menu__header-name">{{ auth.username }}</span>
          <span class="menu__header-id">ID {{ shortId }}</span>
        </div>
        <div class="divider" />
        <RouterLink class="menu__item" role="menuitem" :to="profileLink" @click="open = false">
          <AppIcon name="user" :size="16" />
          我的主页
        </RouterLink>
        <RouterLink class="menu__item" role="menuitem" :to="{ name: 'publish' }" @click="open = false">
          <AppIcon name="compose" :size="16" />
          发布帖子
        </RouterLink>
        <RouterLink class="menu__item" role="menuitem" :to="{ name: 'home' }" @click="open = false">
          <AppIcon name="home" :size="16" />
          回到首页
        </RouterLink>
        <div class="divider" />
        <button class="menu__item menu__item--danger" type="button" role="menuitem" @click="handleLogout">
          <AppIcon name="logout" :size="16" />
          退出登录
        </button>
      </div>
    </Transition>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { useRouter } from 'vue-router'

import AppIcon from '@/components/AppIcon.vue'
import { useClickOutside } from '@/composables/useClickOutside'
import { useAuthStore } from '@/stores/auth'
import { useToastStore } from '@/stores/toast'
import { useVoteStore } from '@/stores/votes'
import { avatarText, avatarTone } from '@/utils/format'

const auth = useAuthStore()
const toast = useToastStore()
const votes = useVoteStore()
const router = useRouter()

const open = ref(false)
const root = ref<HTMLElement | null>(null)

useClickOutside(root, () => {
  open.value = false
})

const initial = computed(() => avatarText(auth.username ?? ''))
const avatarStyle = computed(() => avatarTone(auth.username ?? ''))
const shortId = computed(() => (auth.userID ? `${auth.userID.slice(0, 6)}…${auth.userID.slice(-4)}` : ''))
/** 自己的主页; 没有 userID 时回首页, 避免拼出一个空的 /user/ 链接 */
const profileLink = computed(() =>
  auth.userID ? { name: 'user', params: { id: auth.userID } } : { name: 'home' },
)

function handleLogout(): void {
  const userID = auth.userID
  if (userID) {
    // 清掉本机记录的投票状态, 避免下一个登录的人看到上一个用户的投票痕迹
    votes.clearUser(userID)
  }
  auth.logout()
  open.value = false
  toast.notifyThenRedirect('success', '已退出登录', () => {
    void router.push({ name: 'home' })
  })
}
</script>

<style scoped lang="scss">
.menu {
  position: relative;
}

.menu__trigger {
  display: inline-flex;
  align-items: center;
  height: 40px;
  padding: 0 4px;
  border: 1px solid var(--border);
  border-radius: var(--radius-pill);
  background: var(--surface);
  transition: border-color var(--transition), background-color var(--transition);

  &:hover,
  &.is-open {
    border-color: var(--border-strong);
    background: var(--surface-hover);
  }
}

.menu__avatar-link {
  display: inline-flex;
  align-items: center;
  border-radius: var(--radius-pill);
}

.menu__toggle {
  display: inline-flex;
  height: 100%;
  align-items: center;
  gap: 8px;
  padding: 0 10px 0 8px;
  color: var(--text-secondary);
  font-size: var(--text-sm);
  font-weight: var(--weight-medium);
  transition: color var(--transition);

  &:hover {
    color: var(--text);
  }
}

.menu__name {
  max-width: 96px;
  overflow: hidden;
  color: var(--text);
  text-overflow: ellipsis;
  white-space: nowrap;
}
/* 窄屏: 顶栏右侧有四个元素, 收起用户名只留头像, 避免把顶栏挤变形 */
@media (max-width: 479px) {
  .menu__name {
    display: none;
  }

  .menu__toggle {
    padding: 0 8px 0 4px;
  }
}

.menu__panel {
  position: absolute;
  right: 0;
  top: calc(100% + 10px);
  z-index: 30;
  width: 232px;
  padding: 6px;
  border: 1px solid var(--border);
  border-radius: var(--radius-md);
  background: var(--surface);
  box-shadow: var(--shadow-md);
}

.menu__header {
  display: flex;
  flex-direction: column;
  gap: 2px;
  padding: 10px 12px;
}

.menu__header-name {
  font-size: var(--text-base);
  font-weight: var(--weight-bold);
}

.menu__header-id {
  color: var(--text-tertiary);
  font-family: var(--font-mono);
  font-size: var(--text-xs);
}

.menu__item {
  display: flex;
  width: 100%;
  align-items: center;
  gap: 10px;
  padding: 10px 12px;
  border-radius: var(--radius-xs);
  color: var(--text-secondary);
  font-size: var(--text-base);
  text-align: left;
  transition: background-color var(--transition), color var(--transition);

  &:hover {
    background: var(--surface-hover);
    color: var(--text);
  }
}

.menu__item--danger:hover {
  background: var(--danger-soft);
  color: var(--danger);
}
</style>
