/** 后端统一响应结构: {"code":1000,"message":"success","data":{}} */
export interface ApiEnvelope<T> {
  code: number
  message: string
  data: T
}

/** 业务响应码, 与 inkwell_backend/controller/code.go 保持一致 */
export const CODE = {
  success: 1000,
  invalidParams: 1001,
  userExist: 1002,
  userNotExist: 1003,
  invalidPassword: 1004,
  serverBusy: 1005,
  invalidToken: 1006,
  invalidAuthFormat: 1007,
  notLogin: 1008,
  notFound: 1009,
  /** 请求过于频繁(登录限流), 响应同时带 HTTP 429 */
  tooManyRequests: 1010,
} as const

/** 版块 */
export interface Community {
  community_id: number
  community_name: string
}

/** 版块详情 */
export interface CommunityDetail extends Community {
  introduction?: string
  create_time: string
}

/**
 * 列表页的帖子。榜单接口(Redis)和 /post2(MySQL)返回的是同一个结构:
 * - 榜单接口: vote_num / comment_num 来自 Redis, score 是榜单排序分;
 * - /post2 与 /user/:id/post: vote_num / comment_num 尽力从 Redis 补齐, score 恒为 0。
 * vote_num 是净票数(赞成票 - 反对票), 可能为负。
 * create_time 是 unix 秒。
 */
export interface PostListItem {
  post_id: string
  title: string
  summary: string
  author_id: string
  author_name: string
  community_id: number
  community_name: string
  vote_num: number
  comment_num: number
  score: number
  create_time: number
}

/** 帖子详情, create_time 是 RFC3339 时间字符串 */
export interface PostDetail {
  post_id: string
  title: string
  content: string
  author_id: string
  community_id: number
  status: number
  create_time: string
  author_name: string
  community_name: string
  vote_num: number
  comment_num: number
}

/** 评论 */
export interface PostComment {
  comment_id: string
  post_id: string
  parent_id: string
  author_id: string
  content: string
  create_time: string
  author_name: string
}

/** 登录成功返回 */
export interface LoginResult {
  accessToken: string
  refreshToken: string
  userID: string
  username: string
}

/** 用户主页信息(只有公开字段) */
export interface UserProfile {
  user_id: string
  username: string
  /** 注册时间, RFC3339 字符串 */
  create_time: string
}

/** 刷新 token 返回 */
export interface RefreshTokenResult {
  accessToken: string
  refreshToken: string
}

/** 登录/注册参数 */
export interface LoginPayload {
  username: string
  password: string
}

export interface SignUpPayload {
  username: string
  password: string
  confirm_password: string
}

/** 发帖参数 */
export interface CreatePostPayload {
  title: string
  content: string
  community_id: number
}

/** 发评论参数, parent_id 传 0 表示直接回复帖子 */
export interface CreateCommentPayload {
  post_id: string
  parent_id: string
  content: string
}

/** 投票方向: 1 赞成 / 0 取消 / -1 反对 */
export type VoteDirection = 1 | 0 | -1

export interface VotePayload {
  post_id: string
  direction: VoteDirection
}

/**
 * 投票成功返回。vote_num 是投票后服务端的权威净票数(赞成 - 反对, 可能为负):
 * 本机可能没有(或记错)自己的投票方向, 用服务端返回的值展示才不会和后端对不上。
 */
export interface VoteResult {
  post_id: string
  vote_num: number
}

/** 榜单排序方式 */
export type OrderType = 'score' | 'time'

/** 列表数据来源: 榜单(Redis) / 全部帖子(MySQL) / 某个用户的帖子(MySQL) */
export type FeedSource = 'rank' | 'mysql' | 'user'

export interface PageQuery {
  page?: number
  size?: number
}

export interface RankQuery extends PageQuery {
  order?: OrderType
}
