<template>
  <div class="container">
    <div class="narrow">
      <RouterLink class="back-link" :to="{ name: 'home' }">
        <AppIcon name="arrowLeft" :size="16" />
        返回首页
      </RouterLink>

      <form class="card publish" @submit.prevent="submit">
        <header class="publish__header">
          <h1 class="publish__title font-display">发布帖子</h1>
          <p class="publish__desc">
            正文会写入 MySQL, 同时进入 Redis 榜单; 发帖后 7 天内其他人都可以投票。
          </p>
        </header>

        <div class="field">
          <label class="field__label" for="post-title">标题</label>
          <input
            id="post-title"
            v-model="title"
            class="input"
            type="text"
            maxlength="128"
            placeholder="用一句话说清楚你要讨论什么"
            :disabled="submitting"
          />
          <span class="field__hint">
            <span>最长 128 字</span>
            <span :class="{ 'is-warning': title.length > 110 }">{{ title.length }} / 128</span>
          </span>
        </div>

        <div class="field">
          <label class="field__label" for="post-community">版块</label>
          <select
            id="post-community"
            v-model.number="communityId"
            class="select"
            :disabled="community.loading || submitting"
          >
            <option :value="0" disabled>请选择版块</option>
            <option v-for="item in community.list" :key="item.community_id" :value="item.community_id">
              {{ item.community_name }}
            </option>
          </select>
          <span v-if="community.error" class="field__hint field__hint--error">
            {{ community.error }}
            <button class="btn btn--ghost btn--sm" type="button" @click="community.reload()">重试</button>
          </span>
        </div>

        <div class="field">
          <label class="field__label" for="post-content">正文</label>
          <textarea
            id="post-content"
            v-model="content"
            class="textarea publish__content"
            maxlength="8192"
            placeholder="把背景、你的思路和疑问写清楚, 更容易得到有价值的回复。"
            :disabled="submitting"
          />
          <span class="field__hint">
            <span>支持换行, 暂不支持 Markdown 渲染</span>
            <span :class="{ 'is-warning': content.length > 7600 }">{{ content.length }} / 8192</span>
          </span>
        </div>

        <p class="publish__tip">
          <AppIcon name="info" :size="16" />
          发布成功后你会自动获得一票赞成, 帖子会立刻出现在对应版块的榜单里。
        </p>

        <footer class="publish__footer">
          <span class="publish__invalid">{{ invalidReason }}</span>
          <div class="publish__buttons">
            <RouterLink class="btn btn--secondary" :to="{ name: 'home' }">取消</RouterLink>
            <button class="btn btn--primary" type="submit" :disabled="Boolean(invalidReason) || submitting">
              <AppIcon name="send" :size="16" />
              {{ submitting ? '发布中…' : '发布帖子' }}
            </button>
          </div>
        </footer>
      </form>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import AppIcon from '@/components/AppIcon.vue'
import { postApi, toErrorMessage } from '@/api'
import { useCommunityStore } from '@/stores/community'
import { useToastStore } from '@/stores/toast'

/** 与后端 models.Post 的限制保持一致 */
const TITLE_MAX = 128
const CONTENT_MAX = 8192

const router = useRouter()
const route = useRoute()
const community = useCommunityStore()
const toast = useToastStore()

const title = ref('')
const content = ref('')
const communityId = ref(0)
const submitting = ref(false)

const invalidReason = computed(() => {
  if (!title.value.trim()) {
    return '请填写标题'
  }
  if (title.value.length > TITLE_MAX) {
    return `标题不能超过 ${TITLE_MAX} 字`
  }
  if (!content.value.trim()) {
    return '请填写正文'
  }
  if (content.value.length > CONTENT_MAX) {
    return `正文不能超过 ${CONTENT_MAX} 字`
  }
  if (!communityId.value) {
    return '请选择版块'
  }
  return ''
})

async function submit(): Promise<void> {
  if (invalidReason.value || submitting.value) {
    return
  }
  submitting.value = true
  try {
    const result = await postApi.createPost({
      title: title.value.trim(),
      content: content.value.trim(),
      community_id: communityId.value,
    })
    // 先让"发布成功"的提示渲染出来, 再跳到详情页
    toast.notifyThenRedirect('success', '发布成功', () => {
      void router.push({ name: 'post', params: { id: result.post_id } })
    })
  } catch (error) {
    toast.error(toErrorMessage(error))
    submitting.value = false
  }
}

onMounted(async () => {
  await community.ensureLoaded()
  // 支持从版块页带着 ?community=3 进来直接选中对应版块
  const preselect = Number(route.query.community)
  if (preselect && community.list.some((item) => item.community_id === preselect)) {
    communityId.value = preselect
    return
  }
  if (community.list.length === 1) {
    communityId.value = community.list[0].community_id
  }
})
</script>

<style scoped lang="scss">
.narrow {
  max-width: 820px;
  margin: 0 auto;
}

.back-link {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  margin-bottom: var(--space-5);
  padding: 6px 12px 6px 8px;
  border-radius: var(--radius-pill);
  color: var(--text-secondary);
  font-size: var(--text-sm);
  transition: background-color var(--transition), color var(--transition);

  &:hover {
    background: var(--surface-hover);
    color: var(--text);
  }
}

.publish {
  display: flex;
  flex-direction: column;
  gap: var(--space-6);
  padding: var(--space-6);

  @media (min-width: 768px) {
    padding: var(--space-7);
  }
}

.publish__title {
  font-size: var(--text-xl);
  font-weight: var(--weight-heavy);
  letter-spacing: -0.03em;
}

.publish__desc {
  margin-top: 8px;
  color: var(--text-secondary);
  font-size: var(--text-base);
}

.publish__content {
  min-height: 260px;
}

.publish__tip {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 12px 16px;
  border-radius: var(--radius-md);
  background: var(--accent-soft);
  color: var(--accent);
  font-size: var(--text-sm);
}

.publish__footer {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-3);
  padding-top: var(--space-5);
  border-top: 1px solid var(--divider);
}

.publish__invalid {
  color: var(--text-tertiary);
  font-size: var(--text-sm);
}

.publish__buttons {
  display: flex;
  gap: var(--space-3);
  margin-left: auto;
}

.field__hint--error {
  color: var(--danger);
}

.is-warning {
  color: var(--warning);
}
</style>
