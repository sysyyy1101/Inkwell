import { request } from './http'
import type { PostListItem, UserProfile } from './types'

/** 用户主页信息: GET /user/:id */
export function fetchUserProfile(id: string | number): Promise<UserProfile> {
  return request<UserProfile>({
    method: 'get',
    url: `/user/${id}`,
  })
}

/** 某个用户发布的帖子(按发帖时间倒序): GET /user/:id/post?page=N&size=M */
export function fetchUserPosts(
  id: string | number,
  params: { page?: number; size?: number } = {},
): Promise<PostListItem[]> {
  return request<PostListItem[]>({
    method: 'get',
    url: `/user/${id}/post`,
    params: { page: params.page ?? 1, size: params.size ?? 10 },
  })
}
