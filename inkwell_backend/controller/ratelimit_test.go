package controller

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

// TestLoginRateLimit 登录限流: 桶里的令牌用光之后必须返回 HTTP 429 + 业务码 1010。
// 这里用一个只有限流中间件的路由来验证, 不依赖 MySQL。
func TestLoginRateLimit(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.POST("/login", LoginRateLimit(), func(c *gin.Context) {
		ResponseSuccess(c, "ok")
	})

	burst := int(newLoginLimiter().Capacity())

	doPost := func() *httptest.ResponseRecorder {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader("{}"))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)
		return w
	}

	// 前 burst 次请求应该放行(默认参数下 burst = 10)
	for i := 0; i < burst; i++ {
		w := doPost()
		if w.Code != http.StatusOK {
			t.Fatalf("第 %d 次请求 status = %d, want %d, body = %s", i+1, w.Code, http.StatusOK, w.Body.String())
		}
	}

	// 令牌用光之后再请求就会被限流
	w := doPost()
	if w.Code != http.StatusTooManyRequests {
		t.Fatalf("超出桶容量后 status = %d, want %d", w.Code, http.StatusTooManyRequests)
	}
	body := w.Body.String()
	if !strings.Contains(body, `"code":1010`) {
		t.Errorf("限流响应里应该有业务码 1010, got %s", body)
	}
	if !strings.Contains(body, "请求过于频繁") {
		t.Errorf("限流响应里应该有提示文案, got %s", body)
	}
}
