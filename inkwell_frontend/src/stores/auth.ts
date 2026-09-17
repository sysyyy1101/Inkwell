import { computed, reactive } from 'vue'
import { defineStore } from 'pinia'

import { authApi } from '@/api'
import { emptySession, getSession, setSession, subscribeSession, type AuthSession } from '@/api/session'
import type { LoginPayload, SignUpPayload } from '@/api'

/**
 * 登录态。真正的数据放在 api/session 里(拦截器也要用),
 * 这里只是把它包装成响应式状态, 保证 token 被自动刷新后界面能同步更新。
 */
export const useAuthStore = defineStore('auth', () => {
  const session = reactive<AuthSession>({ ...getSession() })

  // session 变化(登录/登出/拦截器刷新 token)时同步到响应式状态
  subscribeSession((next) => {
    Object.assign(session, next)
  })

  const isLogin = computed(() => Boolean(session.accessToken))
  const userID = computed(() => session.userID)
  const username = computed(() => session.username)

  async function login(payload: LoginPayload): Promise<void> {
    const result = await authApi.login(payload)
    setSession({
      accessToken: result.accessToken,
      refreshToken: result.refreshToken,
      userID: result.userID,
      username: result.username,
    })
  }

  async function register(payload: SignUpPayload): Promise<void> {
    await authApi.signUp(payload)
  }

  function logout(): void {
    setSession(emptySession())
  }

  return {
    session,
    isLogin,
    userID,
    username,
    login,
    register,
    logout,
  }
})
