import { request } from './http'
import type { Community, CommunityDetail, OrderType, PostListItem } from './types'

/** 版块列表: GET /community */
export function fetchCommunities(): Promise<Community[]> {
  return request<Community[]>({
    method: 'get',
    url: '/community',
  })
}

/** 版块详情: GET /community/:id */
export function fetchCommunityDetail(id: number | string): Promise<CommunityDetail> {
  return request<CommunityDetail>({
    method: 'get',
    url: `/community/${id}`,
  })
}

/** 版块下的帖子榜单: GET /community/:id/post, 每页固定 20 条 */
export function fetchCommunityPosts(
  id: number | string,
  params: { order?: OrderType; page?: number } = {},
): Promise<PostListItem[]> {
  return request<PostListItem[]>({
    method: 'get',
    url: `/community/${id}/post`,
    params: { order: params.order ?? 'score', page: params.page ?? 1 },
  })
}
