<template>
  <header class="header" :class="{ 'header--scrolled': scrolled }">
    <span class="header__glass" aria-hidden="true" />
    <div class="container header__inner">
      <!-- 左: 品牌字标单独占一侧 -->
      <RouterLink class="brand" :to="{ name: 'home' }" aria-label="Inkwell 首页" title="回到首页">
        <span class="brand__word font-brand">Inkwell</span>
        <span class="brand__ink" aria-hidden="true" />
      </RouterLink>

      <!-- 右: 房子 -> 发帖 -> 个人 -> 深色模式 -->
      <nav class="header__actions" aria-label="主导航">
        <RouterLink
          class="icon-button"
          :class="{ 'is-active': isHome }"
          :to="{ name: 'home' }"
          aria-label="首页"
          title="首页"
        >
          <AppIcon name="home" :size="19" />
        </RouterLink>

        <template v-if="auth.isLogin">
          <RouterLink
            class="icon-button icon-button--accent"
            :to="{ name: 'publish' }"
            aria-label="发布帖子"
            title="发布帖子"
          >
            <AppIcon name="compose" :size="19" />
          </RouterLink>
          <UserMenu />
        </template>

        <template v-else>
          <RouterLink class="btn btn--ghost btn--sm" :to="{ name: 'login' }">登录</RouterLink>
          <RouterLink class="btn btn--primary btn--sm" :to="{ name: 'signup' }">注册</RouterLink>
        </template>

        <ThemeToggle />
      </nav>
    </div>
  </header>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import { useRoute } from 'vue-router'

import AppIcon from '@/components/AppIcon.vue'
import ThemeToggle from '@/components/ThemeToggle.vue'
import UserMenu from '@/components/UserMenu.vue'
import { useAuthStore } from '@/stores/auth'

const auth = useAuthStore()
const route = useRoute()

const isHome = computed(() => route.name === 'home')

/**
 * 滚动到一定距离后顶栏加深一点并浮起一层很浅的阴影:
 * 既让玻璃层和内容区分开, 又不会有"一条黑带"的割裂感。
 * 用 rAF 节流, 避免滚动时每帧都触发样式计算。
 */
const scrolled = ref(false)
let ticking = false

function handleScroll(): void {
  if (ticking) {
    return
  }
  ticking = true
  window.requestAnimationFrame(() => {
    scrolled.value = window.scrollY > 8
    ticking = false
  })
}

onMounted(() => {
  window.addEventListener('scroll', handleScroll, { passive: true })
  handleScroll()
})

onBeforeUnmount(() => {
  window.removeEventListener('scroll', handleScroll)
})
</script>

<style scoped lang="scss">
.header {
  position: sticky;
  top: 0;
  z-index: 20;
  /* 不要背景、不要边框、不要阴影: 分割感全部交给下面那层带蒙版的模糊 */
  background: transparent;
}

/**
 * iOS 26 那种"渐变模糊"顶栏:
 * 模糊层比内容区高出一截, 再用 mask 让模糊和底色一起向下淡出,
 * 所以内容滚到下面会被自然雾化, 而不是出现一条硬边。
 */
.header__glass {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  height: calc(var(--header-height) + 20px);
  z-index: -1;
  pointer-events: none;
  background: var(--header-bg);
  -webkit-backdrop-filter: blur(20px) saturate(180%);
  backdrop-filter: blur(20px) saturate(180%);
  -webkit-mask-image: linear-gradient(
    180deg,
    rgba(0, 0, 0, 1) 0%,
    rgba(0, 0, 0, 0.94) 56%,
    rgba(0, 0, 0, 0) 100%
  );
  mask-image: linear-gradient(
    180deg,
    rgba(0, 0, 0, 1) 0%,
    rgba(0, 0, 0, 0.94) 56%,
    rgba(0, 0, 0, 0) 100%
  );
  transition: background var(--transition);
}

.header--scrolled .header__glass {
  background: var(--header-bg-scrolled);
}

.header__inner {
  display: flex;
  height: var(--header-height);
  align-items: center;
  justify-content: space-between;
  gap: var(--space-4);
}


.brand {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  padding: 4px 12px 4px 4px;
  border-radius: var(--radius-pill);
  transition: background-color var(--transition), transform var(--transition);

  &:hover {
    background: var(--surface-hover);
  }

  &:active {
    transform: scale(0.97);
  }
}

.brand__word {
  font-size: 21px;
  font-weight: 600;
  letter-spacing: 0.055em;
  line-height: 1.1;
  /* 墨色渐变的文字: 从强调蓝过渡到正文墨色, 像刚蘸过墨 */
  background-image: linear-gradient(104deg, var(--accent) 0%, var(--accent) 22%, var(--text) 78%);
  -webkit-background-clip: text;
  background-clip: text;
  color: transparent;
  -webkit-text-fill-color: transparent;
  /* 极小的一点描边感, 让字在小尺寸下也立得住(不影响渐变填充) */
  text-shadow: 0 0 0.01px var(--accent);
}

/* 字标末尾的"墨点": 呼应 Inkwell 的墨水瓶意象 */
.brand__ink {
  width: 6px;
  height: 6px;
  flex: none;
  margin-left: 1px;
  border-radius: 50% 50% 50% 0;
  transform: rotate(-45deg);
  background: var(--accent);
  box-shadow: 0 0 0 3px var(--accent-soft);
}

@media (max-width: 479px) {
  .brand {
    padding-right: 8px;
  }

  .brand__word {
    font-size: 18px;
    letter-spacing: 0.03em;
  }

  .brand__ink {
    width: 5px;
    height: 5px;
  }
}


.header__actions {
  display: flex;
  align-items: center;
  gap: var(--space-2);
}

</style>
