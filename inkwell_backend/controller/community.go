package controller

import (
	"inkwell_backend/logic"
	"errors"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// 社区

// CommunityHandler 社区列表
//
// @Summary  版块列表
// @Description  查询全部版块, 游客可访问
// @Tags  版块
// @Produce  json
// @Success  200  {object}  controller.ResponseData{data=[]models.Community}  "版块列表, 没有数据时返回空数组"
// @Router  /community [get]
func CommunityHandler(c *gin.Context) {
	communityList, err := logic.GetCommunityList()
	if err != nil {
		zap.L().Error("logic.GetCommunityList() failed", zap.Error(err))
		ResponseError(c, CodeServerBusy)
		return
	}
	ResponseSuccess(c, communityList)
}

// CommunityDetailHandler 社区详情
//
// @Summary  版块详情
// @Description  按版块 ID 查询版块详情, 游客可访问。版块 ID 非法或不存在时返回 code 1009。
// @Tags  版块
// @Produce  json
// @Param  id  path  int  true  "版块ID"
// @Success  200  {object}  controller.ResponseData{data=models.CommunityDetail}  "版块详情"
// @Router  /community/{id} [get]
func CommunityDetailHandler(c *gin.Context) {
	communityID, err := getIDParam(c, "id")
	if err != nil {
		ResponseErrorWithMsg(c, CodeInvalidParams, "非法的版块ID")
		return
	}
	community, err := logic.GetCommunityDetail(communityID)
	if err != nil {
		if errors.Is(err, logic.ErrorCommunityNotExist) {
			ResponseError(c, CodeNotFound)
			return
		}
		zap.L().Error("logic.GetCommunityDetail() failed",
			zap.Uint64("community_id", communityID), zap.Error(err))
		ResponseError(c, CodeServerBusy)
		return
	}
	ResponseSuccess(c, community)
}

// CommunityPostListHandler 某个版块下的帖子榜单
//
// @Summary  版块帖子榜单
// @Description  分页读取某个版块下的帖子榜单, 游客可访问。内部用 ZInterStore 求"版块帖子集合"与榜单的交集, 结果缓存 60 秒。
// @Description  版块 ID 不存在时返回 code 1009, ID 非法或 order 取值不合法时返回 code 1001。
// @Tags  版块
// @Produce  json
// @Param  id  path  int  true  "版块ID"
// @Param  order  query  string  false  "排序方式"  Enums(score, time)  default(score)
// @Param  page  query  int  false  "页码, 从1开始"  default(1)
// @Success  200  {object}  controller.ResponseData{data=[]models.PostListItem}  "帖子列表, 每页20条"
// @Router  /community/{id}/post [get]
func CommunityPostListHandler(c *gin.Context) {
	communityID, err := getIDParam(c, "id")
	if err != nil {
		ResponseErrorWithMsg(c, CodeInvalidParams, "非法的版块ID")
		return
	}
	order := c.DefaultQuery("order", "score")
	page, _ := getPageParams(c, postListDefaultSize, postListMaxSize)
	posts, err := logic.GetCommunityPostList(communityID, order, int64(page))
	if err != nil {
		switch {
		case errors.Is(err, logic.ErrorInvalidOrder):
			ResponseErrorWithMsg(c, CodeInvalidParams, err.Error())
		case errors.Is(err, logic.ErrorCommunityNotExist):
			ResponseError(c, CodeNotFound)
		default:
			zap.L().Error("logic.GetCommunityPostList() failed",
				zap.Uint64("community_id", communityID),
				zap.String("order", order), zap.Int("page", page), zap.Error(err))
			ResponseError(c, CodeServerBusy)
		}
		return
	}
	ResponseSuccess(c, posts)
}
