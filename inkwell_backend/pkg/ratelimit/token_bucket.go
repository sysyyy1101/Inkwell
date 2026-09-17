// Package ratelimit 提供一个极简的令牌桶实现。
//
// 设计取舍:
//   - 惰性补充: 每次 Allow 时按"距离上次取令牌过了多久"补令牌, 不需要后台 goroutine,
//     也就没有 goroutine 泄漏和定时器精度问题;
//   - 只依赖标准库: 不引入 golang.org/x/time/rate, 少一个依赖;
//   - 并发安全: 内部用互斥锁保护, 可以放心地被多个请求 goroutine 共享。
package ratelimit

import (
	"math"
	"sync"
	"time"
)

const (
	// defaultRate 每秒补充的令牌数(配置没填时的兜底值)
	defaultRate = 2.0
	// defaultBurst 桶容量(配置没填时的兜底值)
	defaultBurst = 10
)

// Bucket 令牌桶。零值不可用, 请使用 New 创建。
type Bucket struct {
	mu       sync.Mutex
	rate     float64   // 每秒补充的令牌数
	capacity float64   // 桶容量, 也就是允许的最大突发次数
	tokens   float64   // 当前令牌数
	last     time.Time // 上次补充令牌的时间
	now      func() time.Time
}

// New 创建一个令牌桶。rate <= 0 或 burst < 1 时使用默认值(2 个/秒, 容量 10),
// 保证配置写错时不会退化成"把所有请求都拒掉"或者"完全不限流"。
func New(rate float64, burst int) *Bucket {
	if rate <= 0 || math.IsNaN(rate) || math.IsInf(rate, 0) {
		rate = defaultRate
	}
	if burst < 1 {
		burst = defaultBurst
	}
	return &Bucket{
		rate:     rate,
		capacity: float64(burst),
		// 初始装满, 这样服务刚启动时不会立刻把前几个请求挡掉
		tokens: float64(burst),
		last:   time.Now(),
		now:    time.Now,
	}
}

// Allow 尝试取 1 个令牌, 取到返回 true。
func (b *Bucket) Allow() bool {
	return b.AllowN(1)
}

// AllowN 尝试取 n 个令牌, 取到返回 true。
// n <= 0 视为不消耗令牌; n 大于桶容量时永远返回 false(不可能凑够这么多令牌)。
func (b *Bucket) AllowN(n int) bool {
	if n <= 0 {
		return true
	}
	b.mu.Lock()
	defer b.mu.Unlock()

	b.refillLocked(b.now())
	if float64(n) > b.tokens {
		return false
	}
	b.tokens -= float64(n)
	return true
}

// refillLocked 按流逝的时间补充令牌, 调用前必须持有锁。
// 时钟回拨时 elapsed 为负, 这时既不能补也不能退, 避免"时间倒流白送令牌"。
func (b *Bucket) refillLocked(now time.Time) {
	if elapsed := now.Sub(b.last).Seconds(); elapsed > 0 {
		b.tokens = math.Min(b.capacity, b.tokens+elapsed*b.rate)
		b.last = now
	}
}

// Tokens 返回当前桶里剩余的令牌数(会先按流逝的时间补充一次), 主要用于观测和测试。
func (b *Bucket) Tokens() float64 {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.refillLocked(b.now())
	return b.tokens
}

// Rate 返回每秒补充的令牌数。
func (b *Bucket) Rate() float64 {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.rate
}

// Capacity 返回桶容量。
func (b *Bucket) Capacity() float64 {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.capacity
}
