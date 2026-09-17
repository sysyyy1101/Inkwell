package controller

import (
	"net/http"
	"sync"

	"inkwell_backend/pkg/ratelimit"
	"inkwell_backend/settings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// 登录限流的兜底参数: 配置没填或写错时使用
const (
	loginRateLimitDefaultRate  = 2.0
	loginRateLimitDefaultBurst = 10
)

var (
	// 整个进程只有一个登录令牌桶(不是每个 IP 一个桶)。
	// 桶在第一次请求时创建, 因为那时配置一定已经加载完了。
	loginLimiterOnce sync.Once
	loginLimiter     *ratelimit.Bucket
)

// LoginRateLimit 登录接口的令牌桶限流中间件。
//
// 只用一个桶而不是按 IP 分桶: 实现简单、内存恒定, 目的是挡住脚本化的密码爆破,
// 代价是多个用户同时登录时共享同一份额度(参数见配置 login_rate_limit)。
func LoginRateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		loginLimiterOnce.Do(func() {
			loginLimiter = newLoginLimiter()
		})
		if !loginLimiter.Allow() {
			zap.L().Warn("login rate limited",
				zap.String("ip", c.ClientIP()),
				zap.String("path", c.Request.URL.Path),
			)
			// 限流比较特殊: 除了业务码, 同时返回 HTTP 429, 方便网关/前端区分"被限流"和"参数错误"
			ResponseErrorWithStatus(c, CodeTooManyRequests, http.StatusTooManyRequests)
			c.Abort()
			return
		}
		c.Next()
	}
}

// newLoginLimiter 按配置创建令牌桶
func newLoginLimiter() *ratelimit.Bucket {
	rate, burst := loginRateLimitDefaultRate, loginRateLimitDefaultBurst
	if cfg := settings.Conf.LoginRateLimitConfig; cfg != nil {
		if cfg.Rate > 0 {
			rate = cfg.Rate
		}
		if cfg.Burst > 0 {
			burst = cfg.Burst
		}
	}
	return ratelimit.New(rate, burst)
}
