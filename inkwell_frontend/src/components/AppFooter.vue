<template>
  <footer class="footer">
    <div class="container footer__inner">
      <div class="footer__brand">
        <p class="footer__desc">
          Go + Gin + MySQL + Redis 打造的社区示例, 榜单、投票与评论数都由 Redis 实时计算。
        </p>
      </div>

      <div class="footer__meta">
        <ServiceStatus />
        <a class="footer__link" :href="swaggerUrl" target="_blank" rel="noreferrer">
          接口文档
          <AppIcon name="external" :size="14" />
        </a>
        <span class="footer__copy">© {{ year }} Inkwell</span>
      </div>
    </div>
  </footer>
</template>

<script setup lang="ts">
import AppIcon from '@/components/AppIcon.vue'
import ServiceStatus from '@/components/ServiceStatus.vue'

// 用相对路径: 开发/预览服务器会把 /swagger 代理到后端, 生产部署由网关统一转发,
// 这样换端口、换域名都不用改代码
const swaggerUrl = '/swagger/index.html'
const year = new Date().getFullYear()
</script>

<style scoped lang="scss">
.footer {
  /* 和顶栏一致: 不用硬分割线, 用一段很浅的渐变把页脚托起来 */
  background: linear-gradient(180deg, transparent 0%, var(--surface-2) 28%, var(--surface-2) 100%);
}

.footer__inner {
  display: flex;
  flex-direction: column;
  gap: var(--space-5);
  padding-top: var(--space-7);
  padding-bottom: var(--space-7);

  @media (min-width: 900px) {
    flex-direction: row;
    align-items: center;
    justify-content: space-between;
  }
}

.footer__brand {
  display: flex;
  flex-direction: column;
  gap: var(--space-3);
  max-width: 420px;
}

.footer__desc {
  color: var(--text-tertiary);
  font-size: var(--text-sm);
}

.footer__meta {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: var(--space-4);
  color: var(--text-tertiary);
  font-size: var(--text-sm);
}

.footer__link {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  border-radius: var(--radius-xs);
  color: var(--text-secondary);
  transition: color var(--transition);

  &:hover {
    color: var(--accent);
  }
}

.footer__copy {
  color: var(--text-tertiary);
}
</style>
