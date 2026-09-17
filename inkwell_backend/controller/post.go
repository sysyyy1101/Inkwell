package controller

import (
	"inkwell_backend/logic"
	"inkwell_backend/models"
	"errors"
	"strconv"

	"go.uber.org/zap"

	"github.com/gin-gonic/gin"
)

const (
	// postListDefaultSize 帖子列表默认每页条数
	postListDefaultSize = 10
	// postListMaxSize 帖子列表单页最大条数
	postListMaxSize = 50
)

// CreatePostHandler 创建帖子
//
// @Summary  发帖
// @Description  创建帖子, 需要登录。标题最长 128 字, 内容最长 8192 字。
// @Description  失败时 HTTP 状态码仍是 200, 通过响应体里的 code 判断: 参数错误 1001, 未登录或 token 失效 1006/1008, 服务繁忙 1005。
// @Tags  帖子
// @Accept  json
// @Produce  json
// @Security  ApiKeyAuth
// @Param  body  body  models.CreatePostForm  true  "发帖参数"
// @Success  200  {object}  controller.ResponseData{data=models.PostIDResponse}  "发帖成功, 返回新帖子ID"
// @Router  /post [post]
func CreatePostHandler(c *gin.Context) {
	var post models.Post
	if err := c.ShouldBindJSON(&post); err != nil {
		// 帖子标题/内容/版块的校验都实现在 models.Post.UnmarshalJSON 里
		ResponseErrorWithMsg(c, CodeInvalidParams, err.Error())
		return
	}
	// 获取作者ID，当前请求的UserID
	userID, err := getCurrentUserID(c)
	if err != nil {
		zap.L().Error("getCurrentUserID() failed", zap.Error(err))
		ResponseError(c, CodeNotLogin)
		return
	}
	post.AuthorId = userID

	if err := logic.CreatePost(&post); err != nil {
		switch {
		case errors.Is(err, logic.ErrorAuthorNotExist), errors.Is(err, logic.ErrorCommunityNotExist):
			ResponseErrorWithMsg(c, CodeInvalidParams, err.Error())
		default:
			zap.L().Error("logic.CreatePost() failed", zap.Error(err))
			ResponseError(c, CodeServerBusy)
		}
		return
	}
	// 返回新帖子的ID, 前端可以直接跳转到帖子详情页
	ResponseSuccess(c, gin.H{"post_id": strconv.FormatUint(post.PostID, 10)})
}

// PostListHandler 帖子榜单, 支持 order=time/score 与分页
//
// @Summary  帖子榜单
// @Description  从 Redis 榜单分页读取帖子, order=score 按热度排序, order=time 按发帖时间排序, 游客可访问。
// @Description  order 取值不合法时返回 code 1001。
// @Tags  帖子
// @Produce  json
// @Param  order  query  string  false  "排序方式"  Enums(score, time)  default(score)
// @Param  page  query  int  false  "页码, 从1开始"  default(1)
// @Success  200  {object}  controller.ResponseData{data=[]models.PostListItem}  "帖子列表, 每页20条"
// @Router  /post [get]
func PostListHandler(c *gin.Context) {
	order := c.DefaultQuery("order", "score")
	page, _ := getPageParams(c, postListDefaultSize, postListMaxSize)
	posts, err := logic.GetPostList(order, int64(page))
	if err != nil {
		if errors.Is(err, logic.ErrorInvalidOrder) {
			ResponseErrorWithMsg(c, CodeInvalidParams, err.Error())
			return
		}
		zap.L().Error("logic.GetPostList() failed",
			zap.String("order", order), zap.Int("page", page), zap.Error(err))
		ResponseError(c, CodeServerBusy)
		return
	}
	ResponseSuccess(c, posts)
}

// PostList2Handler 直接读 MySQL 的帖子列表(按发帖时间倒序, 支持分页)
//
// @Summary  帖子列表(MySQL)
// @Description  直接查询 MySQL 的帖子列表, 按发帖时间倒序, 游客可访问
// @Tags  帖子
// @Produce  json
// @Param  page  query  int  false  "页码, 从1开始"  default(1)
// @Param  size  query  int  false  "每页条数, 最大50"  default(10)
// @Success  200  {object}  controller.ResponseData{data=[]models.PostListItem}  "帖子列表"
// @Router  /post2 [get]
func PostList2Handler(c *gin.Context) {
	page, size := getPageParams(c, postListDefaultSize, postListMaxSize)
	posts, err := logic.GetPostListFromMySQL(page, size)
	if err != nil {
		zap.L().Error("logic.GetPostListFromMySQL() failed",
			zap.Int("page", page), zap.Int("size", size), zap.Error(err))
		ResponseError(c, CodeServerBusy)
		return
	}
	ResponseSuccess(c, posts)
}

// PostDetailHandler 帖子详情
//
// @Summary  帖子详情
// @Description  查询帖子详情, 并带上 Redis 里的实时票数与评论数, 游客可访问。帖子不存在时返回 code 1009。
// @Tags  帖子
// @Produce  json
// @Param  id  path  int  true  "帖子ID"
// @Success  200  {object}  controller.ResponseData{data=models.ApiPostDetail}  "帖子详情"
// @Router  /post/{id} [get]
func PostDetailHandler(c *gin.Context) {
	postID, err := getIDParam(c, "id")
	if err != nil {
		ResponseErrorWithMsg(c, CodeInvalidParams, "非法的帖子ID")
		return
	}
	post, err := logic.GetPostDetail(postID)
	if err != nil {
		if errors.Is(err, logic.ErrorPostNotExist) {
			ResponseError(c, CodeNotFound)
			return
		}
		zap.L().Error("logic.GetPostDetail() failed", zap.Uint64("post_id", postID), zap.Error(err))
		ResponseError(c, CodeServerBusy)
		return
	}
	ResponseSuccess(c, post)
}
