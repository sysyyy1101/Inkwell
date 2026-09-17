package controller

import (
	"inkwell_backend/logic"
	"inkwell_backend/models"
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// 评论

// CommentHandler 创建评论
//
// @Summary  发表评论
// @Description  给帖子发表评论, 需要登录。内容最长 2000 字。
// @Description  失败时 HTTP 状态码仍是 200, 通过响应体里的 code 判断: 参数错误 1001, 未登录或 token 失效 1006/1008, 服务繁忙 1005。
// @Tags  评论
// @Accept  json
// @Produce  json
// @Security  ApiKeyAuth
// @Param  body  body  models.CreateCommentForm  true  "评论参数, parent_id 是要回复的评论ID, 不回复传 0"
// @Success  200  {object}  controller.ResponseData{data=models.CommentIDResponse}  "评论成功, 返回新评论ID"
// @Router  /comment [post]
func CommentHandler(c *gin.Context) {
	var comment models.Comment
	if err := c.ShouldBindJSON(&comment); err != nil {
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
	comment.AuthorID = userID

	if err := logic.CreateComment(&comment); err != nil {
		if errors.Is(err, logic.ErrorPostNotExist) {
			ResponseErrorWithMsg(c, CodeInvalidParams, err.Error())
			return
		}
		zap.L().Error("logic.CreateComment() failed", zap.Error(err))
		ResponseError(c, CodeServerBusy)
		return
	}
	ResponseSuccess(c, gin.H{"comment_id": strconv.FormatUint(comment.CommentID, 10)})
}

// CommentListHandler 评论列表。
// 支持 ?post_id=xxx 查询某个帖子的全部评论, 也兼容 ?ids=1&ids=2 批量查询。
//
// @Summary  评论列表
// @Description  按 post_id 查询某个帖子的全部评论(按时间正序), 或者按 ids 批量查询评论, 游客可访问。
// @Description  两个参数都没传或者 ID 非法时返回 code 1001。
// @Tags  评论
// @Produce  json
// @Param  post_id  query  string  false  "帖子ID, 和 ids 二选一"
// @Param  ids  query  []string  false  "评论ID列表, 可以重复传多个, 例如 ?ids=1&ids=2"  collectionFormat(multi)
// @Success  200  {object}  controller.ResponseData{data=[]models.ApiComment}  "评论列表"
// @Router  /comment [get]
func CommentListHandler(c *gin.Context) {
	if postIDStr := c.Query("post_id"); postIDStr != "" {
		postID, err := strconv.ParseUint(postIDStr, 10, 64)
		if err != nil {
			ResponseErrorWithMsg(c, CodeInvalidParams, "非法的帖子ID")
			return
		}
		comments, err := logic.GetCommentListByPostID(postID)
		if err != nil {
			zap.L().Error("logic.GetCommentListByPostID() failed",
				zap.Uint64("post_id", postID), zap.Error(err))
			ResponseError(c, CodeServerBusy)
			return
		}
		ResponseSuccess(c, comments)
		return
	}

	ids, ok := c.GetQueryArray("ids")
	if !ok || len(ids) == 0 {
		ResponseErrorWithMsg(c, CodeInvalidParams, "缺少post_id或ids参数")
		return
	}
	commentIDs := make([]uint64, 0, len(ids))
	for _, id := range ids {
		commentID, err := strconv.ParseUint(id, 10, 64)
		if err != nil {
			ResponseErrorWithMsg(c, CodeInvalidParams, "非法的评论ID")
			return
		}
		commentIDs = append(commentIDs, commentID)
	}
	comments, err := logic.GetCommentListByIDs(commentIDs)
	if err != nil {
		zap.L().Error("logic.GetCommentListByIDs() failed", zap.Error(err))
		ResponseError(c, CodeServerBusy)
		return
	}
	ResponseSuccess(c, comments)
}
