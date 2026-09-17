import { ref } from 'vue'
import { defineStore } from 'pinia'

import { communityApi, toErrorMessage } from '@/api'
import type { Community } from '@/api'

/** 版块列表在整个应用里会用到多次(导航/侧边栏/发帖页), 这里缓存一份 */
export const useCommunityStore = defineStore('community', () => {
  const list = ref<Community[]>([])
  const loading = ref(false)
  const error = ref<string | null>(null)
  let loaded = false

  async function ensureLoaded(): Promise<void> {
    if (loaded || loading.value) {
      return
    }
    await reload()
  }

  async function reload(): Promise<void> {
    loading.value = true
    error.value = null
    try {
      list.value = await communityApi.fetchCommunities()
      loaded = true
    } catch (e) {
      error.value = toErrorMessage(e)
      list.value = []
      loaded = false
    } finally {
      loading.value = false
    }
  }

  function nameOf(id: number): string {
    return list.value.find((item) => item.community_id === id)?.community_name ?? '未知版块'
  }

  return {
    list,
    loading,
    error,
    ensureLoaded,
    reload,
    nameOf,
  }
})
