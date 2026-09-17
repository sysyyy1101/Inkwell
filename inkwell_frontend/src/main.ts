import { createApp } from 'vue'
import { createPinia } from 'pinia'

import App from './App.vue'
import router from './router'
import { setSessionExpiredHandler } from './api/session'
import { useAuthStore } from './stores/auth'
import { useToastStore } from './stores/toast'
// 自托管可变字体: Inter(拉丁正文) / Noto Sans SC(中文正文) / Source Serif 4(拉丁标题) /
// Noto Serif SC(中文标题) / JetBrains Mono(等宽, 用于 ID)。字体栈见 styles/tokens.scss
import '@fontsource-variable/inter'
import '@fontsource-variable/noto-sans-sc/wght.css'
import '@fontsource-variable/noto-serif-sc/wght.css'
import '@fontsource-variable/source-serif-4/wght.css'
import '@fontsource-variable/jetbrains-mono/wght.css'
// 品牌字标专用字体(只在顶栏的 Inkwell 字标上用, 见 tokens.scss 里的 --font-brand)
import '@fontsource-variable/cinzel'
import './styles/base.scss'

const app = createApp(App)
const pinia = createPinia()

app.use(pinia)
app.use(router)

// refresh token 也失效时: 清空登录态, 提示并跳回登录页(带上回跳地址)
const auth = useAuthStore(pinia)
const toast = useToastStore(pinia)

setSessionExpiredHandler(() => {
  if (!auth.isLogin) {
    return
  }
  auth.logout()
  toast.error('登录已过期, 请重新登录')
  const current = router.currentRoute.value
  if (current.name !== 'login') {
    void router.replace({ name: 'login', query: { redirect: current.fullPath } })
  }
})

app.mount('#app')
