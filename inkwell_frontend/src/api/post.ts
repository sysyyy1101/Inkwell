import { request } from './http'
import type {
  CreatePostPayload,
  OrderType,
  PostDetail,
  PostListItem,
  VoteDirection,
  VotePayload,
  VoteResult,
} from './types'

/** 帖子榜单: GET /post?order=score|time&page=N, 每页固定 20 条 */
export function fetchPostRank(params: { order?: OrderType; page?: number } = {}): Promise<PostListItem[]> {
  return request<PostListItem[]>({
    method: 'get',
    url: '/post',
    params: { order: params.order ?? 'score', page: params.page ?? 1 },
  })
}

/** 全部帖子(直接读 MySQL, 按发帖时间倒序): GET /post2?page=N&size=M */
export function fetchPostListFromMySQL(params: { page?: number; size?: number } = {}): Promise<PostListItem[]> {
  return request<PostListItem[]>({
    method: 'get',
    url: '/post2',
    params: { page: params.page ?? 1, size: params.size ?? 10 },
  })
}

/** 帖子详情: GET /post/:id */
export function fetchPostDetail(id: string | number): Promise<PostDetail> {
  return request<PostDetail>({
    method: 'get',
    url: `/post/${id}`,
  })
}

/** 发帖: POST /post, 返回新帖子 ID */
export function createPost(payload: CreatePostPayload): Promise<{ post_id: string }> {
  return request<{ post_id: string }>({
    method: 'post',
    url: '/post',
    data: payload,
  })
}

/** 投票: POST /vote, direction 为 1/0/-1, 返回投票后的权威票数 */
export function votePost(postId: string, direction: VoteDirection): Promise<VoteResult> {
  const payload: VotePayload = { post_id: postId, direction }
  return request<VoteResult>({
    method: 'post',
    url: '/vote',
    data: payload,
  })
}
