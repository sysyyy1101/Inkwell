<template>
  <div class="container">
    <div class="auth">
      <section class="card auth__card">
        <header class="auth__header">
          <h1 class="auth__title font-display">登录 Inkwell</h1>
          <p class="auth__desc">登录后即可发帖、投票和评论。</p>
        </header>

        <form class="auth__form" @submit.prevent="submit">
          <div class="field">
            <label class="field__label" for="login-username">用户名</label>
            <input
              id="login-username"
              v-model="username"
              class="input"
              type="text"
              autocomplete="username"
              placeholder="请输入用户名"
              :disabled="submitting"
            />
          </div>

          <div class="field">
            <label class="field__label" for="login-password">密码</label>
            <div class="auth__password">
              <input
                id="login-password"
                v-model="password"
                class="input"
                :type="showPassword ? 'text' : 'password'"
                autocomplete="current-password"
                placeholder="请输入密码"
                :disabled="submitting"
              />
              <button
                class="auth__toggle"
                type="button"
                :aria-label="showPassword ? '隐藏密码' : '显示密码'"
                @click="showPassword = !showPassword"
              >
                {{ showPassword ? '隐藏' : '显示' }}
              </button>
            </div>
          </div>

          <p v-if="errorMessage" class="auth__error">
            <AppIcon name="alert" :size="16" />
            {{ errorMessage }}
          </p>

          <button class="btn btn--primary btn--lg btn--block" type="submit" :disabled="!canSubmit">
            {{ submitting ? '登录中…' : '登录' }}
          </button>
        </form>

        <p class="auth__footer">
          还没有账号?
          <RouterLink class="auth__link" :to="{ name: 'signup', query: route.query }">立即注册</RouterLink>
        </p>
      </section>

      <p class="auth__hint">
        测试账号来自 <code>init.sql</code>: <code>gopher_zhang</code> / <code>123456</code>
      </p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import AppIcon from '@/components/AppIcon.vue'
import { toErrorMessage } from '@/api'
import { useAuthStore } from '@/stores/auth'
import { useToastStore } from '@/stores/toast'

const auth = useAuthStore()
const toast = useToastStore()
const router = useRouter()
const route = useRoute()

const username = ref('')
const password = ref('')
const showPassword = ref(false)
const submitting = ref(false)
const errorMessage = ref('')

const canSubmit = computed(
  () => username.value.trim().length > 0 && password.value.length > 0 && !submitting.value,
)

/** 只允许回跳到站内相对路径, 避免被 query 参数带到外部站点 */
function redirectTarget(): string {
  const redirect = route.query.redirect
  if (typeof redirect === 'string' && redirect.startsWith('/') && !redirect.startsWith('//')) {
    return redirect
  }
  return '/'
}

async function submit(): Promise<void> {
  if (!canSubmit.value) {
    return
  }
  submitting.value = true
  errorMessage.value = ''
  try {
    await auth.login({ username: username.value.trim(), password: password.value })
    // 先让提示渲染出来再跳转, 否则用户看不到这条反馈
    toast.notifyThenRedirect(
      'success',
      `欢迎回来, ${auth.username}`,
      () => {
        void router.replace(redirectTarget())
      },
      450,
    )
  } catch (error) {
    errorMessage.value = toErrorMessage(error)
    // 失败才需要复位按钮; 成功时保持禁用, 避免跳转前的这段时间被重复提交
    submitting.value = false
  }
}
</script>

<style scoped lang="scss">
.auth {
  display: flex;
  max-width: 420px;
  flex-direction: column;
  gap: var(--space-4);
  margin: var(--space-6) auto 0;
}

.auth__card {
  display: flex;
  flex-direction: column;
  gap: var(--space-6);
  padding: var(--space-7) var(--space-6);
}

.auth__header {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: var(--space-2);
  text-align: center;
}

.auth__title {
  margin-top: var(--space-3);
  font-size: var(--text-xl);
  font-weight: var(--weight-heavy);
  letter-spacing: -0.03em;
}

.auth__desc {
  color: var(--text-secondary);
  font-size: var(--text-sm);
}

.auth__form {
  display: flex;
  flex-direction: column;
  gap: var(--space-5);
}

.auth__password {
  position: relative;
}

.auth__toggle {
  position: absolute;
  right: 12px;
  top: 50%;
  transform: translateY(-50%);
  color: var(--text-tertiary);
  font-size: var(--text-xs);
  transition: color var(--transition);

  &:hover {
    color: var(--accent);
  }
}

.auth__error {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 14px;
  border-radius: var(--radius-md);
  background: var(--danger-soft);
  color: var(--danger);
  font-size: var(--text-sm);
}

.auth__footer {
  color: var(--text-secondary);
  font-size: var(--text-sm);
  text-align: center;
}

.auth__link {
  color: var(--accent);
  font-weight: var(--weight-medium);
}

.auth__hint {
  color: var(--text-tertiary);
  font-size: var(--text-xs);
  line-height: 1.7;
  text-align: center;

  code {
    padding: 1px 6px;
    border-radius: var(--radius-xs);
    background: var(--surface-hover);
    font-family: var(--font-mono);
  }
}
</style>
