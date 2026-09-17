import { request } from './http'
import type { LoginPayload, LoginResult, SignUpPayload } from './types'

/** 登录: POST /login */
export function login(payload: LoginPayload): Promise<LoginResult> {
  return request<LoginResult>({
    method: 'post',
    url: '/login',
    data: payload,
  })
}

/** 注册: POST /signup, 成功时 data 为 null */
export function signUp(payload: SignUpPayload): Promise<null> {
  return request<null>({
    method: 'post',
    url: '/signup',
    data: payload,
  })
}
