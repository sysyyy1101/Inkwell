package controller

import (
	"inkwell_backend/logic"
	"inkwell_backend/models"
	"inkwell_backend/pkg/jwt"
	"errors"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// UserProfileHandler 用户主页信息
//
// @Summary  用户主页
// @Description  查询用户的公开信息(用户名、注册时间), 游客可访问。用户不存在或 ID 非法时返回 code 1009。
// @Tags  用户
// @Produce  json
// @Param  id  path  string  true  "用户ID(雪花 ID, 用字符串传)"
// @Success  200  {object}  controller.ResponseData{data=models.UserProfile}  "用户信息"
// @Router  /user/{id} [get]
func UserProfileHandler(c *gin.Context) {
	userID, err := getIDParam(c, "id")
	if err != nil {
		ResponseErrorWithMsg(c, CodeInvalidParams, "非法的用户ID")
		return
	}
	profile, err := logic.GetUserProfile(userID)
	if err != nil {
		if errors.Is(err, logic.ErrorUserNotExist) {
			ResponseError(c, CodeNotFound)
			return
		}
		zap.L().Error("logic.GetUserProfile() failed", zap.Uint64("user_id", userID), zap.Error(err))
		ResponseError(c, CodeServerBusy)
		return
	}
	ResponseSuccess(c, profile)
}

// UserPostListHandler 某个用户发布的帖子列表
//
// @Summary  用户发布的帖子
// @Description  分页查询某个用户发布的帖子(按发帖时间倒序), 游客可访问, 票数与评论数来自 Redis。
// @Description  用户不存在时返回 code 1009, ID 非法时返回 code 1001。
// @Tags  用户
// @Produce  json
// @Param  id  path  string  true  "用户ID(雪花 ID, 用字符串传)"
// @Param  page  query  int  false  "页码, 从1开始"  default(1)
// @Param  size  query  int  false  "每页条数, 最大50"  default(10)
// @Success  200  {object}  controller.ResponseData{data=[]models.PostListItem}  "帖子列表"
// @Router  /user/{id}/post [get]
func UserPostListHandler(c *gin.Context) {
	userID, err := getIDParam(c, "id")
	if err != nil {
		ResponseErrorWithMsg(c, CodeInvalidParams, "非法的用户ID")
		return
	}
	// 用户不存在时直接返回 1009, 避免前端出现一个"空主页"
	if _, err := logic.GetUserProfile(userID); err != nil {
		if errors.Is(err, logic.ErrorUserNotExist) {
			ResponseError(c, CodeNotFound)
			return
		}
		zap.L().Error("logic.GetUserProfile() failed", zap.Uint64("user_id", userID), zap.Error(err))
		ResponseError(c, CodeServerBusy)
		return
	}
	page, size := getPageParams(c, postListDefaultSize, postListMaxSize)
	posts, err := logic.GetPostListByAuthor(userID, page, size)
	if err != nil {
		zap.L().Error("logic.GetPostListByAuthor() failed",
			zap.Uint64("user_id", userID), zap.Int("page", page), zap.Int("size", size), zap.Error(err))
		ResponseError(c, CodeServerBusy)
		return
	}
	ResponseSuccess(c, posts)
}

// SignUpHandler 注册
//
// @Summary  用户注册
// @Description  注册新用户。用户名长度 2-64, 密码长度 6-64, 两次密码必须一致。
// @Description  失败时 HTTP 状态码仍是 200, 通过响应体里的 code 判断: 参数错误 1001, 用户名重复 1002。
// @Tags  用户
// @Accept  json
// @Produce  json
// @Param  body  body  models.RegisterForm  true  "注册参数, confirm_password 和 re_password 传一个即可"
// @Success  200  {object}  controller.ResponseData  "注册成功, data 为 null"
// @Router  /signup [post]
func SignUpHandler(c *gin.Context) {
	// 1.获取请求参数 2.校验数据有效性
	var fo models.RegisterForm
	if err := c.ShouldBindJSON(&fo); err != nil {
		ResponseErrorWithMsg(c, CodeInvalidParams, err.Error())
		return
	}
	// 3.注册用户
	err := logic.Register(&models.User{
		UserName: fo.UserName,
		Password: fo.Password,
	})
	switch {
	case errors.Is(err, logic.ErrorUserExist):
		ResponseError(c, CodeUserExist)
	case errors.Is(err, logic.ErrorUserNameInvalid), errors.Is(err, logic.ErrorPasswordInvalid):
		ResponseErrorWithMsg(c, CodeInvalidParams, err.Error())
	case err != nil:
		zap.L().Error("logic.Register() failed", zap.Error(err))
		ResponseError(c, CodeServerBusy)
	default:
		ResponseSuccess(c, nil)
	}
}

// LoginHandler 登录
//
// @Summary  用户登录
// @Description  登录成功后返回 accessToken(2 小时有效)与 refreshToken(7 天有效)。
// @Description  用户名不存在和密码错误都返回 code 1004, 避免被用来枚举用户名。
// @Tags  用户
// @Accept  json
// @Produce  json
// @Param  body  body  models.LoginForm  true  "登录参数"
// @Success  200  {object}  controller.ResponseData{data=models.LoginResponse}  "登录成功"
// @Router  /login [post]
func LoginHandler(c *gin.Context) {
	var u models.User
	if err := c.ShouldBindJSON(&u); err != nil {
		zap.L().Error("invalid params", zap.Error(err))
		ResponseErrorWithMsg(c, CodeInvalidParams, err.Error())
		return
	}
	if err := logic.Login(&u); err != nil {
		// 用户不存在和密码错误统一返回同一个错误码, 避免被用来枚举用户名
		zap.L().Error("logic.Login() failed", zap.String("username", u.UserName), zap.Error(err))
		ResponseError(c, CodeInvalidPassword)
		return
	}
	// 生成Token
	aToken, rToken, err := jwt.GenToken(u.UserID)
	if err != nil {
		zap.L().Error("jwt.GenToken() failed", zap.Error(err))
		ResponseError(c, CodeServerBusy)
		return
	}
	ResponseSuccess(c, gin.H{
		"accessToken":  aToken,
		"refreshToken": rToken,
		// userID 是雪花 ID, 用字符串返回避免前端精度丢失
		"userID":   strconv.FormatUint(u.UserID, 10),
		"username": u.UserName,
	})
}

// RefreshTokenHandler 刷新 access token。
// access token 放在请求头里(即使已经过期也要带), refresh token 放在 query 参数里。
//
// @Summary  刷新 access token
// @Description  access token 放在 Authorization 请求头(已过期也要带), refresh token 放在 query 参数里。
// @Description  token 无效或已过期时返回 code 1006。
// @Tags  用户
// @Produce  json
// @Param  Authorization  header  string  true  "Bearer <accessToken>"
// @Param  refresh_token  query  string  true  "refresh token"
// @Success  200  {object}  controller.ResponseData{data=models.RefreshTokenResponse}  "刷新成功"
// @Router  /refresh_token [get]
func RefreshTokenHandler(c *gin.Context) {
	rt := c.Query("refresh_token")
	if rt == "" {
		ResponseErrorWithMsg(c, CodeInvalidParams, "缺少refresh_token参数")
		return
	}
	// 客户端携带Token有三种方式 1.放在请求头 2.放在请求体 3.放在URI
	// 这里假设Token放在Header的Authorization中，并使用Bearer开头
	authHeader := c.Request.Header.Get("Authorization")
	if authHeader == "" {
		ResponseErrorWithMsg(c, CodeInvalidToken, "请求头缺少Auth Token")
		return
	}
	// 按空格分割
	parts := strings.SplitN(authHeader, " ", 2)
	if !(len(parts) == 2 && parts[0] == "Bearer") {
		ResponseErrorWithMsg(c, CodeInvalidToken, "Token格式不对")
		return
	}
	aToken, rToken, err := jwt.RefreshToken(parts[1], rt)
	if err != nil {
		zap.L().Error("jwt.RefreshToken() failed", zap.Error(err))
		ResponseError(c, CodeInvalidToken)
		return
	}
	// 返回结构与登录接口保持一致
	ResponseSuccess(c, gin.H{
		"accessToken":  aToken,
		"refreshToken": rToken,
	})
}
