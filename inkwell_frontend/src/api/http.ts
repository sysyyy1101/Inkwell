import axios, { AxiosError, type AxiosInstance, type AxiosRequestConfig } from 'axios'

import {
  CODE,
  type ApiEnvelope,
  type RefreshTokenResult,
} from './types'
import {
  getAccessToken,
  getRefreshToken,
  getSession,
  notifySessionExpired,
  setSession,
} from './session'

/** 接口前缀, 开发环境由 Vite 代理到后端, 生产环境可用 VITE_API_BASE 覆盖 */
const baseURL = (import.meta.env.VITE_API_BASE || '/api/v1').replace(/\/+$/, '')

/** 业务错误: code 是后端返回的响应码, message 是后端返回的中文提示 */
export class ApiError extends Error {
  readonly code: number

  constructor(code: number, message: string) {
    super(message)
    this.name = 'ApiError'
    this.code = code
  }
}

/** 把任意异常转换成可以直接展示给用户的文案 */
export function toErrorMessage(error: unknown): string {
  if (error instanceof ApiError) {
    return error.message
  }
  if (error instanceof AxiosError) {
    if (error.code === 'ECONNABORTED') {
      return '请求超时, 请稍后重试'
    }
    // 不写死后端地址: 展示实际使用的接口前缀, 换环境(代理/网关/域名)时提示依然准确
    return `无法连接后端服务 (${baseURL}), 请稍后重试`
  }
  if (error instanceof Error) {
    return error.message
  }
  return '请求失败, 请稍后重试'
}

interface RetryableConfig extends AxiosRequestConfig {
  /** 该请求是否跳过"token 失效自动刷新"逻辑 */
  skipAuthRefresh?: boolean
  /** 该请求是否是刷新 token 之后的重放, 用于避免死循环 */
  retried?: boolean
}

const instance: AxiosInstance = axios.create({
  baseURL,
  timeout: 15000,
  headers: {
    'Content-Type': 'application/json;charset=utf-8',
  },
})

instance.interceptors.request.use((config) => {
  const token = getAccessToken()
  if (token) {
    config.headers.set('Authorization', `Bearer ${token}`)
  }
  return config
})

/** 并发请求同时遇到 token 过期时, 只发一次刷新请求, 其余请求等这一次的结果 */
let refreshing: Promise<unknown> | null = null

function refreshAccessToken(): Promise<unknown> {
  if (refreshing) {
    return refreshing
  }
  const currentAccessToken = getAccessToken()
  const currentRefreshToken = getRefreshToken()
  if (!currentRefreshToken) {
    notifySessionExpired()
    return Promise.reject(new ApiError(CODE.invalidToken, '登录已过期, 请重新登录'))
  }

  const task = instance
    .request<ApiEnvelope<RefreshTokenResult>>({
      method: 'get',
      url: '/refresh_token',
      params: { refresh_token: currentRefreshToken },
      // access token 可能已经过期, 但刷新接口要求仍然带上它
      headers: currentAccessToken ? { Authorization: `Bearer ${currentAccessToken}` } : undefined,
      skipAuthRefresh: true,
    } as RetryableConfig)
    .then((response) => {
      const body = response.data
      if (!body || body.code !== CODE.success || !body.data) {
        throw new ApiError(body?.code ?? CODE.invalidToken, body?.message || '登录已过期')
      }
      // 刷新成功后更新登录态: 拦截器改的是 api/session, store 通过订阅同步, 界面无需手动处理
      setSession({
        ...getSession(),
        accessToken: body.data.accessToken,
        refreshToken: body.data.refreshToken,
      })
      return body.data
    })

  refreshing = task.finally(() => {
    // 无论成功失败都要复位, 否则后续请求会一直拿到这个已经结束的 promise
    refreshing = null
  })

  return refreshing.catch((error: unknown) => {
    // 刷新失败说明 refresh token 也失效了, 清空登录态并回到登录页
    notifySessionExpired()
    throw error
  })
}

instance.interceptors.response.use(
  (response) => {
    const body = response.data as ApiEnvelope<unknown> | undefined
    const config = response.config as RetryableConfig
    const tokenExpired = body?.code === CODE.invalidToken || body?.code === CODE.notLogin

    // token 失效时先刷新再自动重放一次, 业务代码不用关心这个流程
    if (tokenExpired && !config.skipAuthRefresh && !config.retried) {
      return refreshAccessToken().then(() =>
        instance.request({
          ...config,
          retried: true,
          headers: { ...config.headers, Authorization: `Bearer ${getAccessToken() ?? ''}` },
        } as RetryableConfig),
      )
    }

    return response
  },
  (error: AxiosError) => {
    // 访问了不存在的路由时后端会返回 HTTP 404 + {"code":1009,...},
    // 这里统一转成 ApiError, 让调用方拿到的是后端的提示而不是 axios 的错误信息。
    const body = error.response?.data as ApiEnvelope<unknown> | undefined
    if (body && typeof body.code === 'number') {
      return Promise.reject(new ApiError(body.code, body.message || '请求失败'))
    }
    return Promise.reject(error)
  },
)

/**
 * 发起请求并拆开统一响应结构:
 * 成功时直接拿到 data, 失败时抛出带 code/message 的 ApiError。
 */
export async function request<T>(config: AxiosRequestConfig & RetryableConfig): Promise<T> {
  const response = await instance.request<ApiEnvelope<T>>(config)
  const body = response.data

  if (!body || typeof body.code !== 'number') {
    throw new ApiError(-1, '后端返回的数据格式不正确')
  }
  if (body.code !== CODE.success) {
    throw new ApiError(body.code, body.message || '请求失败')
  }
  return body.data
}

export default instance
