import { request } from './http'
import type { CreateCommentPayload, PostComment } from './types'

/** 某个帖子下的全部评论(按时间正序): GET /comment?post_id=xxx */
export function fetchCommentsByPost(postId: string): Promise<PostComment[]> {
  return request<PostComment[]>({
    method: 'get',
    url: '/comment',
    params: { post_id: postId },
  })
}

/** 按评论 ID 批量查询: GET /comment?ids=1&ids=2 */
export function fetchCommentsByIds(ids: Array<string | number>): Promise<PostComment[]> {
  return request<PostComment[]>({
    method: 'get',
    url: '/comment',
    // 同名参数重复出现, axios 会序列化成 ?ids=1&ids=2
    params: { ids },
    paramsSerializer: {
      indexes: null,
    },
  })
}

/** 发表评论/回复: POST /comment, 返回新评论 ID */
export function createComment(payload: CreateCommentPayload): Promise<{ comment_id: string }> {
  return request<{ comment_id: string }>({
    method: 'post',
    url: '/comment',
    data: payload,
  })
}
