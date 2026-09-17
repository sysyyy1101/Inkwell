package routers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

// swaggerDoc swagger.json 里本测试关心的字段
type swaggerDoc struct {
	Swagger             string                                `json:"swagger"`
	BasePath            string                                `json:"basePath"`
	Paths               map[string]map[string]json.RawMessage `json:"paths"`
	SecurityDefinitions map[string]json.RawMessage            `json:"securityDefinitions"`
}

// doGet 发起一次 GET 请求并返回响应
func doGet(t *testing.T, r *gin.Engine, path string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, path, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

// TestSwaggerHandler swagger 页面、描述文件和静态资源都要能正常访问
func TestSwaggerHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := SetupRouter()

	// 1. 页面能打开, 并且引用了 swagger-ui 的脚本
	w := doGet(t, r, "/swagger/index.html")
	if w.Code != http.StatusOK {
		t.Fatalf("GET /swagger/index.html status = %d, want %d", w.Code, http.StatusOK)
	}
	if body := w.Body.String(); !strings.Contains(body, "SwaggerUIBundle") ||
		!strings.Contains(body, `"./doc.json"`) {
		t.Errorf("GET /swagger/index.html body 没有正确引用 doc.json: %s", head(body, 200))
	}
	if ct := w.Header().Get("Content-Type"); !strings.HasPrefix(ct, "text/html") {
		t.Errorf("GET /swagger/index.html Content-Type = %q, want text/html", ct)
	}

	// 2. /swagger 要能跳转到 /swagger/
	if w = doGet(t, r, "/swagger"); w.Code != http.StatusMovedPermanently &&
		w.Code != http.StatusTemporaryRedirect && w.Code != http.StatusOK {
		t.Errorf("GET /swagger status = %d, want 200 或 30x", w.Code)
	}

	// 3. 静态资源(官方 swaggo/files 提供)能被访问到
	for _, path := range []string{"/swagger/swagger-ui.css", "/swagger/swagger-ui-bundle.js"} {
		if w = doGet(t, r, path); w.Code != http.StatusOK || w.Body.Len() == 0 {
			t.Errorf("GET %s status = %d, bodyLen = %d, want 200 且非空", path, w.Code, w.Body.Len())
		}
	}

	// 4. 不存在的静态资源返回 404
	if w = doGet(t, r, "/swagger/not-exist.js"); w.Code != http.StatusNotFound {
		t.Errorf("GET /swagger/not-exist.js status = %d, want %d", w.Code, http.StatusNotFound)
	}
}

// TestSwaggerDocCoversAllRoutes 文档里的接口要和路由注册保持一致,
// 以后新增接口忘了写注解时这个测试会失败。
func TestSwaggerDocCoversAllRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := SetupRouter()

	w := doGet(t, r, "/swagger/doc.json")
	if w.Code != http.StatusOK {
		t.Fatalf("GET /swagger/doc.json status = %d, want %d", w.Code, http.StatusOK)
	}
	var doc swaggerDoc
	if err := json.Unmarshal(w.Body.Bytes(), &doc); err != nil {
		t.Fatalf("doc.json 不是合法JSON, err = %v", err)
	}
	if doc.Swagger != "2.0" || doc.BasePath != "/api/v1" {
		t.Errorf("doc = {swagger:%q basePath:%q}, want {2.0 /api/v1}", doc.Swagger, doc.BasePath)
	}
	if len(doc.SecurityDefinitions) == 0 {
		t.Error("doc.json 缺少 securityDefinitions, 需要登录的接口在文档里无法填写 token")
	}

	// 路由里注册的接口(去掉 /api/v1 前缀和 :id 转成 {id})都应该在文档里出现
	for _, route := range r.Routes() {
		if strings.HasPrefix(route.Path, swaggerPath) {
			continue
		}
		path := strings.TrimPrefix(route.Path, "/api/v1")
		path = strings.ReplaceAll(path, ":id", "{id}")
		methods, ok := doc.Paths[path]
		if !ok {
			t.Errorf("路由 %s %s 没有出现在 swagger 文档里", route.Method, route.Path)
			continue
		}
		if _, ok := methods[strings.ToLower(route.Method)]; !ok {
			t.Errorf("路由 %s %s 的 %s 方法没有出现在 swagger 文档里",
				route.Method, route.Path, strings.ToLower(route.Method))
		}
	}
}

// head 返回 s 的前 n 个字符, 只用于测试失败时打印日志
func head(s string, n int) string {
	if len(s) > n {
		return s[:n]
	}
	return s
}
