import { ref } from 'vue'
import { defineStore } from 'pinia'

import type { VoteDirection } from '@/api'

const STORAGE_KEY = 'inkwell.votes'

type VoteMap = Record<string, VoteDirection>

/**
 * 后端没有"查询我对某帖投过什么票"的接口, 而重复投同一方向的票会被拒绝,
 * 所以这里在本地按 `${userID}:${postID}` 记录投票状态, 用来决定点击时该发什么方向:
 * 赞/踩互斥, 再点一次表示取消(0)。
 */
export const useVoteStore = defineStore('votes', () => {
  const votes = ref<VoteMap>(readVotes())

  function readVotes(): VoteMap {
    try {
      const raw = localStorage.getItem(STORAGE_KEY)
      if (!raw) {
        return {}
      }
      const parsed = JSON.parse(raw) as VoteMap
      return parsed && typeof parsed === 'object' ? parsed : {}
    } catch {
      localStorage.removeItem(STORAGE_KEY)
      return {}
    }
  }

  function persist(): void {
    try {
      localStorage.setItem(STORAGE_KEY, JSON.stringify(votes.value))
    } catch {
      // 忽略存储失败
    }
  }

  /** 返回 null 表示"本机没有这个用户对这帖的投票记录" */
  function get(userID: string | null, postID: string): VoteDirection | null {
    if (!userID) {
      return null
    }
    return votes.value[`${userID}:${postID}`] ?? null
  }

  /**
   * 记录投票方向。0(取消投票)也要落库,
   * 否则会退化成"没有记录", 界面又会把作者默认的赞成票显示出来。
   */
  function record(userID: string, postID: string, directionValue: VoteDirection): void {
    votes.value = { ...votes.value, [`${userID}:${postID}`]: directionValue }
    persist()
  }

  function clearUser(userID: string): void {
    const prefix = `${userID}:`
    const next: VoteMap = {}
    Object.entries(votes.value).forEach(([key, value]) => {
      if (!key.startsWith(prefix)) {
        next[key] = value
      }
    })
    votes.value = next
    persist()
  }

  return {
    votes,
    get,
    record,
    clearUser,
  }
})
