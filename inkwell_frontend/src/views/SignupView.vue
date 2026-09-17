<template>
  <div class="container">
    <div class="auth">
      <section class="card auth__card">
        <header class="auth__header">
          <h1 class="auth__title font-display">创建账号</h1>
          <p class="auth__desc">注册后就可以发帖、投票和评论。</p>
        </header>

        <form class="auth__form" @submit.prevent="submit">
          <div class="field">
            <label class="field__label" for="signup-username">用户名</label>
            <input
              id="signup-username"
              v-model="username"
              class="input"
              type="text"
              autocomplete="username"
              placeholder="2-64 个字符"
              :disabled="submitting"
            />
            <span class="field__hint">
              <span>用户名唯一, 注册后不可修改</span>
            </span>
          </div>

          <div class="field">
            <label class="field__label" for="signup-password">密码</label>
            <input
              id="signup-password"
              v-model="password"
              class="input"
              type="password"
              autocomplete="new-password"
              placeholder="6-64 位"
              :disabled="submitting"
            />
          </div>

          <div class="field">
            <label class="field__label" for="signup-confirm">确认密码</label>
            <input
              id="signup-confirm"
              v-model="confirmPassword"
              class="input"
              type="password"
              autocomplete="new-password"
              placeholder="再输入一次密码"
              :disabled="submitting"
            />
          </div>

          <p v-if="errorMessage" class="auth__error">
            <AppIcon name="alert" :size="16" />
            {{ errorMessage }}
          </p>

          <button class="btn btn--primary btn--lg btn--block" type="submit" :disabled="!canSubmit">
            {{ submitting ? '注册中…' : '注册' }}
          </button>
        </form>

        <p class="auth__footer">
          已经有账号了?
          <RouterLink class="auth__link" :to="{ name: 'login', query: route.query }">去登录</RouterLink>
        </p>
      </section>
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
const confirmPassword = ref('')
const submitting = ref(false)
const errorMessage = ref('')

const canSubmit = computed(
  () =>
    username.value.trim().length >= 2 &&
    password.value.length >= 6 &&
    confirmPassword.value.length > 0 &&
    !submitting.value,
)

async function submit(): Promise<void> {
  if (!canSubmit.value) {
    return
  }
  errorMessage.value = ''
  if (password.value !== confirmPassword.value) {
    errorMessage.value = '两次输入的密码不一致'
    return
  }
  if (password.value.length > 64) {
    errorMessage.value = '密码不能超过 64 位'
    return
  }

  submitting.value = true
  try {
    await auth.register({
      username: username.value.trim(),
      password: password.value,
      confirm_password: confirmPassword.value,
    })
    toast.notifyThenRedirect(
      'success',
      '注册成功, 请登录',
      () => {
        void router.replace({ name: 'login', query: route.query })
      },
      450,
    )
  } catch (error) {
    errorMessage.value = toErrorMessage(error)
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
</style>
