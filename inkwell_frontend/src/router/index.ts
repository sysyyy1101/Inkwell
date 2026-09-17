import { createRouter, createWebHistory } from 'vue-router'

import { getSession } from '@/api/session'

const SITE_NAME = 'Inkwell'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/',
      name: 'home',
      component: () => import('@/views/HomeView.vue'),
      meta: { title: '首页' },
    },
    {
      path: '/post/:id',
      name: 'post',
      component: () => import('@/views/PostDetailView.vue'),
      meta: { title: '帖子详情' },
    },
    {
      path: '/community/:id',
      name: 'community',
      component: () => import('@/views/CommunityView.vue'),
      meta: { title: '版块' },
    },
    {
      path: '/user/:id',
      name: 'user',
      component: () => import('@/views/UserView.vue'),
      meta: { title: '用户主页' },
    },
    {
      path: '/publish',
      name: 'publish',
      component: () => import('@/views/PublishView.vue'),
      meta: { title: '发布帖子', requireAuth: true },
    },
    {
      path: '/login',
      name: 'login',
      component: () => import('@/views/LoginView.vue'),
      meta: { title: '登录', guestOnly: true },
    },
    {
      path: '/signup',
      name: 'signup',
      component: () => import('@/views/SignupView.vue'),
      meta: { title: '注册', guestOnly: true },
    },
    {
      path: '/:pathMatch(.*)*',
      name: 'not-found',
      component: () => import('@/views/NotFoundView.vue'),
      meta: { title: '页面不存在' },
    },
  ],
  scrollBehavior(_to, _from, savedPosition) {
    return savedPosition ?? { top: 0 }
  },
})

// 需要登录的页面: 未登录时跳登录页并带上回跳地址
router.beforeEach((to) => {
  const isLogin = Boolean(getSession().accessToken)
  if (to.meta.requireAuth && !isLogin) {
    return { name: 'login', query: { redirect: to.fullPath } }
  }
  if (to.meta.guestOnly && isLogin) {
    return { name: 'home' }
  }
  return true
})

router.afterEach((to) => {
  const title = to.meta.title as string | undefined
  document.title = title ? `${title} · ${SITE_NAME}` : `${SITE_NAME} · 技术社区`
})

export default router
