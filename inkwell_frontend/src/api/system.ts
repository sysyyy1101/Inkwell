import { request } from './http'

/** 健康检查: GET /ping, 正常时返回 "pong" */
export function ping(timeout = 6000): Promise<string> {
  return request<string>({
    method: 'get',
    url: '/ping',
    timeout,
    // 健康检查不需要带 token, 失败了也不要去刷新登录态
    skipAuthRefresh: true,
  })
}
