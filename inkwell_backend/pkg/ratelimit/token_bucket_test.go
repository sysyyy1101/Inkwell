package ratelimit

import (
	"testing"
	"time"
)

// newTestBucket 创建一个时钟可控的桶, 返回桶和"把时间往前拨"的函数
func newTestBucket(rate float64, burst int) (*Bucket, func(time.Duration)) {
	b := New(rate, burst)
	current := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	b.last = current
	b.now = func() time.Time { return current }
	advance := func(d time.Duration) { current = current.Add(d) }
	return b, advance
}

func TestNewFallsBackToDefaults(t *testing.T) {
	for _, c := range []struct {
		name  string
		rate  float64
		burst int
	}{
		{name: "零值", rate: 0, burst: 0},
		{name: "负数", rate: -1, burst: -5},
	} {
		t.Run(c.name, func(t *testing.T) {
			b := New(c.rate, c.burst)
			if b.Rate() != defaultRate || b.Capacity() != float64(defaultBurst) {
				t.Fatalf("New(%v, %d) = rate %v capacity %v, want rate %v capacity %v",
					c.rate, c.burst, b.Rate(), b.Capacity(), defaultRate, float64(defaultBurst))
			}
			// 配置写错时也不能一上来就把请求全拒了
			if !b.Allow() {
				t.Error("使用默认参数时第一次 Allow 应该通过")
			}
		})
	}
}

// 桶容量决定突发上限: 连续取满 burst 次后必须被拦住
func TestBucketAllowsBurstThenBlocks(t *testing.T) {
	b, _ := newTestBucket(1, 3)

	for i := 0; i < 3; i++ {
		if !b.Allow() {
			t.Fatalf("第 %d 次 Allow 应该通过(桶容量 3)", i+1)
		}
	}
	if b.Allow() {
		t.Error("桶已经空了, 再 Allow 必须被拒绝")
	}
	if got := b.Tokens(); got != 0 {
		t.Errorf("tokens = %v, want 0", got)
	}
}

// 时间推移后按 rate 补充令牌
func TestBucketRefillsByRate(t *testing.T) {
	b, advance := newTestBucket(2, 2) // 2 个/秒, 容量 2

	b.Allow()
	b.Allow()
	if b.Allow() {
		t.Fatal("刚取空就应该被拒绝")
	}

	advance(250 * time.Millisecond) // 0.25s * 2 = 0.5 个令牌, 还不够
	if b.Allow() {
		t.Error("令牌不足 1 个时不应通过")
	}

	advance(250 * time.Millisecond) // 累计 1 个令牌
	if !b.Allow() {
		t.Error("补齐 1 个令牌后应该通过")
	}
	if b.Allow() {
		t.Error("刚用掉唯一一个令牌, 应该被拒绝")
	}
}

// 长时间不访问也只补到桶容量, 不会无限攒令牌
func TestBucketRefillIsCappedByCapacity(t *testing.T) {
	b, advance := newTestBucket(10, 5)

	for i := 0; i < 5; i++ {
		b.Allow()
	}
	advance(10 * time.Minute)
	if got := b.Tokens(); got != 5 {
		t.Errorf("tokens = %v, want 5(容量上限)", got)
	}
}

// 时钟回拨不应该白送令牌
func TestBucketClockGoingBackwards(t *testing.T) {
	b, advance := newTestBucket(1, 1)

	if !b.Allow() {
		t.Fatal("第一次 Allow 应该通过")
	}
	advance(-time.Hour)
	if b.Allow() {
		t.Error("时钟回拨后不应该补出令牌")
	}
}

func TestAllowN(t *testing.T) {
	b, _ := newTestBucket(1, 3)

	if !b.AllowN(0) || !b.AllowN(-1) {
		t.Error("n <= 0 不应该消耗令牌, 而且应该通过")
	}
	if !b.AllowN(3) {
		t.Error("桶容量 3 时 AllowN(3) 应该通过")
	}
	if b.AllowN(1) {
		t.Error("桶已空, 不应通过")
	}
	// 一次要得比桶容量还多, 永远不可能满足
	if b.AllowN(4) {
		t.Error("n 超过桶容量时必须返回 false")
	}
}

// 并发调用不能把令牌数取超: 100 个 goroutine 抢容量 10 的桶, 只应有 10 个通过
func TestBucketConcurrentAllow(t *testing.T) {
	b := New(0.0001, 10)
	b.last = time.Now()

	const goroutines = 100
	start := make(chan struct{})
	results := make(chan bool, goroutines)
	for i := 0; i < goroutines; i++ {
		go func() {
			<-start
			results <- b.Allow()
		}()
	}
	close(start)

	allowed := 0
	for i := 0; i < goroutines; i++ {
		if <-results {
			allowed++
		}
	}
	if allowed != 10 {
		t.Errorf("通过数 = %d, want 10", allowed)
	}
}
