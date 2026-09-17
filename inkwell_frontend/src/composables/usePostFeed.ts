import { computed, ref, type Ref } from 'vue'

import { communityApi, postApi, toErrorMessage, userApi } from '@/api'
import type { FeedSource, OrderType, PostListItem } from '@/api'

/** 榜单接口(Redis)固定每页 20 条, 与后端 dao/redis 里的 PostPerAge 保持一致 */
export const RANK_PAGE_SIZE = 20
/** /post2 与 /user/:id/post 接口默认每页 10 条 */
export const MYSQL_PAGE_SIZE = 10

interface UsePostFeedOptions {
  source: FeedSource
  /** 版块 ID, 只有版块页需要传 */
  communityId?: Ref<string | number | null>
  /** 用户 ID, 只有用户主页需要传 */
  userId?: Ref<string | null>
}

/**
 * 帖子列表的分页逻辑, 首页和版块页共用。
 * - source = 'rank': 走 Redis 榜单(/post 或 /community/:id/post), 支持 score/time 排序, 每页固定 20 条;
 * - source = 'mysql': 走 /post2, 按发帖时间倒序, 每页 10 条;
 * - source = 'user': 走 /user/:id/post, 某个用户发过的帖子, 按发帖时间倒序, 每页 10 条。
 * 两个接口都不返回总数, 所以用"本页是否取满"判断有没有下一页。
 */
export function usePostFeed(options: UsePostFeedOptions) {
  const pageSize = options.source === 'rank' ? RANK_PAGE_SIZE : MYSQL_PAGE_SIZE

  const items = ref<PostListItem[]>([])
  const order = ref<OrderType>('score')
  const page = ref(1)
  const loading = ref(false)
  const loadingMore = ref(false)
  const hasMore = ref(false)
  const error = ref<string | null>(null)
  /** 只有榜单接口有排序参数 */
  const supportsOrder = options.source === 'rank'

  // 请求序号: 快速切 Tab 时丢弃过期响应, 避免旧数据覆盖新数据
  let requestId = 0

  const isEmpty = computed(() => !loading.value && items.value.length === 0)

  async function fetchPage(targetPage: number): Promise<PostListItem[]> {
    if (options.source === 'rank') {
      if (options.communityId) {
        const id = options.communityId.value
        if (id === null || id === undefined) {
          return []
        }
        return communityApi.fetchCommunityPosts(id, { order: order.value, page: targetPage })
      }
      return postApi.fetchPostRank({ order: order.value, page: targetPage })
    }
    if (options.source === 'user') {
      const id = options.userId?.value
      if (!id) {
        return []
      }
      return userApi.fetchUserPosts(id, { page: targetPage, size: pageSize })
    }
    return postApi.fetchPostListFromMySQL({ page: targetPage, size: pageSize })
  }

  /** 首次加载/重新加载(切排序、切数据源时调用) */
  async function reload(): Promise<void> {
    const current = (requestId += 1)
    loading.value = true
    error.value = null
    try {
      const list = await fetchPage(1)
      if (current !== requestId) {
        return
      }
      items.value = list
      page.value = 1
      hasMore.value = list.length >= pageSize
    } catch (e) {
      if (current !== requestId) {
        return
      }
      items.value = []
      hasMore.value = false
      error.value = toErrorMessage(e)
    } finally {
      if (current === requestId) {
        loading.value = false
      }
    }
  }

  /** 加载下一页并追加到列表末尾 */
  async function loadMore(): Promise<void> {
    if (loadingMore.value || loading.value || !hasMore.value) {
      return
    }
    const current = requestId
    const nextPage = page.value + 1
    loadingMore.value = true
    try {
      const list = await fetchPage(nextPage)
      if (current !== requestId) {
        return
      }
      // 榜单里可能有已过期的帖子被后端跳过, 这里按 post_id 去重, 避免 key 冲突
      const seen = new Set(items.value.map((item) => item.post_id))
      items.value = [...items.value, ...list.filter((item) => !seen.has(item.post_id))]
      page.value = nextPage
      hasMore.value = list.length >= pageSize
    } catch (e) {
      error.value = toErrorMessage(e)
    } finally {
      loadingMore.value = false
    }
  }

  /** 切换排序并重新加载 */
  async function setOrder(next: OrderType): Promise<void> {
    if (order.value === next) {
      return
    }
    order.value = next
    await reload()
  }

  /** 本地更新某条帖子的票数(投票后立即反馈, 不等接口重新拉列表) */
  function patchVoteCount(postID: string, delta: number): void {
    items.value = items.value.map((item) =>
      item.post_id === postID ? { ...item, vote_num: item.vote_num + delta } : item,
    )
  }

  return {
    items,
    order,
    page,
    loading,
    loadingMore,
    hasMore,
    error,
    isEmpty,
    supportsOrder,
    pageSize,
    reload,
    loadMore,
    setOrder,
    patchVoteCount,
  }
}
