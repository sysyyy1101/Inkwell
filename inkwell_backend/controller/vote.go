package controller

import (
	"inkwell_backend/logic"
	"inkwell_backend/models"
	"encoding/json"
	"errors"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Direction 投票方向, 兼容 1 和 "1" 两种 JSON 写法
type Direction int8

const (
	// DirectionDown 反对票
	DirectionDown Direction = -1
	// DirectionCancel 取消投票
	DirectionCancel Direction = 0
	// DirectionUp 赞成票
	DirectionUp Direction = 1
)

func (d *Direction) UnmarshalJSON(data []byte) error {
	// 去掉可能存在的引号, 兼容前端传字符串的老写法
	raw := strings.Trim(strings.TrimSpace(string(data)), `"`)
	v, err := strconv.ParseInt(raw, 10, 8)
	if err != nil {
		return errors.New("direction 必须是 -1、0 或 1")
	}
	switch Direction(v) {
	case DirectionDown, DirectionCancel, DirectionUp:
		*d = Direction(v)
	default:
		return errors.New("direction 必须是 -1、0 或 1")
	}
	return nil
}

// VoteData 投票请求参数
type VoteData struct {
	PostID    uint64    `json:"post_id"`
	Direction Direction `json:"direction"`
}

func (v *VoteData) UnmarshalJSON(data []byte) (err error) {
	required := struct {
		PostID    json.RawMessage `json:"post_id"`
		Direction json.RawMessage `json:"direction"`
	}{}
	if err = json.Unmarshal(data, &required); err != nil {
		return
	}
	postID, err := models.ParseID(required.PostID)
	if err != nil {
		return
	}
	if postID == 0 {
		return errors.New("缺少必填字段post_id")
	}
	// direction 必须显式传入, 它的取值在 Direction.UnmarshalJSON 里校验
	if len(required.Direction) == 0 {
		return errors.New("缺少必填字段direction")
	}
	var direction Direction
	if err = json.Unmarshal(required.Direction, &direction); err != nil {
		return
	}
	v.PostID = postID
	v.Direction = direction
	return
}

// VoteHandler 给帖子投票
//
// @Summary  给帖子投票
// @Description  点赞/点踩/取消投票, 需要登录。发帖超过一周的帖子不能再投票。direction: 1 赞成 / 0 取消 / -1 反对。
// @Description  失败时 HTTP 状态码仍是 200, 通过响应体里的 code 判断: 参数错误/重复投票/已过投票时间 1001, 未登录或 token 失效 1006/1008。
// @Tags  投票
// @Accept  json
// @Produce  json
// @Security  ApiKeyAuth
// @Param  body  body  models.VoteForm  true  "投票参数"
// @Success  200  {object}  controller.ResponseData{data=models.VoteResponse}  "投票成功, 返回投票后的净票数(赞成 - 反对)"
// @Router  /vote [post]
func VoteHandler(c *gin.Context) {
	// 给哪个文章投什么票
	var vote VoteData
	if err := c.ShouldBindJSON(&vote); err != nil {
		ResponseErrorWithMsg(c, CodeInvalidParams, err.Error())
		return
	}
	userID, err := getCurrentUserID(c)
	if err != nil {
		ResponseError(c, CodeNotLogin)
		return
	}
	votes, err := logic.Vote(vote.PostID, userID, float64(vote.Direction))
	if err != nil {
		switch {
		case errors.Is(err, logic.ErrorPostNotExist),
			errors.Is(err, logic.ErrorVoted),
			errors.Is(err, logic.ErrorVoteTimeExpire),
			errors.Is(err, logic.ErrorInvalidVoteDirection):
			ResponseErrorWithMsg(c, CodeInvalidParams, err.Error())
		default:
			zap.L().Error("logic.Vote() failed",
				zap.Uint64("post_id", vote.PostID), zap.Error(err))
			ResponseError(c, CodeServerBusy)
		}
		return
	}
	// 返回投票后的票数: 前端直接用它刷新展示, 不用按本地记录推算增量
	// (换设备/清缓存后本地记录可能是错的, 推算出来的票数会和服务端不一致)
	ResponseSuccess(c, gin.H{
		"post_id":  strconv.FormatUint(vote.PostID, 10),
		"vote_num": votes,
	})
}
