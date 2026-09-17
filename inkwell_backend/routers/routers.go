package routers

import (
	"inkwell_backend/controller"
	"net/http"

	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	//gin.SetMode(gin.ReleaseMode)
	//r := gin.New()
	//r.Use(logger.GinLogger(), logger.GinRecovery(true))
	r := gin.Default()
	v1 := r.Group("/api/v1")
	{
		// 不需要登录的接口
		// 登录接口挂令牌桶限流(整个进程只有一个桶), 用来挡脚本化的密码爆破
		v1.POST("/login", controller.LoginRateLimit(), controller.LoginHandler)
		v1.POST("/signup", controller.SignUpHandler)
		// 刷新token时 access token 可能已经过期, 所以不能走 JWTAuthMiddleware
		v1.GET("/refresh_token", controller.RefreshTokenHandler)
		v1.GET("/ping", controller.PingHandler)

		// 帖子/版块/评论的读接口对游客开放, 否则未登录用户打开首页看不到任何内容
		v1.GET("/community", controller.CommunityHandler)
		v1.GET("/community/:id", controller.CommunityDetailHandler)
		v1.GET("/community/:id/post", controller.CommunityPostListHandler)
		v1.GET("/post", controller.PostListHandler)
		v1.GET("/post2", controller.PostList2Handler)
		v1.GET("/post/:id", controller.PostDetailHandler)
		v1.GET("/comment", controller.CommentListHandler)
		// 用户主页: 自己/别人的发帖记录, 游客也能看
		v1.GET("/user/:id", controller.UserProfileHandler)
		v1.GET("/user/:id/post", controller.UserPostListHandler)
	}

	// 需要登录才能访问的接口
	auth := v1.Group("", controller.JWTAuthMiddleware())
	{
		auth.POST("/post", controller.CreatePostHandler)
		auth.POST("/vote", controller.VoteHandler)
		auth.POST("/comment", controller.CommentHandler)
	}

	// 接口文档: GET /swagger/index.html
	registerSwagger(r)

	r.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, controller.ResponseData{
			Code:    controller.CodeNotFound,
			Message: controller.CodeNotFound.Msg(),
			Data:    nil,
		})
	})
	return r
}
