/**
 * 登录态的单一数据源。
 *
 * 放在 api 层而不是 store 里, 是因为 axios 拦截器(刷新 token 时)也要读写它,
 * 而拦截器不能依赖 Pinia store, 否则会形成 store -> api -> store 的循环依赖。
 * store 通过 subscribeSession 订阅变化, 因此 token 刷新后界面状态会自动同步。
 */
export interface AuthSession {
  accessToken: string | null
  refreshToken: string | null
  userID: string | null
  username: string | null
}

const STORAGE_KEY = 'inkwell.auth'

export function emptySession(): AuthSession {
  return {
    accessToken: null,
    refreshToken: null,
    userID: null,
    username: null,
  }
}

function pickString(value: unknown): string | null {
  return typeof value === 'string' && value.length > 0 ? value : null
}

function readSession(): AuthSession {
  try {
    const raw = localStorage.getItem(STORAGE_KEY)
    if (!raw) {
      return emptySession()
    }
    const parsed = JSON.parse(raw) as Partial<AuthSession>
    return {
      accessToken: pickString(parsed.accessToken),
      refreshToken: pickString(parsed.refreshToken),
      userID: pickString(parsed.userID),
      username: pickString(parsed.username),
    }
  } catch {
    // 本地数据被改坏时降级成"未登录", 而不是让整个应用报错
    localStorage.removeItem(STORAGE_KEY)
    return emptySession()
  }
}

let current: AuthSession = readSession()

type SessionListener = (session: AuthSession) => void
const listeners = new Set<SessionListener>()

export function getSession(): AuthSession {
  return current
}

/** 更新登录态并持久化; 传 null 表示登出 */
export function setSession(next: Partial<AuthSession> | null): AuthSession {
  current = next ? { ...emptySession(), ...next } : emptySession()
  try {
    if (current.accessToken) {
      localStorage.setItem(STORAGE_KEY, JSON.stringify(current))
    } else {
      localStorage.removeItem(STORAGE_KEY)
    }
  } catch {
    // 无痕模式等场景下 localStorage 可能不可写, 此时只保留内存中的登录态
  }
  listeners.forEach((listener) => listener(current))
  return current
}

export function subscribeSession(listener: SessionListener): () => void {
  listeners.add(listener)
  return () => {
    listeners.delete(listener)
  }
}

export function getAccessToken(): string | null {
  return current.accessToken
}

export function getRefreshToken(): string | null {
  return current.refreshToken
}

let sessionExpiredHandler: (() => void) | null = null

/** 由 main.ts 注册: refresh token 也失效时清空登录态并跳回登录页 */
export function setSessionExpiredHandler(handler: (() => void) | null): void {
  sessionExpiredHandler = handler
}

export function notifySessionExpired(): void {
  sessionExpiredHandler?.()
}
